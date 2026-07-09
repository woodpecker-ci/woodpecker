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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

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

			t.Run("IgnoresAgentStartTime", func(t *testing.T) {
				t.Parallel()
				before := time.Now().Unix()
				step := &model.Step{State: model.StatusPending}
				// The agent reports a start time, but the server must ignore it
				// and stamp its own clock so Started/Finished share one time
				// source (#6808).
				state := rpc.StepState{Started: 42, Finished: 0}

				err := UpdateStepStatus(t.Context(), mockStoreStep(t), step, state)

				assert.NoError(t, err)
				assert.Equal(t, model.StatusRunning, step.State)
				assert.NotEqual(t, int64(42), step.Started)
				assert.GreaterOrEqual(t, step.Started, before)
				assert.Equal(t, int64(0), step.Finished)
			})

			t.Run("WithoutStartTime", func(t *testing.T) {
				t.Parallel()
				step := &model.Step{State: model.StatusPending}
				state := rpc.StepState{Started: 0, Finished: 0}

				err := UpdateStepStatus(t.Context(), mockStoreStep(t), step, state)

				assert.NoError(t, err)
				assert.Equal(t, model.StatusRunning, step.State)
				assert.Greater(t, step.Started, int64(0))
			})
		})

		t.Run("DirectToSuccess", func(t *testing.T) {
			t.Parallel()

			t.Run("IgnoresAgentTimes", func(t *testing.T) {
				t.Parallel()
				before := time.Now().Unix()
				step := &model.Step{State: model.StatusPending}
				state := rpc.StepState{Started: 42, Exited: true, Finished: 100, ExitCode: 0, Error: ""}

				err := UpdateStepStatus(t.Context(), mockStoreStep(t), step, state)

				assert.NoError(t, err)
				assert.Equal(t, model.StatusSuccess, step.State)
				assert.NotEqual(t, int64(42), step.Started)
				assert.NotEqual(t, int64(100), step.Finished)
				assert.GreaterOrEqual(t, step.Started, before)
				assert.GreaterOrEqual(t, step.Finished, step.Started)
			})

			t.Run("WithoutFinishTime", func(t *testing.T) {
				t.Parallel()
				step := &model.Step{State: model.StatusPending}
				state := rpc.StepState{Started: 42, Exited: true, Finished: 0, ExitCode: 0, Error: ""}

				err := UpdateStepStatus(t.Context(), mockStoreStep(t), step, state)

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

				err := UpdateStepStatus(t.Context(), mockStoreStep(t), step, state)

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

			t.Run("UsesServerFinishTime", func(t *testing.T) {
				t.Parallel()
				before := time.Now().Unix()
				step := &model.Step{State: model.StatusRunning, Started: 42}
				// Agent-reported finish time is ignored; the server stamps its
				// own clock instead (#6808).
				state := rpc.StepState{Exited: true, Finished: 100, ExitCode: 0, Error: ""}

				err := UpdateStepStatus(t.Context(), mockStoreStep(t), step, state)

				assert.NoError(t, err)
				assert.Equal(t, model.StatusSuccess, step.State)
				assert.NotEqual(t, int64(100), step.Finished)
				assert.GreaterOrEqual(t, step.Finished, before)
			})

			t.Run("WithoutFinishTime", func(t *testing.T) {
				t.Parallel()
				step := &model.Step{State: model.StatusRunning, Started: 42}
				state := rpc.StepState{Exited: true, Finished: 0, ExitCode: 0, Error: ""}

				err := UpdateStepStatus(t.Context(), mockStoreStep(t), step, state)

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
				before := time.Now().Unix()
				state := rpc.StepState{Exited: true, Finished: 34, ExitCode: pipeline.ExitCodeKilled, Error: "an error"}

				err := UpdateStepStatus(t.Context(), mockStoreStep(t), step, state)

				assert.NoError(t, err)
				assert.Equal(t, model.StatusFailure, step.State)
				assert.NotEqual(t, int64(34), step.Finished)
				assert.GreaterOrEqual(t, step.Finished, before)
				assert.Equal(t, pipeline.ExitCodeKilled, step.ExitCode)
			})

			t.Run("WithError", func(t *testing.T) {
				t.Parallel()
				step := &model.Step{State: model.StatusRunning, Started: 42}
				state := rpc.StepState{Exited: true, Finished: 34, ExitCode: 0, Error: "an error"}

				err := UpdateStepStatus(t.Context(), mockStoreStep(t), step, state)

				assert.NoError(t, err)
				assert.Equal(t, model.StatusFailure, step.State)
				assert.Equal(t, "an error", step.Error)
			})
		})

		t.Run("StillRunning", func(t *testing.T) {
			t.Parallel()
			step := &model.Step{State: model.StatusRunning, Started: 42}
			state := rpc.StepState{Exited: false, Finished: 0}

			err := UpdateStepStatus(t.Context(), mockStoreStep(t), step, state)

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

			err := UpdateStepStatus(t.Context(), mockStoreStep(t), step, state)

			assert.NoError(t, err)
			assert.Equal(t, model.StatusKilled, step.State)
			assert.Greater(t, step.Finished, int64(0))
		})

		t.Run("WithExitedAndFinishTime", func(t *testing.T) {
			t.Parallel()
			before := time.Now().Unix()
			step := &model.Step{State: model.StatusRunning, Started: 42}
			state := rpc.StepState{Canceled: true, Exited: true, Finished: 100, ExitCode: 1, Error: "canceled"}

			err := UpdateStepStatus(t.Context(), mockStoreStep(t), step, state)

			assert.NoError(t, err)
			assert.Equal(t, model.StatusKilled, step.State)
			assert.NotEqual(t, int64(100), step.Finished)
			assert.GreaterOrEqual(t, step.Finished, before)
			assert.Equal(t, 1, step.ExitCode)
			assert.Equal(t, "canceled", step.Error)
		})
	})

	t.Run("Skipped", func(t *testing.T) {
		t.Parallel()

		// This mirrors exactly what the agent sends when executor.go detects
		// OnSuccess=false or OnFailure=false — only Skipped is set, everything
		// else is zero/false (no Started, no Finished, not Exited).
		t.Run("PendingToSkipped_AgentPayload", func(t *testing.T) {
			t.Parallel()
			step := &model.Step{State: model.StatusPending}
			// Exact payload from: traceStep(&backend.State{Skipped: true}, nil, step)
			// Started=0, Finished=0, Exited=false, Skipped=true
			state := rpc.StepState{
				Skipped:  true,
				Exited:   false,
				Finished: 0,
				Started:  0,
			}

			err := UpdateStepStatus(t.Context(), mockStoreStep(t), step, state)

			assert.NoError(t, err)
			// Must be Skipped, NOT Running (the bug: Finished==0 triggers StatusRunning first)
			assert.Equal(t, model.StatusSkipped, step.State)
			// Started must NOT be set — skipped steps never ran
			assert.Equal(t, int64(0), step.Started)
			// Finished must NOT be set — skipped steps never ran
			assert.Equal(t, int64(0), step.Finished)
		})

		t.Run("PendingToSkipped", func(t *testing.T) {
			t.Parallel()
			step := &model.Step{State: model.StatusPending}
			state := rpc.StepState{Skipped: true}

			err := UpdateStepStatus(t.Context(), mockStoreStep(t), step, state)

			assert.NoError(t, err)
			assert.Equal(t, model.StatusSkipped, step.State)
		})

		t.Run("PendingToSkippedWithFinishTime", func(t *testing.T) {
			t.Parallel()
			step := &model.Step{State: model.StatusPending}
			state := rpc.StepState{Skipped: true, Exited: true, Finished: 50}

			err := UpdateStepStatus(t.Context(), mockStoreStep(t), step, state)

			assert.NoError(t, err)
			assert.Equal(t, model.StatusSkipped, step.State)
			assert.Equal(t, int64(50), step.Finished)
		})
	})

	t.Run("TerminalState", func(t *testing.T) {
		t.Parallel()
		step := &model.Step{State: model.StatusKilled, Started: 42, Finished: 64}
		state := rpc.StepState{Exited: false}

		err := UpdateStepStatus(t.Context(), mocks.NewMockStore(t), step, state)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not expect rpc state updates")
		assert.Equal(t, model.StatusKilled, step.State)
	})
}

// TestStepStatusSkewedAgentClock reproduces #6808: a step is started (server
// records its own clock) and then finished by an agent whose clock is skewed
// far into the past. Before the fix the finish time was copied from the agent
// clock, so step.Finished < step.Started and the reported duration was
// negative. With the server-authoritative fix both timestamps come from the
// server clock, so the duration is always non-negative.
func TestStepStatusSkewedAgentClock(t *testing.T) {
	t.Parallel()

	store := mockStoreStep(t)
	step := &model.Step{State: model.StatusPending}

	// Phase 1 — "started" trace: the agent sends Started=0, so the server
	// stamps its own clock.
	require.NoError(t, UpdateStepStatus(t.Context(), store, step, rpc.StepState{Started: 0, Finished: 0}))
	require.Equal(t, model.StatusRunning, step.State)
	require.Greater(t, step.Started, int64(0))

	// Phase 2 — "finished" trace with a badly skewed agent clock (far in the
	// past relative to the server). The old behaviour would set
	// step.Finished = 1000 < step.Started -> negative duration.
	const skewedAgentFinish = int64(1000)
	require.NoError(t, UpdateStepStatus(t.Context(), store, step, rpc.StepState{Exited: true, Finished: skewedAgentFinish, ExitCode: 0}))

	assert.Equal(t, model.StatusSuccess, step.State)
	assert.NotEqual(t, skewedAgentFinish, step.Finished)
	assert.GreaterOrEqual(t, step.Finished, step.Started, "duration must never be negative")
}

func TestUpdateStepToStatusSkipped(t *testing.T) {
	t.Parallel()

	t.Run("NotStarted", func(t *testing.T) {
		t.Parallel()

		step, err := UpdateStepToStatusSkipped(mockStoreStep(t), model.Step{}, int64(1), model.StatusSkipped)

		assert.NoError(t, err)
		assert.Equal(t, model.StatusSkipped, step.State)
		assert.Equal(t, int64(0), step.Finished)
	})

	t.Run("AlreadyStarted", func(t *testing.T) {
		t.Parallel()

		step, err := UpdateStepToStatusSkipped(mockStoreStep(t), model.Step{Started: 42}, int64(100), model.StatusSkipped)

		assert.NoError(t, err)
		assert.Equal(t, model.StatusSuccess, step.State)
		assert.Equal(t, int64(100), step.Finished)
	})
}
