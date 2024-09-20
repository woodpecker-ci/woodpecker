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
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	grpcproto "google.golang.org/protobuf/proto"

	backend "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc/proto"
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
)

type client struct {
	client proto.WoodpeckerClient
	conn   *grpc.ClientConn
	logs   chan *proto.LogEntry
}

// NewGrpcClient returns a new grpc Client.
func NewGrpcClient(ctx context.Context, conn *grpc.ClientConn) rpc.Peer {
	client := new(client)
	client.client = proto.NewWoodpeckerClient(conn)
	client.conn = conn
	client.logs = make(chan *proto.LogEntry, 10) // max memory use: 10 lines * 1 MiB
	go client.processLogs(ctx)
	return client
}

func (c *client) Close() error {
	close(c.logs)
	return c.conn.Close()
}

func (c *client) newBackOff() backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 0
	b.MaxInterval = 10 * time.Second          //nolint:mnd
	b.InitialInterval = 10 * time.Millisecond //nolint:mnd
	return b
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
	var res *proto.NextResponse
	var err error
	retry := c.newBackOff()
	req := new(proto.NextRequest)
	req.Filter = new(proto.Filter)
	req.Filter.Labels = filter.Labels
	for {
		res, err = c.client.Next(ctx, req)
		if err == nil {
			break
		}

		switch status.Code(err) {
		case codes.Canceled:
			if ctx.Err() != nil {
				// expected as context was canceled
				log.Debug().Err(err).Msgf("grpc error: next(): context canceled")
				return nil, nil
			}
			log.Error().Err(err).Msgf("grpc error: next(): code: %v", status.Code(err))
			return nil, err
		case
			codes.Aborted,
			codes.DataLoss,
			codes.DeadlineExceeded,
			codes.Internal,
			codes.Unavailable:
			// non-fatal errors
			// TODO: remove after adding continuous data exchange by something like #536
			if strings.Contains(err.Error(), "\"too_many_pings\"") {
				// https://github.com/woodpecker-ci/woodpecker/issues/717#issuecomment-1049365104
				log.Trace().Err(err).Msg("grpc: to many keepalive pings without sending data")
			} else {
				log.Warn().Err(err).Msgf("grpc error: next(): code: %v", status.Code(err))
			}
		default:
			log.Error().Err(err).Msgf("grpc error: next(): code: %v", status.Code(err))
			return nil, err
		}

		select {
		case <-time.After(retry.NextBackOff()):
		case <-ctx.Done():
			return nil, nil
		}
	}

	if res.GetWorkflow() == nil {
		return nil, nil
	}

	w := new(rpc.Workflow)
	w.ID = res.GetWorkflow().GetId()
	w.Timeout = res.GetWorkflow().GetTimeout()
	w.Config = new(backend.Config)
	if err := json.Unmarshal(res.GetWorkflow().GetPayload(), w.Config); err != nil {
		log.Error().Err(err).Msgf("could not unmarshal workflow config of '%s'", w.ID)
	}
	return w, nil
}

// Wait blocks until the workflow is complete.
func (c *client) Wait(ctx context.Context, workflowID string) (err error) {
	retry := c.newBackOff()
	req := new(proto.WaitRequest)
	req.Id = workflowID
	for {
		_, err = c.client.Wait(ctx, req)
		if err == nil {
			break
		}

		switch status.Code(err) {
		case codes.Canceled:
			if ctx.Err() != nil {
				// expected as context was canceled
				log.Debug().Err(err).Msgf("grpc error: wait(): context canceled")
				return nil
			}
			log.Error().Err(err).Msgf("grpc error: wait(): code: %v", status.Code(err))
			return err
		case
			codes.Aborted,
			codes.DataLoss,
			codes.DeadlineExceeded,
			codes.Internal,
			codes.Unavailable:
			// non-fatal errors
			log.Warn().Err(err).Msgf("grpc error: wait(): code: %v", status.Code(err))
		default:
			log.Error().Err(err).Msgf("grpc error: wait(): code: %v", status.Code(err))
			return err
		}

		select {
		case <-time.After(retry.NextBackOff()):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// Init signals the workflow is initialized.
func (c *client) Init(ctx context.Context, workflowID string, state rpc.WorkflowState) (err error) {
	retry := c.newBackOff()
	req := new(proto.InitRequest)
	req.Id = workflowID
	req.State = new(proto.WorkflowState)
	req.State.Started = state.Started
	req.State.Finished = state.Finished
	req.State.Error = state.Error
	for {
		_, err = c.client.Init(ctx, req)
		if err == nil {
			break
		}

		log.Error().Err(err).Msgf("grpc error: init(): code: %v", status.Code(err))

		switch status.Code(err) {
		case codes.Canceled:
			if ctx.Err() != nil {
				// expected as context was canceled
				log.Debug().Err(err).Msgf("grpc error: init(): context canceled")
				return nil
			}
			log.Error().Err(err).Msgf("grpc error: init(): code: %v", status.Code(err))
			return err
		case
			codes.Aborted,
			codes.DataLoss,
			codes.DeadlineExceeded,
			codes.Internal,
			codes.Unavailable:
			// non-fatal errors
			log.Warn().Err(err).Msgf("grpc error: init(): code: %v", status.Code(err))
		default:
			log.Error().Err(err).Msgf("grpc error: init(): code: %v", status.Code(err))
			return err
		}

		select {
		case <-time.After(retry.NextBackOff()):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// Done signals the workflow is complete.
func (c *client) Done(ctx context.Context, workflowID string, state rpc.WorkflowState) (err error) {
	retry := c.newBackOff()
	req := new(proto.DoneRequest)
	req.Id = workflowID
	req.State = new(proto.WorkflowState)
	req.State.Started = state.Started
	req.State.Finished = state.Finished
	req.State.Error = state.Error
	for {
		_, err = c.client.Done(ctx, req)
		if err == nil {
			break
		}

		log.Error().Err(err).Msgf("grpc error: done(): code: %v", status.Code(err))

		switch status.Code(err) {
		case codes.Canceled:
			if ctx.Err() != nil {
				// expected as context was canceled
				log.Debug().Err(err).Msgf("grpc error: done(): context canceled")
				return nil
			}
			log.Error().Err(err).Msgf("grpc error: done(): code: %v", status.Code(err))
			return err
		case
			codes.Aborted,
			codes.DataLoss,
			codes.DeadlineExceeded,
			codes.Internal,
			codes.Unavailable:
			// non-fatal errors
			log.Warn().Err(err).Msgf("grpc error: done(): code: %v", status.Code(err))
		default:
			log.Error().Err(err).Msgf("grpc error: done(): code: %v", status.Code(err))
			return err
		}

		select {
		case <-time.After(retry.NextBackOff()):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// Extend extends the workflow deadline.
func (c *client) Extend(ctx context.Context, workflowID string) (err error) {
	retry := c.newBackOff()
	req := new(proto.ExtendRequest)
	req.Id = workflowID
	for {
		_, err = c.client.Extend(ctx, req)
		if err == nil {
			break
		}

		log.Error().Err(err).Msgf("grpc error: extend(): code: %v", status.Code(err))

		switch status.Code(err) {
		case codes.Canceled:
			if ctx.Err() != nil {
				// expected as context was canceled
				log.Debug().Err(err).Msgf("grpc error: extend(): context canceled")
				return nil
			}
			log.Error().Err(err).Msgf("grpc error: extend(): code: %v", status.Code(err))
			return err
		case
			codes.Aborted,
			codes.DataLoss,
			codes.DeadlineExceeded,
			codes.Internal,
			codes.Unavailable:
			// non-fatal errors
			log.Warn().Err(err).Msgf("grpc error: extend(): code: %v", status.Code(err))
		default:
			log.Error().Err(err).Msgf("grpc error: extend(): code: %v", status.Code(err))
			return err
		}

		select {
		case <-time.After(retry.NextBackOff()):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// Update updates the workflow state.
func (c *client) Update(ctx context.Context, workflowID string, state rpc.StepState) (err error) {
	retry := c.newBackOff()
	req := new(proto.UpdateRequest)
	req.Id = workflowID
	req.State = new(proto.StepState)
	req.State.StepUuid = state.StepUUID
	req.State.Started = state.Started
	req.State.Finished = state.Finished
	req.State.Exited = state.Exited
	req.State.ExitCode = int32(state.ExitCode)
	req.State.Error = state.Error
	for {
		_, err = c.client.Update(ctx, req)
		if err == nil {
			break
		}

		log.Error().Err(err).Msgf("grpc error: update(): code: %v", status.Code(err))

		switch status.Code(err) {
		case codes.Canceled:
			if ctx.Err() != nil {
				// expected as context was canceled
				log.Debug().Err(err).Msgf("grpc error: update(): context canceled")
				return nil
			}
			log.Error().Err(err).Msgf("grpc error: update(): code: %v", status.Code(err))
			return err
		case
			codes.Aborted,
			codes.DataLoss,
			codes.DeadlineExceeded,
			codes.Internal,
			codes.Unavailable:
			// non-fatal errors
			log.Warn().Err(err).Msgf("grpc error: update(): code: %v", status.Code(err))
		default:
			log.Error().Err(err).Msgf("grpc error: update(): code: %v", status.Code(err))
			return err
		}

		select {
		case <-time.After(retry.NextBackOff()):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
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

	// ctx.Done() is covered by the log channel being closed
	for {
		select {
		case entry, ok := <-c.logs:
			if !ok {
				log.Info().Msg("log drain: channel closed")
				send()
				return
			}

			entries = append(entries, entry)
			bytes += grpcproto.Size(entry) // cspell:words grpcproto

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
	retry := c.newBackOff()

	for {
		_, err := c.client.Log(ctx, req)
		if err == nil {
			break
		}

		switch status.Code(err) {
		case codes.Canceled:
			if ctx.Err() != nil {
				// expected as context was canceled
				log.Debug().Err(err).Msgf("grpc error: log(): context canceled")
				return nil
			}
			log.Error().Err(err).Msgf("grpc error: log(): code: %v", status.Code(err))
			return err
		case
			codes.Aborted,
			codes.DataLoss,
			codes.DeadlineExceeded,
			codes.Internal,
			codes.Unavailable:
			// non-fatal errors
			log.Warn().Err(err).Msgf("grpc error: log(): code: %v", status.Code(err))
		default:
			log.Error().Err(err).Msgf("grpc error: log(): code: %v", status.Code(err))
			return err
		}

		select {
		case <-time.After(retry.NextBackOff()):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (c *client) RegisterAgent(ctx context.Context, platform, backend, version string, capacity int) (int64, error) {
	req := new(proto.RegisterAgentRequest)
	req.Platform = platform
	req.Backend = backend
	req.Version = version
	req.Capacity = int32(capacity)

	res, err := c.client.RegisterAgent(ctx, req)
	return res.GetAgentId(), err
}

func (c *client) UnregisterAgent(ctx context.Context) error {
	_, err := c.client.UnregisterAgent(ctx, &proto.Empty{})
	return err
}

func (c *client) ReportHealth(ctx context.Context) (err error) {
	retry := c.newBackOff()
	req := new(proto.ReportHealthRequest)
	req.Status = "I am alive!"

	for {
		_, err = c.client.ReportHealth(ctx, req)
		if err == nil {
			return nil
		}
		switch status.Code(err) {
		case codes.Canceled:
			if ctx.Err() != nil {
				// expected as context was canceled
				log.Debug().Err(err).Msgf("grpc error: report_health(): context canceled")
				return nil
			}
			log.Error().Err(err).Msgf("grpc error: report_health(): code: %v", status.Code(err))
			return err
		case
			codes.Aborted,
			codes.DataLoss,
			codes.DeadlineExceeded,
			codes.Internal,
			codes.Unavailable:
			// non-fatal errors
			log.Warn().Err(err).Msgf("grpc error: report_health(): code: %v", status.Code(err))
		default:
			log.Error().Err(err).Msgf("grpc error: report_health(): code: %v", status.Code(err))
			return err
		}

		select {
		case <-time.After(retry.NextBackOff()):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
