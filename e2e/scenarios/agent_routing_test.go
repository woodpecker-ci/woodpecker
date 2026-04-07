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
	server_pipeline "go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
)

// labelRoutingYAML is a single-workflow pipeline that requires the label
// gpu=true. Only the gpu-agent should pick it up; the plain agent must not.
var labelRoutingYAML = []byte(`
labels:
  gpu: "true"

steps:
  - name: gpu-step
    image: dummy
    commands:
      - echo running on gpu agent
`)

// TestAgentLabelRouting starts two agents — one plain, one with gpu=true —
// and asserts that the pipeline with labels: gpu: "true" is always picked up
// by the gpu agent and never by the plain agent.
func TestAgentLabelRouting(t *testing.T) {
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker.yaml", Data: labelRoutingYAML},
	})

	// Plain agent: wildcard repo label only — cannot satisfy gpu=true.
	plainAgent := setup.StartAgent(t.Context(), t, env.GRPCAddr,
		setup.WithHostname("plain-agent"),
	)

	// GPU agent: carries gpu=true — the only agent that can accept the task.
	gpuAgent := setup.StartAgent(t.Context(), t, env.GRPCAddr,
		setup.WithHostname("gpu-agent"),
		setup.WithCustomLabels(map[string]string{"gpu": "true"}),
	)

	setup.WaitForAgentRegistered(t, env.Store, plainAgent, gpuAgent)

	created, err := server_pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, &model.Pipeline{
		Event:  model.EventPush,
		Branch: "main",
		Commit: "deadbeef",
		Ref:    "refs/heads/main",
		Author: env.Fixtures.Owner.Login,
		Sender: env.Fixtures.Owner.Login,
	})
	require.NoError(t, err, "create pipeline")

	finished := setup.WaitForPipeline(t, env.Store, created.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status, "pipeline should succeed")

	// The single workflow (name="woodpecker" from SanitizePath(".woodpecker.yaml"))
	// must have been executed by the gpu agent, not the plain agent.
	setup.AssertWorkflowRanOnAgent(t, env.Store, finished, "woodpecker", gpuAgent)
}

// orgPipelineYAML is a plain single-step pipeline used for org-preference tests.
var orgPipelineYAML = []byte(`
steps:
  - name: build
    image: dummy
    commands:
      - echo building
`)

// TestOrgAgentPreferredOverGlobal starts a global agent and an org-scoped agent
// for the same org as the test repo. It asserts that the org agent is always
// preferred by the queue (score 10 vs 1) and picks up the pipeline.
func TestOrgAgentPreferredOverGlobal(t *testing.T) {
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker.yaml", Data: orgPipelineYAML},
	})

	// Global agent: matches org-id=* (score 1).
	globalAgent := setup.StartAgent(t.Context(), t, env.GRPCAddr,
		setup.WithHostname("global-agent"),
	)

	// Org agent: will be patched with the repo's OrgID (score 10).
	orgAgent := setup.StartAgent(t.Context(), t, env.GRPCAddr,
		setup.WithHostname("org-agent"),
		setup.WithOrgID(env.Fixtures.Repo.OrgID),
	)

	setup.WaitForAgentRegistered(t, env.Store, globalAgent, orgAgent)

	created, err := server_pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, &model.Pipeline{
		Event:  model.EventPush,
		Branch: "main",
		Commit: "deadbeef",
		Ref:    "refs/heads/main",
		Author: env.Fixtures.Owner.Login,
		Sender: env.Fixtures.Owner.Login,
	})
	require.NoError(t, err, "create pipeline")

	finished := setup.WaitForPipeline(t, env.Store, created.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status, "pipeline should succeed")

	// The workflow must have been picked up by the org-scoped agent, not the
	// global one — the queue scores exact org-id matches 10× higher.
	setup.AssertWorkflowRanOnAgent(t, env.Store, finished, "woodpecker", orgAgent)
}
