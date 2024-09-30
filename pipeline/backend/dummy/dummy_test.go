// Copyright 2024 Woodpecker Authors
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

package dummy_test

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/dummy"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

func TestSmalPipelineDummyRun(t *testing.T) {
	dummyEngine := dummy.New()
	ctx := context.Background()

	assert.True(t, dummyEngine.IsAvailable(ctx))
	assert.EqualValues(t, "dummy", dummyEngine.Name())
	_, err := dummyEngine.Load(ctx)
	assert.NoError(t, err)

	assert.Error(t, dummyEngine.SetupWorkflow(ctx, nil, dummy.WorkflowSetupFailUUID))

	t.Run("expect fail of step func with non setup workflow", func(t *testing.T) {
		step := &types.Step{Name: "step1", UUID: "SID_1"}
		nonExistWorkflowID := "WID_NONE"

		err := dummyEngine.StartStep(ctx, step, nonExistWorkflowID)
		assert.Error(t, err)

		_, err = dummyEngine.TailStep(ctx, step, nonExistWorkflowID)
		assert.Error(t, err)

		_, err = dummyEngine.WaitStep(ctx, step, nonExistWorkflowID)
		assert.Error(t, err)

		err = dummyEngine.DestroyStep(ctx, step, nonExistWorkflowID)
		assert.Error(t, err)
	})

	t.Run("step exec successfully", func(t *testing.T) {
		step := &types.Step{
			Name:        "step1",
			UUID:        "SID_1",
			Type:        types.StepTypeCommands,
			Environment: map[string]string{},
			Commands:    []string{"echo ja", "echo nein"},
		}
		workflowUUID := "WID_1"

		assert.NoError(t, dummyEngine.SetupWorkflow(ctx, nil, workflowUUID))

		assert.NoError(t, dummyEngine.StartStep(ctx, step, workflowUUID))

		reader, err := dummyEngine.TailStep(ctx, step, workflowUUID)
		assert.NoError(t, err)
		log, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.EqualValues(t, `StepName: step1
StepType: commands
StepUUID: SID_1
StepCommands:
------------------
echo ja
echo nein
------------------
`, string(log))

		state, err := dummyEngine.WaitStep(ctx, step, workflowUUID)
		assert.NoError(t, err)
		assert.NoError(t, state.Error)
		assert.EqualValues(t, 0, state.ExitCode)

		assert.NoError(t, dummyEngine.DestroyStep(ctx, step, workflowUUID))

		assert.NoError(t, dummyEngine.DestroyWorkflow(ctx, nil, workflowUUID))
	})

	t.Run("step exec error", func(t *testing.T) {
		step := &types.Step{
			Name:        "dummy",
			UUID:        "SID_2",
			Type:        types.StepTypePlugin,
			Environment: map[string]string{dummy.EnvKeyStepType: "plugin", dummy.EnvKeyStepExitCode: "1"},
		}
		workflowUUID := "WID_1"

		assert.NoError(t, dummyEngine.SetupWorkflow(ctx, nil, workflowUUID))

		assert.NoError(t, dummyEngine.StartStep(ctx, step, workflowUUID))

		_, err := dummyEngine.TailStep(ctx, step, workflowUUID)
		assert.NoError(t, err)

		state, err := dummyEngine.WaitStep(ctx, step, workflowUUID)
		assert.NoError(t, err)
		assert.NoError(t, state.Error)
		assert.EqualValues(t, 1, state.ExitCode)

		assert.NoError(t, dummyEngine.DestroyStep(ctx, step, workflowUUID))

		assert.NoError(t, dummyEngine.DestroyWorkflow(ctx, nil, workflowUUID))
	})

	t.Run("step tail error", func(t *testing.T) {
		step := &types.Step{
			Name:        "dummy",
			UUID:        "SID_2",
			Environment: map[string]string{dummy.EnvKeyStepTailFail: "true"},
		}
		workflowUUID := "WID_1"

		assert.NoError(t, dummyEngine.SetupWorkflow(ctx, nil, workflowUUID))

		assert.NoError(t, dummyEngine.StartStep(ctx, step, workflowUUID))

		_, err := dummyEngine.TailStep(ctx, step, workflowUUID)
		assert.Error(t, err)

		_, err = dummyEngine.WaitStep(ctx, step, workflowUUID)
		assert.NoError(t, err)

		assert.NoError(t, dummyEngine.DestroyStep(ctx, step, workflowUUID))

		assert.NoError(t, dummyEngine.DestroyWorkflow(ctx, nil, workflowUUID))
	})

	t.Run("step start fail", func(t *testing.T) {
		step := &types.Step{
			Name:        "dummy",
			UUID:        "SID_2",
			Type:        types.StepTypeService,
			Environment: map[string]string{dummy.EnvKeyStepType: "service", dummy.EnvKeyStepStartFail: "true"},
		}
		workflowUUID := "WID_1"

		assert.NoError(t, dummyEngine.SetupWorkflow(ctx, nil, workflowUUID))

		assert.Error(t, dummyEngine.StartStep(ctx, step, workflowUUID))

		_, err := dummyEngine.TailStep(ctx, step, workflowUUID)
		assert.Error(t, err)

		state, err := dummyEngine.WaitStep(ctx, step, workflowUUID)
		assert.Error(t, err)
		assert.Error(t, state.Error)
		assert.EqualValues(t, 0, state.ExitCode)

		assert.Error(t, dummyEngine.DestroyStep(ctx, step, workflowUUID))

		assert.NoError(t, dummyEngine.DestroyWorkflow(ctx, nil, workflowUUID))
	})
}
