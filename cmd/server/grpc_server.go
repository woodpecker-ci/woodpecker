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
	"os"
	"path/filepath"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	server_rpc "go.woodpecker-ci.org/woodpecker/v3/server/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

func runGrpcServer(ctx context.Context, c *cli.Command, _store store.Store) error {
	network := "tcp"
	addr := c.String("grpc-addr")

	if strings.HasPrefix(addr, "unix://") {
		network = "unix"
		addr, _ = filepath.Abs(strings.TrimPrefix(addr, "unix://"))
		if _, err := os.Stat(filepath.Dir(addr)); os.IsNotExist(err) {
			return fmt.Errorf("can not listen to unix socket, parent folder %q not exist", filepath.Dir(addr))
		}
	}

	lis, err := net.Listen(network, addr)
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
