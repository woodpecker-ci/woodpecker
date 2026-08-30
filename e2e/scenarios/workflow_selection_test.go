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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/e2e/setup"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
)

var (
	selectionAYAML = []byte(`
steps:
  - name: a
    image: dummy
    commands:
      - echo a
`)

	selectionBYAML = []byte(`
steps:
  - name: b
    image: dummy
    commands:
      - echo b
`)

	selectionCYAML = []byte(`
steps:
  - name: c
    image: dummy
    commands:
      - echo c
`)

	// Requires workflow "a", so selecting it alone must be rejected rather than
	// silently producing an empty pipeline.
	selectionDependentYAML = []byte(`
depends_on:
  - a

steps:
  - name: dependent
    image: dummy
    commands:
      - echo dependent
`)
)

func selectionFiles() []*forge_types.FileMeta {
	return []*forge_types.FileMeta{
		{Name: ".woodpecker/a.yaml", Data: selectionAYAML},
		{Name: ".woodpecker/b.yaml", Data: selectionBYAML},
		{Name: ".woodpecker/c.yaml", Data: selectionCYAML},
	}
}

func workflowNames(t *testing.T, env *setup.ServerEnv, pl *model.Pipeline) []string {
	t.Helper()
	workflows, err := env.Store.WorkflowGetTree(pl)
	require.NoError(t, err)

	names := make([]string, 0, len(workflows))
	for _, workflow := range workflows {
		names = append(names, workflow.Name)
	}
	sort.Strings(names)
	return names
}

// TestManualPipelineRunsAllWorkflowsByDefault pins the pre-existing behavior:
// a trigger that selects nothing still runs every workflow.
func TestManualPipelineRunsAllWorkflowsByDefault(t *testing.T) {
	env := setup.StartServer(t.Context(), t, selectionFiles())
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, env.DummyPipeline(model.EventManual))
	require.NoError(t, err, "create pipeline without selection")

	finished := setup.WaitForPipeline(t, env.Store, created.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status)
	assert.Equal(t, []string{"a", "b", "c"}, workflowNames(t, env, finished))
}

// TestManualPipelineSelectsWorkflows is the feature itself: only the selected
// workflows run, and only their configs are linked to the pipeline.
func TestManualPipelineSelectsWorkflows(t *testing.T) {
	env := setup.StartServer(t.Context(), t, selectionFiles())
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	pl := env.DummyPipeline(model.EventManual)
	pl.SelectedWorkflows = []string{"b", "c"}

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, pl)
	require.NoError(t, err, "create pipeline with selection")

	finished := setup.WaitForPipeline(t, env.Store, created.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status)
	assert.Equal(t, []string{"b", "c"}, workflowNames(t, env, finished))

	// Only the selected configs are persisted, so the config tab and any
	// restart see the same subset.
	configs, err := env.Store.ConfigsForPipeline(finished.ID)
	require.NoError(t, err)
	assert.Len(t, configs, 2, "only the selected configs should be linked to the pipeline")
}

// TestSelectionByForgePath checks the full config path is accepted as well as
// the sanitized workflow name.
func TestSelectionByForgePath(t *testing.T) {
	env := setup.StartServer(t.Context(), t, selectionFiles())
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	pl := env.DummyPipeline(model.EventManual)
	pl.SelectedWorkflows = []string{".woodpecker/a.yaml"}

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, pl)
	require.NoError(t, err)

	finished := setup.WaitForPipeline(t, env.Store, created.ID)
	assert.Equal(t, []string{"a"}, workflowNames(t, env, finished))
}

// TestRestartReplaysSelection is the reason the filtering happens before the
// configs are persisted: a restart must not silently widen back to every
// workflow.
func TestRestartReplaysSelection(t *testing.T) {
	env := setup.StartServer(t.Context(), t, selectionFiles())
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	pl := env.DummyPipeline(model.EventManual)
	pl.SelectedWorkflows = []string{"b"}

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, pl)
	require.NoError(t, err)
	original := setup.WaitForPipeline(t, env.Store, created.ID)
	require.Equal(t, model.StatusSuccess, original.Status)

	restarted, err := pipeline.Restart(t.Context(), env.Store, original, env.Fixtures.Owner, env.Fixtures.Repo, nil)
	require.NoError(t, err, "restart selected pipeline")

	finished := setup.WaitForPipeline(t, env.Store, restarted.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status)
	assert.Equal(t, []string{"b"}, workflowNames(t, env, finished), "restart must replay the original selection")
}

// TestSelectionRejectsUnknownWorkflow: a typo must be reported, not silently
// swallowed into an empty run.
func TestSelectionRejectsUnknownWorkflow(t *testing.T) {
	env := setup.StartServer(t.Context(), t, selectionFiles())

	pl := env.DummyPipeline(model.EventManual)
	pl.SelectedWorkflows = []string{"nope"}

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, pl)

	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown workflow(s) "nope"`)
	assert.Contains(t, err.Error(), `"a", "b", "c"`, "the error should list what is available")
	assert.Nil(t, created)
}

// TestSelectionRejectsMissingDependency covers the trap: without this check the
// builder drops the dependent workflow and the run collapses to nothing.
func TestSelectionRejectsMissingDependency(t *testing.T) {
	files := append(selectionFiles(), &forge_types.FileMeta{
		Name: ".woodpecker/dependent.yaml", Data: selectionDependentYAML,
	})
	env := setup.StartServer(t.Context(), t, files)

	pl := env.DummyPipeline(model.EventManual)
	pl.SelectedWorkflows = []string{"dependent"}

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, pl)

	require.Error(t, err)
	assert.Contains(t, err.Error(), `depend on "a"`)
	assert.Nil(t, created)
}

// TestSelectionWithDependencyIncluded is the same setup, but selecting the
// dependency too, which must succeed.
func TestSelectionWithDependencyIncluded(t *testing.T) {
	files := append(selectionFiles(), &forge_types.FileMeta{
		Name: ".woodpecker/dependent.yaml", Data: selectionDependentYAML,
	})
	env := setup.StartServer(t.Context(), t, files)
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	pl := env.DummyPipeline(model.EventManual)
	pl.SelectedWorkflows = []string{"dependent", "a"}

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, pl)
	require.NoError(t, err)

	finished := setup.WaitForPipeline(t, env.Store, created.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status)
	assert.Equal(t, []string{"a", "dependent"}, workflowNames(t, env, finished))
}
