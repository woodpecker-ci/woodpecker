// Copyright 2022 Woodpecker Authors
// Copyright 2019 mhmxs
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

package pipeline

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/mocks"
)

func mockStoreStep(t *testing.T) store.Store {
	s := mocks.NewStore(t)
	s.On("StepUpdate", mock.Anything).Return(nil)
	return s
}

func TestUpdateStepStatusNotExited(t *testing.T) {
	t.Parallel()
	// step in db before update
	step := &model.Step{}

	// advertised step status
	state := rpc.StepState{
		Started: int64(42),
		Exited:  false,
		// Dummy data
		Finished: int64(1),
		ExitCode: pipeline.ExitCodeKilled,
		Error:    "not an error",
	}

	err := UpdateStepStatus(mockStoreStep(t), step, state)
	assert.NoError(t, err)
	assert.EqualValues(t, model.StatusRunning, step.State)
	assert.EqualValues(t, 42, step.Started)
	assert.EqualValues(t, 0, step.Finished)
	assert.EqualValues(t, 0, step.ExitCode)
	assert.EqualValues(t, "", step.Error)
}

func TestUpdateStepStatusNotExitedButStopped(t *testing.T) {
	t.Parallel()

	// step in db before update
	step := &model.Step{Started: 42, Finished: 64, State: model.StatusKilled}

	// advertised step status
	state := rpc.StepState{
		Exited: false,
		// Dummy data
		Finished: int64(1),
		ExitCode: pipeline.ExitCodeKilled,
		Error:    "not an error",
	}

	err := UpdateStepStatus(mockStoreStep(t), step, state)
	assert.NoError(t, err)
	assert.EqualValues(t, model.StatusKilled, step.State)
	assert.EqualValues(t, 42, step.Started)
	assert.EqualValues(t, 64, step.Finished)
	assert.EqualValues(t, 0, step.ExitCode)
	assert.EqualValues(t, "", step.Error)
}

func TestUpdateStepStatusExited(t *testing.T) {
	t.Parallel()

	// step in db before update
	step := &model.Step{Started: 42}

	// advertised step status
	state := rpc.StepState{
		Started:  int64(42),
		Exited:   true,
		Finished: int64(34),
		ExitCode: pipeline.ExitCodeKilled,
		Error:    "an error",
	}

	err := UpdateStepStatus(mockStoreStep(t), step, state)
	assert.NoError(t, err)
	assert.EqualValues(t, model.StatusKilled, step.State)
	assert.EqualValues(t, 42, step.Started)
	assert.EqualValues(t, 34, step.Finished)
	assert.EqualValues(t, pipeline.ExitCodeKilled, step.ExitCode)
	assert.EqualValues(t, "an error", step.Error)
}

func TestUpdateStepStatusExitedButNot137(t *testing.T) {
	t.Parallel()

	// step in db before update
	step := &model.Step{Started: 42}

	// advertised step status
	state := rpc.StepState{
		Started:  int64(42),
		Exited:   true,
		Finished: int64(34),
		Error:    "an error",
	}

	err := UpdateStepStatus(mockStoreStep(t), step, state)
	assert.NoError(t, err)
	assert.EqualValues(t, model.StatusFailure, step.State)
	assert.EqualValues(t, 42, step.Started)
	assert.EqualValues(t, 34, step.Finished)
	assert.EqualValues(t, 0, step.ExitCode)
	assert.EqualValues(t, "an error", step.Error)
}

func TestUpdateStepStatusExitedWithCode(t *testing.T) {
	t.Parallel()

	// advertised step status
	state := rpc.StepState{
		Started:  int64(42),
		Exited:   true,
		Finished: int64(34),
		ExitCode: 1,
		Error:    "an error",
	}
	step := &model.Step{}
	err := UpdateStepStatus(mockStoreStep(t), step, state)
	assert.NoError(t, err)

	assert.Equal(t, model.StatusFailure, step.State)
	assert.Equal(t, 1, step.ExitCode)
}

func TestUpdateStepToStatusStarted(t *testing.T) {
	t.Parallel()

	state := rpc.StepState{Started: int64(42)}
	step, _ := UpdateStepToStatusStarted(mockStoreStep(t), model.Step{}, state)

	assert.Equal(t, model.StatusRunning, step.State)
	assert.EqualValues(t, 42, step.Started)
}

func TestUpdateStepToStatusSkipped(t *testing.T) {
	t.Parallel()

	step, _ := UpdateStepToStatusSkipped(mockStoreStep(t), model.Step{}, int64(1))

	assert.Equal(t, model.StatusSkipped, step.State)
	assert.EqualValues(t, 0, step.Finished)
}

func TestUpdateStepToStatusSkippedButStarted(t *testing.T) {
	t.Parallel()

	step := &model.Step{
		Started: int64(42),
	}

	step, _ = UpdateStepToStatusSkipped(mockStoreStep(t), *step, int64(1))

	assert.Equal(t, model.StatusSuccess, step.State)
	assert.EqualValues(t, 1, step.Finished)
}

func TestUpdateStepStatusToDoneSkipped(t *testing.T) {
	t.Parallel()

	state := rpc.StepState{
		Finished: int64(34),
	}

	step, _ := UpdateStepStatusToDone(mockStoreStep(t), model.Step{}, state)

	assert.Equal(t, model.StatusSkipped, step.State)
	assert.EqualValues(t, 34, step.Finished)
	assert.Empty(t, step.Error)
	assert.Equal(t, 0, step.ExitCode)
}

func TestUpdateStepStatusToDoneSuccess(t *testing.T) {
	t.Parallel()

	state := rpc.StepState{
		Started:  int64(42),
		Finished: int64(34),
	}

	step, _ := UpdateStepStatusToDone(mockStoreStep(t), model.Step{}, state)

	assert.Equal(t, model.StatusSuccess, step.State)
	assert.EqualValues(t, 34, step.Finished)
	assert.Empty(t, step.Error)
	assert.Equal(t, 0, step.ExitCode)
}

func TestUpdateStepStatusToDoneFailureWithError(t *testing.T) {
	t.Parallel()

	state := rpc.StepState{Error: "an error"}

	step, _ := UpdateStepStatusToDone(mockStoreStep(t), model.Step{}, state)

	assert.Equal(t, model.StatusFailure, step.State)
}

func TestUpdateStepStatusToDoneFailureWithExitCode(t *testing.T) {
	t.Parallel()

	state := rpc.StepState{ExitCode: 43}

	step, _ := UpdateStepStatusToDone(mockStoreStep(t), model.Step{}, state)

	assert.Equal(t, model.StatusFailure, step.State)
}

func TestUpdateStepToStatusKilledStarted(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	step, _ := UpdateStepToStatusKilled(mockStoreStep(t), model.Step{})

	assert.Equal(t, model.StatusKilled, step.State)
	assert.LessOrEqual(t, now, step.Finished)
	assert.Equal(t, step.Finished, step.Started)
	assert.Equal(t, 137, step.ExitCode)
}

func TestUpdateStepToStatusKilledNotStarted(t *testing.T) {
	t.Parallel()

	step, _ := UpdateStepToStatusKilled(mockStoreStep(t), model.Step{Started: int64(1)})

	assert.EqualValues(t, 1, step.Started)
}
