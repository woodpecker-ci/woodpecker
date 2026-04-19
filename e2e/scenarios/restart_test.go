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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/e2e/setup"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
)

// TestRestartPipeline verifies pipeline.Restart produces a distinct pipeline
// linked to the original via Parent, with its own fresh workflow rows, and
// that the original's workflows are untouched.
func TestRestartPipeline(t *testing.T) {
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker.yaml", Data: simpleSuccessYAML},
	})
	agent := setup.StartAgent(t.Context(), t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	// First run.
	original, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, &model.Pipeline{
		Event:  model.EventPush,
		Branch: "main",
		Commit: "deadbeef",
		Ref:    "refs/heads/main",
		Author: env.Fixtures.Owner.Login,
		Sender: env.Fixtures.Owner.Login,
	})
	require.NoError(t, err, "create original pipeline")
	originalFinished := setup.WaitForPipeline(t, env.Store, original.ID)
	require.Equal(t, model.StatusSuccess, originalFinished.Status, "original should succeed")

	originalWorkflows, err := env.Store.WorkflowGetTree(originalFinished)
	require.NoError(t, err)
	require.Len(t, originalWorkflows, 1, "original should have exactly one workflow")
	originalWorkflowID := originalWorkflows[0].ID

	// Restart it.
	restarted, err := pipeline.Restart(t.Context(), env.Store, originalFinished, env.Fixtures.Owner, env.Fixtures.Repo, nil)
	require.NoError(t, err, "restart pipeline")
	require.NotNil(t, restarted)

	// Parent/ID invariants.
	assert.NotEqual(t, originalFinished.ID, restarted.ID, "restart should have a new ID")
	assert.NotEqual(t, originalFinished.Number, restarted.Number, "restart should have a new number")
	assert.Equal(t, originalFinished.Number, restarted.Parent, "restart.Parent should point at original.Number")

	// The restart runs through the same start path — wait for it to finish.
	restartedFinished := setup.WaitForPipeline(t, env.Store, restarted.ID)
	assert.Equal(t, model.StatusSuccess, restartedFinished.Status, "restarted pipeline should succeed")

	// Restart should have its OWN workflows, not reuse the originals.
	restartedWorkflows, err := env.Store.WorkflowGetTree(restartedFinished)
	require.NoError(t, err)
	require.Len(t, restartedWorkflows, 1, "restart should produce its own workflow")
	assert.NotEqual(t, originalWorkflowID, restartedWorkflows[0].ID,
		"restart should insert a new workflow row, not reassign the original")
	assert.Equal(t, restartedFinished.ID, restartedWorkflows[0].PipelineID,
		"restarted workflow must be linked to the restarted pipeline")
	assert.Equal(t, model.StatusSuccess, restartedWorkflows[0].State)
	assert.Greater(t, restartedWorkflows[0].AgentID, int64(0))

	// Original's workflows must remain pointing at the original pipeline.
	originalAfter, err := env.Store.WorkflowGetTree(originalFinished)
	require.NoError(t, err)
	require.Len(t, originalAfter, 1)
	assert.Equal(t, originalWorkflowID, originalAfter[0].ID,
		"restart must not mutate the original's workflow row")
	assert.Equal(t, originalFinished.ID, originalAfter[0].PipelineID,
		"original's workflow must still be linked to the original pipeline")
}
