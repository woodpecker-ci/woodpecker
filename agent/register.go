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

package agent

import (
	"context"
	"sync/atomic"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/version"
)

// registerAgent registers this agent with the server and returns the assigned
// agent ID. It reports the backend name, platform, capacity and any custom
// labels so the server can route workflows appropriately.
func registerAgent(ctx context.Context, client rpc.Peer, cfg Config, platform string) (int64, error) {
	agentID, err := client.RegisterAgent(ctx, rpc.AgentInfo{
		Version:      version.String(),
		Backend:      cfg.BackendEngine,
		Platform:     platform,
		Capacity:     cfg.MaxWorkflows,
		CustomLabels: cfg.CustomLabels,
	})
	if err != nil {
		return 0, err
	}

	log.Debug().Msgf("agent registered with ID %d", agentID)
	return agentID, nil
}

// persistAgentID calls writeFn with the assigned agent ID so the caller can
// persist it (e.g. write to an on-disk config file). Returns true on success,
// which suppresses automatic unregistration on shutdown.
//
// The write logic deliberately lives in cmd/agent/core (which owns AgentConfig
// and writeAgentConfig) to keep the agent package free of config-file I/O.
func persistAgentID(agentID int64, writeFn func(int64) error) bool {
	if writeFn == nil {
		return false
	}
	if err := writeFn(agentID); err != nil {
		log.Error().Err(err).Msg("failed to persist agent ID; agent will unregister on shutdown")
		return false
	}
	return true
}

// startUnregisterOnShutdown spawns a goroutine that waits for agentCtx to be
// cancelled and then, if the agent was not persisted, unregisters it from the
// server using the still-valid grpcCtx.
//
// It returns a channel that is closed once the goroutine has completed so the
// caller can sequence the gRPC connection teardown after it.
func startUnregisterOnShutdown(agentCtx, grpcCtx context.Context, client rpc.Peer, persisted *atomic.Bool) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		<-agentCtx.Done()

		if persisted.Load() {
			// Stateful agent — server keeps the registration across restarts.
			return
		}

		log.Debug().Msg("unregistering stateless agent from server")
		if err := client.UnregisterAgent(grpcCtx); err != nil {
			log.Error().Err(err).Msg("failed to unregister agent from server")
		} else {
			log.Info().Msg("agent unregistered from server")
		}
	}()
	return done
}
