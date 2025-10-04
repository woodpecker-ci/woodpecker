package pipeline

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/stepbuilder"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestSetPipelineStepsOnPipeline(t *testing.T) {
	t.Parallel()

	pipeline := &model.Pipeline{
		ID:    1,
		Event: model.EventPush,
	}

	workflow := &model.Workflow{
		ID: 1,
	}

	pipelineItems := []*stepbuilder.Item{{
		Workflow: &stepbuilder.Workflow{
			ID:  1,
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

	s := mocks.NewStore(t)
	s.On("WorkflowLoad", mock.Anything).Return(workflow, nil)

	pipeline = applyWorkflowsFromStepBuilder(s, pipeline, pipelineItems)
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
