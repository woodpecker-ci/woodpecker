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
	"slices"
	"testing"

	"github.com/franela/goblin"
	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
	yaml_base_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/types/base"
)

func TestParse(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Parser", func() {
		g.Describe("Given a yaml file", func() {
			g.It("Should unmarshal a string", func() {
				out, err := ParseString(sampleYaml)
				if err != nil {
					g.Fail(err)
				}

				g.Assert(slices.Contains(out.When.Constraints[0].Event, "tester")).Equal(true)

				g.Assert(out.Workspace.Base).Equal("/go")
				g.Assert(out.Workspace.Path).Equal("src/github.com/octocat/hello-world")
				g.Assert(out.Volumes.WorkflowVolumes[0].Name).Equal("custom")
				g.Assert(out.Volumes.WorkflowVolumes[0].Driver).Equal("blockbridge")
				g.Assert(out.Networks.WorkflowNetworks[0].Name).Equal("custom")
				g.Assert(out.Networks.WorkflowNetworks[0].Driver).Equal("overlay")
				g.Assert(out.Services.ContainerList[0].Name).Equal("database")
				g.Assert(out.Services.ContainerList[0].Image).Equal("mysql")
				g.Assert(out.Steps.ContainerList[0].Name).Equal("test")
				g.Assert(out.Steps.ContainerList[0].Image).Equal("golang")
				g.Assert(out.Steps.ContainerList[0].Commands).Equal(yaml_base_types.StringOrSlice{"go install", "go test"})
				g.Assert(out.Steps.ContainerList[1].Name).Equal("build")
				g.Assert(out.Steps.ContainerList[1].Image).Equal("golang")
				g.Assert(out.Steps.ContainerList[1].Commands).Equal(yaml_base_types.StringOrSlice{"go build"})
				g.Assert(out.Steps.ContainerList[2].Name).Equal("notify")
				g.Assert(out.Steps.ContainerList[2].Image).Equal("slack")
				// g.Assert(out.Steps.ContainerList[2].NetworkMode).Equal("container:name")
				g.Assert(out.Labels["com.example.team"]).Equal("frontend")
				g.Assert(out.Labels["com.example.type"]).Equal("build")
				g.Assert(out.DependsOn[0]).Equal("lint")
				g.Assert(out.DependsOn[1]).Equal("test")
				g.Assert(out.RunsOn[0]).Equal("success")
				g.Assert(out.RunsOn[1]).Equal("failure")
				g.Assert(out.SkipClone).Equal(false)
			})

			g.It("Should handle simple yaml anchors", func() {
				out, err := ParseString(simpleYamlAnchors)
				if err != nil {
					g.Fail(err)
				}
				g.Assert(out.Steps.ContainerList[0].Name).Equal("notify_success")
				g.Assert(out.Steps.ContainerList[0].Image).Equal("plugins/slack")
			})

			g.It("Should unmarshal variables", func() {
				out, err := ParseString(sampleVarYaml)
				if err != nil {
					g.Fail(err)
				}
				g.Assert(out.Steps.ContainerList[0].Name).Equal("notify_fail")
				g.Assert(out.Steps.ContainerList[0].Image).Equal("plugins/slack")
				g.Assert(out.Steps.ContainerList[1].Name).Equal("notify_success")
				g.Assert(out.Steps.ContainerList[1].Image).Equal("plugins/slack")

				g.Assert(len(out.Steps.ContainerList[0].When.Constraints)).Equal(0)
				g.Assert(out.Steps.ContainerList[1].Name).Equal("notify_success")
				g.Assert(out.Steps.ContainerList[1].Image).Equal("plugins/slack")
				g.Assert(out.Steps.ContainerList[1].When.Constraints[0].Event).Equal(yaml_base_types.StringOrSlice{"success"})
			})

			matchConfig, err := ParseString(sampleYaml)
			if err != nil {
				g.Fail(err)
			}

			g.It("Should match event tester", func() {
				match, err := matchConfig.When.Match(metadata.Metadata{
					Curr: metadata.Pipeline{
						Event: "tester",
					},
				}, false, nil)
				g.Assert(match).Equal(true)
				g.Assert(err).IsNil()
			})

			g.It("Should match event tester2", func() {
				match, err := matchConfig.When.Match(metadata.Metadata{
					Curr: metadata.Pipeline{
						Event: "tester2",
					},
				}, false, nil)
				g.Assert(match).Equal(true)
				g.Assert(err).IsNil()
			})

			g.It("Should match branch tester", func() {
				match, err := matchConfig.When.Match(metadata.Metadata{
					Curr: metadata.Pipeline{
						Commit: metadata.Commit{
							Branch: "tester",
						},
					},
				}, true, nil)
				g.Assert(match).Equal(true)
				g.Assert(err).IsNil()
			})

			g.It("Should not match event push", func() {
				match, err := matchConfig.When.Match(metadata.Metadata{
					Curr: metadata.Pipeline{
						Event: "push",
					},
				}, false, nil)
				g.Assert(match).Equal(false)
				g.Assert(err).IsNil()
			})
		})
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
	g := goblin.Goblin(t)

	g.Describe("Parser", func() {
		g.It("should marshal a not set slice to nil", func() {
			out, err := ParseString(sampleSliceYaml)
			if err != nil {
				g.Fail(err)
			}

			g.Assert(out.Steps.ContainerList[0].DependsOn).IsNil()
			g.Assert(len(out.Steps.ContainerList[0].DependsOn)).Equal(0)
		})

		g.It("should marshal an empty slice", func() {
			out, err := ParseString(sampleSliceYaml)
			if err != nil {
				g.Fail(err)
			}

			g.Assert(out.Steps.ContainerList[1].DependsOn).IsNotNil()
			g.Assert(len(out.Steps.ContainerList[1].DependsOn)).Equal(0)
		})
	})
}
