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

// TestGatedPipeline verifies the approval gate on a repo that requires
// approval for every event. The pipeline is created in StatusBlocked, and
// once pipeline.Approve releases it the pipeline runs to completion on an
// agent like any normal pipeline, steps included.
func TestGatedPipeline(t *testing.T) {
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker.yaml", Data: simpleSuccessYAML},
	})
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	// Require approval for every event, gating every pipeline regardless of author.
	env.Fixtures.Repo.RequireApproval = model.RequireApprovalAllEvents
	require.NoError(t, env.Store.UpdateRepo(env.Fixtures.Repo), "enable repo approval")

	// Pipeline must come back blocked, not running.
	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, env.DummyPipeline(model.EventPush))
	require.NoError(t, err, "create gated pipeline")
	require.NotNil(t, created)
	require.Equal(t, model.StatusBlocked, created.Status, "untrusted author pipeline must be blocked")

	// Approve as the repo owner, releasing the gate.
	approved, err := pipeline.Approve(t.Context(), env.Store, created, env.Fixtures.Owner, env.Fixtures.Repo)
	require.NoError(t, err, "approve gated pipeline")
	require.NotNil(t, approved)
	assert.Equal(t, env.Fixtures.Owner.Login, approved.Reviewer, "reviewer should be the approver")
	assert.NotZero(t, approved.Reviewed, "reviewed timestamp should be set")

	// Wait for the agent to actually pick it up and run it to a terminal state.
	finished := setup.WaitForPipeline(t, env.Store, approved.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status, "approved gated pipeline should succeed")

	// Workflow outcome: one workflow, succeeded, assigned to an agent.
	workflows, err := env.Store.WorkflowGetTree(finished)
	require.NoError(t, err, "get workflow tree")
	require.Len(t, workflows, 1, "approved pipeline should produce exactly one workflow")
	assert.Equal(t, model.StatusSuccess, workflows[0].State, "workflow should succeed")
	assert.Greater(t, workflows[0].AgentID, int64(0), "workflow should record the agent that ran it")

	// Step outcome: every step from simpleSuccessYAML ran and exited cleanly.
	steps, err := env.Store.StepList(finished.ID)
	require.NoError(t, err, "list steps")
	require.ElementsMatch(t, []string{"clone", "step-one", "step-two"}, modelStepsToName(steps),
		"approved pipeline should run exactly the steps from the YAML")
	for _, step := range steps {
		assert.Equalf(t, model.StatusSuccess, step.State, "step %q status", step.Name)
		assert.Equalf(t, 0, step.ExitCode, "step %q exit code", step.Name)
	}
}
