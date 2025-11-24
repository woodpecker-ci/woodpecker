// Copyright 2022 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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

package stepbuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/compiler"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestGlobalEnvsubst(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Forge: getMockForge(t),
		Envs: map[string]string{
			"KEY_K": "VALUE_V",
			"IMAGE": "scratch",
		},
		RepoTrusted: &metadata.TrustedConfiguration{},
		Repo:        &model.Repo{},
		Curr: &model.Pipeline{
			Message: "aaa",
			Event:   model.EventPush,
		},
		Prev: &model.Pipeline{},
		Host: "",
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push
steps:
  - name: build
    image: ${IMAGE}
    settings:
      yyy: ${CI_COMMIT_MESSAGE}
`)},
		},
	}

	_, err := b.Build()
	assert.NoError(t, err)
}

func TestMissingGlobalEnvsubst(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Forge: getMockForge(t),
		Envs: map[string]string{
			"KEY_K":    "VALUE_V",
			"NO_IMAGE": "scratch",
		},
		RepoTrusted: &metadata.TrustedConfiguration{},
		Repo:        &model.Repo{},
		Curr: &model.Pipeline{
			Message: "aaa",
			Event:   model.EventPush,
		},
		Prev: &model.Pipeline{},
		Host: "",
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push
steps:
  - name: build
    image: ${IMAGE}
    settings:
      yyy: ${CI_COMMIT_MESSAGE}
`)},
		},
	}

	_, err := b.Build()
	assert.Error(t, err, "test erroneously succeeded")
}

func TestMultilineEnvsubst(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Forge:       getMockForge(t),
		RepoTrusted: &metadata.TrustedConfiguration{},
		Repo:        &model.Repo{},
		Curr: &model.Pipeline{
			Message: `aaa
bbb`,
		},
		Prev: &model.Pipeline{},
		Host: "",
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push
steps:
  - name: xxx
    image: scratch
    settings:
      yyy: ${CI_COMMIT_MESSAGE}
`)},
			{Data: []byte(`
when:
  event: push
steps:
  - name: build
    image: scratch
    settings:
      yyy: ${CI_COMMIT_MESSAGE}
`)},
		},
	}

	_, err := b.Build()
	assert.NoError(t, err)
}

func TestMultiPipeline(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Forge:       getMockForge(t),
		Repo:        &model.Repo{},
		RepoTrusted: &metadata.TrustedConfiguration{},
		Curr: &model.Pipeline{
			Event: model.EventPush,
		},
		Prev: &model.Pipeline{},
		Host: "",
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push
steps:
  - name: xxx
    image: scratch
`)},
			{Data: []byte(`
when:
  event: push
steps:
  - name: build
    image: scratch
`)},
		},
	}

	items, err := b.Build()
	assert.NoError(t, err)
	assert.Len(t, items, 2, "Should have generated 2 items")
}

func TestDependsOn(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Forge:       getMockForge(t),
		Repo:        &model.Repo{},
		RepoTrusted: &metadata.TrustedConfiguration{},
		Curr: &model.Pipeline{
			Event: model.EventPush,
		},
		Prev: &model.Pipeline{},
		Host: "",
		Yamls: []*forge_types.FileMeta{
			{Name: "lint", Data: []byte(`
when:
  event: push
steps:
  - name: build
    image: scratch
`)},
			{Name: "test", Data: []byte(`
when:
  event: push
steps:
  - name: build
    image: scratch
`)},
			{Name: "deploy", Data: []byte(`
when:
  event: push
steps:
  - name: deploy
    image: scratch

depends_on:
  - lint
  - test
`)},
			{Name: "missing dependencies", Data: []byte(`
when:
  event: push
steps:
  - name: deploy
    image: scratch

depends_on:
  - missing
`)},
		},
	}

	items, err := b.Build()
	assert.NoError(t, err)
	assert.Len(t, items, 3, "Should have generated 3 items")
	assert.Len(t, items[0].DependsOn, 2, "Should have 2 dependencies")
	assert.Equal(t, "test", items[0].DependsOn[1], "Should depend on test")
}

func TestRunsOn(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Forge:       getMockForge(t),
		RepoTrusted: &metadata.TrustedConfiguration{},
		Repo:        &model.Repo{},
		Curr: &model.Pipeline{
			Event: model.EventPush,
		},
		Prev: &model.Pipeline{},
		Host: "",
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push
steps:
  - name: deploy
    image: scratch

runs_on:
  - success
  - failure
`)},
		},
	}

	items, err := b.Build()
	assert.NoError(t, err)
	assert.Len(t, items[0].RunsOn, 2, "Should run on success and failure")
	assert.Equal(t, "failure", items[0].RunsOn[1], "Should run on failure")
}

func TestPipelineName(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Forge:       getMockForge(t),
		RepoTrusted: &metadata.TrustedConfiguration{},
		Repo:        &model.Repo{Config: ".woodpecker"},
		Curr: &model.Pipeline{
			Event: model.EventPush,
		},
		Prev: &model.Pipeline{},
		Host: "",
		Yamls: []*forge_types.FileMeta{
			{Name: ".woodpecker/lint.yml", Data: []byte(`
when:
  event: push
steps:
  - name: build
    image: scratch
`)},
			{Name: ".woodpecker/.test.yml", Data: []byte(`
when:
  event: push
steps:
  - name: build
    image: scratch
`)},
		},
	}

	items, err := b.Build()
	assert.NoError(t, err)
	pipelineNames := []string{items[0].Workflow.Name, items[1].Workflow.Name}
	assert.True(t, containsItemWithName("lint", items) && containsItemWithName("test", items),
		"Pipeline name should be 'lint' and 'test' but are '%v'", pipelineNames)
}

func TestBranchFilter(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Forge:       getMockForge(t),
		RepoTrusted: &metadata.TrustedConfiguration{},
		Repo:        &model.Repo{},
		Curr: &model.Pipeline{
			Branch: "dev",
			Event:  model.EventPush,
		},
		Prev: &model.Pipeline{},
		Host: "",
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push
  branch: main
steps:
  - name: xxx
    image: scratch
`)},
			{Data: []byte(`
when:
  event: push
steps:
  - name: build
    image: scratch
`)},
		},
	}

	items, err := b.Build()
	assert.NoError(t, err)
	assert.Len(t, items, 1, "Should have generated 1 pipeline")
	assert.Equal(t, model.StatusPending, items[0].Workflow.State, "Should run on dev branch")
}

func TestRootWhenFilter(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Forge:       getMockForge(t),
		RepoTrusted: &metadata.TrustedConfiguration{},
		Repo:        &model.Repo{},
		Curr:        &model.Pipeline{Event: "tag"},
		Prev:        &model.Pipeline{},
		Host:        "",
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event:
    - tag
steps:
  - name: xxx
    image: scratch
`)},
			{Data: []byte(`
when:
  event:
    - push
steps:
  - name: xxx
    image: scratch
`)},
			{Data: []byte(`
steps:
  - name: build
    image: scratch
`)},
		},
	}

	items, err := b.Build()
	assert.False(t, errors.HasBlockingErrors(err))
	assert.Len(t, items, 2, "Should have generated 2 items")
}

func TestZeroSteps(t *testing.T) {
	t.Parallel()

	pipeline := &model.Pipeline{
		Branch: "dev",
		Event:  model.EventPush,
	}

	b := StepBuilder{
		Forge:       getMockForge(t),
		RepoTrusted: &metadata.TrustedConfiguration{},
		Repo:        &model.Repo{},
		Curr:        pipeline,
		Prev:        &model.Pipeline{},
		Host:        "",
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push
skip_clone: true
steps:
  - name: build
    when:
      branch: notdev
    image: scratch
`)},
		},
	}

	items, err := b.Build()
	assert.NoError(t, err)
	assert.Empty(t, items, "Should not generate a pipeline item if there are no steps")
}

func TestZeroStepsAsMultiPipelineTransitiveDeps(t *testing.T) {
	t.Parallel()

	pipeline := &model.Pipeline{
		Branch: "dev",
		Event:  model.EventPush,
	}

	b := StepBuilder{
		Forge:       getMockForge(t),
		RepoTrusted: &metadata.TrustedConfiguration{},
		Repo:        &model.Repo{},
		Curr:        pipeline,
		Prev:        &model.Pipeline{},
		Host:        "",
		Yamls: []*forge_types.FileMeta{
			{Name: "zerostep", Data: []byte(`
when:
  event: push
skip_clone: true
steps:
  - name: build
    when:
      branch: notdev
    image: scratch
`)},
			{Name: "justastep", Data: []byte(`
when:
  event: push
steps:
  - name: build
    image: scratch
`)},
			{Name: "shouldbefiltered", Data: []byte(`
when:
  event: push
steps:
  - name: build
    image: scratch
depends_on: [ zerostep ]
`)},
			{Name: "shouldbefilteredtoo", Data: []byte(`
when:
  event: push
steps:
  - name: build
    image: scratch
depends_on: [ shouldbefiltered ]
`)},
		},
	}

	items, err := b.Build()
	assert.NoError(t, err)
	assert.Len(t, items, 1, "Zerostep and the step that depends on it, and the one depending on it should not generate a pipeline item")
	assert.Equal(t, "justastep", items[0].Workflow.Name, "justastep should have been generated")
}

func TestSanitizePath(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		path          string
		sanitizedPath string
	}{
		{
			path:          ".woodpecker/test.yml",
			sanitizedPath: "test",
		},
		{
			path:          ".woodpecker.yml",
			sanitizedPath: "woodpecker",
		},
		{
			path:          "folder/sub-folder/test.yml",
			sanitizedPath: "test",
		},
		{
			path:          ".woodpecker/test.yaml",
			sanitizedPath: "test",
		},
		{
			path:          ".woodpecker.yaml",
			sanitizedPath: "woodpecker",
		},
		{
			path:          "folder/sub-folder/test.yaml",
			sanitizedPath: "test",
		},
	}

	for _, test := range testTable {
		assert.Equal(t, test.sanitizedPath, SanitizePath(test.path), "Path hasn't been sanitized correctly")
	}
}

func TestMatrix(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Forge:       getMockForge(t),
		RepoTrusted: &metadata.TrustedConfiguration{},
		Repo:        &model.Repo{},
		Curr:        &model.Pipeline{Event: model.EventPush},
		Prev:        &model.Pipeline{},
		Host:        "",
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push

matrix:
  GO_VERSION:
    - 1.14
    - 1.15

steps:
  - name: build
    image: golang:${GO_VERSION}
    commands:
      - go build
`)},
		},
	}

	items, err := b.Build()
	assert.NoError(t, err)
	assert.Len(t, items, 2)

	// Check AxisID and Environ
	assert.Equal(t, 1, items[0].Workflow.AxisID)
	assert.Equal(t, "1.14", items[0].Workflow.Environ["GO_VERSION"])

	assert.Equal(t, 2, items[1].Workflow.AxisID)
	assert.Equal(t, "1.15", items[1].Workflow.Environ["GO_VERSION"])
}

func TestMissingWorkflowDeps(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Forge:       getMockForge(t),
		RepoTrusted: &metadata.TrustedConfiguration{},
		Repo:        &model.Repo{},
		Curr:        &model.Pipeline{Event: model.EventPush},
		Prev:        &model.Pipeline{},
		Host:        "",
		Yamls: []*forge_types.FileMeta{
			{
				Name: "workflow-with-missing-deps",
				Data: []byte(`
when:
  event: push
steps:
  - name: build
    image: scratch
depends_on:
  - non-existing
`),
			},
		},
	}

	items, err := b.Build()
	assert.NoError(t, err)
	assert.Empty(t, items, "Workflows with missing dependencies should be filtered out")
}

func TestInvalidYAML(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Forge:       getMockForge(t),
		RepoTrusted: &metadata.TrustedConfiguration{},
		Repo:        &model.Repo{},
		Curr:        &model.Pipeline{Event: model.EventPush},
		Prev:        &model.Pipeline{},
		Yamls: []*forge_types.FileMeta{
			{Name: "broken-yaml", Data: []byte(`
when:
  event: push
steps:
  - name: build
    image: scratch
	invalid yaml indentation
`)},
		},
	}

	_, err := b.Build()
	assert.Error(t, err, "Invalid YAML should return an error")
}

func TestEnvVarPrecedence(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Forge: getMockForge(t),
		Envs: map[string]string{
			"CUSTOM_VAR":     "global-value",
			"CI_REPO_NAME":   "should-not-override",
			"ANOTHER_CUSTOM": "global-value-2",
		},
		RepoTrusted: &metadata.TrustedConfiguration{},
		Repo:        &model.Repo{Name: "actual-repo"},
		Curr: &model.Pipeline{
			Event:   model.EventPush,
			Message: "test",
		},
		Prev: &model.Pipeline{},
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push
steps:
  - name: test-env
    image: scratch
    environment:
      CUSTOM_VAR: ${CUSTOM_VAR}
      REPO_NAME: ${CI_REPO_NAME}
      ANOTHER: ${ANOTHER_CUSTOM}
`)},
		},
	}

	items, err := b.Build()
	assert.NoError(t, err)
	assert.Len(t, items, 1)

	// Verify CI_REPO_NAME wasn't overridden by global env
	assert.Equal(t, "actual-repo", items[0].Config.Stages[0].Steps[0].Environment["CI_REPO_NAME"])
}

func TestLabelMerging(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Forge:       getMockForge(t),
		RepoTrusted: &metadata.TrustedConfiguration{},
		Repo:        &model.Repo{Name: "test-repo"},
		Curr:        &model.Pipeline{Event: model.EventPush},
		Prev:        &model.Pipeline{},
		DefaultLabels: map[string]string{
			"default-label": "default-value",
			"override-me":   "default",
		},
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push

labels:
  override-me: "custom-value"
  workflow-label: "workflow-value"

steps:
  - name: build
    image: scratch
`)},
			{Data: []byte(`
when:
  event: push

steps:
  - name: build
    image: scratch
`)},
		},
	}

	items, err := b.Build()
	assert.NoError(t, err)
	assert.Len(t, items, 2)

	assert.Equal(t, "custom-value", items[0].Labels["override-me"], "Workflow label should override default")
	assert.Equal(t, "workflow-value", items[0].Labels["workflow-label"], "Workflow-specific label should be present")
	assert.Equal(t, "default-value", items[1].Labels["default-label"], "Default label should be present")
}

func TestCompilerOptions(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Forge:       getMockForge(t),
		RepoTrusted: &metadata.TrustedConfiguration{},
		Repo:        &model.Repo{},
		Curr:        &model.Pipeline{Event: model.EventPush},
		Prev:        &model.Pipeline{},
		CompilerOptions: []compiler.Option{
			compiler.WithEnviron(map[string]string{
				"KEY": "VALUE",
			}),
		},
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
skip_clone: true
when:
  event: push
steps:
  - name: build
    image: scratch
`)},
		},
	}

	items, err := b.Build()
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Len(t, items[0].Config.Stages, 1, "Should have 1 stage")
	assert.Len(t, items[0].Config.Stages[0].Steps, 1, "Should have 1 step in first stage")
	assert.Equal(t, "VALUE", items[0].Config.Stages[0].Steps[0].Environment["KEY"], "Environment variable should be set")
}

func getMockForge(t *testing.T) forge.Forge {
	forge := mocks.NewMockForge(t)
	forge.On("Name").Return("mock")
	forge.On("URL").Return("https://codeberg.org")
	return forge
}
