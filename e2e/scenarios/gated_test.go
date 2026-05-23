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

func TestApprovedGatedPipelineRuns(t *testing.T) {
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker.yaml", Data: simpleSuccessYAML},
	})
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	env.Fixtures.Repo.RequireApproval = model.RequireApprovalAllEvents
	require.NoError(t, env.Store.UpdateRepo(env.Fixtures.Repo))

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, &model.Pipeline{
		Event:  model.EventPush,
		Branch: "main",
		Commit: "deadbeef",
		Ref:    "refs/heads/main",
		Author: "external-contributor",
		Sender: "external-contributor",
	})
	require.NoError(t, err, "create gated pipeline")
	require.NotNil(t, created)
	require.Equal(t, model.StatusBlocked, created.Status)

	blockedWorkflows, err := env.Store.WorkflowGetTree(created)
	require.NoError(t, err)
	require.Len(t, blockedWorkflows, 1)
	assert.Equal(t, model.StatusBlocked, blockedWorkflows[0].State)
	require.NotEmpty(t, blockedWorkflows[0].Children)
	for _, step := range blockedWorkflows[0].Children {
		assert.Equal(t, model.StatusBlocked, step.State)
	}

	approved, err := pipeline.Approve(t.Context(), env.Store, created, env.Fixtures.Owner, env.Fixtures.Repo)
	require.NoError(t, err, "approve gated pipeline")
	require.NotNil(t, approved)
	assert.Equal(t, env.Fixtures.Owner.Login, approved.Reviewer)
	assert.NotZero(t, approved.Reviewed)

	finished := setup.WaitForPipeline(t, env.Store, approved.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status, "approved gated pipeline should run")

	workflows, err := env.Store.WorkflowGetTree(finished)
	require.NoError(t, err)
	require.Len(t, workflows, 1)
	assert.Equal(t, model.StatusSuccess, workflows[0].State)
	assert.Greater(t, workflows[0].AgentID, int64(0), "approved workflow should be assigned to an agent")
	require.Len(t, workflows[0].Children, len(blockedWorkflows[0].Children))
	for _, step := range workflows[0].Children {
		assert.Equal(t, model.StatusSuccess, step.State)
	}
}
