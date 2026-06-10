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

package scenarios

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/e2e/setup"
	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// infraFailYAML has a single step that the dummy backend reports as an
// infrastructure failure (the equivalent of a spot-node preemption: the step
// "failed" but InfraFailure is set, so the server should auto-restart it).
var infraFailYAML = []byte(`
steps:
  - name: preempted
    image: dummy
    commands:
      - echo boom
    environment:
      STEP_INFRA_FAILURE: "true"
      STEP_EXIT_CODE: "1"
`)

// genuineFailYAML has a single step that fails for a real reason (exit 1, no
// infra flag). The server must NOT auto-restart it.
var genuineFailYAML = []byte(`
steps:
  - name: broken
    image: dummy
    commands:
      - echo nope
    environment:
      STEP_EXIT_CODE: "1"
`)

// withInfraRetryMax sets the global infra-retry budget for the duration of a
// test and restores it afterwards. server.Config is process-global, so these
// tests must not run with t.Parallel().
func withInfraRetryMax(t *testing.T, n int64) {
	t.Helper()
	old := server.Config.Pipeline.InfraRetryMaxAttempts
	server.Config.Pipeline.InfraRetryMaxAttempts = n
	t.Cleanup(func() { server.Config.Pipeline.InfraRetryMaxAttempts = old })
}

// pipelinesFor returns all pipelines for the repo, oldest first.
func pipelinesFor(t *testing.T, s store.Store, repo *model.Repo) []*model.Pipeline {
	t.Helper()
	list, err := s.GetPipelineList(repo, &model.ListOptions{Page: 1, PerPage: 100}, nil)
	require.NoError(t, err)
	sort.Slice(list, func(i, j int) bool { return list[i].Number < list[j].Number })
	return list
}

// waitForStablePipelineChain waits until the repo has at least one pipeline,
// every pipeline is terminal, and the count holds steady across a grace window
// (so any pending auto-retry has had time to materialize). It returns the
// chain oldest-first.
func waitForStablePipelineChain(t *testing.T, s store.Store, repo *model.Repo) []*model.Pipeline {
	t.Helper()

	deadline := time.Now().Add(60 * time.Second)
	var stableSince time.Time
	for time.Now().Before(deadline) {
		list := pipelinesFor(t, s, repo)
		allTerminal := len(list) > 0
		for _, p := range list {
			switch p.Status {
			case model.StatusSuccess, model.StatusFailure, model.StatusKilled,
				model.StatusError, model.StatusDeclined, model.StatusBlocked:
			default:
				allTerminal = false
			}
		}
		if allTerminal {
			if stableSince.IsZero() {
				stableSince = time.Now()
			}
			// Hold steady for a grace window: long enough that a queued
			// retry would have been created and started.
			if time.Since(stableSince) > 1500*time.Millisecond {
				return list
			}
		} else {
			stableSince = time.Time{}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("pipeline chain never stabilized for repo %s", repo.FullName)
	return nil
}

func triggerPush(t *testing.T, env *setup.ServerEnv) *model.Pipeline {
	t.Helper()
	draft := &model.Pipeline{
		Event:  model.EventPush,
		Branch: "main",
		Commit: "deadbeef",
		Ref:    "refs/heads/main",
		Author: env.Fixtures.Owner.Login,
		Sender: env.Fixtures.Owner.Login,
	}
	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, draft)
	require.NoError(t, err, "create pipeline")
	require.NotNil(t, created)
	return created
}

// TestInfraFailureTriggersAutoRetry proves the end-to-end path: a step the
// backend flags as an infrastructure failure flows agent -> server -> an
// automatic restart, bounded by the budget. With a budget of 1 we expect
// exactly one retry: the original pipeline plus one descendant.
func TestInfraFailureTriggersAutoRetry(t *testing.T) {
	withInfraRetryMax(t, 1)

	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker.yaml", Data: infraFailYAML},
	})
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	original := triggerPush(t, env)
	setup.WaitForPipelineStatus(t, env.Store, original.ID, model.StatusFailure, 20*time.Second)

	chain := waitForStablePipelineChain(t, env.Store, env.Fixtures.Repo)
	require.Len(t, chain, 2, "expected the original pipeline plus exactly one infra retry")

	orig, retry := chain[0], chain[1]
	assert.Equal(t, model.StatusFailure, orig.Status)
	assert.Equal(t, int64(0), orig.InfraRetryCount, "original pipeline has no prior infra retries")

	assert.Equal(t, model.StatusFailure, retry.Status, "the retry re-runs the same infra-failing config")
	assert.Equal(t, orig.Number, retry.Parent, "retry should be parented to the original")
	assert.Equal(t, int64(1), retry.InfraRetryCount, "retry records one infra attempt")
	assert.Greater(t, retry.RerunCount, int64(0), "restart bumps RerunCount too")
}

// TestGenuineFailureNoAutoRetry proves a real (non-infra) failure is never
// auto-restarted, even with budget available.
func TestGenuineFailureNoAutoRetry(t *testing.T) {
	withInfraRetryMax(t, 2)

	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker.yaml", Data: genuineFailYAML},
	})
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	original := triggerPush(t, env)
	setup.WaitForPipelineStatus(t, env.Store, original.ID, model.StatusFailure, 20*time.Second)

	chain := waitForStablePipelineChain(t, env.Store, env.Fixtures.Repo)
	require.Len(t, chain, 1, "a genuine failure must not be auto-retried")
	assert.Equal(t, int64(0), chain[0].InfraRetryCount)
}

// TestInfraRetryRespectsMaxAttempts proves the budget is a hard ceiling: a
// perpetually infra-failing pipeline with budget 2 yields exactly the original
// plus two retries, then stops.
func TestInfraRetryRespectsMaxAttempts(t *testing.T) {
	withInfraRetryMax(t, 2)

	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker.yaml", Data: infraFailYAML},
	})
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	original := triggerPush(t, env)
	setup.WaitForPipelineStatus(t, env.Store, original.ID, model.StatusFailure, 20*time.Second)

	chain := waitForStablePipelineChain(t, env.Store, env.Fixtures.Repo)
	require.Len(t, chain, 3, "original + 2 retries (budget = 2)")
	assert.Equal(t, int64(0), chain[0].InfraRetryCount)
	assert.Equal(t, int64(1), chain[1].InfraRetryCount)
	assert.Equal(t, int64(2), chain[2].InfraRetryCount, "last retry hits the ceiling and is not restarted again")
}
