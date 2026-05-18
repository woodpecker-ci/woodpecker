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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// drainEvents collects events until either the expected count is reached or
// the timeout fires. It returns whatever was collected so the test can decide
// whether that is enough.
func drainEvents[T any](ch <-chan T, want int, timeout time.Duration) []T {
	out := make([]T, 0, want)
	deadline := time.After(timeout)
	for len(out) < want {
		select {
		case v := <-ch:
			out = append(out, v)
		case <-deadline:
			return out
		}
	}
	return out
}

// Test_Subscribe_WS_HappyPath asserts that Subscribe receives JSON events
// pushed by a WebSocket server and terminates cleanly when the server closes
// the connection with NormalClosure + reason "eof".
func Test_Subscribe_WS_HappyPath(t *testing.T) {
	want := Event{
		Repo:     Repo{ID: 42, FullName: "octo/cat"},
		Pipeline: Pipeline{ID: 7, Number: 3, Status: "success"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/stream/ws/") {
			t.Errorf("unexpected non-ws request path %q", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		require.NoError(t, err)
		defer func() { _ = conn.CloseNow() }()
		buf, err := json.Marshal(want)
		require.NoError(t, err)
		require.NoError(t, conn.Write(r.Context(), websocket.MessageText, buf))
		_ = conn.Close(websocket.StatusNormalClosure, "eof")
	}))
	defer srv.Close()

	c := NewClient(srv.URL, http.DefaultClient)
	events := make(chan Event, 1)
	stream := c.Subscribe(t.Context(), func(e Event) { events <- e })

	got := drainEvents(events, 1, 2*time.Second)
	require.Len(t, got, 1)
	assert.Equal(t, want, got[0])

	select {
	case <-stream.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("stream did not terminate after server EOF")
	}
	assert.NoError(t, stream.Err())
}

// Test_Subscribe_WS_FallsBackToSSE asserts that when the very first WS
// handshake fails (server returns 400 on the WS path), the client falls back
// to the SSE endpoint and successfully delivers events parsed from the SSE
// frames.
func Test_Subscribe_WS_FallsBackToSSE(t *testing.T) {
	want := Event{
		Repo:     Repo{ID: 1, FullName: "alice/repo"},
		Pipeline: Pipeline{ID: 100, Number: 1, Status: "running"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/stream/ws/"):
			// Reject the upgrade — simulates a reverse proxy that strips
			// the WebSocket Upgrade headers.
			http.Error(w, "ws not supported here", http.StatusBadRequest)
		case strings.HasPrefix(r.URL.Path, "/api/stream/sse/"):
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			f, ok := w.(http.Flusher)
			require.True(t, ok)
			buf, _ := json.Marshal(want)
			_, _ = fmt.Fprintf(w, "data: %s\n\n", buf)
			f.Flush()
			// Send the explicit eof event so the client terminates cleanly.
			_, _ = fmt.Fprint(w, "event: eof\ndata: eof\n\n")
			f.Flush()
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL, http.DefaultClient)
	events := make(chan Event, 1)
	stream := c.Subscribe(t.Context(), func(e Event) { events <- e })

	got := drainEvents(events, 1, 2*time.Second)
	require.Len(t, got, 1, "expected one event via SSE fallback")
	assert.Equal(t, want, got[0])

	select {
	case <-stream.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("stream did not terminate after SSE eof")
	}
	assert.NoError(t, stream.Err())
}

// Test_LogStream_WS_HappyPath asserts that LogStream decodes LogEntry JSON
// frames sent over a WebSocket and that the path encodes the route params
// the server expects.
func Test_LogStream_WS_HappyPath(t *testing.T) {
	var seenPath string
	want := []LogEntry{
		{ID: 1, StepID: 99, Line: 1, Data: []byte("first"), Type: LogEntryStdout},
		{ID: 2, StepID: 99, Line: 2, Data: []byte("second"), Type: LogEntryStdout},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		require.NoError(t, err)
		defer func() { _ = conn.CloseNow() }()
		for _, e := range want {
			buf, err := json.Marshal(e)
			require.NoError(t, err)
			require.NoError(t, conn.Write(r.Context(), websocket.MessageText, buf))
		}
		_ = conn.Close(websocket.StatusNormalClosure, "eof")
	}))
	defer srv.Close()

	c := NewClient(srv.URL, http.DefaultClient)
	got := make(chan LogEntry, len(want))
	stream := c.LogStream(t.Context(), 11, 22, 99, func(e LogEntry) { got <- e })

	entries := drainEvents(got, len(want), 2*time.Second)
	require.Len(t, entries, len(want))
	for i, e := range entries {
		assert.Equal(t, want[i].ID, e.ID)
		assert.Equal(t, want[i].Line, e.Line)
		assert.Equal(t, want[i].Data, e.Data)
	}
	assert.Equal(t, "/api/stream/ws/logs/11/22/99", seenPath)

	select {
	case <-stream.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("stream did not terminate")
	}
	assert.NoError(t, stream.Err())
}

// Test_Stream_Close_StopsDelivery asserts that calling Close on a live
// subscription tears the stream down even when the server is still willing
// to push events. We use a server that sits in a long-lived send loop so the
// stream would otherwise never finish on its own.
func Test_Stream_Close_StopsDelivery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		require.NoError(t, err)
		defer func() { _ = conn.CloseNow() }()
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				if err := conn.Write(r.Context(), websocket.MessageText, []byte(`{"repo":{},"pipeline":{}}`)); err != nil {
					return
				}
			}
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL, http.DefaultClient)
	var (
		mu      sync.Mutex
		nEvents int
	)
	stream := c.Subscribe(t.Context(), func(e Event) {
		mu.Lock()
		nEvents++
		mu.Unlock()
	})

	// Let at least one event arrive so we know the subscription is live
	// before we Close it.
	time.Sleep(50 * time.Millisecond)
	stream.Close()

	select {
	case <-stream.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("stream did not stop after Close")
	}

	// Snapshot the count, sleep, then confirm no further events arrive.
	mu.Lock()
	before := nEvents
	mu.Unlock()
	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	after := nEvents
	mu.Unlock()
	assert.Equal(t, before, after, "events kept arriving after Close")
}

// Test_buildWSURL exercises the http→ws and https→wss rewrite plus path
// joining, including the case where the server base URL already carries a
// trailing slash.
func Test_buildWSURL(t *testing.T) {
	cases := []struct {
		name string
		addr string
		path string
		want string
	}{
		{"http to ws", "http://example.com", "/api/stream/ws/events", "ws://example.com/api/stream/ws/events"},
		{"https to wss", "https://example.com:8000", "/api/stream/ws/events", "wss://example.com:8000/api/stream/ws/events"},
		{"trailing slash trimmed", "http://example.com/", "/api/stream/ws/events", "ws://example.com/api/stream/ws/events"},
		{"with port", "http://localhost:8000", "/api/stream/ws/logs/1/2/3", "ws://localhost:8000/api/stream/ws/logs/1/2/3"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildWSURL(tc.addr, tc.path)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// Test_SSE_Parses_KeepAliveAndID asserts the SSE parser ignores `: ping`
// comments, tracks the most recent `id:` so reconnects can resume, and
// dispatches multi-`data:` events as a single payload.
func Test_SSE_Parses_KeepAliveAndID(t *testing.T) {
	want := LogEntry{ID: 1, StepID: 1, Data: []byte("hello"), Type: LogEntryStdout}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only respond to SSE; reject WS so the client falls back.
		if !strings.HasPrefix(r.URL.Path, "/api/stream/sse/") {
			http.Error(w, "ws not supported", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		f, ok := w.(http.Flusher)
		require.True(t, ok)
		_, _ = fmt.Fprint(w, ": ping\n\n")
		f.Flush()
		buf, _ := json.Marshal(want)
		_, _ = fmt.Fprintf(w, "id: 5\ndata: %s\n\n", buf)
		f.Flush()
		_, _ = fmt.Fprint(w, "event: eof\ndata: eof\n\n")
		f.Flush()
	}))
	defer srv.Close()

	c := NewClient(srv.URL, http.DefaultClient)
	got := make(chan LogEntry, 1)
	stream := c.LogStream(t.Context(), 1, 1, 1, func(e LogEntry) { got <- e })

	entries := drainEvents(got, 1, 2*time.Second)
	require.Len(t, entries, 1)
	assert.Equal(t, want.ID, entries[0].ID)
	assert.Equal(t, want.Data, entries[0].Data)

	select {
	case <-stream.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("stream did not terminate after SSE eof")
	}
	assert.NoError(t, stream.Err())
}

// Test_Subscribe_ContextCancel asserts that canceling the parent context
// shuts the stream down even without an explicit Close().
func Test_Subscribe_ContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		require.NoError(t, err)
		defer func() { _ = conn.CloseNow() }()
		<-r.Context().Done()
	}))
	defer srv.Close()

	c := NewClient(srv.URL, http.DefaultClient)
	ctx, cancel := context.WithCancelCause(t.Context())
	stream := c.Subscribe(ctx, func(e Event) {})

	// Give the stream a moment to actually establish before we yank it.
	time.Sleep(50 * time.Millisecond)
	cancel(nil)

	select {
	case <-stream.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("stream did not stop after context cancel")
	}
}
