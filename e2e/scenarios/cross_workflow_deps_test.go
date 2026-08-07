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

// The whole point of cross-workflow STEP dependencies (vs workflow-level
// depends_on) is overlap: the expensive early steps of the consumer workflow
// run in parallel with the producer workflow, and only the gated steps
// synchronize. This test proves both halves with recorded step timestamps:
//
//	auxiliaries:  resolve-pins (1s) ──> publish (1s)
//	base:         build ──> pin-deps ──> wait-deps
//	                        ▲ waits for resolve-pins   ▲ waits for auxiliaries
//
//  1. overlap:   base/build STARTS before auxiliaries FINISHES
//  2. gating:    pin-deps starts only after resolve-pins finished
//  3. gating:    wait-deps starts only after ALL of auxiliaries finished

var crossAuxiliariesYAML = []byte(`
skip_clone: true

steps:
  - name: resolve-pins
    image: dummy
    environment:
      SLEEP: '1s'
    commands:
      - echo resolving pins

  - name: publish
    image: dummy
    depends_on:
      - resolve-pins
    environment:
      SLEEP: '1s'
    commands:
      - echo publishing deps
`)

var crossBaseYAML = []byte(`
skip_clone: true

steps:
  - name: build
    image: dummy
    commands:
      - echo building

  - name: pin-deps
    image: dummy
    depends_on:
      - build
      - workflow: auxiliaries
        step: resolve-pins
    commands:
      - echo pinning deps

  - name: wait-deps
    image: dummy
    depends_on:
      - pin-deps
      - workflow: auxiliaries
    commands:
      - echo deps published
`)

// TestCrossWorkflowStepDepsOrdering asserts that a step-level dependency on
// another workflow gates only the declaring step while the rest of its
// workflow overlaps the dependency.
func TestCrossWorkflowStepDepsOrdering(t *testing.T) {
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker/auxiliaries.yaml", Data: crossAuxiliariesYAML},
		{Name: ".woodpecker/base.yaml", Data: crossBaseYAML},
	})
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, env.DummyPipeline(model.EventPush))
	require.NoError(t, err, "create pipeline")
	require.NotNil(t, created)

	finished := setup.WaitForPipeline(t, env.Store, created.ID)
	require.Equal(t, model.StatusSuccess, finished.Status, "pipeline should succeed")

	workflows, err := env.Store.WorkflowGetTree(finished)
	require.NoError(t, err, "list workflows")
	byWorkflow := make(map[string]*model.Workflow, len(workflows))
	for _, w := range workflows {
		byWorkflow[w.Name] = w
	}
	require.Contains(t, byWorkflow, "auxiliaries")
	require.Contains(t, byWorkflow, "base")
	assert.Equal(t, model.StatusSuccess, byWorkflow["auxiliaries"].State)
	assert.Equal(t, model.StatusSuccess, byWorkflow["base"].State)

	resolvePins := setup.WaitForStep(t, env.Store, finished, "resolve-pins")
	publish := setup.WaitForStep(t, env.Store, finished, "publish")
	build := setup.WaitForStep(t, env.Store, finished, "build")
	pinDeps := setup.WaitForStep(t, env.Store, finished, "pin-deps")
	waitDeps := setup.WaitForStep(t, env.Store, finished, "wait-deps")

	for name, step := range map[string]*model.Step{
		"resolve-pins": resolvePins, "publish": publish,
		"build": build, "pin-deps": pinDeps, "wait-deps": waitDeps,
	} {
		require.NotZerof(t, step.Started, "%s must record a start time", name)
		require.NotZerof(t, step.Finished, "%s must record a finish time", name)
	}

	// 1. Overlap: base is NOT serialized behind auxiliaries. build must start
	// while auxiliaries is still sleeping (resolve-pins alone takes 1s).
	assert.Lessf(t, build.Started, byWorkflow["auxiliaries"].Finished,
		"build started at %d only after auxiliaries finished at %d — the consumer workflow was serialized instead of overlapping",
		build.Started, byWorkflow["auxiliaries"].Finished)

	// 2. Step-granular gate: pin-deps waits for resolve-pins, but NOT for
	// publish (which is still sleeping when resolve-pins is done).
	assert.GreaterOrEqualf(t, pinDeps.Started, resolvePins.Finished,
		"pin-deps started at %d before its dependency resolve-pins finished at %d",
		pinDeps.Started, resolvePins.Finished)
	assert.Lessf(t, pinDeps.Started, publish.Finished,
		"pin-deps started at %d only after publish finished at %d — the step-granular dependency waited for the whole workflow",
		pinDeps.Started, publish.Finished)

	// 3. Whole-workflow gate: wait-deps waits for all of auxiliaries.
	assert.GreaterOrEqualf(t, waitDeps.Started, publish.Finished,
		"wait-deps started at %d before auxiliaries' last step finished at %d",
		waitDeps.Started, publish.Finished)
}
