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

package scheduler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestCompleteRunningChildren(t *testing.T) {
	t.Run("a still-running service step is finished so the workflow succeeds", func(t *testing.T) {
		successStep := &model.Step{
			ID:      1,
			State:   model.StatusSuccess,
			Started: 1234567800,
		}
		runningService := &model.Step{
			ID:      2,
			State:   model.StatusRunning,
			Started: 1234567800,
		}
		workflow := model.Workflow{
			ID:       7,
			State:    model.StatusRunning,
			Children: []*model.Step{successStep, runningService},
		}

		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("StepUpdate", mock.Anything).Return(nil)

		p := &impl{store: mockStore}
		p.completeRunningChildren(&workflow, 1234567900)

		assert.Equal(t, model.StatusSuccess, runningService.State)
		assert.Equal(t, int64(1234567900), runningService.Finished)

		updateWorkflowStateToDone(&workflow, rpc.WorkflowState{
			Started:  1234567800,
			Finished: 1234567900,
		})
		assert.Equal(t, model.StatusSuccess, workflow.State)
	})
}
