// Copyright 2024 Woodpecker Authors
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

package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
)

// item builds a minimal *builder.Item with enough populated for
// injectLocalWorkspaceMounts to operate on. The only fields the
// function touches are Config.Volume and the Steps inside Stages.
func testItem(name, volume string, steps ...string) *builder.Item {
	stage := &backend_types.Stage{}
	for _, sn := range steps {
		stage.Steps = append(stage.Steps, &backend_types.Step{Name: sn})
	}
	return &builder.Item{
		Workflow: &builder.Workflow{Name: name},
		Config: &backend_types.Config{
			Volume:  volume,
			Network: volume, // same name, mirrors compiler behavior
			Stages:  []*backend_types.Stage{stage},
		},
	}
}

func TestInjectLocalWorkspaceMountsPerWorkflow(t *testing.T) {
	// This is the regression test for the parallel-execution bug:
	// two workflows must end up with DIFFERENT workspace mounts
	// because their Config.Volume names differ, even though they
	// share the same workspace-base.
	items := []*builder.Item{
		testItem("build", "wp_A_1", "compile", "test"),
		testItem("deploy", "wp_B_2", "push"),
	}

	injectLocalWorkspaceMounts(items, "/woodpecker")

	// Workflow "build" steps must mount wp_A_1:/woodpecker.
	for _, step := range items[0].Config.Stages[0].Steps {
		assert.Contains(t, step.Volumes, "wp_A_1:/woodpecker",
			"step %q in workflow 'build' missing per-workflow workspace mount",
			step.Name)
		assert.NotContains(t, step.Volumes, "wp_B_2:/woodpecker",
			"step %q in workflow 'build' wrongly got workflow 'deploy's mount",
			step.Name)
	}

	// Workflow "deploy" steps must mount wp_B_2:/woodpecker.
	for _, step := range items[1].Config.Stages[0].Steps {
		assert.Contains(t, step.Volumes, "wp_B_2:/woodpecker")
		assert.NotContains(t, step.Volumes, "wp_A_1:/woodpecker")
	}
}

func TestInjectLocalWorkspaceMountsAllSteps(t *testing.T) {
	// Every step in every stage should get the mount — a workflow
	// with three steps ends up with three mounted steps.
	it := testItem("build", "wp_X_1", "a", "b", "c")
	injectLocalWorkspaceMounts([]*builder.Item{it}, "/ws")

	for _, step := range it.Config.Stages[0].Steps {
		assert.Equal(t, []string{"wp_X_1:/ws"}, step.Volumes,
			"step %q missing mount", step.Name)
	}
}

func TestInjectLocalWorkspaceMountsAppendsNotReplaces(t *testing.T) {
	// If a step already has existing volumes (user-configured mounts,
	// secrets-as-files, etc.), the workspace mount is appended, not
	// substituted.
	it := testItem("build", "wp_Y_1", "step")
	it.Config.Stages[0].Steps[0].Volumes = []string{"/etc/ssl:/etc/ssl:ro"}

	injectLocalWorkspaceMounts([]*builder.Item{it}, "/ws")

	assert.Equal(t,
		[]string{"/etc/ssl:/etc/ssl:ro", "wp_Y_1:/ws"},
		it.Config.Stages[0].Steps[0].Volumes,
		"existing volumes must be preserved with the mount appended",
	)
}

func TestInjectLocalWorkspaceMountsIgnoresItemsWithoutVolume(t *testing.T) {
	// Defensive: items lacking Config or Config.Volume should not
	// panic and should not get a bogus ":workspace-base" mount.
	items := []*builder.Item{
		testItem("ok", "wp_Z_1", "step"),
		{Workflow: &builder.Workflow{Name: "no-config"}}, // Config == nil
		{
			Workflow: &builder.Workflow{Name: "empty-volume"},
			Config:   &backend_types.Config{Volume: ""},
		},
		nil, // nil item
	}

	// Must not panic.
	injectLocalWorkspaceMounts(items, "/ws")

	// Only the first item gets a mount.
	assert.Equal(t, []string{"wp_Z_1:/ws"},
		items[0].Config.Stages[0].Steps[0].Volumes)
}

func TestInjectLocalWorkspaceMountsMultipleStages(t *testing.T) {
	// A workflow with multiple stages (e.g. services + pipeline):
	// each step across all stages must get the mount.
	it := &builder.Item{
		Workflow: &builder.Workflow{Name: "build"},
		Config: &backend_types.Config{
			Volume: "wp_M_1",
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{{Name: "svc1"}}},
				{Steps: []*backend_types.Step{{Name: "step1"}, {Name: "step2"}}},
			},
		},
	}

	injectLocalWorkspaceMounts([]*builder.Item{it}, "/ws")

	for _, stage := range it.Config.Stages {
		for _, step := range stage.Steps {
			assert.Contains(t, step.Volumes, "wp_M_1:/ws",
				"step %q across multi-stage workflow missing mount",
				step.Name)
		}
	}
}
