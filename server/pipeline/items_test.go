package pipeline

import (
	"testing"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	sharedPipeline "go.woodpecker-ci.org/woodpecker/v2/server/pipeline/stepbuilder"
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
