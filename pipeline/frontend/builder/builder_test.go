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

package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/compiler"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/constraint"
)

func TestGlobalEnvsubst(t *testing.T) {
	t.Parallel()

	m := &testMetadata{}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		Envs: map[string]string{
			"KEY_K": "VALUE_V",
			"IMAGE": "scratch",
		},
		RepoTrusted: &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
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

	m := &testMetadata{}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		Envs: map[string]string{
			"KEY_K":    "VALUE_V",
			"NO_IMAGE": "scratch",
		},
		RepoTrusted: &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
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

func TestAdditionalEnvsSubstitution(t *testing.T) {
	t.Parallel()

	m := &testMetadata{
		pipelineEvent: "manual",
	}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		AdditionalEnvs: map[string]string{
			"TAG": "1.2.3",
		},
		RepoTrusted:         &metadata.TrustedConfiguration{},
		TrustedClonePlugins: []string{"woodpeckerci/plugin-git"},
		Yamls: []*YamlFile{
			{Data: []byte(`
when:
  event: manual
clone:
  - name: clone
    image: woodpeckerci/plugin-git
    settings:
      ref: refs/tags/${TAG}
steps:
  - name: build
    image: scratch
    commands:
      - go build
`)},
		},
	}

	items, err := b.Build()
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	cloneStep := items[0].Config.Stages[0].Steps[0]
	assert.Equal(t, "clone", cloneStep.Name)
	assert.Equal(t, "refs/tags/1.2.3", cloneStep.Environment["PLUGIN_REF"],
		"additional (manual trigger) variables must be substituted in the yaml")
}

func TestAdditionalAndMatrixEnvsNotInjectedIntoPlugins(t *testing.T) {
	t.Parallel()

	m := &testMetadata{
		pipelineEvent: "manual",
	}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		AdditionalEnvs: map[string]string{
			"TAG": "1.2.3",
		},
		RepoTrusted: &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
			{Data: []byte(`
when:
  event: manual
skip_clone: true
matrix:
  GO_VERSION:
    - "1.22"
steps:
  - name: build
    image: scratch
    commands:
      - go build
  - name: publish
    image: scratch
    settings:
      ref: refs/tags/${TAG}
`)},
		},
	}

	items, err := b.Build()
	assert.NoError(t, err)
	assert.Len(t, items, 1)

	assert.Len(t, items[0].Config.Stages, 2)
	buildStep := items[0].Config.Stages[0].Steps[0]
	publishStep := items[0].Config.Stages[1].Steps[0]
	assert.Equal(t, "build", buildStep.Name)
	assert.Equal(t, "publish", publishStep.Name)

	// non-plugin steps get additional (manual trigger) and matrix envs
	assert.Equal(t, "1.2.3", buildStep.Environment["TAG"])
	assert.Equal(t, "1.22", buildStep.Environment["GO_VERSION"])

	// plugin steps must not have them injected to avoid smuggling settings
	assert.NotContains(t, publishStep.Environment, "TAG")
	assert.NotContains(t, publishStep.Environment, "GO_VERSION")
}

func TestAdditionalEnvsWhenEvaluate(t *testing.T) {
	t.Parallel()

	yamlData := []byte(`
when:
  event: manual
  evaluate: 'TAG == "1.2.3"'
steps:
  - name: build
    image: scratch
    commands:
      - go build
`)

	m := &testMetadata{
		pipelineEvent: "manual",
	}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		AdditionalEnvs: map[string]string{
			"TAG": "1.2.3",
		},
		RepoTrusted: &metadata.TrustedConfiguration{},
		Yamls:       []*YamlFile{{Data: yamlData}},
	}

	items, err := b.Build()
	assert.NoError(t, err)
	assert.Len(t, items, 1,
		"workflow must not be filtered when evaluate matches an additional (manual trigger) variable")
}

func TestMultilineEnvsubst(t *testing.T) {
	t.Parallel()

	m := &testMetadata{}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
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

	m := &testMetadata{
		pipelineEvent: "push",
	}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
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

	m := &testMetadata{
		pipelineEvent: "push",
	}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
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
	assert.Equal(t, "test", items[0].DependsOn[1].Name, "Should depend on test")
}

func TestRunsOn(t *testing.T) {
	t.Parallel()

	m := &testMetadata{
		pipelineEvent: "push",
	}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
			{Data: []byte(`
when:
  event: push
  status: [ success, failure ]

steps:
  - name: deploy
    image: scratch
`)},
		},
	}

	items, err := b.Build()
	assert.NoError(t, err)
	assert.Len(t, items, 1, "Should have generated 1 pipeline")
	assert.Len(t, items[0].RunsOn, 2, "Should run on success and failure")
	assert.ElementsMatchf(t, []string{"success", "failure"}, items[0].RunsOn, "Should run on failure")
}

func TestRunsOnStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		when string
		want []string
	}{
		{
			name: "only failure",
			when: "when:\n  event: push\n  status: [ failure ]\n",
			want: []string{"failure"},
		},
		{
			name: "only success",
			when: "when:\n  event: push\n  status: [ success ]\n",
			want: []string{"success"},
		},
		{
			name: "no status at all",
			when: "when:\n  event: push\n",
			want: []string{"success"},
		},
		{
			// The deprecated spelling must keep deciding on its own, even
			// though the when above it sets no status.
			name: "deprecated runs_on wins over when",
			when: "when:\n  event: push\nruns_on: [ failure ]\n",
			want: []string{"failure"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &testMetadata{
				pipelineEvent: "push",
			}

			b := PipelineBuilder{
				GetWorkflowMetadata: m.GetWorkflowMetadata,
				RepoTrusted:         &metadata.TrustedConfiguration{},
				Yamls: []*YamlFile{
					{Data: []byte(tt.when + "\nsteps:\n  - name: deploy\n    image: scratch\n")},
				},
			}

			// runs_on is deprecated, so Build reports a warning for that case.
			items, _ := b.Build()
			assert.Len(t, items, 1, "Should have generated 1 pipeline")
			assert.ElementsMatch(t, tt.want, items[0].RunsOn)
		})
	}
}

func TestPipelineName(t *testing.T) {
	t.Parallel()

	m := &testMetadata{
		pipelineEvent: "push",
	}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
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
	assert.Len(t, items, 2, "Should have generated 2 pipelines")
	pipelineNames := []string{items[0].Workflow.Name, items[1].Workflow.Name}
	assert.True(t, ContainsItemWithName("lint", items) && ContainsItemWithName("test", items),
		"Pipeline name should be 'lint' and 'test' but are '%v'", pipelineNames)
}

func TestBranchFilter(t *testing.T) {
	t.Parallel()

	m := &testMetadata{
		pipelineEvent: "push",
		branch:        "dev",
	}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
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
}

func TestRootWhenFilter(t *testing.T) {
	t.Parallel()

	m := &testMetadata{
		pipelineEvent: "tag",
	}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
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

	m := &testMetadata{}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
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

	m := &testMetadata{
		pipelineEvent: "push",
	}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
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

	for _, tt := range testTable {
		assert.Equal(t, tt.sanitizedPath, SanitizePath(tt.path), "Path hasn't been sanitized correctly")
	}
}

func TestMatrix(t *testing.T) {
	t.Parallel()

	m := &testMetadata{
		pipelineEvent: "push",
	}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
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

	m := &testMetadata{}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
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

func TestDependsOnOptionalFlag(t *testing.T) {
	t.Parallel()

	m := &testMetadata{
		pipelineEvent: metadata.EventPush,
		branch:        "main",
		repo:          "test-repo",
	}

	t.Run("missing optional dep is dropped", func(t *testing.T) {
		b := PipelineBuilder{
			GetWorkflowMetadata: m.GetWorkflowMetadata,
			RepoTrusted:         &metadata.TrustedConfiguration{},
			Yamls: []*YamlFile{
				{Name: "check-a", Data: []byte(`
when:
  event: push
steps:
  - name: check
    image: scratch
`)},
				{Name: "deploy", Data: []byte(`
when:
  event: push
steps:
  - name: deploy
    image: scratch
depends_on:
  - check-a
  - name: check-b
    optional: true
`)},
			},
		}

		items, err := b.Build()
		assert.NoError(t, err)
		assert.Len(t, items, 2, "deploy should not be filtered out")
		deploy := items[0]
		if deploy.Workflow.Name != "deploy" {
			deploy = items[1]
		}
		assert.Equal(t, constraint.DependsOn{{Name: "check-a"}}, deploy.DependsOn, "only check-a should remain as dependency")
	})

	t.Run("present optional dep is promoted", func(t *testing.T) {
		b := PipelineBuilder{
			GetWorkflowMetadata: m.GetWorkflowMetadata,
			RepoTrusted:         &metadata.TrustedConfiguration{},
			Yamls: []*YamlFile{
				{Name: "check-a", Data: []byte(`
when:
  event: push
steps:
  - name: check
    image: scratch
`)},
				{Name: "check-b", Data: []byte(`
when:
  event: push
steps:
  - name: check
    image: scratch
`)},
				{Name: "deploy", Data: []byte(`
when:
  event: push
steps:
  - name: deploy
    image: scratch
depends_on:
  - check-a
  - name: check-b
    optional: true
`)},
			},
		}

		items, err := b.Build()
		assert.NoError(t, err)
		assert.Len(t, items, 3)
		deploy := items[0]
		for _, item := range items {
			if item.Workflow.Name == "deploy" {
				deploy = item
			}
		}
		assert.ElementsMatch(t, []string{"check-a", "check-b"}, deploy.DependsOn.Names(), "both deps should be present")
	})

	t.Run("missing required dep still removes workflow", func(t *testing.T) {
		b := PipelineBuilder{
			GetWorkflowMetadata: m.GetWorkflowMetadata,
			RepoTrusted:         &metadata.TrustedConfiguration{},
			Yamls: []*YamlFile{
				{Name: "check-a", Data: []byte(`
when:
  event: push
steps:
  - name: check
    image: scratch
`)},
				{Name: "deploy", Data: []byte(`
when:
  event: push
steps:
  - name: deploy
    image: scratch
depends_on:
  - missing
  - name: check-a
    optional: true
`)},
			},
		}

		items, err := b.Build()
		assert.NoError(t, err)
		assert.Len(t, items, 1, "deploy should be filtered out due to missing required dep")
		assert.Equal(t, "check-a", items[0].Workflow.Name)
	})

	t.Run("optional dep filtered by when is dropped", func(t *testing.T) {
		b := PipelineBuilder{
			GetWorkflowMetadata: m.GetWorkflowMetadata,
			RepoTrusted:         &metadata.TrustedConfiguration{},
			Yamls: []*YamlFile{
				{Name: "check-a", Data: []byte(`
when:
  event: push
steps:
  - name: check
    image: scratch
`)},
				{Name: "check-b", Data: []byte(`
when:
  event: tag
steps:
  - name: check
    image: scratch
`)},
				{Name: "deploy", Data: []byte(`
when:
  event: push
steps:
  - name: deploy
    image: scratch
depends_on:
  - check-a
  - name: check-b
    optional: true
`)},
			},
		}

		items, err := b.Build()
		assert.NoError(t, err)
		assert.Len(t, items, 2, "check-b filtered by when, deploy should still run")
		deploy := items[0]
		for _, item := range items {
			if item.Workflow.Name == "deploy" {
				deploy = item
			}
		}
		assert.Equal(t, constraint.DependsOn{{Name: "check-a"}}, deploy.DependsOn)
	})

	t.Run("optional dep on workflow removed for own missing required dep is dropped", func(t *testing.T) {
		b := PipelineBuilder{
			GetWorkflowMetadata: m.GetWorkflowMetadata,
			RepoTrusted:         &metadata.TrustedConfiguration{},
			Yamls: []*YamlFile{
				{Name: "broken", Data: []byte(`
when:
  event: push
steps:
  - name: check
    image: scratch
depends_on:
  - missing
`)},
				{Name: "deploy", Data: []byte(`
when:
  event: push
steps:
  - name: deploy
    image: scratch
depends_on:
  - name: broken
    optional: true
`)},
			},
		}

		items, err := b.Build()
		assert.NoError(t, err)
		assert.Len(t, items, 1, "deploy should survive: its only dep is optional and the target was removed")
		assert.Equal(t, "deploy", items[0].Workflow.Name)
		assert.Empty(t, items[0].DependsOn, "optional dep on a removed workflow should be dropped")
	})
}

func TestInvalidYAML(t *testing.T) {
	t.Parallel()

	m := &testMetadata{}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
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
	assert.ErrorContains(t, err, "found a tab character that violates indentation")
}

func TestEnvVarPrecedence(t *testing.T) {
	t.Parallel()

	m := &testMetadata{
		pipelineEvent: "push",
		repo:          "actual-repo",
	}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		Envs: map[string]string{
			"CUSTOM_VAR":     "global-value",
			"CI_REPO_NAME":   "should-not-override",
			"ANOTHER_CUSTOM": "global-value-2",
		},
		RepoTrusted: &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
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

	m := &testMetadata{
		pipelineEvent: "push",
	}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		DefaultLabels: map[string]string{
			"default-label": "default-value",
			"override-me":   "default",
		},
		Yamls: []*YamlFile{
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

	m := &testMetadata{
		pipelineEvent: "push",
	}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		CompilerOptions: []compiler.Option{
			compiler.WithEnviron(map[string]string{
				"KEY": "VALUE",
			}),
		},
		Yamls: []*YamlFile{
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

type testMetadata struct {
	pipelineEvent metadata.Event
	branch        string
	repo          string
}

func (t *testMetadata) GetWorkflowMetadata(w *Workflow) metadata.Metadata {
	return metadata.Metadata{
		Repo: metadata.Repo{
			Name: t.repo,
		},
		Curr: metadata.Pipeline{
			Event: t.pipelineEvent,
			Commit: metadata.Commit{
				Branch: t.branch,
			},
		},
	}
}

func TestInternalCommitBranchAndEventLabels(t *testing.T) {
	t.Parallel()

	m := &testMetadata{
		pipelineEvent: "push",
		branch:        "main",
	}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
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
	assert.Len(t, items, 1)

	assert.Equal(t, "main", items[0].Labels[pipeline.LabelCommitBranch])
	assert.Equal(t, "push", items[0].Labels[pipeline.LabelEvent])
}

func selectedWorkflowYamls() []*YamlFile {
	return []*YamlFile{
		{Name: "lint", Data: []byte(`
steps:
  - name: lint
    image: scratch
`)},
		{Name: "build", Data: []byte(`
steps:
  - name: build
    image: scratch
depends_on: [lint]
`)},
		{Name: "deploy", Data: []byte(`
when:
  event: manual
steps:
  - name: deploy
    image: scratch
depends_on: [build]
`)},
		{Name: "push-only", Data: []byte(`
when:
  event: push
steps:
  - name: push
    image: scratch
`)},
		{Name: "prod-only", Data: []byte(`
when:
  event: manual
  evaluate: 'DEPLOY_ENV == "prod"'
steps:
  - name: prod
    image: scratch
`)},
	}
}

func TestSelectedWorkflow(t *testing.T) {
	t.Parallel()

	m := &testMetadata{pipelineEvent: "manual"}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls:               selectedWorkflowYamls(),
		SelectedWorkflow:    "deploy",
	}

	items, err := b.Build()
	assert.False(t, errors.HasBlockingErrors(err))
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Workflow.Name)
	}
	assert.Equal(t, []string{"build", "deploy", "lint"}, names, "should run the selected workflow and its transitive dependencies")
}

func TestSelectedWorkflowSkipsOptionalDependencies(t *testing.T) {
	t.Parallel()

	m := &testMetadata{pipelineEvent: "manual"}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls: []*YamlFile{
			{Name: "lint", Data: []byte(`
steps:
  - name: lint
    image: scratch
`)},
			{Name: "test", Data: []byte(`
steps:
  - name: test
    image: scratch
`)},
			{Name: "deploy", Data: []byte(`
steps:
  - name: deploy
    image: scratch
depends_on:
  - lint
  - name: test
    optional: true
`)},
		},
		SelectedWorkflow: "deploy",
	}

	items, err := b.Build()
	assert.False(t, errors.HasBlockingErrors(err))
	assert.Len(t, items, 2)
	assert.Equal(t, "deploy", items[0].Workflow.Name)
	assert.Equal(t, []string{"lint"}, items[0].DependsOn.Names(), "optional dependency should be dropped, not started")
	assert.Equal(t, "lint", items[1].Workflow.Name)
}

func TestSelectedWorkflowByPath(t *testing.T) {
	t.Parallel()

	m := &testMetadata{pipelineEvent: "manual"}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls:               selectedWorkflowYamls(),
		SelectedWorkflow:    ".woodpecker/deploy.yml",
	}

	items, err := b.Build()
	assert.False(t, errors.HasBlockingErrors(err))
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Workflow.Name)
	}
	assert.Equal(t, []string{"build", "deploy", "lint"}, names, "selection should be normalized like workflow names")
}

func TestListOnly(t *testing.T) {
	t.Parallel()

	m := &testMetadata{pipelineEvent: "manual"}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls:               selectedWorkflowYamls(),
		ListOnly:            true,
	}

	items, err := b.Build()
	assert.False(t, errors.HasBlockingErrors(err))
	names := make([]string, 0, len(items))
	for _, item := range items {
		assert.Nil(t, item.Config, "should not compile")
		names = append(names, item.Workflow.Name)
	}
	assert.Equal(t, []string{"build", "deploy", "lint", "prod-only"}, names, "should list workflows gated by evaluate, but not by other conditions")
}

func TestSelectedWorkflowNotFound(t *testing.T) {
	t.Parallel()

	m := &testMetadata{pipelineEvent: "manual"}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls:               selectedWorkflowYamls(),
		SelectedWorkflow:    "push-only",
	}

	items, err := b.Build()
	assert.Nil(t, items)
	assert.True(t, errors.HasBlockingErrors(err))
	assert.ErrorContains(t, err, "selected workflow 'push-only' does not match its conditions")
}

func TestSelectedWorkflowDoesNotExist(t *testing.T) {
	t.Parallel()

	m := &testMetadata{pipelineEvent: "manual"}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls:               selectedWorkflowYamls(),
		SelectedWorkflow:    "release",
	}

	items, err := b.Build()
	assert.Nil(t, items)
	assert.ErrorContains(t, err, "selected workflow 'release' does not exist")
}

// workflowsWithUnrelatedBroken has a lint workflow next to a deploy workflow
// that needs manual variables to substitute and to evaluate its conditions.
func workflowsWithUnrelatedBroken() []*YamlFile {
	return []*YamlFile{
		{Name: "lint", Data: []byte(`
steps:
  - name: lint
    image: scratch
`)},
		{Name: "deploy", Data: []byte(`
when:
  event: manual
  evaluate: 'int(DEPLOY_COUNT) > 0'
steps:
  - name: deploy
    image: ${BUILD_IMAGE}
`)},
	}
}

func TestSelectedWorkflowIgnoresUnrelatedWorkflows(t *testing.T) {
	t.Parallel()

	m := &testMetadata{pipelineEvent: "manual"}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls:               workflowsWithUnrelatedBroken(),
	}

	_, err := b.Build()
	assert.True(t, errors.HasBlockingErrors(err), "without the variables the full build should fail")

	b.SelectedWorkflow = "lint"
	items, err := b.Build()
	assert.False(t, errors.HasBlockingErrors(err), "unrelated workflows should not be substituted, linted or compiled")
	assert.Len(t, items, 1)
	assert.Equal(t, "lint", items[0].Workflow.Name)
}

func TestListOnlyToleratesMissingVariables(t *testing.T) {
	t.Parallel()

	m := &testMetadata{pipelineEvent: "manual"}

	b := PipelineBuilder{
		GetWorkflowMetadata: m.GetWorkflowMetadata,
		RepoTrusted:         &metadata.TrustedConfiguration{},
		Yamls:               workflowsWithUnrelatedBroken(),
		ListOnly:            true,
	}

	items, err := b.Build()
	assert.False(t, errors.HasBlockingErrors(err))
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Workflow.Name)
	}
	assert.Equal(t, []string{"deploy", "lint"}, names, "workflows needing variables should still be listed")
}

func TestSelectedWorkflowSubstitutedDependencies(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name       string
		dependency string
		matrix     string
		envs       map[string]string
		variables  map[string]string
		want       []string
		wantError  bool
	}{
		{
			name:       "manual variable overrides global",
			dependency: "${DEPENDENCY}",
			envs:       map[string]string{"DEPENDENCY": "ignored"},
			variables:  map[string]string{"DEPENDENCY": "build"},
			want:       []string{"build", "deploy", "lint"},
		},
		{
			name:       "global variable",
			dependency: "${DEPENDENCY}",
			envs:       map[string]string{"DEPENDENCY": "build"},
			want:       []string{"build", "deploy", "lint"},
		},
		{
			name:       "metadata overrides manual variable",
			dependency: "${CI_REPO_NAME}",
			variables:  map[string]string{"CI_REPO_NAME": "ignored"},
			want:       []string{"deploy", "lint"},
		},
		{
			name:       "missing substituted dependency still fails",
			dependency: "${DEPENDENCY}",
			variables:  map[string]string{"DEPENDENCY": "missing"},
			wantError:  true,
		},
		{
			name:       "matrix axes override manual variable",
			dependency: "${DEPENDENCY}",
			matrix:     "matrix:\n  DEPENDENCY: [lint, build]\n",
			variables:  map[string]string{"DEPENDENCY": "ignored"},
			want:       []string{"build", "deploy", "deploy", "lint"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &testMetadata{pipelineEvent: "manual", repo: "lint"}
			b := PipelineBuilder{
				GetWorkflowMetadata: m.GetWorkflowMetadata,
				RepoTrusted:         &metadata.TrustedConfiguration{},
				Envs:                tt.envs,
				AdditionalEnvs:      tt.variables,
				SelectedWorkflow:    "deploy",
				Yamls: []*YamlFile{
					{Name: "lint", Data: []byte("steps:\n  - name: lint\n    image: scratch\n")},
					{Name: "build", Data: []byte("depends_on: [\"${PARENT:-lint}\"]\nsteps:\n  - name: build\n    image: scratch\n")},
					{Name: "deploy", Data: []byte(tt.matrix + "depends_on:\n  - name: \"" + tt.dependency + "\"\n  - name: ignored\n    optional: true\nsteps:\n  - name: deploy\n    image: scratch\n")},
					{Name: "ignored", Data: []byte("when:\n  evaluate: 'int(DEPLOY_COUNT) > 0'\nsteps:\n  - name: ignored\n    image: scratch\n")},
				},
			}

			items, err := b.Build()
			if tt.wantError {
				assert.True(t, errors.HasBlockingErrors(err))
				assert.ErrorContains(t, err, "misses a dependency")
				return
			}
			if errors.HasBlockingErrors(err) {
				t.Fatalf("selected workflow must retain its substituted hard dependencies: %v", err)
			}
			names := make([]string, 0, len(items))
			for _, item := range items {
				names = append(names, item.Workflow.Name)
			}
			assert.Equal(t, tt.want, names)
		})
	}
}

func TestSelectedWorkflowInvalidConfig(t *testing.T) {
	t.Parallel()

	for _, config := range []string{"steps: [", "skip_clone: invalid\nsteps:\n  - name: deploy\n    image: scratch\n"} {
		t.Run(config, func(t *testing.T) {
			t.Parallel()

			m := &testMetadata{pipelineEvent: "manual"}
			b := PipelineBuilder{
				GetWorkflowMetadata: m.GetWorkflowMetadata,
				RepoTrusted:         &metadata.TrustedConfiguration{},
				SelectedWorkflow:    "deploy",
				Yamls:               []*YamlFile{{Name: "deploy", Data: []byte(config)}},
			}
			items, err := b.Build()
			assert.True(t, errors.HasBlockingErrors(err))
			assert.Nil(t, items)
		})
	}
}
