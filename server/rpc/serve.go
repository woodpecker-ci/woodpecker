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
	"fmt"
	"net"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

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

// Serve registers Woodpecker's gRPC services on cfg.Listener and blocks
// until ctx is canceled or Serve returns an error. GracefulStop is
// triggered on ctx cancellation. The listener is owned by Serve — it is
// closed when grpc.Server.Serve returns.
func Serve(ctx context.Context, cfg ServeConfig) error {
	jwtManager := NewJWTManager(cfg.JWTSecret)
	authorizer := NewAuthorizer(jwtManager)

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(authorizer.StreamInterceptor),
		grpc.UnaryInterceptor(authorizer.UnaryInterceptor),
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

	grpcCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

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
