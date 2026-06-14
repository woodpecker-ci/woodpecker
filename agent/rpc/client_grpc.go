// Copyright 2023 Woodpecker Authors
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

package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"
	grpc_proto "google.golang.org/protobuf/proto"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/rpc/proto"
)

var (
	ErrConnectionLost = errors.New("connection to server lost")
	errNotConnected   = errors.New("grpc: not connected")
)

const (
	// Set grpc version on compile time to compare against server version response.
	ClientGrpcVersion int32 = proto.Version

	// Maximum size of an outgoing log message.
	// Picked to prevent it from going over GRPC size limit (4 MiB) with a large safety margin.
	maxLogBatchSize int = 1 * 1024 * 1024

	// Maximum amount of time between sending consecutive batched log messages.
	// Controls the delay between the CI job generating a log record, and web users receiving it.
	maxLogFlushPeriod time.Duration = time.Second

	// Maximum number of times a single RPC will re-authenticate and retry after
	// the server rejects its access token with codes.Unauthenticated. It
	// prevents an infinite refresh/retry loop when re-authentication does not
	// resolve the rejection (e.g. a revoked agent secret) — even when
	// connectionRetryTimeout is infinite.
	maxAuthRefreshes int = 2
)

// tokenRefresher lets the RPC client force a re-authentication after the
// server rejects the current access token (most commonly because it expired).
// It is implemented by *AuthInterceptor.
type tokenRefresher interface {
	// Token returns the access token currently in use.
	Token() string
	// RefreshToken re-authenticates and replaces the stored access token.
	// staleToken is the token the caller last used; if the stored token has
	// already changed, the refresh is skipped.
	RefreshToken(ctx context.Context, staleToken string) error
}

type client struct {
	client proto.WoodpeckerClient
	conn   *grpc.ClientConn
	logs   chan *proto.LogEntry
	// auth re-authenticates the agent when an RPC fails with
	// codes.Unauthenticated. It may be nil (e.g. in tests), in which case
	// Unauthenticated errors are treated as permanent.
	auth tokenRefresher
	// connectionRetryTimeout is the maximum time to wait for a connection to be
	// restored before the agent gives up and exits. Zero means infinite.
	// Maps directly onto backoff.WithMaxElapsedTime.
	connectionRetryTimeout time.Duration
}

// NewGrpcClient returns a new grpc Client.
func NewGrpcClient(ctx context.Context, conn *grpc.ClientConn, opts ...ClientOption) rpc.Peer {
	client := new(client)
	client.client = proto.NewWoodpeckerClient(conn)
	client.conn = conn
	client.logs = make(chan *proto.LogEntry, 10) // max memory use: 10 lines * 1 MiB

	for _, opt := range opts {
		opt(client)
	}

	go client.processLogs(ctx)
	return client
}

type ClientOption func(c *client)

func SetConnectionRetryTimeout(d time.Duration) ClientOption {
	if d == 0 {
		log.Warn().Msg("connection retry timeout set to infinite")
	}
	return func(c *client) {
		c.connectionRetryTimeout = d
	}
}

// SetAuthRefresher wires the auth interceptor into the client so that RPCs
// rejected with codes.Unauthenticated trigger a re-authentication and retry.
func SetAuthRefresher(a tokenRefresher) ClientOption {
	return func(c *client) {
		c.auth = a
	}
}

// IsConnected reports whether the underlying gRPC connection is currently up.
// It is a pure observer with no side effects.
func (c *client) IsConnected() bool {
	state := c.conn.GetState()
	return state == connectivity.Ready || state == connectivity.Idle
}

// retryOpts returns the backoff options used for every retry loop in this
// file. The exponential backoff parameters preserve the original tuning
// (10 ms initial, 10 s cap), and connectionRetryTimeout is wired straight into
// WithMaxElapsedTime — when it elapses, backoff.Retry returns the last error,
// which we translate into ErrConnectionLost in retryRPC.
func (c *client) retryOpts(op string) []backoff.RetryOption {
	b := backoff.NewExponentialBackOff()
	b.MaxInterval = 10 * time.Second          //nolint:mnd
	b.InitialInterval = 10 * time.Millisecond //nolint:mnd

	notify := func(err error, next time.Duration) {
		// The "too_many_pings" GOAWAY is well-known noise; demote to trace.
		// See https://github.com/woodpecker-ci/woodpecker/issues/717
		if strings.Contains(err.Error(), `"too_many_pings"`) {
			log.Trace().Err(err).Dur("retry_in", next).Msgf("grpc: %s(): too many keepalive pings without sending data", op)
			return
		}
		if errors.Is(err, errNotConnected) {
			log.Warn().Dur("retry_in", next).Msgf("grpc: %s() waiting for server connection...", op)
			return
		}
		log.Warn().Err(err).Dur("retry_in", next).Msgf("grpc error: %s(): code: %v", op, status.Code(err))
	}

	return []backoff.RetryOption{
		backoff.WithBackOff(b),
		backoff.WithMaxElapsedTime(c.connectionRetryTimeout),
		backoff.WithNotify(notify),
	}
}

// retryRPC is the workhorse used by every RPC method in this file. It runs op
// under backoff.Retry with the standard options, and translates the few
// special outcomes the callers care about:
//
//   - op succeeds          -> (result, nil)
//   - ctx canceled         -> (zero, nil)            same contract as before
//   - MaxElapsedTime hit   -> (zero, ErrConnectionLost)
//   - permanent (fatal)    -> (zero, underlying err)
//
// The op closure is responsible for:
//   - returning errNotConnected when IsConnected() is false (Retry will sleep
//     and call again — same effect as the old "if !c.IsConnected()" preamble)
//   - returning backoff.Permanent(err) for unrecoverable gRPC codes
//   - returning the raw error for retryable codes (Aborted/DataLoss/...)
func retryRPC[T any](ctx context.Context, c *client, opName string, op backoff.Operation[T]) (T, error) {
	res, err := backoff.Retry(ctx, withAuthRefresh(ctx, c, opName, op), c.retryOpts(opName)...)
	if err == nil {
		return res, nil
	}

	var zero T

	// Context canceled while inside Retry: callers historically swallowed this
	// and returned a zero-value error, so preserve that contract.
	if ctxErr := context.Cause(ctx); ctxErr != nil && errors.Is(err, ctxErr) {
		log.Debug().Err(err).Msgf("grpc: %s(): context canceled", opName)
		return zero, nil
	}

	// MaxElapsedTime exhausted while we were still in errNotConnected — give up.
	if errors.Is(err, errNotConnected) {
		log.Error().Msg("grpc: connection lost, giving up")
		return zero, ErrConnectionLost
	}

	log.Error().Err(err).Msgf("grpc error: %s(): code: %v", opName, status.Code(err))
	return zero, err
}

// withAuthRefresh wraps an RPC operation so that, when it fails with
// codes.Unauthenticated (typically an expired access token), the client
// re-authenticates and lets backoff retry the call. Without this, an expired
// token would never be replaced on demand and the agent would stay stuck
// retrying with a dead token (see issue #4144).
//
// Re-authentication is attempted at most maxAuthRefreshes times per RPC; once
// exhausted the error is made permanent so the call returns and the failure
// surfaces to the caller (and ultimately the agent's supervisor). If no
// refresher is configured, the wrapper is a no-op and Unauthenticated stays
// permanent.
func withAuthRefresh[T any](ctx context.Context, c *client, opName string, op backoff.Operation[T]) backoff.Operation[T] {
	if c.auth == nil {
		// Without a refresher there is no way to recover from an auth
		// rejection, so keep codes.Unauthenticated permanent (the behavior
		// before on-demand refresh existed).
		return func() (T, error) {
			res, err := op()
			if err != nil && status.Code(err) == codes.Unauthenticated {
				return res, backoff.Permanent(err)
			}
			return res, err
		}
	}

	refreshes := 0
	return func() (T, error) {
		res, err := op()
		if err == nil || status.Code(err) != codes.Unauthenticated {
			return res, err
		}

		if refreshes >= maxAuthRefreshes {
			log.Error().Msgf("grpc: %s(): access token still rejected after %d re-authentication attempts, giving up", opName, refreshes)
			return res, backoff.Permanent(err)
		}
		refreshes++

		staleToken := c.auth.Token()
		log.Warn().Msgf("grpc: %s(): access token rejected by server, re-authenticating (attempt %d/%d)", opName, refreshes, maxAuthRefreshes)
		if rerr := c.auth.RefreshToken(ctx, staleToken); rerr != nil {
			// Re-auth itself failed (e.g. server briefly unreachable); keep the
			// error retryable so backoff waits and we try again on the next
			// iteration, up to maxAuthRefreshes.
			log.Error().Err(rerr).Msgf("grpc: %s(): re-authentication failed", opName)
		}
		// Return the original (retryable) error so backoff retries the RPC,
		// this time with the refreshed token attached by the interceptor.
		return res, err
	}
}

// classifyRPCErr inspects a gRPC error and returns either the same error (for
// retryable codes) or a backoff.Permanent wrapping it (for fatal codes). It is
// the single source of truth for which gRPC codes are worth retrying.
func classifyRPCErr(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	switch status.Code(err) {
	case codes.Canceled:
		// If our own ctx is dead, surface that as the cause so Retry's
		// context.Cause(ctx) check exits cleanly. Otherwise it's a server-side
		// cancel that we treat as permanent.
		if ctx.Err() != nil {
			return backoff.Permanent(ctx.Err())
		}
		return backoff.Permanent(err)
	case codes.Unauthenticated:
		// The access token was rejected, most commonly because it expired.
		// Keep it retryable so withAuthRefresh can re-authenticate and retry;
		// the retry count there bounds how long we keep trying.
		return err
	case codes.Aborted,
		codes.DataLoss,
		codes.DeadlineExceeded,
		codes.Internal,
		codes.Unavailable:
		return err
	default:
		return backoff.Permanent(err)
	}
}

// Version returns the server- & grpc-version.
func (c *client) Version(ctx context.Context) (*rpc.Version, error) {
	res, err := c.client.Version(ctx, &proto.Empty{})
	if err != nil {
		return nil, err
	}
	return &rpc.Version{
		GrpcVersion:   res.GrpcVersion,
		ServerVersion: res.ServerVersion,
	}, nil
}

// Next returns the next workflow in the queue.
func (c *client) Next(ctx context.Context, filter rpc.Filter) (*rpc.Workflow, error) {
	req := &proto.NextRequest{Filter: &proto.Filter{Labels: filter.Labels}}

	res, err := retryRPC(ctx, c, "next", func() (*proto.NextResponse, error) {
		if !c.IsConnected() {
			return nil, errNotConnected
		}
		r, err := c.client.Next(ctx, req)
		return r, classifyRPCErr(ctx, err)
	})
	if err != nil {
		return nil, err
	}
	if res == nil || res.GetWorkflow() == nil {
		return nil, nil
	}

	w := &rpc.Workflow{
		ID:      res.GetWorkflow().GetId(),
		Timeout: res.GetWorkflow().GetTimeout(),
		Config:  new(backend_types.Config),
	}
	if err := json.Unmarshal(res.GetWorkflow().GetPayload(), w.Config); err != nil {
		log.Error().Err(err).Msgf("could not unmarshal workflow config of '%s'", w.ID)
	}
	return w, nil
}

// Wait blocks until the workflow with the given ID is marked as completed or canceled by the server.
func (c *client) Wait(ctx context.Context, workflowID string) (canceled bool, err error) {
	req := &proto.WaitRequest{Id: workflowID}

	resp, err := retryRPC(ctx, c, "wait", func() (*proto.WaitResponse, error) {
		if !c.IsConnected() {
			return nil, errNotConnected
		}
		r, err := c.client.Wait(ctx, req)
		return r, classifyRPCErr(ctx, err)
	})
	if err != nil {
		return false, err
	}
	if resp == nil {
		return false, nil
	}
	return resp.GetCanceled(), nil
}

// Init signals the workflow is initialized.
func (c *client) Init(ctx context.Context, workflowID string, state rpc.WorkflowState) error {
	req := &proto.InitRequest{
		Id: workflowID,
		State: &proto.WorkflowState{
			Started:  state.Started,
			Finished: state.Finished,
			Error:    state.Error,
			Canceled: state.Canceled,
		},
	}

	_, err := retryRPC(ctx, c, "init", func() (*proto.Empty, error) {
		if !c.IsConnected() {
			return nil, errNotConnected
		}
		r, err := c.client.Init(ctx, req)
		return r, classifyRPCErr(ctx, err)
	})
	return err
}

// Done let agent signal to server the workflow has stopped.
func (c *client) Done(ctx context.Context, workflowID string, state rpc.WorkflowState) error {
	req := &proto.DoneRequest{
		Id: workflowID,
		State: &proto.WorkflowState{
			Started:  state.Started,
			Finished: state.Finished,
			Error:    state.Error,
			Canceled: state.Canceled,
		},
	}

	_, err := retryRPC(ctx, c, "done", func() (*proto.Empty, error) {
		if !c.IsConnected() {
			return nil, errNotConnected
		}
		r, err := c.client.Done(ctx, req)
		return r, classifyRPCErr(ctx, err)
	})
	return err
}

// Extend extends the workflow deadline.
func (c *client) Extend(ctx context.Context, workflowID string) error {
	req := &proto.ExtendRequest{Id: workflowID}

	_, err := retryRPC(ctx, c, "extend", func() (*proto.Empty, error) {
		if !c.IsConnected() {
			return nil, errNotConnected
		}
		r, err := c.client.Extend(ctx, req)
		return r, classifyRPCErr(ctx, err)
	})
	return err
}

// Update let agent updates the step state at the server.
func (c *client) Update(ctx context.Context, workflowID string, state rpc.StepState) error {
	req := &proto.UpdateRequest{
		Id: workflowID,
		State: &proto.StepState{
			StepUuid: state.StepUUID,
			Started:  state.Started,
			Finished: state.Finished,
			Exited:   state.Exited,
			ExitCode: int32(state.ExitCode),
			Error:    state.Error,
			Canceled: state.Canceled,
			Skipped:  state.Skipped,
		},
	}

	_, err := retryRPC(ctx, c, "update", func() (*proto.Empty, error) {
		if !c.IsConnected() {
			return nil, errNotConnected
		}
		r, err := c.client.Update(ctx, req)
		return r, classifyRPCErr(ctx, err)
	})
	return err
}

// EnqueueLog queues the log entry to be written in a batch later.
func (c *client) EnqueueLog(logEntry *rpc.LogEntry) {
	c.logs <- &proto.LogEntry{
		StepUuid: logEntry.StepUUID,
		Data:     logEntry.Data,
		Line:     int32(logEntry.Line),
		Time:     logEntry.Time,
		Type:     int32(logEntry.Type),
	}
}

func (c *client) processLogs(ctx context.Context) {
	var entries []*proto.LogEntry
	var bytes int

	send := func() {
		if len(entries) == 0 {
			return
		}

		log.Debug().
			Int("entries", len(entries)).
			Int("bytes", bytes).
			Msg("log drain: sending queued logs")

		if err := c.sendLogs(ctx, entries); err != nil {
			log.Error().Err(err).Msg("log drain: could not send logs to server")
		}

		// even if send failed, we don't have infinite memory; retry has already been used
		entries = entries[:0]
		bytes = 0
	}

	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-c.logs:
			if !ok {
				log.Info().Msg("log drain: channel closed")
				send()
				return
			}

			entries = append(entries, entry)
			bytes += grpc_proto.Size(entry)

			if bytes >= maxLogBatchSize {
				send()
			}

		case <-time.After(maxLogFlushPeriod):
			send()
		}
	}
}

func (c *client) sendLogs(ctx context.Context, entries []*proto.LogEntry) error {
	req := &proto.LogRequest{LogEntries: entries}

	// sendLogs intentionally does not gate on IsConnected — the original code
	// didn't either. backoff.Retry will keep trying through transient transport
	// errors until MaxElapsedTime elapses.
	_, err := retryRPC(ctx, c, "log", func() (*proto.Empty, error) {
		r, err := c.client.Log(ctx, req)
		return r, classifyRPCErr(ctx, err)
	})
	return err
}

func (c *client) RegisterAgent(ctx context.Context, info rpc.AgentInfo) (int64, error) {
	req := &proto.RegisterAgentRequest{
		Info: &proto.AgentInfo{
			Platform:     info.Platform,
			Backend:      info.Backend,
			Version:      info.Version,
			Capacity:     int32(info.Capacity),
			CustomLabels: info.CustomLabels,
		},
	}

	res, err := c.client.RegisterAgent(ctx, req)
	return res.GetAgentId(), err
}

func (c *client) UnregisterAgent(ctx context.Context) error {
	_, err := c.client.UnregisterAgent(ctx, &proto.Empty{})
	return err
}

func (c *client) ReportHealth(ctx context.Context) error {
	req := &proto.ReportHealthRequest{Status: "I am alive!"}

	_, err := retryRPC(ctx, c, "report_health", func() (*proto.Empty, error) {
		if !c.IsConnected() {
			return nil, errNotConnected
		}
		r, err := c.client.ReportHealth(ctx, req)
		return r, classifyRPCErr(ctx, err)
	})
	return err
}
