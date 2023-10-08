package pipeline

import (
	"testing"

	sharedPipeline "github.com/woodpecker-ci/woodpecker/pipeline"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	forge_types "github.com/woodpecker-ci/woodpecker/server/forge/types"
	"github.com/woodpecker-ci/woodpecker/server/model"
	mocks_store "github.com/woodpecker-ci/woodpecker/server/store/mocks"
)

func TestSetPipelineStepsOnPipeline(t *testing.T) {
	t.Parallel()

	pipeline := &model.Pipeline{
		ID:    1,
		Event: model.EventPush,
	}

	pipelineItems := []*sharedPipeline.Item{{
		Workflow: &model.Workflow{
			PID: 1,
		},
		Config: &types.Config{
			Stages: []*types.Stage{
				{
					Steps: []*types.Step{
						{
							Name: "clone",
						},
					},
				},
				{
					Steps: []*types.Step{
						{
							Name: "step",
						},
					},
				},
			},
		},
	}}
	pipeline = setPipelineStepsOnPipeline(pipeline, pipelineItems)
	if len(pipeline.Workflows) != 1 {
		t.Fatal("Should generate three in total")
	}
	if pipeline.Workflows[0].PipelineID != 1 {
		t.Fatal("Should set workflow's pipeline ID")
	}
	if pipeline.Workflows[0].Children[0].PPID != 1 {
		t.Fatal("Should set step PPID")
	}
}

func TestParsePipeline(t *testing.T) {
	store := mocks_store.NewStore(t)

	pipeline := &model.Pipeline{}
	user := &model.User{}
	repo := &model.Repo{}
	yamls := []*forge_types.FileMeta{
		{
			Name: ".woodpecker.yml",
			Data: []byte(sampleYaml),
		},
	}

	items, err := parsePipeline(store, pipeline, user, repo, yamls, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 1 {
		t.Fatal("Should generate three in total")
	}
}

var sampleYaml = `
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
  deploy-preview:
    image: woodpeckerci/plugin-surge-preview:next
    settings:
      path: "docs/build/"
      surge_token:
        from_secret: SURGE_TOKEN
      forge_type: github
      forge_url: "https://github.com"
      forge_repo_token:
        from_secret: GITHUB_TOKEN_SURGE

services:
  database:
    image: mysql

networks:
  custom:
    driver: overlay

volumes:
  custom:
    driver: blockbridge

depends_on:
  - lint
  - test
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
