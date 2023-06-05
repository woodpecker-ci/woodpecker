package yaml

import (
	"testing"

	"github.com/franela/goblin"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/metadata"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
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

				g.Assert(out.When.Constraints[0].Event.Match("tester")).Equal(true)

				g.Assert(out.Workspace.Base).Equal("/go")
				g.Assert(out.Workspace.Path).Equal("src/github.com/octocat/hello-world")
				g.Assert(out.Volumes.WorkflowVolumes[0].Name).Equal("custom")
				g.Assert(out.Volumes.WorkflowVolumes[0].Driver).Equal("blockbridge")
				g.Assert(out.Networks.Networks[0].Name).Equal("custom")
				g.Assert(out.Networks.Networks[0].Driver).Equal("overlay")
				g.Assert(out.Services.ContainerList[0].Name).Equal("database")
				g.Assert(out.Services.ContainerList[0].Image).Equal("mysql")
				g.Assert(out.Steps.ContainerList[0].Name).Equal("test")
				g.Assert(out.Steps.ContainerList[0].Image).Equal("golang")
				g.Assert(out.Steps.ContainerList[0].Commands).Equal(types.StringOrSlice{"go install", "go test"})
				g.Assert(out.Steps.ContainerList[1].Name).Equal("build")
				g.Assert(out.Steps.ContainerList[1].Image).Equal("golang")
				g.Assert(out.Steps.ContainerList[1].Commands).Equal(types.StringOrSlice{"go build"})
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
				g.Assert(out.Steps.ContainerList[1].When.Constraints[0].Event.Include).Equal([]string{"success"})
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
				}, false)
				g.Assert(match).Equal(true)
				g.Assert(err).IsNil()
			})

			g.It("Should match event tester2", func() {
				match, err := matchConfig.When.Match(metadata.Metadata{
					Curr: metadata.Pipeline{
						Event: "tester2",
					},
				}, false)
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
				}, true)
				g.Assert(match).Equal(true)
				g.Assert(err).IsNil()
			})

			g.It("Should not match event push", func() {
				match, err := matchConfig.When.Match(metadata.Metadata{
					Curr: metadata.Pipeline{
						Event: "push",
					},
				}, false)
				g.Assert(match).Equal(false)
				g.Assert(err).IsNil()
			})
		})
	})
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
pipeline:
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
pipeline:
  notify_success:
    image: *image
`

var sampleVarYaml = `
_slack: &SLACK
  image: plugins/slack
pipeline:
  notify_fail: *SLACK
  notify_success:
    << : *SLACK
    when:
      event: success
`
