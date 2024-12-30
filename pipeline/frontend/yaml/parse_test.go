// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	yaml_base_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/types/base"
)

func TestParse(t *testing.T) {
	t.Run("Should unmarshal a string", func(t *testing.T) {
		out, err := ParseString(sampleYaml)
		assert.NoError(t, err)

		assert.Contains(t, out.When.Constraints[0].Event, "tester")

		assert.Equal(t, "/go", out.Workspace.Base)
		assert.Equal(t, "src/github.com/octocat/hello-world", out.Workspace.Path)
		assert.Equal(t, "custom", out.Volumes.WorkflowVolumes[0].Name)
		assert.Equal(t, "blockbridge", out.Volumes.WorkflowVolumes[0].Driver)
		assert.Equal(t, "custom", out.Networks.WorkflowNetworks[0].Name)
		assert.Equal(t, "overlay", out.Networks.WorkflowNetworks[0].Driver)
		assert.Equal(t, "database", out.Services.ContainerList[0].Name)
		assert.Equal(t, "mysql", out.Services.ContainerList[0].Image)
		assert.Equal(t, "test", out.Steps.ContainerList[0].Name)
		assert.Equal(t, "golang", out.Steps.ContainerList[0].Image)
		assert.Equal(t, yaml_base_types.StringOrSlice{"go install", "go test"}, out.Steps.ContainerList[0].Commands)
		assert.Equal(t, "build", out.Steps.ContainerList[1].Name)
		assert.Equal(t, "golang", out.Steps.ContainerList[1].Image)
		assert.Equal(t, yaml_base_types.StringOrSlice{"go build"}, out.Steps.ContainerList[1].Commands)
		assert.Equal(t, "notify", out.Steps.ContainerList[2].Name)
		assert.Equal(t, "slack", out.Steps.ContainerList[2].Image)
		assert.Equal(t, "frontend", out.Labels["com.example.team"])
		assert.Equal(t, "build", out.Labels["com.example.type"])
		assert.Equal(t, "lint", out.DependsOn[0])
		assert.Equal(t, "test", out.DependsOn[1])
		assert.Equal(t, ("success"), out.RunsOn[0])
		assert.Equal(t, ("failure"), out.RunsOn[1])
		assert.False(t, out.SkipClone)
	})

	t.Run("Should handle simple yaml anchors", func(t *testing.T) {
		out, err := ParseString(simpleYamlAnchors)
		assert.NoError(t, err)
		assert.Equal(t, "notify_success", out.Steps.ContainerList[0].Name)
		assert.Equal(t, "plugins/slack", out.Steps.ContainerList[0].Image)
	})

	t.Run("Should unmarshal variables", func(t *testing.T) {
		out, err := ParseString(sampleVarYaml)
		assert.NoError(t, err)
		assert.Equal(t, "notify_fail", out.Steps.ContainerList[0].Name)
		assert.Equal(t, "plugins/slack", out.Steps.ContainerList[0].Image)
		assert.Equal(t, "notify_success", out.Steps.ContainerList[1].Name)
		assert.Equal(t, "plugins/slack", out.Steps.ContainerList[1].Image)

		assert.Empty(t, out.Steps.ContainerList[0].When.Constraints)
		assert.Equal(t, "notify_success", out.Steps.ContainerList[1].Name)
		assert.Equal(t, "plugins/slack", out.Steps.ContainerList[1].Image)
		assert.Equal(t, yaml_base_types.StringOrSlice{"success"}, out.Steps.ContainerList[1].When.Constraints[0].Event)
	})
}

func TestMatch(t *testing.T) {
	matchConfig, err := ParseString(sampleYaml)
	assert.NoError(t, err)

	t.Run("Should match event tester", func(t *testing.T) {
		match, err := matchConfig.When.Match(metadata.Metadata{
			Curr: metadata.Pipeline{
				Event: "tester",
			},
		}, false, nil)
		assert.True(t, match)
		assert.NoError(t, err)
	})

	t.Run("Should match event tester2", func(t *testing.T) {
		match, err := matchConfig.When.Match(metadata.Metadata{
			Curr: metadata.Pipeline{
				Event: "tester2",
			},
		}, false, nil)
		assert.True(t, match)
		assert.NoError(t, err)
	})

	t.Run("Should match branch tester", func(t *testing.T) {
		match, err := matchConfig.When.Match(metadata.Metadata{
			Curr: metadata.Pipeline{
				Commit: metadata.Commit{
					Branch: "tester",
				},
			},
		}, true, nil)
		assert.True(t, match)
		assert.NoError(t, err)
	})

	t.Run("Should not match event push", func(t *testing.T) {
		match, err := matchConfig.When.Match(metadata.Metadata{
			Curr: metadata.Pipeline{
				Event: "push",
			},
		}, false, nil)
		assert.False(t, match)
		assert.NoError(t, err)
	})
}

func TestParseLegacy(t *testing.T) {
	sampleYamlPipeline := `
labels:
  platform: linux/amd64

steps:
  say hello:
    image: bash
    commands: echo hello
`

	sampleYamlPipelineLegacyIgnore := `
platform: windows/amd64
labels:
  platform: linux/amd64

steps:
  say hello:
    image: bash
    commands: echo hello

pipeline:
  old crap:
    image: bash
    commands: meh!
`

	workflow1, err := ParseString(sampleYamlPipeline)
	if !assert.NoError(t, err) {
		return
	}

	workflow2, err := ParseString(sampleYamlPipelineLegacyIgnore)
	if !assert.NoError(t, err) {
		return
	}

	assert.EqualValues(t, workflow1, workflow2)
	assert.Len(t, workflow1.Steps.ContainerList, 1)
	assert.EqualValues(t, "say hello", workflow1.Steps.ContainerList[0].Name)
}

var sampleYaml = `
image: hello-world
when:
  - event:
    - tester
    - tester2
  - branch:
    - tester
build:
  context: .
  dockerfile: Dockerfile
workspace:
  path: src/github.com/octocat/hello-world
  base: /go
steps:
  test:
    image: golang
    commands:
      - go install
      - go test
  build:
    image: golang
    network_mode: container:name
    commands:
      - go build
    when:
      event: push
  notify:
    image: slack
    channel: dev
    when:
      event: failure
services:
  database:
    image: mysql
networks:
  custom:
    driver: overlay
volumes:
  custom:
    driver: blockbridge
labels:
  com.example.type: "build"
  com.example.team: "frontend"
depends_on:
  - lint
  - test
runs_on:
  - success
  - failure
`

var simpleYamlAnchors = `
vars:
  image: &image plugins/slack
steps:
  notify_success:
    image: *image
`

var sampleVarYaml = `
_slack: &SLACK
  image: plugins/slack
steps:
  notify_fail: *SLACK
  notify_success:
    << : *SLACK
    when:
      event: success
`

var sampleSliceYaml = `
steps:
  nil_slice:
    image: plugins/slack
  empty_slice:
    image: plugins/slack
    depends_on: []
`

func TestSlice(t *testing.T) {
	t.Run("should marshal a not set slice to nil", func(t *testing.T) {
		out, err := ParseString(sampleSliceYaml)
		assert.NoError(t, err)

		assert.Nil(t, out.Steps.ContainerList[0].DependsOn)
		assert.Empty(t, out.Steps.ContainerList[0].DependsOn)
	})

	t.Run("should marshal an empty slice", func(t *testing.T) {
		out, err := ParseString(sampleSliceYaml)
		assert.NoError(t, err)

		assert.NotNil(t, out.Steps.ContainerList[1].DependsOn)
		assert.Empty(t, (out.Steps.ContainerList[1].DependsOn))
	})
}
