// Copyright 2023 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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

package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.woodpecker-ci.org/woodpecker/v3/agent"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/shared/logger"
)

func run(ctx context.Context, c *cli.Command, backends []types.Backend) error {
	agentConfigPath := c.String("agent-config")

	hostname := c.String("hostname")
	if len(hostname) == 0 {
		hostname, _ = os.Hostname()
	}

	customLabels := make(map[string]string)
	if err := agent.StringSliceAddToMap(c.StringSlice("labels"), customLabels); err != nil {
		return err
	}

	// Read the persisted agent ID (if any) from disk.
	agentConfig := readAgentConfig(agentConfigPath)

	// Inject the CLI command into the context so that backends which need raw
	// flag access (e.g. Kubernetes) can retrieve it via types.CliCommand.
	ctx = context.WithValue(ctx, types.CliCommand, c)

	cfg := agent.Config{
		Server:           c.String("server"),
		GRPCToken:        c.String("grpc-token"),
		GRPCSecure:       c.Bool("grpc-secure"),
		GRPCVerify:       c.Bool("grpc-skip-insecure"),
		KeepaliveTime:    c.Duration("keepalive-time"),
		KeepaliveTimeout: c.Duration("keepalive-timeout"),
		Hostname:         hostname,
		AgentID:          agentConfig.AgentID,
		MaxWorkflows:     c.Int("max-workflows"),
		BackendEngine:    c.String("backend-engine"),
		CustomLabels:     customLabels,
		HealthcheckAddr:  c.String("healthcheck-addr"),

		// PersistAgentID writes the server-assigned ID back to the config file.
		// When agentConfigPath is empty the agent runs stateless and this is nil,
		// which causes agent.Run to unregister on shutdown instead.
		PersistAgentID: func(id int64) error {
			if agentConfigPath == "" {
				return fmt.Errorf("no agent config path configured; running stateless")
			}
			return writeAgentConfig(AgentConfig{AgentID: id}, agentConfigPath)
		},
	}

	counter.Polling = cfg.MaxWorkflows
	counter.Running = 0

	if c.Bool("healthcheck") {
		go serveHealthcheck(ctx, c.String("healthcheck-addr"))
	}

	return agent.Run(ctx, cfg, backends)
}

// serveHealthcheck starts the HTTP health endpoint and shuts it down when ctx
// is cancelled. It is intentionally non-blocking — the caller runs it in a
// goroutine.
func serveHealthcheck(ctx context.Context, addr string) {
	server := &http.Server{Addr: addr}
	go func() {
		<-ctx.Done()
		log.Info().Msg("shutdown healthcheck server ...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), agent.DefaultShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("shutdown healthcheck server failed")
		} else {
			log.Info().Msg("healthcheck server stopped")
		}
	}()
	if err := server.ListenAndServe(); err != nil {
		log.Error().Err(err).Msgf("cannot listen on address %s", addr)
	}
}

func runWithRetry(backendEngines []types.Backend) func(ctx context.Context, c *cli.Command) error {
	return func(ctx context.Context, c *cli.Command) error {
		if err := logger.SetupGlobalLogger(ctx, c, true); err != nil {
			return err
		}

		initHealth()

		retryCount := c.Int("connect-retry-count")
		retryDelay := c.Duration("connect-retry-delay")
		var err error
		for range retryCount {
			if err = run(ctx, c, backendEngines); status.Code(err) == codes.Unavailable {
				log.Warn().Err(err).Msg(fmt.Sprintf("cannot connect to %s, retrying in %v", c.String("server"), retryDelay))
				time.Sleep(retryDelay)
			} else {
				break
			}
		}
		return err
	}
}
