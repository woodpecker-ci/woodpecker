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
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	server_metadata "go.woodpecker-ci.org/woodpecker/v3/server/pipeline/metadata"
)

func getMockMetadata(t *testing.T) func(*model.Workflow) metadata.Metadata {
	repo := &model.Repo{}
	curr := &model.Pipeline{
		Message: "aaa",
		Event:   model.EventPush,
	}
	last := &model.Pipeline{}
	host := ""
	return server_metadata.MetadataFromStruct(getMockForge(t), repo, curr, last, host)
}

func TestGlobalEnvsubst(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Envs: map[string]string{
			"KEY_K": "VALUE_V",
			"IMAGE": "scratch",
		},
		WorkflowMetadataFunc: getMockMetadata(t),
		Host:                 "",
		RepoTrusted: &metadata.TrustedConfiguration{
			Network:  false,
			Volumes:  false,
			Security: false,
		},
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push
steps:
  build:
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
		Envs: map[string]string{
			"KEY_K":    "VALUE_V",
			"NO_IMAGE": "scratch",
		},
		WorkflowMetadataFunc: getMockMetadata(t),
		Host:                 "",
		RepoTrusted: &metadata.TrustedConfiguration{
			Network:  false,
			Volumes:  false,
			Security: false,
		},
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push
steps:
  build:
    image: ${IMAGE}
    settings:
      yyy: ${CI_COMMIT_MESSAGE}
`)},
		},
	}

	_, err := b.Build()
	assert.ErrorContains(t, err, "Invalid or missing image")
}

func TestMultilineEnvsubst(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Host:                 "",
		WorkflowMetadataFunc: getMockMetadata(t),
		RepoTrusted: &metadata.TrustedConfiguration{
			Network:  false,
			Volumes:  false,
			Security: false,
		},
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push
steps:
  xxx:
    image: scratch
    settings:
      yyy: ${CI_COMMIT_MESSAGE}
`)},
			{Data: []byte(`
when:
  event: push
steps:
  build:
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
		Host:                 "",
		WorkflowMetadataFunc: getMockMetadata(t),
		RepoTrusted: &metadata.TrustedConfiguration{
			Network:  false,
			Volumes:  false,
			Security: false,
		},
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push
steps:
  xxx:
    image: scratch
`)},
			{Data: []byte(`
when:
  event: push
steps:
  build:
    image: scratch
`)},
		},
	}

	pipelineItems, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(pipelineItems) != 2 {
		t.Fatal("Should have generated 2 pipelineItems")
	}
}

func TestDependsOn(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Host:                 "",
		WorkflowMetadataFunc: getMockMetadata(t),
		RepoTrusted: &metadata.TrustedConfiguration{
			Network:  false,
			Volumes:  false,
			Security: false,
		},
		Yamls: []*forge_types.FileMeta{
			{Name: "lint", Data: []byte(`
when:
  event: push
steps:
  build:
    image: scratch
`)},
			{Name: "test", Data: []byte(`
when:
  event: push
steps:
  build:
    image: scratch
`)},
			{Data: []byte(`
when:
  event: push
steps:
  deploy:
    image: scratch

depends_on:
  - lint
  - test
`)},
		},
	}

	pipelineItems, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(pipelineItems[0].DependsOn) != 2 {
		t.Fatal("Should have 3 dependencies")
	}
	if pipelineItems[0].DependsOn[1] != "test" {
		t.Fatal("Should depend on test")
	}
}

func TestRunsOn(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Host:                 "",
		WorkflowMetadataFunc: getMockMetadata(t),
		RepoTrusted: &metadata.TrustedConfiguration{
			Network:  false,
			Volumes:  false,
			Security: false,
		},
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push
steps:
  deploy:
    image: scratch

runs_on:
  - success
  - failure
`)},
		},
	}

	pipelineItems, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(pipelineItems[0].RunsOn) != 2 {
		t.Fatal("Should run on success and failure")
	}
	if pipelineItems[0].RunsOn[1] != "failure" {
		t.Fatal("Should run on failure")
	}
}

func TestPipelineName(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Host:                 "",
		WorkflowMetadataFunc: getMockMetadata(t),
		RepoTrusted: &metadata.TrustedConfiguration{
			Network:  false,
			Volumes:  false,
			Security: false,
		},
		Yamls: []*forge_types.FileMeta{
			{Name: ".woodpecker/lint.yml", Data: []byte(`
when:
  event: push
steps:
  build:
    image: scratch
`)},
			{Name: ".woodpecker/.test.yml", Data: []byte(`
when:
  event: push
steps:
  build:
    image: scratch
`)},
		},
	}

	pipelineItems, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	pipelineNames := []string{pipelineItems[0].Workflow.Name, pipelineItems[1].Workflow.Name}
	if !containsItemWithName("lint", pipelineItems) || !containsItemWithName("test", pipelineItems) {
		t.Fatalf("Pipeline name should be 'lint' and 'test' but are '%v'", pipelineNames)
	}
}

func TestBranchFilter(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Host:                 "",
		WorkflowMetadataFunc: getMockMetadata(t),
		RepoTrusted: &metadata.TrustedConfiguration{
			Network:  false,
			Volumes:  false,
			Security: false,
		},
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push
  branch: main
steps:
  xxx:
    image: scratch
`)},
			{Data: []byte(`
when:
  event: push
steps:
  build:
    image: scratch
`)},
		},
	}

	pipelineItems, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Len(t, pipelineItems, 1) {
		t.Fatal("Should have generated 1 pipeline")
	}
	if pipelineItems[0].Workflow.State != model.StatusPending {
		t.Fatal("Should run on dev branch")
	}
}

func TestRootWhenFilter(t *testing.T) {
	t.Parallel()

	b := StepBuilder{
		Host:                 "",
		WorkflowMetadataFunc: getMockMetadata(t),
		RepoTrusted: &metadata.TrustedConfiguration{
			Network:  false,
			Volumes:  false,
			Security: false,
		},
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event:
    - tag
steps:
  xxx:
    image: scratch
`)},
			{Data: []byte(`
when:
  event:
    - push
steps:
  xxx:
    image: scratch
`)},
			{Data: []byte(`
steps:
  build:
    image: scratch
`)},
		},
	}

	pipelineItems, err := b.Build()
	assert.False(t, errors.HasBlockingErrors(err))

	if len(pipelineItems) != 2 {
		t.Fatal("Should have generated 2 pipelineItems")
	}
}

func TestZeroSteps(t *testing.T) {
	t.Parallel()

	// pipeline := &model.Pipeline{
	// 	Branch: "dev",
	// 	Event:  model.EventPush,
	// }

	b := StepBuilder{
		Host:                 "",
		WorkflowMetadataFunc: getMockMetadata(t),
		RepoTrusted: &metadata.TrustedConfiguration{
			Network:  false,
			Volumes:  false,
			Security: false,
		},
		Yamls: []*forge_types.FileMeta{
			{Data: []byte(`
when:
  event: push
skip_clone: true
steps:
  build:
    when:
      branch: notdev
    image: scratch
`)},
		},
	}

	pipelineItems, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(pipelineItems) != 0 {
		t.Fatal("Should not generate a pipeline item if there are no steps")
	}
}

func TestZeroStepsAsMultiPipelineDeps(t *testing.T) {
	t.Parallel()

	// pipeline := &model.Pipeline{
	// 	Branch: "dev",
	// 	Event:  model.EventPush,
	// }

	b := StepBuilder{
		Host:                 "",
		WorkflowMetadataFunc: getMockMetadata(t),
		RepoTrusted: &metadata.TrustedConfiguration{
			Network:  false,
			Volumes:  false,
			Security: false,
		},
		Yamls: []*forge_types.FileMeta{
			{Name: "zerostep", Data: []byte(`
when:
  event: push
skip_clone: true
steps:
  build:
    when:
      branch: notdev
    image: scratch
`)},
			{Name: "justastep", Data: []byte(`
when:
  event: push
steps:
  build:
    image: scratch
`)},
			{Name: "shouldbefiltered", Data: []byte(`
when:
  event: push
steps:
  build:
    image: scratch
depends_on: [ zerostep ]
`)},
		},
	}

	pipelineItems, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(pipelineItems) != 1 {
		t.Fatal("Zerostep and the step that depends on it should not generate a pipeline item")
	}
	if pipelineItems[0].Workflow.Name != "justastep" {
		t.Fatal("justastep should have been generated")
	}
}

func TestZeroStepsAsMultiPipelineTransitiveDeps(t *testing.T) {
	t.Parallel()

	// pipeline := &model.Pipeline{
	// 	Branch: "dev",
	// 	Event:  model.EventPush,
	// }

	b := StepBuilder{
		Host:                 "",
		WorkflowMetadataFunc: getMockMetadata(t),
		RepoTrusted: &metadata.TrustedConfiguration{
			Network:  false,
			Volumes:  false,
			Security: false,
		},
		Yamls: []*forge_types.FileMeta{
			{Name: "zerostep", Data: []byte(`
when:
  event: push
skip_clone: true
steps:
  build:
    when:
      branch: notdev
    image: scratch
`)},
			{Name: "justastep", Data: []byte(`
when:
  event: push
steps:
  build:
    image: scratch
`)},
			{Name: "shouldbefiltered", Data: []byte(`
when:
  event: push
steps:
  build:
    image: scratch
depends_on: [ zerostep ]
`)},
			{Name: "shouldbefilteredtoo", Data: []byte(`
when:
  event: push
steps:
  build:
    image: scratch
depends_on: [ shouldbefiltered ]
`)},
		},
	}

	pipelineItems, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(pipelineItems) != 1 {
		t.Fatal("Zerostep and the step that depends on it, and the one depending on it should not generate a pipeline item")
	}
	if pipelineItems[0].Workflow.Name != "justastep" {
		t.Fatal("justastep should have been generated")
	}
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
		if test.sanitizedPath != SanitizePath(test.path) {
			t.Fatal("Path hasn't been sanitized correctly")
		}
	}
}

func getMockForge(t *testing.T) forge.Forge {
	forge := mocks.NewForge(t)
	forge.On("Name").Return("mock")
	forge.On("URL").Return("https://codeberg.org")
	return forge
}
