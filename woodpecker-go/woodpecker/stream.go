// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package woodpecker

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// Event is the payload delivered by Subscribe whenever the server pushes a
// pipeline / repo state change. The shape mirrors the JSON the server emits on
// both the WebSocket and SSE event streams.
type Event struct {
	Repo     Repo     `json:"repo"`
	Pipeline Pipeline `json:"pipeline"`
}

// Stream is the handle returned by Subscribe and LogStream. Callers Close it
// when they no longer want to receive updates. Closing is idempotent.
//
// Done returns a channel that is closed once the stream is fully torn down,
// either because the caller invoked Close or because the server signaled the
// end of the stream (EOF) and the stream is configured not to reconnect.
// After Done is closed, Err returns the terminating error, or nil on a clean
// EOF / user-initiated close.
type Stream interface {
	Close()
	Done() <-chan struct{}
	Err() error
}

// Default tuning constants for the reconnect / handshake loop. The numbers
// mirror the web client (web/src/lib/api/client.ts) so a Woodpecker server
// sees the same retry / fallback shape from both clients.
const (
	defaultWSHandshakeTimeout = 5 * time.Second
	defaultInitialBackoff     = 1 * time.Second
	defaultMaxBackoff         = 30 * time.Second

	// SSE log frames can be larger than bufio.Scanner's default 64 KiB line
	// limit. Bump the initial buffer to 64 KiB and the maximum to 1 MiB;
	// the latter mirrors how much a single log line can grow without us
	// treating it as a stream-level error.
	sseScannerBufSize    = 64 * 1024
	sseScannerMaxBufSize = 1024 * 1024
)

// streamOptions tunes the reconnect and handshake behavior. They mirror the
// constants used by the web client so the user-visible behavior stays in
// lockstep across both libraries.
type streamOptions struct {
	// reconnect controls whether transient drops on an already-established
	// connection are retried with exponential backoff. The very first
	// handshake is always attempted regardless of this flag.
	reconnect bool

	// handshakeTimeout caps how long we wait for the initial WS upgrade
	// before treating it as a failure. Some proxies accept the TCP
	// connection but never finish the HTTP upgrade, which would otherwise
	// hang forever and prevent us from falling back to SSE.
	handshakeTimeout time.Duration

	// initialBackoff is the wait before the first reconnect attempt after a
	// drop on an established connection. It doubles up to maxBackoff.
	initialBackoff time.Duration
	maxBackoff     time.Duration
}

func defaultStreamOptions() streamOptions {
	return streamOptions{
		reconnect:        true,
		handshakeTimeout: defaultWSHandshakeTimeout,
		initialBackoff:   defaultInitialBackoff,
		maxBackoff:       defaultMaxBackoff,
	}
}

// stream is the concrete implementation of Stream shared by Subscribe and
// LogStream. The two entry points only differ in the URL paths they feed to
// the run loop and the payload type the caller wants.
type stream struct {
	close     context.CancelCauseFunc
	doneChan  chan struct{}
	errMu     sync.Mutex
	finalErr  error
	closeOnce sync.Once
}

func (s *stream) Close() {
	s.closeOnce.Do(func() {
		s.close(nil)
	})
}

func (s *stream) Done() <-chan struct{} { return s.doneChan }

func (s *stream) Err() error {
	s.errMu.Lock()
	defer s.errMu.Unlock()
	return s.finalErr
}

func (s *stream) setErr(err error) {
	s.errMu.Lock()
	defer s.errMu.Unlock()
	if s.finalErr == nil {
		s.finalErr = err
	}
}

// Subscribe opens an event stream that delivers every Event the server pushes
// for repositories the caller has access to. It prefers WebSocket and falls
// back to SSE if the very first WS handshake fails — the same logic the web
// UI uses, kept in sync so both clients behave identically against a given
// deployment / reverse proxy.
//
// The callback runs in the stream's internal goroutine; long-running work
// inside it will block delivery of subsequent events. If the caller needs to
// do anything non-trivial, hand the event off to a worker goroutine.
//
// The returned Stream's Done channel closes when the stream is fully torn
// down; Err returns the terminating error (nil on a clean close).
func (c *client) Subscribe(ctx context.Context, callback func(Event)) Stream {
	return c.subscribeStream(
		ctx,
		"/api/stream/ws/events",
		"/api/stream/sse/events",
		func(data []byte) error {
			var e Event
			if err := json.Unmarshal(data, &e); err != nil {
				return fmt.Errorf("unmarshal event: %w", err)
			}
			callback(e)
			return nil
		},
	)
}

// LogStream opens a log stream for the given pipeline step. As with Subscribe
// it prefers WebSocket and falls back to SSE on a first-attempt handshake
// failure. The stream terminates on EOF (the server closes with a normal
// status once the step finishes); the resulting Stream's Done channel will
// close at that point and Err will be nil.
func (c *client) LogStream(ctx context.Context, repoID, pipelineNumber, stepID int64, callback func(LogEntry)) Stream {
	wsPath := fmt.Sprintf("/api/stream/ws/logs/%d/%d/%d", repoID, pipelineNumber, stepID)
	ssePath := fmt.Sprintf("/api/stream/sse/logs/%d/%d/%d", repoID, pipelineNumber, stepID)
	return c.subscribeStream(
		ctx,
		wsPath,
		ssePath,
		func(data []byte) error {
			var l LogEntry
			if err := json.Unmarshal(data, &l); err != nil {
				return fmt.Errorf("unmarshal log entry: %w", err)
			}
			callback(l)
			return nil
		},
	)
}

// subscribeStream is the shared engine behind Subscribe and LogStream. The
// dispatcher is given the raw JSON payload of a single message and is
// responsible for decoding it into the caller's type and invoking the
// caller's callback.
//
// The control flow mirrors the web client's _subscribeStream:
//   - try WS first;
//   - if the very first WS handshake fails, fall back to SSE for the lifetime
//     of this subscription (one-way, per-subscription fallback);
//   - transient drops on an already-opened WS keep retrying with WS rather
//     than falling back, since a working WS that briefly drops is a network
//     blip, not a "WS not supported here" condition.
func (c *client) subscribeStream(ctx context.Context, wsPath, ssePath string, dispatch func([]byte) error) Stream {
	opts := defaultStreamOptions()

	streamCtx, cancel := context.WithCancelCause(ctx)
	s := &stream{
		close:    cancel,
		doneChan: make(chan struct{}),
	}

	go func() {
		defer close(s.doneChan)

		// First attempt is WS. If the handshake fails before we ever get a
		// connected socket, we switch to SSE and never look back for this
		// subscription's lifetime.
		opened, err := c.runWS(streamCtx, wsPath, dispatch, opts)
		if streamCtx.Err() != nil {
			return
		}
		if !opened {
			// Handshake never succeeded — fall back to SSE.
			err = c.runSSE(streamCtx, ssePath, dispatch, opts)
			s.setErr(err)
			return
		}
		// WS opened at least once before terminating. Honor the terminal
		// error path (clean EOF returns nil).
		s.setErr(err)
	}()

	return s
}

// runWS drives a WebSocket subscription with reconnect / backoff. It returns
// (opened, err) where opened indicates whether the socket was successfully
// accepted at least once. Callers use opened == false to decide whether to
// fall back to SSE (only triggered when the first handshake fails). When
// opened is true, err carries the terminal error of the last attempt; nil
// means a clean EOF (server closed with NormalClosure + reason "eof").
func (c *client) runWS(ctx context.Context, path string, dispatch func([]byte) error, opts streamOptions) (bool, error) {
	wsURL, err := buildWSURL(c.addr, path)
	if err != nil {
		return false, err
	}

	everOpened := false
	backoff := opts.initialBackoff
	for {
		if ctx.Err() != nil {
			return everOpened, nil
		}

		opened, eof, err := c.dialAndReadWS(ctx, wsURL, dispatch, opts.handshakeTimeout)
		if opened {
			everOpened = true
			// Reset backoff after any successful (re)connect so the next
			// failure starts low again.
			backoff = opts.initialBackoff
		}

		// Server signaled normal EOF — stream is legitimately done.
		if eof {
			return everOpened, nil
		}

		// First-attempt handshake failure — surrender so the caller can
		// fall back to SSE.
		if !opened {
			return false, err
		}

		// Drop after a successful open — retry with WS if asked to.
		if !opts.reconnect {
			return true, err
		}

		select {
		case <-ctx.Done():
			return everOpened, nil
		case <-time.After(backoff):
		}
		backoff *= 2
		if backoff > opts.maxBackoff {
			backoff = opts.maxBackoff
		}
	}
}

// dialAndReadWS performs a single dial + read loop. The returned (opened,
// eof, err) tuple lets the caller distinguish:
//   - opened=false: dial / handshake failed (consider fallback);
//   - opened=true,  eof=true: server closed cleanly with EOF reason;
//   - opened=true,  eof=false: transient error after open (reconnect).
func (c *client) dialAndReadWS(ctx context.Context, wsURL string, dispatch func([]byte) error, handshakeTimeout time.Duration) (opened, eof bool, err error) {
	dialCtx, dialCancel := context.WithTimeout(ctx, handshakeTimeout)
	defer dialCancel()

	conn, resp, err := websocket.Dial(dialCtx, wsURL, &websocket.DialOptions{
		HTTPClient: c.client,
	})
	// On a successful upgrade the response body is already closed by the
	// library, but Dial may still return a response on error (e.g. a 400
	// from a proxy that rejects the upgrade) and that body must be closed
	// to avoid leaking the connection.
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return false, false, fmt.Errorf("ws dial %s: %w", wsURL, err)
	}
	defer func() { _ = conn.CloseNow() }()

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			// Close code + reason "eof" is the server's signal that the
			// stream legitimately ended (e.g. the step finished). Treat that
			// as a clean termination rather than a transient drop.
			closeStatus := websocket.CloseStatus(err)
			if closeStatus == websocket.StatusNormalClosure {
				reason := websocket.CloseError{}
				if errors.As(err, &reason) && reason.Reason == "eof" {
					return true, true, nil
				}
				// Normal closure without an explicit eof reason — also done.
				return true, true, nil
			}
			if ctx.Err() != nil {
				return true, true, nil
			}
			return true, false, fmt.Errorf("ws read: %w", err)
		}
		if err := dispatch(data); err != nil {
			// A malformed frame shouldn't tear down the whole subscription;
			// the server only ever sends JSON, but during shutdown a partial
			// frame is possible. Skip and keep reading.
			continue
		}
	}
}

// runSSE drives an EventSource-style stream parsed straight out of the
// response body. The server emits:
//
//	data: <json>\n\n           — payload
//	: ping\n\n                  — keep-alive (ignored)
//	id: <n>\nid: ...            — sequential id for resumption
//	event: eof\ndata: eof\n\n   — end of stream
//	event: error\ndata: ...\n\n — server-side error
//
// We honor the id: header on the log stream via Last-Event-ID to resume from
// where we left off across reconnects.
func (c *client) runSSE(ctx context.Context, path string, dispatch func([]byte) error, opts streamOptions) error {
	endpoint := strings.TrimSuffix(c.addr, "/") + path
	lastEventID := ""
	backoff := opts.initialBackoff

	for {
		if ctx.Err() != nil {
			return nil
		}

		eof, lastSeen, err := c.readSSEOnce(ctx, endpoint, lastEventID, dispatch)
		if lastSeen != "" {
			lastEventID = lastSeen
		}
		if eof {
			return nil
		}
		if !opts.reconnect {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff):
		}
		backoff *= 2
		if backoff > opts.maxBackoff {
			backoff = opts.maxBackoff
		}
	}
}

// readSSEOnce opens a single SSE connection and reads it to completion. It
// returns (eof, lastEventID, err): eof=true means a clean end-of-stream from
// the server; lastEventID is the most recent id: value seen, suitable for
// passing back on a reconnect via the Last-Event-ID header.
func (c *client) readSSEOnce(ctx context.Context, endpoint, lastEventID string, dispatch func([]byte) error) (bool, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false, lastEventID, err
	}
	req.Header.Set("Accept", "text/event-stream")
	if lastEventID != "" {
		req.Header.Set("Last-Event-ID", lastEventID)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false, lastEventID, fmt.Errorf("sse dial %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, lastEventID, &ClientError{
			StatusCode: resp.StatusCode,
			Message:    strings.TrimSpace(string(body)),
		}
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, sseScannerBufSize), sseScannerMaxBufSize)

	var (
		eventType string
		dataBuf   strings.Builder
		seenID    = lastEventID
	)

	flush := func() error {
		defer func() {
			eventType = ""
			dataBuf.Reset()
		}()
		if dataBuf.Len() == 0 {
			return nil
		}
		switch eventType {
		case "eof":
			return io.EOF
		case "error":
			return &ClientError{
				StatusCode: http.StatusInternalServerError,
				Message:    dataBuf.String(),
			}
		default:
			return dispatch([]byte(dataBuf.String()))
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			// Empty line = dispatch the buffered event.
			if err := flush(); err != nil {
				if errors.Is(err, io.EOF) {
					return true, seenID, nil
				}
				if ce := (&ClientError{}); errors.As(err, &ce) {
					return false, seenID, err
				}
				// Decoding error on a single frame — skip and continue.
				continue
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			// Comment / ping — ignored.
			continue
		}
		field, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = strings.TrimPrefix(value, " ")
		switch field {
		case "data":
			if dataBuf.Len() > 0 {
				dataBuf.WriteByte('\n')
			}
			dataBuf.WriteString(value)
		case "event":
			eventType = value
		case "id":
			seenID = value
		}
	}
	if err := scanner.Err(); err != nil {
		return false, seenID, fmt.Errorf("sse read: %w", err)
	}
	// Body closed without an explicit eof frame — treat as a transient drop
	// so the caller's reconnect loop can pick it back up.
	return false, seenID, io.ErrUnexpectedEOF
}

// buildWSURL converts the configured server address ("http://...") and a path
// like "/api/stream/ws/events" into the matching ws:// or wss:// URL.
func buildWSURL(addr, path string) (string, error) {
	base := strings.TrimSuffix(addr, "/") + path
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("parse ws url %s: %w", base, err)
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	case "wss", "ws":
		// already a websocket scheme — accept as-is
	default:
		return "", fmt.Errorf("unsupported scheme %q for ws url", u.Scheme)
	}
	return u.String(), nil
}
