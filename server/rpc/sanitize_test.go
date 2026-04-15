// Copyright 2026 Woodpecker Authors
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

package rpc

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestCheckWorkflowAllowsStepUpdate(t *testing.T) {
	t.Parallel()

	t.Run("workflow running allows any step update", func(t *testing.T) {
		t.Parallel()
		step := &model.Step{State: model.StatusRunning}
		// Non-terminal update (step stays running)
		assert.NoError(t, checkWorkflowAllowsStepUpdate(model.StatusRunning, step, rpc.StepState{}))
	})

	t.Run("workflow pending allows any step update", func(t *testing.T) {
		t.Parallel()
		step := &model.Step{State: model.StatusPending}
		assert.NoError(t, checkWorkflowAllowsStepUpdate(model.StatusPending, step, rpc.StepState{}))
	})

	t.Run("workflow created allows any step update", func(t *testing.T) {
		t.Parallel()
		step := &model.Step{State: model.StatusPending}
		assert.NoError(t, checkWorkflowAllowsStepUpdate(model.StatusCreated, step, rpc.StepState{}))
	})

	t.Run("workflow finished allows terminal step update", func(t *testing.T) {
		t.Parallel()
		step := &model.Step{State: model.StatusRunning}
		// Step exits with code 0 → CalcStepStatus produces StatusSuccess (terminal)
		state := rpc.StepState{Exited: true, ExitCode: 0}
		assert.NoError(t, checkWorkflowAllowsStepUpdate(model.StatusSuccess, step, state))
	})

	t.Run("workflow finished allows failed step update", func(t *testing.T) {
		t.Parallel()
		step := &model.Step{State: model.StatusRunning}
		state := rpc.StepState{Exited: true, ExitCode: 1}
		assert.NoError(t, checkWorkflowAllowsStepUpdate(model.StatusFailure, step, state))
	})

	t.Run("workflow finished allows canceled step update", func(t *testing.T) {
		t.Parallel()
		step := &model.Step{State: model.StatusRunning}
		state := rpc.StepState{Canceled: true}
		assert.NoError(t, checkWorkflowAllowsStepUpdate(model.StatusKilled, step, state))
	})

	t.Run("workflow finished allows skipped step update", func(t *testing.T) {
		t.Parallel()
		step := &model.Step{State: model.StatusPending}
		state := rpc.StepState{Skipped: true}
		assert.NoError(t, checkWorkflowAllowsStepUpdate(model.StatusSuccess, step, state))
	})

	t.Run("workflow finished rejects non-terminal step update", func(t *testing.T) {
		t.Parallel()
		step := &model.Step{State: model.StatusRunning}
		// No exit, no cancel → step stays Running (non-terminal)
		state := rpc.StepState{}
		assert.ErrorIs(t, checkWorkflowAllowsStepUpdate(model.StatusSuccess, step, state), ErrAgentIllegalWorkflowReRunStateChange)
	})

	t.Run("workflow killed rejects non-terminal step update", func(t *testing.T) {
		t.Parallel()
		step := &model.Step{State: model.StatusRunning}
		state := rpc.StepState{}
		assert.ErrorIs(t, checkWorkflowAllowsStepUpdate(model.StatusKilled, step, state), ErrAgentIllegalWorkflowReRunStateChange)
	})

	t.Run("workflow blocked rejects non-terminal step update", func(t *testing.T) {
		t.Parallel()
		step := &model.Step{State: model.StatusRunning}
		state := rpc.StepState{}
		assert.ErrorIs(t, checkWorkflowAllowsStepUpdate(model.StatusBlocked, step, state), ErrAgentIllegalWorkflowReRunStateChange)
	})

	t.Run("workflow finished rejects pending-to-running transition", func(t *testing.T) {
		t.Parallel()
		step := &model.Step{State: model.StatusPending}
		// No skip, no exit → CalcStepStatus produces Running (non-terminal)
		state := rpc.StepState{Started: 100}
		assert.ErrorIs(t, checkWorkflowAllowsStepUpdate(model.StatusSuccess, step, state), ErrAgentIllegalWorkflowReRunStateChange)
	})
}

func TestCheckWorkflowState(t *testing.T) {
	t.Parallel()

	t.Run("allowed states", func(t *testing.T) {
		t.Parallel()
		for _, s := range []model.StatusValue{
			model.StatusCreated,
			model.StatusPending,
			model.StatusRunning,
		} {
			t.Run(string(s), func(t *testing.T) {
				t.Parallel()
				assert.NoError(t, checkWorkflowState(s))
			})
		}
	})

	t.Run("blocked rejects", func(t *testing.T) {
		t.Parallel()
		assert.ErrorIs(t, checkWorkflowState(model.StatusBlocked), ErrAgentIllegalWorkflowRun)
	})

	t.Run("terminal states reject", func(t *testing.T) {
		t.Parallel()
		for _, s := range []model.StatusValue{
			model.StatusSuccess,
			model.StatusFailure,
			model.StatusKilled,
			model.StatusError,
			model.StatusSkipped,
			model.StatusCanceled,
			model.StatusDeclined,
		} {
			t.Run(string(s), func(t *testing.T) {
				t.Parallel()
				assert.ErrorIs(t, checkWorkflowState(s), ErrAgentIllegalWorkflowReRunStateChange)
			})
		}
	})
}

func TestIsActiveState(t *testing.T) {
	t.Parallel()

	active := []model.StatusValue{model.StatusCreated, model.StatusPending, model.StatusRunning}
	inactive := []model.StatusValue{
		model.StatusSuccess, model.StatusFailure, model.StatusKilled,
		model.StatusBlocked, model.StatusCanceled, model.StatusSkipped,
		model.StatusError, model.StatusDeclined,
	}

	for _, s := range active {
		t.Run(fmt.Sprintf("%s is active", s), func(t *testing.T) {
			t.Parallel()
			assert.True(t, isActiveState(s))
		})
	}
	for _, s := range inactive {
		t.Run(fmt.Sprintf("%s is not active", s), func(t *testing.T) {
			t.Parallel()
			assert.False(t, isActiveState(s))
		})
	}
}

func TestIsDoneState(t *testing.T) {
	t.Parallel()

	done := []model.StatusValue{
		model.StatusSuccess, model.StatusFailure, model.StatusKilled,
		model.StatusCanceled, model.StatusSkipped, model.StatusError,
		model.StatusDeclined,
	}
	notDone := []model.StatusValue{
		model.StatusCreated, model.StatusPending, model.StatusRunning,
		model.StatusBlocked,
	}

	for _, s := range done {
		t.Run(fmt.Sprintf("%s is done", s), func(t *testing.T) {
			t.Parallel()
			assert.True(t, isDoneState(s))
		})
	}
	for _, s := range notDone {
		t.Run(fmt.Sprintf("%s is not done", s), func(t *testing.T) {
			t.Parallel()
			assert.False(t, isDoneState(s))
		})
	}
}

func TestAllowAppendingLogs(t *testing.T) {
	t.Parallel()

	recentFinish := time.Now().Add(-30 * time.Second).Unix()
	staleFinish := time.Now().Add(-10 * time.Minute).Unix()

	// Step running always allows logs, regardless of pipeline state or age.
	t.Run("step running always allowed", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name   string
			status model.StatusValue
			finish int64
		}{
			{"pipeline running", model.StatusRunning, 0},
			{"pipeline success stale", model.StatusSuccess, staleFinish},
			{"pipeline failure stale", model.StatusFailure, staleFinish},
			{"pipeline killed stale", model.StatusKilled, staleFinish},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				p := &model.Pipeline{Status: tc.status, Finished: tc.finish}
				assert.NoError(t, allowAppendingLogs(p, &model.Step{State: model.StatusRunning}))
			})
		}
	})

	// Pipeline running allows logs for any step state.
	t.Run("pipeline running any step allowed", func(t *testing.T) {
		t.Parallel()

		for _, ss := range []model.StatusValue{
			model.StatusSuccess, model.StatusFailure, model.StatusPending, model.StatusKilled,
		} {
			t.Run(string(ss), func(t *testing.T) {
				t.Parallel()
				p := &model.Pipeline{Status: model.StatusRunning}
				assert.NoError(t, allowAppendingLogs(p, &model.Step{State: ss}))
			})
		}
	})

	// Recent finish → drain window allows logs.
	t.Run("recent finish drain allowed", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			pStatus model.StatusValue
			sState  model.StatusValue
		}{
			{model.StatusSuccess, model.StatusSuccess},
			{model.StatusFailure, model.StatusFailure},
			{model.StatusKilled, model.StatusPending},
		} {
			t.Run(fmt.Sprintf("%s/%s", tc.pStatus, tc.sState), func(t *testing.T) {
				t.Parallel()
				p := &model.Pipeline{Status: tc.pStatus, Finished: recentFinish}
				assert.NoError(t, allowAppendingLogs(p, &model.Step{State: tc.sState}))
			})
		}
	})

	// Stale finish → drain window expired → reject.
	t.Run("stale finish drain rejected", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			pStatus model.StatusValue
			sState  model.StatusValue
			finish  int64
		}{
			{model.StatusSuccess, model.StatusSuccess, staleFinish},
			{model.StatusFailure, model.StatusFailure, staleFinish},
			{model.StatusKilled, model.StatusPending, staleFinish},
			{model.StatusError, model.StatusCreated, staleFinish},
			{model.StatusSuccess, model.StatusSuccess, 0}, // zero = never recorded
		} {
			t.Run(fmt.Sprintf("%s/%s/fin=%d", tc.pStatus, tc.sState, tc.finish), func(t *testing.T) {
				t.Parallel()
				p := &model.Pipeline{Status: tc.pStatus, Finished: tc.finish}
				assert.ErrorIs(t, allowAppendingLogs(p, &model.Step{State: tc.sState}), ErrAgentIllegalLogStreaming)
			})
		}
	})
}

// TestAllowAppendingLogsDrainBoundary guards the exact edge of the 5-minute
// drain window against off-by-one errors.
func TestAllowAppendingLogsDrainBoundary(t *testing.T) {
	t.Parallel()

	step := &model.Step{State: model.StatusSuccess}

	t.Run("just inside drain window allowed", func(t *testing.T) {
		t.Parallel()
		p := &model.Pipeline{
			Status:   model.StatusSuccess,
			Finished: time.Now().Add(-(logStreamDelayAllowed - time.Second)).Unix(),
		}
		assert.NoError(t, allowAppendingLogs(p, step))
	})

	t.Run("just outside drain window rejected", func(t *testing.T) {
		t.Parallel()
		p := &model.Pipeline{
			Status:   model.StatusSuccess,
			Finished: time.Now().Add(-(logStreamDelayAllowed + time.Second)).Unix(),
		}
		assert.ErrorIs(t, allowAppendingLogs(p, step), ErrAgentIllegalLogStreaming)
	})
}
