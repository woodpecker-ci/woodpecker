// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"go.woodpecker-ci.org/woodpecker/v3/rpc/proto"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/scheduler/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestServeReleasesPollingAgentsOnShutdown(t *testing.T) {
	const jwtSecret = "serve-test-secret"

	store := store_mocks.NewMockStore(t)
	store.On("AgentFind", int64(1)).Return(&model.Agent{ID: 1}, nil)

	polling := make(chan struct{})
	sched := mocks.NewMockScheduler(t)
	sched.On("Poll", mock.Anything, int64(1), mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			close(polling)
			ctx, _ := args.Get(0).(context.Context)
			<-ctx.Done()
		}).
		Return(nil, context.Canceled)

	lis := bufconn.Listen(1024 * 1024)
	serveCtx, stopServe := context.WithCancelCause(t.Context())
	defer stopServe(nil)
	served := make(chan error, 1)
	go func() {
		served <- Serve(serveCtx, ServeConfig{
			Listener:   lis,
			Store:      store,
			Scheduler:  sched,
			JWTSecret:  jwtSecret,
			Registerer: prometheus.NewRegistry(),
		})
	}()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	token, err := NewJWTManager(jwtSecret).Generate(1)
	require.NoError(t, err)
	// like the agent, Next is called without a deadline
	clientCtx := metadata.AppendToOutgoingContext(t.Context(), "token", token)

	nextErr := make(chan error, 1)
	go func() {
		_, err := proto.NewWoodpeckerClient(conn).Next(clientCtx, &proto.NextRequest{
			Filter: &proto.Filter{Labels: map[string]string{"platform": "linux/amd64"}},
		})
		nextErr <- err
	}()

	select {
	case <-polling:
	case <-time.After(5 * time.Second):
		t.Fatal("Next did not reach the scheduler")
	}

	stopServe(nil)

	select {
	case err := <-served:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not return while an agent was polling")
	}

	select {
	case err := <-nextErr:
		assert.Equal(t, codes.Unavailable, status.Code(err), "unexpected error: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("polling agent was not released")
	}
}

func TestReleaseOnShutdownInterceptor(t *testing.T) {
	// blocking mimics the long polls: it waits for its context and then
	// returns resp without error, like the queue does for Wait.
	blocking := func(resp any) grpc.UnaryHandler {
		return func(ctx context.Context, _ any) (any, error) {
			<-ctx.Done()
			return resp, nil
		}
	}

	shutdown := func() context.Context {
		ctx, cancel := context.WithCancelCause(t.Context())
		cancel(nil)
		return ctx
	}

	t.Run("released Wait returns Unavailable", func(t *testing.T) {
		interceptor := releaseOnShutdownInterceptor(shutdown())
		info := &grpc.UnaryServerInfo{FullMethod: proto.Woodpecker_Wait_FullMethodName}

		resp, err := interceptor(t.Context(), nil, info, blocking(&proto.WaitResponse{}))
		assert.Nil(t, resp)
		assert.Equal(t, codes.Unavailable, status.Code(err))
	})

	t.Run("cancel signal still reaches the agent", func(t *testing.T) {
		interceptor := releaseOnShutdownInterceptor(shutdown())
		info := &grpc.UnaryServerInfo{FullMethod: proto.Woodpecker_Wait_FullMethodName}

		resp, err := interceptor(t.Context(), nil, info, blocking(&proto.WaitResponse{Canceled: true}))
		assert.NoError(t, err)
		assert.Equal(t, &proto.WaitResponse{Canceled: true}, resp)
	})

	t.Run("assigned workflow still reaches the agent", func(t *testing.T) {
		interceptor := releaseOnShutdownInterceptor(shutdown())
		info := &grpc.UnaryServerInfo{FullMethod: proto.Woodpecker_Next_FullMethodName}
		want := &proto.NextResponse{Workflow: &proto.Workflow{Id: "1"}}

		resp, err := interceptor(t.Context(), nil, info, blocking(want))
		assert.NoError(t, err)
		assert.Equal(t, want, resp)
	})

	t.Run("other calls are not canceled", func(t *testing.T) {
		interceptor := releaseOnShutdownInterceptor(shutdown())
		info := &grpc.UnaryServerInfo{FullMethod: proto.Woodpecker_Done_FullMethodName}

		resp, err := interceptor(t.Context(), nil, info, func(ctx context.Context, _ any) (any, error) {
			return &proto.Empty{}, ctx.Err()
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("agent disconnect is passed through", func(t *testing.T) {
		interceptor := releaseOnShutdownInterceptor(t.Context())
		info := &grpc.UnaryServerInfo{FullMethod: proto.Woodpecker_Next_FullMethodName}
		ctx, cancel := context.WithCancelCause(t.Context())
		cancel(nil)

		_, err := interceptor(ctx, nil, info, func(ctx context.Context, _ any) (any, error) {
			return nil, ctx.Err()
		})
		assert.ErrorIs(t, err, context.Canceled)
	})
}
