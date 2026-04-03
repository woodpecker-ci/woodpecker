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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestCheckPipelineState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    model.StatusValue
		wantErr   error
		expectNil bool
	}{
		{
			name:      "created is allowed",
			status:    model.StatusCreated,
			expectNil: true,
		},
		{
			name:      "pending is allowed",
			status:    model.StatusPending,
			expectNil: true,
		},
		{
			name:      "running is allowed",
			status:    model.StatusRunning,
			expectNil: true,
		},
		{
			name:    "blocked is rejected",
			status:  model.StatusBlocked,
			wantErr: ErrAgentIllegalPipelineWorkflowRun,
		},
		{
			name:    "success is rejected as re-run",
			status:  model.StatusSuccess,
			wantErr: ErrAgentIllegalPipelineWorkflowReRunStateChange,
		},
		{
			name:    "failure is rejected as re-run",
			status:  model.StatusFailure,
			wantErr: ErrAgentIllegalPipelineWorkflowReRunStateChange,
		},
		{
			name:    "killed is rejected as re-run",
			status:  model.StatusKilled,
			wantErr: ErrAgentIllegalPipelineWorkflowReRunStateChange,
		},
		{
			name:    "error is rejected as re-run",
			status:  model.StatusError,
			wantErr: ErrAgentIllegalPipelineWorkflowReRunStateChange,
		},
		{
			name:    "skipped is rejected as re-run",
			status:  model.StatusSkipped,
			wantErr: ErrAgentIllegalPipelineWorkflowReRunStateChange,
		},
		{
			name:    "declined is rejected as re-run",
			status:  model.StatusDeclined,
			wantErr: ErrAgentIllegalPipelineWorkflowReRunStateChange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pipeline := &model.Pipeline{Status: tt.status}
			err := checkPipelineState(pipeline)

			if tt.expectNil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func TestCheckWorkflowStepStates(t *testing.T) {
	t.Parallel()

	t.Run("workflow only", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			state   model.StatusValue
			wantErr error
		}{
			{"created allows", model.StatusCreated, nil},
			{"pending allows", model.StatusPending, nil},
			{"running allows", model.StatusRunning, nil},
			{"blocked rejects", model.StatusBlocked, ErrAgentIllegalWorkflowRun},
			{"success rejects", model.StatusSuccess, ErrAgentIllegalWorkflowReRunStateChange},
			{"failure rejects", model.StatusFailure, ErrAgentIllegalWorkflowReRunStateChange},
			{"killed rejects", model.StatusKilled, ErrAgentIllegalWorkflowReRunStateChange},
			{"error rejects", model.StatusError, ErrAgentIllegalWorkflowReRunStateChange},
			{"skipped rejects", model.StatusSkipped, ErrAgentIllegalWorkflowReRunStateChange},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				workflow := &model.Workflow{State: tt.state}
				err := checkWorkflowStepStates(workflow, nil)

				if tt.wantErr == nil {
					assert.NoError(t, err)
				} else {
					assert.ErrorIs(t, err, tt.wantErr)
				}
			})
		}
	})

	t.Run("step only (nil workflow)", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			state   model.StatusValue
			wantErr error
		}{
			{"created allows", model.StatusCreated, nil},
			{"pending allows", model.StatusPending, nil},
			{"running allows", model.StatusRunning, nil},
			{"blocked rejects", model.StatusBlocked, ErrAgentIllegalStepRun},
			{"success rejects", model.StatusSuccess, ErrAgentIllegalStepReRunStateChange},
			{"failure rejects", model.StatusFailure, ErrAgentIllegalStepReRunStateChange},
			{"killed rejects", model.StatusKilled, ErrAgentIllegalStepReRunStateChange},
			{"error rejects", model.StatusError, ErrAgentIllegalStepReRunStateChange},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				step := &model.Step{State: tt.state}
				err := checkWorkflowStepStates(nil, step)

				if tt.wantErr == nil {
					assert.NoError(t, err)
				} else {
					assert.ErrorIs(t, err, tt.wantErr)
				}
			})
		}
	})

	t.Run("nil workflow and nil step", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, checkWorkflowStepStates(nil, nil))
	})

	t.Run("workflow running, step running", func(t *testing.T) {
		t.Parallel()

		workflow := &model.Workflow{State: model.StatusRunning}
		step := &model.Step{State: model.StatusRunning}
		assert.NoError(t, checkWorkflowStepStates(workflow, step))
	})

	t.Run("workflow running, step finished", func(t *testing.T) {
		t.Parallel()

		workflow := &model.Workflow{State: model.StatusRunning}
		step := &model.Step{State: model.StatusSuccess}
		err := checkWorkflowStepStates(workflow, step)
		assert.ErrorIs(t, err, ErrAgentIllegalStepReRunStateChange)
		// should not contain workflow error
		assert.False(t, errors.Is(err, ErrAgentIllegalWorkflowReRunStateChange))
	})

	t.Run("workflow running, step blocked", func(t *testing.T) {
		t.Parallel()

		workflow := &model.Workflow{State: model.StatusRunning}
		step := &model.Step{State: model.StatusBlocked}
		err := checkWorkflowStepStates(workflow, step)
		assert.ErrorIs(t, err, ErrAgentIllegalStepRun)
	})

	t.Run("both finished - joined errors", func(t *testing.T) {
		t.Parallel()

		workflow := &model.Workflow{State: model.StatusSuccess}
		step := &model.Step{State: model.StatusSuccess}
		err := checkWorkflowStepStates(workflow, step)
		assert.ErrorIs(t, err, ErrAgentIllegalWorkflowReRunStateChange)
		assert.ErrorIs(t, err, ErrAgentIllegalStepReRunStateChange)
	})

	t.Run("both blocked - joined errors", func(t *testing.T) {
		t.Parallel()

		workflow := &model.Workflow{State: model.StatusBlocked}
		step := &model.Step{State: model.StatusBlocked}
		err := checkWorkflowStepStates(workflow, step)
		assert.ErrorIs(t, err, ErrAgentIllegalWorkflowRun)
		assert.ErrorIs(t, err, ErrAgentIllegalStepRun)
	})

	t.Run("workflow finished, step blocked - joined errors", func(t *testing.T) {
		t.Parallel()

		workflow := &model.Workflow{State: model.StatusKilled}
		step := &model.Step{State: model.StatusBlocked}
		err := checkWorkflowStepStates(workflow, step)
		assert.ErrorIs(t, err, ErrAgentIllegalWorkflowReRunStateChange)
		assert.ErrorIs(t, err, ErrAgentIllegalStepRun)
	})

	t.Run("workflow finished (failure), step finished (failure) - joined errors", func(t *testing.T) {
		t.Parallel()

		workflow := &model.Workflow{State: model.StatusFailure}
		step := &model.Step{State: model.StatusFailure}
		err := checkWorkflowStepStates(workflow, step)
		assert.ErrorIs(t, err, ErrAgentIllegalWorkflowReRunStateChange)
		assert.ErrorIs(t, err, ErrAgentIllegalStepReRunStateChange)
	})
}

// AllowAppendingLogs — updated for the new (pipeline, step) signature
//
// New logic:
//   Allow if step.State == Running  (step is actively running)
//   Allow if pipeline.Status == Running  (pipeline still running, step may
//     have just finished but pipeline hasn't caught up yet)
//   Allow if pipeline.Finished is within the last logStreamDelayAllowed
//     (drain window after a server restart / network blip)
//   Reject otherwise.

func TestAllowAppendingLogs(t *testing.T) {
	t.Parallel()

	// recentFinish is a pipeline.Finished timestamp just 30 seconds ago —
	// well within the 5-minute drain window.
	recentFinish := time.Now().Add(-30 * time.Second).Unix()

	// staleFinish is a pipeline.Finished timestamp 10 minutes ago —
	// outside the drain window.
	staleFinish := time.Now().Add(-10 * time.Minute).Unix()

	tests := []struct {
		name           string
		pipelineStatus model.StatusValue
		pipelineFinish int64
		stepState      model.StatusValue
		wantErr        error
	}{
		// --- step is running: always allowed regardless of pipeline state ----
		{
			name:           "step running, pipeline running → allow",
			pipelineStatus: model.StatusRunning,
			stepState:      model.StatusRunning,
		},
		{
			name:           "step running, pipeline success → allow (step takes priority)",
			pipelineStatus: model.StatusSuccess,
			pipelineFinish: staleFinish,
			stepState:      model.StatusRunning,
		},
		{
			name:           "step running, pipeline failure → allow",
			pipelineStatus: model.StatusFailure,
			pipelineFinish: staleFinish,
			stepState:      model.StatusRunning,
		},
		{
			name:           "step running, pipeline killed → allow",
			pipelineStatus: model.StatusKilled,
			pipelineFinish: staleFinish,
			stepState:      model.StatusRunning,
		},

		// --- pipeline still running: allow even if step finished ------------
		{
			name:           "step success, pipeline still running → allow",
			pipelineStatus: model.StatusRunning,
			stepState:      model.StatusSuccess,
		},
		{
			name:           "step failure, pipeline still running → allow",
			pipelineStatus: model.StatusRunning,
			stepState:      model.StatusFailure,
		},
		{
			name:           "step pending, pipeline still running → allow",
			pipelineStatus: model.StatusRunning,
			stepState:      model.StatusPending,
		},
		{
			name:           "step killed, pipeline still running → allow",
			pipelineStatus: model.StatusRunning,
			stepState:      model.StatusKilled,
		},

		// --- pipeline finished recently: drain window allows logs -----------
		{
			name:           "step success, pipeline finished recently → allow (drain window)",
			pipelineStatus: model.StatusSuccess,
			pipelineFinish: recentFinish,
			stepState:      model.StatusSuccess,
		},
		{
			name:           "step failure, pipeline failed recently → allow (drain window)",
			pipelineStatus: model.StatusFailure,
			pipelineFinish: recentFinish,
			stepState:      model.StatusFailure,
		},
		{
			name:           "step pending, pipeline killed recently → allow (drain window)",
			pipelineStatus: model.StatusKilled,
			pipelineFinish: recentFinish,
			stepState:      model.StatusPending,
		},

		// --- pipeline finished and drain window expired: reject -------------
		{
			name:           "step success, pipeline success, stale finish → reject",
			pipelineStatus: model.StatusSuccess,
			pipelineFinish: staleFinish,
			stepState:      model.StatusSuccess,
			wantErr:        ErrAgentIllegalLogStreaming,
		},
		{
			name:           "step failure, pipeline failure, stale finish → reject",
			pipelineStatus: model.StatusFailure,
			pipelineFinish: staleFinish,
			stepState:      model.StatusFailure,
			wantErr:        ErrAgentIllegalLogStreaming,
		},
		{
			name:           "step pending, pipeline killed, stale finish → reject",
			pipelineStatus: model.StatusKilled,
			pipelineFinish: staleFinish,
			stepState:      model.StatusPending,
			wantErr:        ErrAgentIllegalLogStreaming,
		},
		{
			name:           "step created, pipeline error, stale finish → reject",
			pipelineStatus: model.StatusError,
			pipelineFinish: staleFinish,
			stepState:      model.StatusCreated,
			wantErr:        ErrAgentIllegalLogStreaming,
		},

		// --- zero Finished timestamp (never recorded): outside drain window -
		{
			name:           "step success, pipeline success, Finished=0 → reject",
			pipelineStatus: model.StatusSuccess,
			pipelineFinish: 0,
			stepState:      model.StatusSuccess,
			wantErr:        ErrAgentIllegalLogStreaming,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pipeline := &model.Pipeline{
				Status:   tt.pipelineStatus,
				Finished: tt.pipelineFinish,
			}
			step := &model.Step{State: tt.stepState}

			err := allowAppendingLogs(pipeline, step)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

// TestAllowAppendingLogsDrainBoundary checks the exact boundary of the
// 5-minute drain window to guard against off-by-one errors.
func TestAllowAppendingLogsDrainBoundary(t *testing.T) {
	t.Parallel()

	step := &model.Step{State: model.StatusSuccess}

	t.Run("finished exactly at drain window boundary is allowed", func(t *testing.T) {
		t.Parallel()

		// Finished just barely inside the window (1 second of headroom).
		finishedAt := time.Now().Add(-(logStreamDelayAllowed - time.Second)).Unix()
		pipeline := &model.Pipeline{Status: model.StatusSuccess, Finished: finishedAt}

		assert.NoError(t, allowAppendingLogs(pipeline, step))
	})

	t.Run("finished just outside drain window is rejected", func(t *testing.T) {
		t.Parallel()

		// Finished 1 second past the allowed window.
		finishedAt := time.Now().Add(-(logStreamDelayAllowed + time.Second)).Unix()
		pipeline := &model.Pipeline{Status: model.StatusSuccess, Finished: finishedAt}

		assert.ErrorIs(t, allowAppendingLogs(pipeline, step), ErrAgentIllegalLogStreaming)
	})
}
