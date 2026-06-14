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

package api

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestProcessQueueTasks(t *testing.T) {
	t.Run("stale agent ID does not return error", func(t *testing.T) {
		// Regression test for woodpecker-ci/woodpecker#6615:
		// a task whose agent row was deleted must not cause processQueueTasks
		// to return an error that would make /api/queue/info return HTTP 500.
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("AgentFind", int64(99)).Return(nil, errors.New("record not found"))

		tasks := []*model.Task{{ID: "task-1", AgentID: 99}}

		result, err := processQueueTasks(mockStore, tasks, make(map[int64]string))
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "", result[0].AgentName)
	})

	t.Run("known agent ID populates agent name", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("AgentFind", int64(1)).Return(&model.Agent{ID: 1, Name: "my-agent"}, nil)

		tasks := []*model.Task{{ID: "task-1", AgentID: 1}}

		result, err := processQueueTasks(mockStore, tasks, make(map[int64]string))
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "my-agent", result[0].AgentName)
	})

	t.Run("task without agent ID has empty agent name", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)

		tasks := []*model.Task{{ID: "task-1", AgentID: 0}}

		result, err := processQueueTasks(mockStore, tasks, make(map[int64]string))
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "", result[0].AgentName)
	})

	t.Run("pipeline ID is resolved to pipeline number", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("GetPipeline", int64(42)).Return(&model.Pipeline{ID: 42, Number: 7}, nil)

		tasks := []*model.Task{{ID: "task-1", PipelineID: 42}}

		result, err := processQueueTasks(mockStore, tasks, make(map[int64]string))
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, int64(7), result[0].PipelineNumber)
	})

	t.Run("agent name is cached across tasks", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		// AgentFind should only be called once even with two tasks sharing the same AgentID.
		mockStore.On("AgentFind", int64(5)).Return(&model.Agent{ID: 5, Name: "cached-agent"}, nil).Once()

		tasks := []*model.Task{
			{ID: "task-1", AgentID: 5},
			{ID: "task-2", AgentID: 5},
		}

		result, err := processQueueTasks(mockStore, tasks, make(map[int64]string))
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "cached-agent", result[0].AgentName)
		assert.Equal(t, "cached-agent", result[1].AgentName)
		mockStore.AssertNumberOfCalls(t, "AgentFind", 1)
	})
}
