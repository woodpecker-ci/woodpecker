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

func TestAllowAppendingLogs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		state   model.StatusValue
		wantErr error
	}{
		{"running allows", model.StatusRunning, nil},
		{"pending rejects", model.StatusPending, ErrAgentIllegalLogStreaming},
		{"created rejects", model.StatusCreated, ErrAgentIllegalLogStreaming},
		{"success rejects", model.StatusSuccess, ErrAgentIllegalLogStreaming},
		{"failure rejects", model.StatusFailure, ErrAgentIllegalLogStreaming},
		{"killed rejects", model.StatusKilled, ErrAgentIllegalLogStreaming},
		{"error rejects", model.StatusError, ErrAgentIllegalLogStreaming},
		{"skipped rejects", model.StatusSkipped, ErrAgentIllegalLogStreaming},
		{"blocked rejects", model.StatusBlocked, ErrAgentIllegalLogStreaming},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			step := &model.Step{State: tt.state}
			err := allowAppendingLogs(step)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
