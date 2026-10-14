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

package rpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"

	"go.woodpecker-ci.org/woodpecker/v3/rpc/proto"
	"go.woodpecker-ci.org/woodpecker/v3/server/logging"
	"go.woodpecker-ci.org/woodpecker/v3/server/scheduler"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// ServeConfig bundles everything Serve needs. Callers build this from a
// *cli.Command (production) or with literals (tests).
type ServeConfig struct {
	Listener         net.Listener
	Store            store.Store
	Scheduler        scheduler.Scheduler
	Logger           logging.Log
	JWTSecret        string
	AgentToken       string
	KeepaliveMinTime time.Duration
	// Registerer is where the server's prometheus metrics are registered.
	// Pass prometheus.DefaultRegisterer in production; pass a fresh
	// prometheus.NewRegistry() in tests to avoid duplicate-registration
	// panics when the server is created multiple times.
	Registerer prometheus.Registerer
}

// errServerShutdown is the cancel cause of long-polling RPCs that got
// released because the server is shutting down.
var errServerShutdown = errors.New("server is shutting down")

// Serve registers Woodpecker's gRPC services on cfg.Listener and blocks
// until ctx is canceled or Serve returns an error. GracefulStop is
// triggered on ctx cancellation, releasing agents that wait in Next or Wait
// (see releaseOnShutdownInterceptor). The listener is owned by Serve — it is
// closed when grpc.Server.Serve returns.
func Serve(ctx context.Context, cfg ServeConfig) error {
	jwtManager := NewJWTManager(cfg.JWTSecret)
	authorizer := NewAuthorizer(jwtManager)

	grpcCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(authorizer.StreamInterceptor),
		grpc.ChainUnaryInterceptor(
			authorizer.UnaryInterceptor,
			releaseOnShutdownInterceptor(grpcCtx),
		),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime: cfg.KeepaliveMinTime,
		}),
	)

	proto.RegisterWoodpeckerServer(grpcServer, NewWoodpeckerServer(
		cfg.Scheduler, cfg.Logger, cfg.Store, cfg.Registerer,
	))
	proto.RegisterWoodpeckerAuthServer(grpcServer, NewWoodpeckerAuthServer(
		jwtManager, cfg.AgentToken, cfg.Store,
	))

	go func() {
		<-grpcCtx.Done()
		log.Info().Msg("terminating grpc service gracefully")
		grpcServer.GracefulStop()
		log.Info().Msg("grpc service stopped")
	}()

	if err := grpcServer.Serve(cfg.Listener); err != nil {
		return fmt.Errorf("grpc server failed: %w", err)
	}
	return nil
}

// releaseOnShutdownInterceptor ends the long-polling Next and Wait calls as
// soon as shutdown is canceled, so GracefulStop does not wait for connected
// agents to go away. Released calls fail with codes.Unavailable, which agents
// retry until they reach the server again.
func releaseOnShutdownInterceptor(shutdown context.Context) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if info.FullMethod != proto.Woodpecker_Next_FullMethodName &&
			info.FullMethod != proto.Woodpecker_Wait_FullMethodName {
			return handler(ctx, req)
		}

		ctx, cancel := context.WithCancelCause(ctx)
		defer cancel(nil)
		stop := context.AfterFunc(shutdown, func() { cancel(errServerShutdown) })
		defer stop()

		resp, err := handler(ctx, req)
		if !errors.Is(context.Cause(ctx), errServerShutdown) {
			return resp, err
		}

		// Results that raced with the shutdown still have to reach the agent:
		// a workflow already assigned to it or a cancel signal.
		if err == nil {
			switch r := resp.(type) {
			case *proto.NextResponse:
				if r.GetWorkflow() != nil {
					return resp, nil
				}
			case *proto.WaitResponse:
				if r.GetCanceled() {
					return resp, nil
				}
			}
		}

		return nil, status.Error(codes.Unavailable, errServerShutdown.Error())
	}
}
