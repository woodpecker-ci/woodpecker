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
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils/wasm/wasmtest"
)

// startServerWithConfigExtension starts a server that runs the test
// configuration extension as wasm module. The extension turns ".steps" files
// into workflows, see its source for the format.
func startServerWithConfigExtension(t *testing.T, files ...*forge_types.FileMeta) *setup.ServerEnv {
	t.Helper()

	return setup.StartServer(t.Context(), t, files,
		setup.WithConfigExtensionWasm(wasmtest.File(t, wasmtest.ConfigGuest)),
		setup.WithPipelineConfigExtensions(".yaml", ".steps"),
	)
}

// TestWasmConfigExtensionConverts verifies the whole way of a config that only
// the extension understands: forge -> wasm module -> compiler -> agent. The
// step names prove that the module got the pipeline and repo of the request.
// A plain YAML workflow next to it must survive the round trip unchanged.
func TestWasmConfigExtensionConverts(t *testing.T) {
	env := startServerWithConfigExtension(t,
		&forge_types.FileMeta{Name: ".woodpecker/build.steps", Data: []byte("compile\non-{event}-in-{repo}\n")},
		&forge_types.FileMeta{Name: ".woodpecker/plain.yaml", Data: simpleSuccessYAML},
	)
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, env.DummyPipeline(model.EventPush))
	require.NoError(t, err, "create pipeline")
	require.NotNil(t, created)

	finished := setup.WaitForPipeline(t, env.Store, created.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status, "pipeline final status")

	workflows, err := env.Store.WorkflowGetTree(finished)
	require.NoError(t, err, "list workflows")

	steps := make(map[string][]string, len(workflows))
	for _, workflow := range workflows {
		for _, step := range workflow.Children {
			assert.Equalf(t, model.StatusSuccess, step.State, "step %q of workflow %q", step.Name, workflow.Name)
			steps[workflow.Name] = append(steps[workflow.Name], step.Name)
		}
	}
	assert.Equal(t, map[string][]string{
		"build": {"clone", "compile", "on-push-in-" + env.Fixtures.Repo.Name},
		"plain": {"clone", "step-one", "step-two"},
	}, steps)
}

// TestWasmConfigExtensionKeepsConfig verifies that a repo without anything to
// convert is not affected by the extension.
func TestWasmConfigExtensionKeepsConfig(t *testing.T) {
	env := startServerWithConfigExtension(t,
		&forge_types.FileMeta{Name: ".woodpecker.yaml", Data: simpleSuccessYAML},
	)
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, env.DummyPipeline(model.EventPush))
	require.NoError(t, err, "create pipeline")
	require.NotNil(t, created)

	finished := setup.WaitForPipeline(t, env.Store, created.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status, "pipeline final status")
}

// TestWasmConfigExtensionFailsClosed verifies that a failing extension stops
// the pipeline before anything is scheduled and that its message reaches the
// user. The plain workflow next to the broken one must not run on its own.
func TestWasmConfigExtensionFailsClosed(t *testing.T) {
	env := startServerWithConfigExtension(t,
		&forge_types.FileMeta{Name: ".woodpecker/build.steps", Data: []byte("compile\n!fail line 2: this is not allowed\n")},
		&forge_types.FileMeta{Name: ".woodpecker/plain.yaml", Data: simpleSuccessYAML},
	)

	_, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, env.DummyPipeline(model.EventPush))
	require.NoError(t, err, "a broken config is reported on the pipeline, not as error")

	failed, err := env.Store.GetPipelineNumber(env.Fixtures.Repo, 1)
	require.NoError(t, err, "the pipeline must exist to show the error")
	assert.Equal(t, model.StatusError, failed.Status, "pipeline final status")
	require.Len(t, failed.Errors, 1)
	assert.Contains(t, failed.Errors[0].Message, "line 2: this is not allowed")

	workflows, err := env.Store.WorkflowGetTree(failed)
	require.NoError(t, err, "list workflows")
	assert.Empty(t, workflows, "nothing must be scheduled")
}
