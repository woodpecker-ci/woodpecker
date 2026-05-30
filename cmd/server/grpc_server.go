// Copyright 2024 Woodpecker Authors
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

package main

import (
	"context"
	"fmt"
	"net"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	server_rpc "go.woodpecker-ci.org/woodpecker/v3/server/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

func runGrpcServer(ctx context.Context, c *cli.Command, _store store.Store) error {
	lis, err := net.Listen("tcp", c.String("grpc-addr"))
	if err != nil {
		return fmt.Errorf("failed to listen on grpc-addr: %w", err)
	}

	return server_rpc.Serve(ctx, server_rpc.ServeConfig{
		Listener:         lis,
		Store:            _store,
		Scheduler:        server.Config.Services.Scheduler,
		Logger:           server.Config.Services.Logs,
		JWTSecret:        c.String("grpc-secret"),
		AgentToken:       server.Config.Server.AgentToken,
		KeepaliveMinTime: c.Duration("keepalive-min-time"),
		Registerer:       prometheus.DefaultRegisterer,
	})
}
