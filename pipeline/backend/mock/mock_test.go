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

package mock_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/mock"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

func TestSmalPipelineMockRun(t *testing.T) {
	mockEngine := mock.New()
	ctx := context.Background()

	assert.True(t, mockEngine.IsAvailable(ctx))
	assert.EqualValues(t, "mock", mockEngine.Name())
	_, err := mockEngine.Load(ctx)
	assert.NoError(t, err)

	t.Run("expect fail of step func with non setup workflow", func(t *testing.T) {
		step := &types.Step{Name: "step1", UUID: "SID_1"}
		nonExistWorkflowID := "WID_NONE"

		err := mockEngine.StartStep(ctx, step, nonExistWorkflowID)
		assert.Error(t, err)

		_, err = mockEngine.TailStep(ctx, step, nonExistWorkflowID)
		assert.Error(t, err)

		_, err = mockEngine.WaitStep(ctx, step, nonExistWorkflowID)
		assert.Error(t, err)

		err = mockEngine.DestroyStep(ctx, step, nonExistWorkflowID)
		assert.Error(t, err)
	})

	t.Run("step exec successfully", func(t *testing.T) {
		step := &types.Step{
			Name:        "step1",
			UUID:        "SID_1",
			Environment: map[string]string{},
			Commands:    []string{"echo ja", "echo nein"},
		}
		workflowUUID := "WID_1"

		assert.NoError(t, mockEngine.SetupWorkflow(ctx, nil, workflowUUID))

		assert.NoError(t, mockEngine.StartStep(ctx, step, workflowUUID))

		reader, err := mockEngine.TailStep(ctx, step, workflowUUID)
		assert.NoError(t, err)
		log, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.EqualValues(t, strings.Join(step.Commands, "\n"), string(log))

		state, err := mockEngine.WaitStep(ctx, step, workflowUUID)
		assert.NoError(t, err)
		assert.NoError(t, state.Error)
		assert.EqualValues(t, 0, state.ExitCode)

		assert.NoError(t, mockEngine.DestroyStep(ctx, step, workflowUUID))

		assert.NoError(t, mockEngine.DestroyWorkflow(ctx, nil, workflowUUID))
	})

	t.Run("step exec fail", func(t *testing.T) {
		step := &types.Step{
			Name:        mock.StepExecError,
			UUID:        "SID_2",
			Type:        types.StepTypePlugin,
			Environment: map[string]string{mock.EnvKeyStepType: "plugin"},
		}
		workflowUUID := "WID_1"

		assert.NoError(t, mockEngine.SetupWorkflow(ctx, nil, workflowUUID))

		assert.NoError(t, mockEngine.StartStep(ctx, step, workflowUUID))

		_, err := mockEngine.TailStep(ctx, step, workflowUUID)
		assert.NoError(t, err)

		state, err := mockEngine.WaitStep(ctx, step, workflowUUID)
		assert.NoError(t, err)
		assert.NoError(t, state.Error)
		assert.EqualValues(t, 1, state.ExitCode)

		assert.NoError(t, mockEngine.DestroyStep(ctx, step, workflowUUID))

		assert.NoError(t, mockEngine.DestroyWorkflow(ctx, nil, workflowUUID))
	})
}