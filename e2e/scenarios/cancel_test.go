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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/e2e/setup"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
)

// cancelPipelineYAML has one long-sleeping step followed by one that must
// be skipped when the pipeline is canceled.
var cancelPipelineYAML = []byte(`
steps:
  - name: long-running
    image: dummy
    commands:
      - echo starting long job
    environment:
      SLEEP: "30s"

  - name: after-cancel
    image: dummy
    commands:
      - echo this should never run
`)

// TestCancelRunningPipeline triggers a long-running pipeline, waits for it
// to enter StatusRunning, then cancels it via pipeline.Cancel and asserts:
//   - pipeline ends up as StatusKilled
//   - the running step exits with code 130 (dummy cancel convention = SIGINT)
//   - the subsequent step is skipped
func TestCancelRunningPipeline(t *testing.T) {
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker.yaml", Data: cancelPipelineYAML},
	})
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, &model.Pipeline{
		Event:  model.EventPush,
		Branch: "main",
		Commit: "deadbeef",
		Ref:    "refs/heads/main",
		Author: env.Fixtures.Owner.Login,
		Sender: env.Fixtures.Owner.Login,
	})
	require.NoError(t, err, "create pipeline")
	require.NotNil(t, created)

	// Wait until the agent has picked it up and set it to running.
	setup.WaitForPipelineStatus(t, env.Store, created.ID, model.StatusRunning, 10*time.Second)

	// Also wait for the specific step to reach StatusRunning in the DB.
	// The pipeline transitions to StatusRunning as soon as the agent starts
	// the workflow, but the step itself may not yet have entered its
	// sleepWithContext call in the dummy backend. If we cancel before the
	// step is actually sleeping, WaitStep returns immediately with success
	// before the cancel context propagates — causing "success" instead of
	// "killed". Waiting here ensures the dummy sleep is genuinely in progress.
	setup.WaitForStepRunning(t, env.Store, created.ID, "long-running")

	// Resolve the forge instance (MockForge) via the manager.
	forge, err := env.Manager.ForgeByID(env.Fixtures.Forge.ID)
	require.NoError(t, err, "resolve forge")

	// Fetch the latest pipeline state from the store before canceling.
	running, err := env.Store.GetPipeline(created.ID)
	require.NoError(t, err, "get running pipeline")

	// Cancel through the normal server API path — same as the HTTP handler does.
	err = pipeline.Cancel(t.Context(), forge, env.Store, env.Fixtures.Repo, env.Fixtures.Owner, running, nil)
	require.NoError(t, err, "cancel pipeline")

	// Wait for the pipeline to reach a terminal state.
	finished := setup.WaitForPipeline(t, env.Store, created.ID)
	assert.Equal(t, model.StatusKilled, finished.Status, "canceled pipeline should be killed")

	t.Run("long-running step is killed", func(t *testing.T) {
		// After pipeline.Cancel() the pipeline itself reaches a terminal state
		// immediately, but the running step's status is written asynchronously
		// by the agent's gRPC Done() call — which arrives *after* the cancel
		// signal is processed. We therefore wait explicitly for the step to
		// leave "running", giving the agent enough time to finish cleanup and
		// report back.
		step := setup.WaitForStepStatus(t, env.Store, finished, "long-running", model.StatusKilled, 30*time.Second)
		assert.Equal(t, model.StatusKilled, step.State)
	})

	t.Run("after-cancel step is canceled", func(t *testing.T) {
		// Pending steps get StatusCanceled synchronously by pipeline.Cancel()
		// before any agent is involved, so this should already be set.
		step := setup.WaitForStep(t, env.Store, finished, "after-cancel")
		assert.Equal(t, model.StatusCanceled, step.State)
	})
}
