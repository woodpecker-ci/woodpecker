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

//go:build test

package setup

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

const (
	defaultTimeout  = 30 * time.Second
	defaultInterval = 100 * time.Millisecond
)

// isTerminal returns true if the status is a final (non-running) state.
func isTerminal(s model.StatusValue) bool {
	switch s {
	case model.StatusSuccess, model.StatusFailure, model.StatusKilled,
		model.StatusError, model.StatusDeclined, model.StatusCanceled:
		return true
	}
	return false
}

// WaitForPipeline polls the store until the pipeline with the given ID reaches
// a terminal status, then returns it. Fails the test if timeout is exceeded.
func WaitForPipeline(t *testing.T, s store.Store, pipelineID int64) *model.Pipeline {
	t.Helper()
	return WaitForPipelineStatus(t, s, pipelineID, "", defaultTimeout)
}

// WaitForPipelineStatus polls until the pipeline reaches wantStatus (or any
// terminal status if wantStatus is empty). Fails the test on timeout.
func WaitForPipelineStatus(t *testing.T, s store.Store, pipelineID int64, wantStatus model.StatusValue, timeout time.Duration) *model.Pipeline {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		p, err := s.GetPipeline(pipelineID)
		require.NoError(t, err, "get pipeline %d", pipelineID)

		if wantStatus != "" {
			if p.Status == wantStatus {
				return p
			}
		} else if isTerminal(p.Status) {
			return p
		}

		time.Sleep(defaultInterval)
	}

	// Fetch final state for a useful failure message.
	p, _ := s.GetPipeline(pipelineID)
	t.Fatalf("timeout waiting for pipeline %d: last status=%q (want %q)", pipelineID, p.Status, wantStatus)
	return nil
}

// WaitForAgentRegistered polls the store until at least one agent is registered.
// This ensures the agent is ready to accept work before a test triggers a pipeline.
func WaitForAgentRegistered(t *testing.T, s store.Store) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		agents, err := s.AgentList(&model.ListOptions{All: true})
		require.NoError(t, err, "list agents")
		if len(agents) > 0 {
			return
		}
		time.Sleep(defaultInterval)
	}
	t.Fatal("timeout: no agent registered with the server")
}
