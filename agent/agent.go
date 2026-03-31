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

package agent

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/metadata"

	agent_rpc "go.woodpecker-ci.org/woodpecker/v3/agent/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/agent/runner"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/version"
)

// Run is the main agent lifecycle:
//
//  1. Establish gRPC connections and verify protocol version.
//  2. Initialize the pipeline backend.
//  3. Register with the server; persist the assigned AgentID if configured.
//  4. Start background goroutines: health reporting, unregister-on-shutdown,
//     and one runner goroutine per MaxWorkflows slot.
//  5. Block until all goroutines finish (i.e. until ctx is canceled).
//
// cfg is expected to have been fully populated by the caller (typically
// cmd/agent/core) from CLI flags / environment variables.
func Run(ctx context.Context, cfg Config, backends []types.Backend) error {
	log.Info().
		Str("version", version.String()).
		Msg("starting Woodpecker agent")

	// Apply defaults for optional timing fields.
	if cfg.AuthInterceptorRefreshInterval == 0 {
		cfg.AuthInterceptorRefreshInterval = DefaultAuthInterceptorRefreshInterval
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = DefaultShutdownTimeout
	}

	// agentCtx is canceled when the agent should stop accepting new work.
	// shutdownCtx gives in-flight RPCs a grace period after agentCtx is done.
	agentCtx, cancelAgent := context.WithCancelCause(ctx)
	defer cancelAgent(nil)

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelShutdown()

	// -------------------------------------------------------------------------
	// gRPC connections
	// -------------------------------------------------------------------------

	// grpcClientCtx is independent of agentCtx so we can still call
	// UnregisterAgent after agentCtx is canceled.
	grpcClientCtx, cancelGRPC := context.WithCancelCause(context.Background())
	defer cancelGRPC(nil)

	conns, _, err := ConnectGRPC(grpcClientCtx, cfg) //nolint:contextcheck
	if err != nil {
		return err
	}
	defer conns.Close()

	client := agent_rpc.NewGrpcClient(ctx, conns.Conn)

	grpcCtx := metadata.NewOutgoingContext(grpcClientCtx, metadata.Pairs("hostname", cfg.Hostname))

	if err := CheckGRPCVersion(grpcCtx, client); err != nil { //nolint:contextcheck
		log.Error().Err(err).Msg("gRPC version check failed")
		return err
	}

	// -------------------------------------------------------------------------
	// Backend
	// -------------------------------------------------------------------------

	// types.CliCommand is consumed by backends that need raw CLI flag access
	// (e.g. the Kubernetes backend). cmd/agent/core must inject the *cli.Command
	// into the context it passes to Run so backends can retrieve it via this key.
	backendCtx := context.WithValue(agentCtx, types.CliCommand, ctx.Value(types.CliCommand))
	backendEngine, err := backend.FindBackend(backendCtx, backends, cfg.BackendEngine)
	if err != nil {
		log.Error().Err(err).Msgf("cannot find backend engine '%s'", cfg.BackendEngine)
		return err
	}
	if !backendEngine.IsAvailable(backendCtx) {
		return fmt.Errorf("selected backend engine %s is unavailable", backendEngine.Name())
	}

	engInfo, err := backendEngine.Load(backendCtx)
	if err != nil {
		log.Error().Err(err).Msg("cannot load backend engine")
		return err
	}
	log.Debug().Msgf("loaded %s backend engine", backendEngine.Name())

	// -------------------------------------------------------------------------
	// Registration
	// -------------------------------------------------------------------------

	agentID, err := registerAgent(grpcCtx, client, cfg, engInfo.Platform) //nolint:contextcheck
	if err != nil {
		return err
	}
	cfg.AgentID = agentID

	agentConfigPersisted := atomic.Bool{}
	if persistAgentID(agentID, cfg.PersistAgentID) {
		agentConfigPersisted.Store(true)
	}

	// -------------------------------------------------------------------------
	// Runner state counter (shared across all runner goroutines)
	// -------------------------------------------------------------------------

	counter := &runner.State{
		Polling:  cfg.MaxWorkflows,
		Running:  0,
		Metadata: make(map[string]runner.Info),
	}

	// -------------------------------------------------------------------------
	// Background services
	// -------------------------------------------------------------------------

	filter := buildFilter(cfg.Hostname, engInfo.Platform, backendEngine.Name(), cfg.CustomLabels)

	svcGroup := errgroup.Group{}

	// Unregister stateless agents when the agent context is done.
	unregDone := startUnregisterOnShutdown(agentCtx, grpcCtx, client, &agentConfigPersisted)
	svcGroup.Go(func() error {
		<-unregDone
		// Cancel the gRPC client context once unregistration is handled so
		// that the gRPC connections can be closed cleanly.
		cancelGRPC(nil)
		return nil
	})

	// Periodic health reports to the server.
	svcGroup.Go(func() error {
		return runHealthReporter(agentCtx, grpcCtx, client)
	})

	// One runner goroutine per workflow slot.
	for i := range cfg.MaxWorkflows {
		svcGroup.Go(func() error {
			r := runner.NewRunner(client, filter, cfg.Hostname, counter, backendEngine)
			log.Debug().Msgf("created runner %d", i)
			return runWorkflowLoop(agentCtx, shutdownCtx, r)
		})
	}

	log.Info().Msgf(
		"Woodpecker agent '%s' ready: backend=%s platform=%s parallel-workflows=%d",
		version.String(), backendEngine.Name(), engInfo.Platform, cfg.MaxWorkflows,
	)

	return svcGroup.Wait()
}

// runHealthReporter periodically calls ReportHealth until agentCtx is done.
func runHealthReporter(agentCtx, grpcCtx context.Context, client rpc.Peer) error {
	for {
		if err := client.ReportHealth(grpcCtx); err != nil {
			log.Error().Err(err).Msg("failed to report health")
			if grpcCtx.Err() != nil || agentCtx.Err() != nil {
				log.Debug().Msg("terminating health reporting due to context cancellation")
				return nil
			}
		}

		select {
		case <-agentCtx.Done():
			log.Debug().Msg("terminating health reporting")
			return nil
		case <-time.After(DefaultReportHealthInterval):
		}
	}
}

// runWorkflowLoop polls for and executes workflows until agentCtx is done.
// On transient errors it backs off briefly and retries rather than returning.
func runWorkflowLoop(agentCtx, shutdownCtx context.Context, r runner.Runner) error {
	for {
		if agentCtx.Err() != nil {
			return nil
		}

		log.Debug().Msg("polling new workflow")
		if err := r.Run(agentCtx, shutdownCtx); err != nil {
			log.Error().Err(err).Msg("runner error, retrying...")
			if agentCtx.Err() != nil {
				return nil
			}
			// Brief back-off to avoid hammering the server on repeated errors.
			select {
			case <-agentCtx.Done():
				return nil
			case <-time.After(time.Second * 5):
			}
		}
	}
}
