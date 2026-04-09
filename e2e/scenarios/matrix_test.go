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

// matrixPipelineYAML defines a 2×2 matrix (GO_VERSION × OS), yielding 4
// workflows. Each step echoes its matrix variables so we can confirm the
// dummy backend receives the interpolated values via the step environment.
var matrixPipelineYAML = []byte(`
matrix:
  GO_VERSION:
    - "1.24"
    - "1.26"
  OS:
    - linux
    - windows

steps:
  - name: build
    image: dummy
    commands:
      - echo "go=${GO_VERSION} os=${OS}"
`)

// matrixIncludePipelineYAML uses the matrix.include form to specify exact
// combinations, verifying the alternative matrix syntax is also handled.
var matrixIncludePipelineYAML = []byte(`
matrix:
  include:
    - GO_VERSION: "1.24"
      OS: linux
    - GO_VERSION: "1.26"
      OS: linux
    - GO_VERSION: "1.26"
      OS: windows

steps:
  - name: build
    image: dummy
    commands:
      - echo "go=${GO_VERSION} os=${OS}"
`)

// TestMatrixPipeline verifies that a matrix YAML expands into the correct
// number of workflows, that every workflow succeeds, and that each workflow's
// Environ map carries the right variable combination.
func TestMatrixPipeline(t *testing.T) {
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker.yaml", Data: matrixPipelineYAML},
	})
	agent := setup.StartAgent(t.Context(), t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, &model.Pipeline{
		Event:  model.EventPush,
		Branch: "main",
		Commit: "deadbeef",
		Ref:    "refs/heads/main",
		Author: env.Fixtures.Owner.Login,
		Sender: env.Fixtures.Owner.Login,
	})
	require.NoError(t, err, "create matrix pipeline")
	require.NotNil(t, created)

	finished := setup.WaitForPipeline(t, env.Store, created.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status, "matrix pipeline should succeed")

	workflows, err := env.Store.WorkflowGetTree(finished)
	require.NoError(t, err, "get workflow tree")

	// 2 GO_VERSION values × 2 OS values = 4 workflows
	const wantWorkflows = 4
	assert.Len(t, workflows, wantWorkflows,
		"matrix should expand to %d workflows", wantWorkflows)

	// Build the set of expected (GO_VERSION, OS) pairs and verify each
	// workflow accounts for exactly one, with no duplicates.
	type combo struct{ goVersion, os string }
	expected := map[combo]bool{
		{"1.24", "linux"}:   true,
		{"1.24", "windows"}: true,
		{"1.26", "linux"}:   true,
		{"1.26", "windows"}: true,
	}

	seen := make(map[combo]bool, len(workflows))
	for _, wf := range workflows {
		assert.Equal(t, model.StatusSuccess, wf.State,
			"workflow axis %d should succeed", wf.AxisID)
		assert.NotZero(t, wf.AxisID,
			"matrix workflows must have a non-zero AxisID")

		goVer := wf.Environ["GO_VERSION"]
		os := wf.Environ["OS"]
		c := combo{goVer, os}

		assert.True(t, expected[c],
			"unexpected matrix combination GO_VERSION=%q OS=%q", goVer, os)
		assert.False(t, seen[c],
			"duplicate matrix combination GO_VERSION=%q OS=%q", goVer, os)
		seen[c] = true
	}

	// Every expected combination must have been present.
	for c := range expected {
		assert.True(t, seen[c],
			"missing matrix combination GO_VERSION=%q OS=%q", c.goVersion, c.os)
	}
}

// TestMatrixIncludePipeline verifies the matrix.include syntax produces the
// exact explicit combinations listed (3 workflows, not a full cross product).
func TestMatrixIncludePipeline(t *testing.T) {
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker.yaml", Data: matrixIncludePipelineYAML},
	})
	agent := setup.StartAgent(t.Context(), t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, &model.Pipeline{
		Event:  model.EventPush,
		Branch: "main",
		Commit: "deadbeef",
		Ref:    "refs/heads/main",
		Author: env.Fixtures.Owner.Login,
		Sender: env.Fixtures.Owner.Login,
	})
	require.NoError(t, err, "create matrix include pipeline")
	require.NotNil(t, created)

	finished := setup.WaitForPipeline(t, env.Store, created.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status, "matrix include pipeline should succeed")

	workflows, err := env.Store.WorkflowGetTree(finished)
	require.NoError(t, err, "get workflow tree")

	// matrix.include has 3 explicit entries — no cross product.
	const wantWorkflows = 3
	assert.Len(t, workflows, wantWorkflows,
		"matrix include should produce exactly %d workflows", wantWorkflows)

	type combo struct{ goVersion, os string }
	expected := map[combo]bool{
		{"1.24", "linux"}:   true,
		{"1.26", "linux"}:   true,
		{"1.26", "windows"}: true,
	}

	seen := make(map[combo]bool, len(workflows))
	for _, wf := range workflows {
		assert.Equal(t, model.StatusSuccess, wf.State,
			"workflow (axis %d) should succeed", wf.AxisID)

		c := combo{wf.Environ["GO_VERSION"], wf.Environ["OS"]}
		assert.True(t, expected[c],
			"unexpected combination GO_VERSION=%q OS=%q", c.goVersion, c.os)
		assert.False(t, seen[c],
			"duplicate combination GO_VERSION=%q OS=%q", c.goVersion, c.os)
		seen[c] = true
	}

	for c := range expected {
		assert.True(t, seen[c],
			"missing combination GO_VERSION=%q OS=%q", c.goVersion, c.os)
	}
}

// TestMatrixSingleAxis verifies a single-axis matrix (TAG: [1.7, 1.8, latest])
// — the simplest possible matrix — to ensure no edge cases in the axis
// calculation code.
func TestMatrixSingleAxis(t *testing.T) {
	yaml := []byte(`
matrix:
  TAG:
    - "1.7"
    - "1.8"
    - latest

steps:
  - name: build
    image: dummy
    commands:
      - echo "tag=${TAG}"
`)

	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker.yaml", Data: yaml},
	})
	agent := setup.StartAgent(t.Context(), t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, &model.Pipeline{
		Event:  model.EventPush,
		Branch: "main",
		Commit: "deadbeef",
		Ref:    "refs/heads/main",
		Author: env.Fixtures.Owner.Login,
		Sender: env.Fixtures.Owner.Login,
	})
	require.NoError(t, err, "create single-axis matrix pipeline")
	require.NotNil(t, created)

	finished := setup.WaitForPipeline(t, env.Store, created.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status, "single-axis matrix pipeline should succeed")

	workflows, err := env.Store.WorkflowGetTree(finished)
	require.NoError(t, err, "get workflow tree")

	assert.Len(t, workflows, 3, "single-axis matrix [1.7, 1.8, latest] should produce 3 workflows")

	wantTags := map[string]bool{"1.7": true, "1.8": true, "latest": true}
	seenTags := make(map[string]bool, 3)
	for _, wf := range workflows {
		assert.Equal(t, model.StatusSuccess, wf.State,
			"workflow for TAG=%q should succeed", wf.Environ["TAG"])
		tag := wf.Environ["TAG"]
		assert.True(t, wantTags[tag], "unexpected TAG value %q", tag)
		assert.False(t, seenTags[tag], "duplicate TAG value %q", tag)
		seenTags[tag] = true
	}
}

// TestMatrixNoMatrix is a regression guard: a YAML without a matrix section
// must produce exactly one workflow (the existing behavior must not break).
func TestMatrixNoMatrix(t *testing.T) {
	yaml := []byte(`
steps:
  - name: build
    image: dummy
    commands:
      - echo "no matrix"
`)

	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker.yaml", Data: yaml},
	})
	agent := setup.StartAgent(t.Context(), t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, &model.Pipeline{
		Event:  model.EventPush,
		Branch: "main",
		Commit: "deadbeef",
		Ref:    "refs/heads/main",
		Author: env.Fixtures.Owner.Login,
		Sender: env.Fixtures.Owner.Login,
	})
	require.NoError(t, err, "create non-matrix pipeline")
	require.NotNil(t, created)

	finished := setup.WaitForPipeline(t, env.Store, created.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status)

	workflows, err := env.Store.WorkflowGetTree(finished)
	require.NoError(t, err, "get workflow tree")

	assert.Len(t, workflows, 1, "non-matrix pipeline should produce exactly 1 workflow")
	assert.Zero(t, workflows[0].AxisID,
		"non-matrix workflow should have AxisID=0")
	assert.Empty(t, workflows[0].Environ,
		"non-matrix workflow should have no Environ variables")
}
