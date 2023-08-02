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

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
	"github.com/woodpecker-ci/woodpecker/pipeline/rpc/proto"
)

var backoff = time.Second

// set grpc version on compile time to compare against server version response
const ClientGrpcVersion int32 = proto.Version

type client struct {
	client proto.WoodpeckerClient
	conn   *grpc.ClientConn
}

// NewGrpcClient returns a new grpc Client.
func NewGrpcClient(conn *grpc.ClientConn) rpc.Peer {
	client := new(client)
	client.client = proto.NewWoodpeckerClient(conn)
	client.conn = conn
	return client
}

func (c *client) Close() error {
	return c.conn.Close()
}

// Version returns the server- & grpc-version
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

// Next returns the next pipeline in the queue.
func (c *client) Next(ctx context.Context, f rpc.Filter) (*rpc.Pipeline, error) {
	var res *proto.NextResponse
	var err error
	req := new(proto.NextRequest)
	req.Filter = new(proto.Filter)
	req.Filter.Labels = f.Labels
	for {
		res, err = c.client.Next(ctx, req)
		if err == nil {
			break
		}

		// TODO: remove after adding continuous data exchange by something like #536
		if strings.Contains(err.Error(), "\"too_many_pings\"") {
			// https://github.com/woodpecker-ci/woodpecker/issues/717#issuecomment-1049365104
			log.Trace().Err(err).Msg("grpc: to many keepalive pings without sending data")
		} else {
			log.Err(err).Msgf("grpc error: done(): code: %v: %s", status.Code(err), err)
		}

		switch status.Code(err) {
		case
			codes.Aborted,
			codes.DataLoss,
			codes.DeadlineExceeded,
			codes.Internal,
			codes.Unavailable:
			// non-fatal errors
		default:
			return nil, err
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		<-time.After(backoff)
	}

	if res.GetPipeline() == nil {
		return nil, nil
	}

	p := new(rpc.Pipeline)
	p.ID = res.GetPipeline().GetId()
	p.Timeout = res.GetPipeline().GetTimeout()
	p.Config = new(backend.Config)
	if err := json.Unmarshal(res.GetPipeline().GetPayload(), p.Config); err != nil {
		log.Error().Err(err).Msgf("could not unmarshal pipeline config of '%s'", p.ID)
	}
	return p, nil
}

// Wait blocks until the pipeline is complete.
func (c *client) Wait(ctx context.Context, id string) (err error) {
	req := new(proto.WaitRequest)
	req.Id = id
	for {
		_, err = c.client.Wait(ctx, req)
		if err == nil {
			break
		}

		log.Err(err).Msgf("grpc error: wait(): code: %v: %s", status.Code(err), err)

		switch status.Code(err) {
		case
			codes.Aborted,
			codes.DataLoss,
			codes.DeadlineExceeded,
			codes.Internal,
			codes.Unavailable:
			// non-fatal errors
		default:
			return err
		}
		<-time.After(backoff)
	}
	return nil
}

// Init signals the pipeline is initialized.
func (c *client) Init(ctx context.Context, id string, state rpc.State) (err error) {
	req := new(proto.InitRequest)
	req.Id = id
	req.State = new(proto.State)
	req.State.Error = state.Error
	req.State.ExitCode = int32(state.ExitCode)
	req.State.Exited = state.Exited
	req.State.Finished = state.Finished
	req.State.Started = state.Started
	req.State.Name = state.Step
	for {
		_, err = c.client.Init(ctx, req)
		if err == nil {
			break
		}

		log.Err(err).Msgf("grpc error: init(): code: %v: %s", status.Code(err), err)

		switch status.Code(err) {
		case
			codes.Aborted,
			codes.DataLoss,
			codes.DeadlineExceeded,
			codes.Internal,
			codes.Unavailable:
			// non-fatal errors
		default:
			return err
		}
		<-time.After(backoff)
	}
	return nil
}

// Done signals the pipeline is complete.
func (c *client) Done(ctx context.Context, id string, state rpc.State) (err error) {
	req := new(proto.DoneRequest)
	req.Id = id
	req.State = new(proto.State)
	req.State.Error = state.Error
	req.State.ExitCode = int32(state.ExitCode)
	req.State.Exited = state.Exited
	req.State.Finished = state.Finished
	req.State.Started = state.Started
	req.State.Name = state.Step
	for {
		_, err = c.client.Done(ctx, req)
		if err == nil {
			break
		}

		log.Err(err).Msgf("grpc error: done(): code: %v: %s", status.Code(err), err)

		switch status.Code(err) {
		case
			codes.Aborted,
			codes.DataLoss,
			codes.DeadlineExceeded,
			codes.Internal,
			codes.Unavailable:
			// non-fatal errors
		default:
			return err
		}
		<-time.After(backoff)
	}
	return nil
}

// Extend extends the pipeline deadline
func (c *client) Extend(ctx context.Context, id string) (err error) {
	req := new(proto.ExtendRequest)
	req.Id = id
	for {
		_, err = c.client.Extend(ctx, req)
		if err == nil {
			break
		}

		log.Err(err).Msgf("grpc error: extend(): code: %v: %s", status.Code(err), err)

		switch status.Code(err) {
		case
			codes.Aborted,
			codes.DataLoss,
			codes.DeadlineExceeded,
			codes.Internal,
			codes.Unavailable:
			// non-fatal errors
		default:
			return err
		}
		<-time.After(backoff)
	}
	return nil
}

// Update updates the pipeline state.
func (c *client) Update(ctx context.Context, id string, state rpc.State) (err error) {
	req := new(proto.UpdateRequest)
	req.Id = id
	req.State = new(proto.State)
	req.State.Error = state.Error
	req.State.ExitCode = int32(state.ExitCode)
	req.State.Exited = state.Exited
	req.State.Finished = state.Finished
	req.State.Started = state.Started
	req.State.Name = state.Step
	for {
		_, err = c.client.Update(ctx, req)
		if err == nil {
			break
		}

		log.Err(err).Msgf("grpc error: update(): code: %v: %s", status.Code(err), err)

		switch status.Code(err) {
		case
			codes.Aborted,
			codes.DataLoss,
			codes.DeadlineExceeded,
			codes.Internal,
			codes.Unavailable:
			// non-fatal errors
		default:
			return err
		}
		<-time.After(backoff)
	}
	return nil
}

// Log writes the pipeline log entry.
func (c *client) Log(ctx context.Context, logEntry *rpc.LogEntry) (err error) {
	req := new(proto.LogRequest)
	req.LogEntry = new(proto.LogEntry)
	req.LogEntry.StepUuid = logEntry.StepUUID
	req.LogEntry.Data = logEntry.Data
	req.LogEntry.Line = int32(logEntry.Line)
	req.LogEntry.Time = logEntry.Time
	req.LogEntry.Type = int32(logEntry.Type)
	for {
		_, err = c.client.Log(ctx, req)
		if err == nil {
			break
		}

		log.Err(err).Msgf("grpc error: log(): code: %v: %s", status.Code(err), err)

		switch status.Code(err) {
		case
			codes.Aborted,
			codes.DataLoss,
			codes.DeadlineExceeded,
			codes.Internal,
			codes.Unavailable:
			// non-fatal errors
		default:
			return err
		}
		<-time.After(backoff)
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

func (c *client) ReportHealth(ctx context.Context) (err error) {
	req := new(proto.ReportHealthRequest)
	req.Status = "I am alive!"

	for {
		_, err = c.client.ReportHealth(ctx, req)
		if err == nil {
			return nil
		}
		switch status.Code(err) {
		case
			codes.Aborted,
			codes.DataLoss,
			codes.DeadlineExceeded,
			codes.Internal,
			codes.Unavailable:
			// non-fatal errors
		default:
			return err
		}
		<-time.After(backoff)
	}
}
