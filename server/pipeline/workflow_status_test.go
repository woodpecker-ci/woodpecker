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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestUpdateWorkflowStatusToRunning(t *testing.T) {
	t.Run("should update workflow to running status", func(t *testing.T) {
		workflow := model.Workflow{
			ID:    1,
			State: model.StatusPending,
		}
		state := rpc.WorkflowState{
			Started: 1234567890,
		}

		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("WorkflowUpdate", mock.MatchedBy(func(w *model.Workflow) bool {
			return w.ID == 1 && w.State == model.StatusRunning && w.Started == 1234567890
		})).Return(nil)

		result, err := UpdateWorkflowStatusToRunning(mockStore, workflow, state)

		assert.NoError(t, err)
		assert.Equal(t, model.StatusRunning, result.State)
		assert.Equal(t, int64(1234567890), result.Started)
		mockStore.AssertCalled(t, "WorkflowUpdate", mock.Anything)
	})
}
