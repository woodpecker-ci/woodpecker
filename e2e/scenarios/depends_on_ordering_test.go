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

// Models woodpecker-ci/woodpecker#3858: "depends_on seems to be broken on
// workflow level". A downstream workflow that depends_on an upstream one
// started running before the upstream had finished — in the reporter's case
// the upstream built a docker image (auth-build:${CI_COMMIT_SHA}) that the
// downstream then tried to use, getting "pull access denied ... repository
// does not exist" because the build had not completed yet.
//
// "Build Auth" sleeps for a measurable duration. "Auth server tests" depends_on it.
// Correct behaviour: "Auth server tests" must not START until "Build Auth" has
// FINISHED. We prove this directly from the recorded step timestamps rather
// than just final status, because a broken ordering still ends "success" —
// the steps just overlap in time.

// Workflow and step names match the issue report verbatim so the test
// reads as a direct reproduction of the failure scenario.
var buildAuthYAML = []byte(`
skip_clone: true

steps:
  - name: Build Auth
    image: dummy
    environment:
      SLEEP: '2s'
    commands:
      - echo building auth-build image
`)

var authServerTestsYAML = []byte(`
skip_clone: true

depends_on:
  - Build Auth

steps:
  - name: Auth server tests
    image: dummy
    commands:
      - echo running tests against built image
`)

// TestWorkflowDependsOnOrdering asserts that a workflow with a workflow-level
// depends_on does not begin executing until its dependency has completed.
func TestWorkflowDependsOnOrdering(t *testing.T) {
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		// Filenames with spaces: the workflow name is derived from the
		// filename (minus extension), so "Build Auth.yaml" → workflow "Build Auth",
		// matching exactly what the issue reporter used.
		{Name: ".woodpecker/Build Auth.yaml", Data: buildAuthYAML},
		{Name: ".woodpecker/Auth server tests.yaml", Data: authServerTestsYAML},
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

	finished := setup.WaitForPipeline(t, env.Store, created.ID)
	require.Equal(t, model.StatusSuccess, finished.Status, "pipeline should succeed")

	// Both workflows should have succeeded.
	workflows, err := env.Store.WorkflowGetTree(finished)
	require.NoError(t, err, "list workflows")
	byWorkflow := make(map[string]*model.Workflow, len(workflows))
	for _, w := range workflows {
		byWorkflow[w.Name] = w
	}
	require.Contains(t, byWorkflow, "Build Auth", "Build Auth workflow present")
	require.Contains(t, byWorkflow, "Auth server tests", "Auth server tests workflow present")
	assert.Equal(t, model.StatusSuccess, byWorkflow["Build Auth"].State)
	assert.Equal(t, model.StatusSuccess, byWorkflow["Auth server tests"].State)

	// The core assertion: the dependent step must start only AFTER the
	// dependency step finished. Compare recorded timestamps.
	buildStep := setup.WaitForStep(t, env.Store, finished, "Build Auth")
	testStep := setup.WaitForStep(t, env.Store, finished, "Auth server tests")

	require.NotZero(t, buildStep.Finished, "Build Auth must record a finish time")
	require.NotZero(t, testStep.Started, "Auth server tests must record a start time")

	// This is the line that fails if #3858 regresses: a broken workflow-level
	// depends_on lets Auth server tests start while Build Auth is still
	// sleeping, so testStep.Started < buildStep.Finished.
	assert.GreaterOrEqualf(t, testStep.Started, buildStep.Finished,
		"Auth server tests started at %d but Build Auth only finished at %d — dependent workflow ran before its dependency completed (issue #3858)",
		testStep.Started, buildStep.Finished)
}
