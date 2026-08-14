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

package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
)

func compileBuilder(t *testing.T, yamls ...*YamlFile) PipelineBuilder {
	t.Helper()

	return PipelineBuilder{
		GetWorkflowMetadata: (&testMetadata{pipelineEvent: "push", branch: "main"}).GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls:               yamls,
	}
}

// stepNames lists the steps of an item, in stage order, without the clone step
// and without services.
func stepNames(item *Item) []string {
	var names []string
	for _, stage := range item.Config.Stages {
		for _, step := range stage.Steps {
			if step.Type == backend_types.StepTypeClone || step.Type == backend_types.StepTypeService {
				continue
			}
			names = append(names, step.Name)
		}
	}

	return names
}

func TestBuildSeparatesTheTwoPhases(t *testing.T) {
	t.Parallel()

	b := compileBuilder(t, &YamlFile{Name: "build.yaml", Data: []byte(`
when:
  event: push
compile:
  generate:
    image: alpine
    commands: [./generate.sh]
services:
  database:
    image: postgres
steps:
  test:
    image: golang
    commands: [go test ./...]
`)})

	plan, err := b.Build()
	require.NoError(t, err)

	require.Len(t, plan.Compile, 1)
	require.Len(t, plan.Run, 1)

	assert.Equal(t, PhaseCompile, plan.Compile[0].Workflow.Phase)
	assert.Equal(t, PhaseRun, plan.Run[0].Workflow.Phase)
	assert.Equal(t, "build", plan.Compile[0].Workflow.Name)
	assert.Equal(t, "build", plan.Run[0].Workflow.Name,
		"the two phases are resolved separately, so a shared name does not collide")

	assert.Equal(t, []string{"generate"}, stepNames(plan.Compile[0]),
		"the compile workflow runs the compile steps, not the ordinary ones")
	assert.Equal(t, []string{"test"}, stepNames(plan.Run[0]))

	// a config generator does not need sidecars
	assert.Empty(t, serviceNames(plan.Compile[0]))
	assert.Equal(t, []string{"database"}, serviceNames(plan.Run[0]))
}

// serviceNames lists the service steps of an item.
func serviceNames(item *Item) []string {
	var names []string
	for _, stage := range item.Config.Stages {
		for _, step := range stage.Steps {
			if step.Type == backend_types.StepTypeService {
				names = append(names, step.Name)
			}
		}
	}

	return names
}

func TestBuildWithoutCompileSection(t *testing.T) {
	t.Parallel()

	b := compileBuilder(t, &YamlFile{Name: "build.yaml", Data: []byte(`
when:
  event: push
steps:
  test:
    image: golang
`)})

	plan, err := b.Build()
	require.NoError(t, err)

	assert.Empty(t, plan.Compile, "a config without compile steps contributes nothing to phase 0")
	assert.Len(t, plan.Run, 1)
}

func TestBuildWithOnlyACompileSection(t *testing.T) {
	t.Parallel()

	b := compileBuilder(t, &YamlFile{Name: "build.yaml", Data: []byte(`
when:
  event: push
compile:
  generate:
    image: alpine
    commands: [./generate.sh]
`)})

	plan, err := b.Build()
	require.NoError(t, err)

	assert.Len(t, plan.Compile, 1)
	assert.Empty(t, plan.Run, "a config with no ordinary steps contributes nothing to phase 1")
}

func TestBuildAppliesTheRootWhenToBothPhases(t *testing.T) {
	t.Parallel()

	b := compileBuilder(t, &YamlFile{Name: "build.yaml", Data: []byte(`
when:
  event: tag
compile:
  generate:
    image: alpine
    commands: [./generate.sh]
steps:
  test:
    image: golang
`)})

	plan, err := b.Build()
	require.NoError(t, err)

	assert.Empty(t, plan.Compile, "the root when: gates the compile workflow like any other")
	assert.Empty(t, plan.Run)
}

func TestBuildSkipsACompileWorkflowWithEveryStepFiltered(t *testing.T) {
	t.Parallel()

	// Every compile step is filtered by its own when:, but the clone stage
	// keeps the compiled config non-empty. Running it would emit no config
	// response and fail the pipeline, which would punish filtering on purpose.
	b := compileBuilder(t, &YamlFile{Name: "build.yaml", Data: []byte(`
when:
  event: push
compile:
  generate:
    image: alpine
    commands: [./generate.sh]
    when:
      event: tag
steps:
  test:
    image: golang
`)})

	plan, err := b.Build()
	require.NoError(t, err)

	assert.Empty(t, plan.Compile)
	assert.Len(t, plan.Run, 1)
}

func TestBuildNumbersEachPhaseIndependently(t *testing.T) {
	t.Parallel()

	b := compileBuilder(t,
		&YamlFile{Name: "a.yaml", Data: []byte(`
when:
  event: push
compile:
  generate:
    image: alpine
    commands: [./generate.sh]
steps:
  test:
    image: golang
`)},
		&YamlFile{Name: "b.yaml", Data: []byte(`
when:
  event: push
compile:
  generate:
    image: alpine
    commands: [./generate.sh]
steps:
  test:
    image: golang
`)},
	)

	plan, err := b.Build()
	require.NoError(t, err)

	require.Len(t, plan.Compile, 2)
	require.Len(t, plan.Run, 2)
	assert.Equal(t, []int{1, 2}, []int{plan.Compile[0].Workflow.PID, plan.Compile[1].Workflow.PID})
	assert.Equal(t, []int{1, 2}, []int{plan.Run[0].Workflow.PID, plan.Run[1].Workflow.PID})
}

func TestBuildResolvesDependenciesWithinAPhase(t *testing.T) {
	t.Parallel()

	// The run workflow of "generate" does not exist, so a run workflow that
	// depends on it must not be wired to the compile workflow of the same name.
	b := compileBuilder(t,
		&YamlFile{Name: "generate.yaml", Data: []byte(`
when:
  event: push
compile:
  generate:
    image: alpine
    commands: [./generate.sh]
`)},
		&YamlFile{Name: "test.yaml", Data: []byte(`
when:
  event: push
depends_on: [generate]
steps:
  test:
    image: golang
`)},
	)

	_, err := b.Build()
	assert.ErrorContains(t, err, `workflow "test" depends on "generate"`,
		"phase 1 cannot depend on phase 0")
}
