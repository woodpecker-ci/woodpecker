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
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"

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
		assert.Equal(t, "lint", out.DependsOn[0].Name)
		assert.Equal(t, "test", out.DependsOn[1].Name)
		assert.EqualValues(t, []string{"success", "failure"}, out.When.Constraints[0].Status)
		assert.False(t, out.SkipClone)
	})

	t.Run("Should fail on invalid yaml", func(t *testing.T) {
		_, err := ParseString("notvalid")
		assert.Error(t, err)
	})

	t.Run("Should unmarshal concurrency object", func(t *testing.T) {
		out, err := ParseString(`steps:
  deploy:
    image: alpine
concurrency:
  limit: 2
  group: deploy
`)
		assert.NoError(t, err)
		assert.Equal(t, 2, out.Concurrency.Limit)
		assert.Equal(t, "deploy", out.Concurrency.Group)
	})

	t.Run("Should unmarshal concurrency shorthand", func(t *testing.T) {
		out, err := ParseString(`steps:
  deploy:
    image: alpine
concurrency: 1
`)
		assert.NoError(t, err)
		assert.Equal(t, 1, out.Concurrency.Limit)
		assert.Empty(t, out.Concurrency.Group)
	})

	t.Run("Should default concurrency to disabled", func(t *testing.T) {
		out, err := ParseString(`steps:
  deploy:
    image: alpine
`)
		assert.NoError(t, err)
		assert.True(t, out.Concurrency.IsZero())
		assert.Equal(t, 0, out.Concurrency.Limit)
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
		assert.Equal(t, yaml_base_types.StringOrSlice{"push"}, out.Steps.ContainerList[1].When.Constraints[0].Event)
	})

	t.Run("Should handle deeply nested yaml", func(t *testing.T) {
		_, err := ParseString(sampleDeepYaml)
		assert.NoError(t, err)
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
	require.NoError(t, err)

	workflow2, err := ParseString(sampleYamlPipelineLegacyIgnore)
	require.NoError(t, err)

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
    status: [ success, failure ]
  - branch:
    - tester
    status: [ success, failure ]
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
    commands:
      - go build
    when:
      event: push
    depends_on: []
  notify:
    image: slack
    settings:
      channel: dev
    when:
      event: failure
services:
  database:
    image: mysql
labels:
  com.example.type: "build"
  com.example.team: "frontend"
depends_on:
  - lint
  - test
concurrency: 1
`

var simpleYamlAnchors = `
vars:
  image: &image plugins/slack
steps:
  notify_success:
    image: *image
`

var sampleVarYaml = `
variables: &SLACK
  image: plugins/slack
steps:
  notify_fail: *SLACK
  notify_success:
    << : *SLACK
    when:
      event: push
  echo:
    when:
    - path: wow.sh
      repo: "test"
      branch:
        exclude: main
    - path:
      - test.yaml
      - test.zig
    - path:
        exclude: a
        on_empty: true
    - ref: ref/tags/v1
      path:
  env:
    image: print
    environment:
      DRIVER: next
      PLATFORM: linux
concurrency:
  limit: 1
  group: test
`

var sampleDeepYaml = `
image: hello-world
when:
  - branch:
    - tester
steps:
  test:
    image: golang
    commands:
      - go install
      - go test
    backend_options:
      kubernetes:
        affinity:
          nodeAffinity:
            requiredDuringSchedulingIgnoredDuringExecution:
              nodeSelectorTerms:
                - matchExpressions:
                    - key: accelerator
                      operator: In
                      values:
                        - nvidia-tesla-v100
`

func TestReSerialize(t *testing.T) {
	work1, err := ParseString(sampleVarYaml)
	require.NoError(t, err)

	work1Bin, err := yaml.Marshal(work1)
	require.NoError(t, err)

	assert.EqualValues(t, `steps:
    - name: notify_fail
      image: plugins/slack
    - name: notify_success
      image: plugins/slack
      when:
        event: push
    - name: echo
      when:
        - repo: test
          branch:
            exclude: main
          path: wow.sh
        - path:
            - test.yaml
            - test.zig
        - path:
            exclude: a
        - ref: ref/tags/v1
    - name: env
      image: print
      environment:
        DRIVER: next
        PLATFORM: linux
concurrency:
    limit: 1
    group: test
`, string(work1Bin))

	work2, err := ParseString(sampleYaml)
	require.NoError(t, err)

	workBin2, err := yaml.Marshal(work2)
	require.NoError(t, err)

	// `depends_on: []` on the build step round-trips intact; an empty
	// list keeps the step in DAG mode rather than silently dropping back
	// to sequential.
	assert.EqualValues(t, `when:
    - status:
        - success
        - failure
      event:
        - tester
        - tester2
    - branch: tester
      status:
        - success
        - failure
workspace:
    base: /go
    path: src/github.com/octocat/hello-world
steps:
    - name: test
      image: golang
      commands:
        - go install
        - go test
    - name: build
      image: golang
      commands: go build
      depends_on: []
      when:
        event: push
    - name: notify
      image: slack
      settings:
        channel: dev
      when:
        event: failure
services:
    - name: database
      image: mysql
labels:
    com.example.team: frontend
    com.example.type: build
depends_on:
    - lint
    - test
concurrency: 1
`, string(workBin2))
}

func TestSlice(t *testing.T) {
	out, err := ParseString(sampleYaml)
	require.NoError(t, err)

	t.Run("should marshal a not set slice to nil", func(t *testing.T) {
		assert.Equal(t, "test", out.Steps.ContainerList[0].Name)
		assert.Nil(t, out.Steps.ContainerList[0].DependsOn)
		assert.Empty(t, out.Steps.ContainerList[0].DependsOn)
	})

	t.Run("should marshal an empty slice", func(t *testing.T) {
		assert.Equal(t, "build", out.Steps.ContainerList[1].Name)
		assert.NotNil(t, out.Steps.ContainerList[1].DependsOn)
		assert.Empty(t, out.Steps.ContainerList[1].DependsOn)
	})
}
