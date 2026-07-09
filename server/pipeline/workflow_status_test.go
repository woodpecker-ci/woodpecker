// Copyright 2026 Woodpecker Authors
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

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestWorkflowStatus(t *testing.T) {
	tests := []struct {
		s []*model.Step
		e model.StatusValue
	}{
		{
			s: []*model.Step{
				{
					State:   model.StatusFailure,
					Failure: model.FailureIgnore,
				},
				{
					State:   model.StatusSuccess,
					Failure: model.FailureFail,
				},
			},
			e: model.StatusSuccess,
		},
		{
			s: []*model.Step{
				{
					State:   model.StatusSuccess,
					Failure: model.FailureFail,
				},
				{
					State:   model.StatusSuccess,
					Failure: model.FailureIgnore,
				},
			},
			e: model.StatusSuccess,
		},
		{
			s: []*model.Step{
				{
					State:   model.StatusFailure,
					Failure: model.FailureFail,
				},
				{
					State:   model.StatusSuccess,
					Failure: model.FailureFail,
				},
			},
			e: model.StatusFailure,
		},
		{
			s: []*model.Step{
				{
					State:   model.StatusSuccess,
					Failure: model.FailureFail,
				},
				{
					State:   model.StatusPending,
					Failure: model.FailureFail,
				},
			},
			e: model.StatusPending,
		},
		{
			s: []*model.Step{
				{
					State:   model.StatusSuccess,
					Failure: model.FailureFail,
				},
				{
					State:   model.StatusPending,
					Failure: model.FailureIgnore,
				},
			},
			e: model.StatusPending,
		},
		{
			s: []*model.Step{
				{
					State:   model.StatusSuccess,
					Failure: model.FailureIgnore,
				},
				{
					State:   model.StatusPending,
					Failure: model.FailureFail,
				},
			},
			e: model.StatusPending,
		},
		{
			s: []*model.Step{
				{
					State:   model.StatusSuccess,
					Failure: model.FailureIgnore,
				},
				{
					State:   model.StatusPending,
					Failure: model.FailureIgnore,
				},
			},
			e: model.StatusPending,
		},
		{
			s: []*model.Step{
				{
					State:   model.StatusRunning,
					Failure: model.FailureFail,
				},
				{
					State:   model.StatusPending,
					Failure: model.FailureFail,
				},
			},
			e: model.StatusRunning,
		},
		{
			s: []*model.Step{
				{
					State:   model.StatusRunning,
					Failure: model.FailureIgnore,
				},
				{
					State:   model.StatusPending,
					Failure: model.FailureIgnore,
				},
			},
			e: model.StatusRunning,
		},
		{
			s: []*model.Step{
				{
					State:   model.StatusRunning,
					Failure: model.FailureIgnore,
				},
				{
					State:   model.StatusPending,
					Failure: model.FailureFail,
				},
			},
			e: model.StatusRunning,
		},
		{
			s: []*model.Step{
				{
					State:   model.StatusRunning,
					Failure: model.FailureFail,
				},
				{
					State:   model.StatusPending,
					Failure: model.FailureIgnore,
				},
			},
			e: model.StatusRunning,
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.e, WorkflowStatus(tt.s))
	}
}

func TestUpdateWorkflowStatusToRunning(t *testing.T) {
	t.Run("should update workflow to running status using the server clock", func(t *testing.T) {
		before := time.Now().Unix()
		workflow := model.Workflow{
			ID:    1,
			State: model.StatusPending,
		}
		// Agent reports a start time; the server must ignore it and record its
		// own clock instead (#6808).
		state := rpc.WorkflowState{
			Started: 1234567890,
		}

		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("WorkflowUpdate", mock.MatchedBy(func(w *model.Workflow) bool {
			return w.ID == 1 && w.State == model.StatusRunning && w.Started >= before
		})).Return(nil)

		result, err := UpdateWorkflowStatusToRunning(mockStore, workflow, state)

		assert.NoError(t, err)
		assert.Equal(t, model.StatusRunning, result.State)
		assert.NotEqual(t, int64(1234567890), result.Started)
		assert.GreaterOrEqual(t, result.Started, before)
		mockStore.AssertCalled(t, "WorkflowUpdate", mock.Anything)
	})
}

func TestUpdateWorkflowToStatusSkipped(t *testing.T) {
	t.Run("should update workflow to skipped status", func(t *testing.T) {
		workflow := model.Workflow{
			ID:    2,
			State: model.StatusPending,
		}

		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("WorkflowUpdate", mock.MatchedBy(func(w *model.Workflow) bool {
			return w.ID == 2 && w.State == model.StatusSkipped
		})).Return(nil)

		result, err := UpdateWorkflowToStatusSkipped(mockStore, workflow)

		assert.NoError(t, err)
		assert.Equal(t, model.StatusSkipped, result.State)
		mockStore.AssertCalled(t, "WorkflowUpdate", mock.Anything)
	})
}

func TestUpdateWorkflowStatusToDone(t *testing.T) {
	t.Run("should mark as skipped when not started", func(t *testing.T) {
		before := time.Now().Unix()
		workflow := model.Workflow{
			ID:       3,
			State:    model.StatusRunning,
			Children: []*model.Step{},
		}
		state := rpc.WorkflowState{
			Started:  0, // Not started
			Finished: 1234567900,
			Error:    "",
		}

		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("WorkflowUpdate", mock.MatchedBy(func(w *model.Workflow) bool {
			return w.State == model.StatusSkipped && w.Finished >= before
		})).Return(nil)

		result, err := UpdateWorkflowStatusToDone(mockStore, workflow, state)

		assert.NoError(t, err)
		// Started==0 still marks the workflow skipped (used as a presence flag)...
		assert.Equal(t, model.StatusSkipped, result.State)
		// ...but Finished is stamped from the server clock, not the agent value (#6808).
		assert.NotEqual(t, int64(1234567900), result.Finished)
		assert.GreaterOrEqual(t, result.Finished, before)
	})

	t.Run("should mark as failure when error exists", func(t *testing.T) {
		workflow := model.Workflow{
			ID:       5,
			State:    model.StatusRunning,
			Children: []*model.Step{},
		}
		state := rpc.WorkflowState{
			Started:  1234567800,
			Finished: 1234567900,
			Error:    "some error occurred",
		}

		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("WorkflowUpdate", mock.MatchedBy(func(w *model.Workflow) bool {
			return w.State == model.StatusFailure
		})).Return(nil)

		result, err := UpdateWorkflowStatusToDone(mockStore, workflow, state)

		assert.NoError(t, err)
		assert.Equal(t, model.StatusFailure, result.State)
		assert.Equal(t, "some error occurred", result.Error)
	})

	t.Run("should mark as success when all children are successful", func(t *testing.T) {
		before := time.Now().Unix()
		successStep := &model.Step{
			ID:    1,
			State: model.StatusSuccess,
		}
		workflow := model.Workflow{
			ID:       6,
			State:    model.StatusRunning,
			Children: []*model.Step{successStep},
		}
		state := rpc.WorkflowState{
			Started:  1234567800,
			Finished: 1234567900,
			Error:    "",
		}

		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("WorkflowUpdate", mock.Anything).Return(nil)

		result, err := UpdateWorkflowStatusToDone(mockStore, workflow, state)

		assert.NoError(t, err)
		assert.Equal(t, model.StatusSuccess, result.State)
		// Finished comes from the server clock, not the agent-supplied value (#6808).
		assert.NotEqual(t, int64(1234567900), result.Finished)
		assert.GreaterOrEqual(t, result.Finished, before)
	})
}
