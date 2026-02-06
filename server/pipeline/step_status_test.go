// Copyright 2022 Woodpecker Authors
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func mockStoreStep(t *testing.T) store.Store {
	s := mocks.NewMockStore(t)
	s.On("StepUpdate", mock.Anything).Return(nil)
	return s
}

func TestUpdateStepStatus(t *testing.T) {
	t.Parallel()

	t.Run("Pending", func(t *testing.T) {
		t.Parallel()

		t.Run("TransitionToRunning", func(t *testing.T) {
			t.Parallel()

			t.Run("WithStartTime", func(t *testing.T) {
				t.Parallel()
				step := &model.Step{State: model.StatusPending}
				state := rpc.StepState{Started: 42, Finished: 0}

				err := UpdateStepStatus(mockStoreStep(t), step, state)

				assert.NoError(t, err)
				assert.Equal(t, model.StatusRunning, step.State)
				assert.Equal(t, int64(42), step.Started)
				assert.Equal(t, int64(0), step.Finished)
			})

			t.Run("WithoutStartTime", func(t *testing.T) {
				t.Parallel()
				step := &model.Step{State: model.StatusPending}
				state := rpc.StepState{Started: 0, Finished: 0}

				err := UpdateStepStatus(mockStoreStep(t), step, state)

				assert.NoError(t, err)
				assert.Equal(t, model.StatusRunning, step.State)
				assert.Greater(t, step.Started, int64(0))
			})
		})

		t.Run("DirectToSuccess", func(t *testing.T) {
			t.Parallel()

			t.Run("WithFinishTime", func(t *testing.T) {
				t.Parallel()
				step := &model.Step{State: model.StatusPending}
				state := rpc.StepState{Started: 42, Exited: true, Finished: 100, ExitCode: 0, Error: ""}

				err := UpdateStepStatus(mockStoreStep(t), step, state)

				assert.NoError(t, err)
				assert.Equal(t, model.StatusSuccess, step.State)
				assert.Equal(t, int64(42), step.Started)
				assert.Equal(t, int64(100), step.Finished)
			})

			t.Run("WithoutFinishTime", func(t *testing.T) {
				t.Parallel()
				step := &model.Step{State: model.StatusPending}
				state := rpc.StepState{Started: 42, Exited: true, Finished: 0, ExitCode: 0, Error: ""}

				err := UpdateStepStatus(mockStoreStep(t), step, state)

				assert.NoError(t, err)
				assert.Equal(t, model.StatusSuccess, step.State)
				assert.Greater(t, step.Finished, int64(0))
			})
		})

		t.Run("DirectToFailure", func(t *testing.T) {
			t.Parallel()

			t.Run("WithExitCode", func(t *testing.T) {
				t.Parallel()
				step := &model.Step{State: model.StatusPending}
				state := rpc.StepState{Started: 42, Exited: true, Finished: 34, ExitCode: 1, Error: "an error"}

				err := UpdateStepStatus(mockStoreStep(t), step, state)

				assert.NoError(t, err)
				assert.Equal(t, model.StatusFailure, step.State)
				assert.Equal(t, 1, step.ExitCode)
				assert.Equal(t, "an error", step.Error)
			})
		})
	})

	t.Run("Running", func(t *testing.T) {
		t.Parallel()

		t.Run("ToSuccess", func(t *testing.T) {
			t.Parallel()

			t.Run("WithFinishTime", func(t *testing.T) {
				t.Parallel()
				step := &model.Step{State: model.StatusRunning, Started: 42}
				state := rpc.StepState{Exited: true, Finished: 100, ExitCode: 0, Error: ""}

				err := UpdateStepStatus(mockStoreStep(t), step, state)

				assert.NoError(t, err)
				assert.Equal(t, model.StatusSuccess, step.State)
				assert.Equal(t, int64(100), step.Finished)
			})

			t.Run("WithoutFinishTime", func(t *testing.T) {
				t.Parallel()
				step := &model.Step{State: model.StatusRunning, Started: 42}
				state := rpc.StepState{Exited: true, Finished: 0, ExitCode: 0, Error: ""}

				err := UpdateStepStatus(mockStoreStep(t), step, state)

				assert.NoError(t, err)
				assert.Equal(t, model.StatusSuccess, step.State)
				assert.Greater(t, step.Finished, int64(0))
			})
		})

		t.Run("ToFailure", func(t *testing.T) {
			t.Parallel()

			t.Run("WithExitCode137", func(t *testing.T) {
				t.Parallel()
				step := &model.Step{State: model.StatusRunning, Started: 42}
				state := rpc.StepState{Exited: true, Finished: 34, ExitCode: pipeline.ExitCodeKilled, Error: "an error"}

				err := UpdateStepStatus(mockStoreStep(t), step, state)

				assert.NoError(t, err)
				assert.Equal(t, model.StatusFailure, step.State)
				assert.Equal(t, int64(34), step.Finished)
				assert.Equal(t, pipeline.ExitCodeKilled, step.ExitCode)
			})

			t.Run("WithError", func(t *testing.T) {
				t.Parallel()
				step := &model.Step{State: model.StatusRunning, Started: 42}
				state := rpc.StepState{Exited: true, Finished: 34, ExitCode: 0, Error: "an error"}

				err := UpdateStepStatus(mockStoreStep(t), step, state)

				assert.NoError(t, err)
				assert.Equal(t, model.StatusFailure, step.State)
				assert.Equal(t, "an error", step.Error)
			})
		})

		t.Run("StillRunning", func(t *testing.T) {
			t.Parallel()
			step := &model.Step{State: model.StatusRunning, Started: 42}
			state := rpc.StepState{Exited: false, Finished: 0}

			err := UpdateStepStatus(mockStoreStep(t), step, state)

			assert.NoError(t, err)
			assert.Equal(t, model.StatusRunning, step.State)
			assert.Equal(t, int64(0), step.Finished)
		})
	})

	t.Run("Canceled", func(t *testing.T) {
		t.Parallel()

		t.Run("WithoutFinishTime", func(t *testing.T) {
			t.Parallel()
			step := &model.Step{State: model.StatusRunning, Started: 42}
			state := rpc.StepState{Canceled: true}

			err := UpdateStepStatus(mockStoreStep(t), step, state)

			assert.NoError(t, err)
			assert.Equal(t, model.StatusKilled, step.State)
			assert.Greater(t, step.Finished, int64(0))
		})

		t.Run("WithExitedAndFinishTime", func(t *testing.T) {
			t.Parallel()
			step := &model.Step{State: model.StatusRunning, Started: 42}
			state := rpc.StepState{Canceled: true, Exited: true, Finished: 100, ExitCode: 1, Error: "canceled"}

			err := UpdateStepStatus(mockStoreStep(t), step, state)

			assert.NoError(t, err)
			assert.Equal(t, model.StatusKilled, step.State)
			assert.Equal(t, int64(100), step.Finished)
			assert.Equal(t, 1, step.ExitCode)
			assert.Equal(t, "canceled", step.Error)
		})
	})

	t.Run("TerminalState", func(t *testing.T) {
		t.Parallel()
		step := &model.Step{State: model.StatusKilled, Started: 42, Finished: 64}
		state := rpc.StepState{Exited: false}

		err := UpdateStepStatus(mocks.NewMockStore(t), step, state)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not expect rpc state updates")
		assert.Equal(t, model.StatusKilled, step.State)
	})
}

func TestUpdateStepToStatusSkipped(t *testing.T) {
	t.Parallel()

	t.Run("NotStarted", func(t *testing.T) {
		t.Parallel()

		step, err := UpdateStepToStatusSkipped(mockStoreStep(t), model.Step{}, int64(1))

		assert.NoError(t, err)
		assert.Equal(t, model.StatusSkipped, step.State)
		assert.Equal(t, int64(0), step.Finished)
	})

	t.Run("AlreadyStarted", func(t *testing.T) {
		t.Parallel()

		step, err := UpdateStepToStatusSkipped(mockStoreStep(t), model.Step{Started: 42}, int64(100))

		assert.NoError(t, err)
		assert.Equal(t, model.StatusSuccess, step.State)
		assert.Equal(t, int64(100), step.Finished)
	})
}
