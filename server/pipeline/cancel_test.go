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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
	scheduler_mocks "go.woodpecker-ci.org/woodpecker/v3/server/scheduler/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestCancel(t *testing.T) {
	repo := &model.Repo{ID: 5, Owner: "o", Name: "r", FullName: "o/r"}
	user := &model.User{ID: 1}

	t.Run("publishes the forge status for every workflow", func(t *testing.T) {
		// Regression test. updatePipelineStatus is a loop over
		// pipeline.Workflows, and the pipeline Cancel is handed does not carry
		// them — so while the tree was loaded AFTER the call, the loop ran zero
		// times and cancelling never touched the forge status at all.
		pipeline := &model.Pipeline{ID: 10, Number: 259, Status: model.StatusRunning}
		workflow := &model.Workflow{ID: 1, PipelineID: 10, Name: "ci", State: model.StatusPending}

		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("WorkflowGetTree", mock.Anything).Return([]*model.Workflow{workflow}, nil)
		mockStore.On("WorkflowUpdate", mock.Anything).Return(nil)
		mockStore.On("UpdatePipeline", mock.Anything).Return(nil)

		mockScheduler := scheduler_mocks.NewMockScheduler(t)
		mockScheduler.On("CancelWorkflows", mock.Anything, mock.Anything).Return(nil)
		mockScheduler.On("PublishPipelineEvent", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		server.Config.Services.Scheduler = mockScheduler

		mockForge := forge_mocks.NewMockForge(t)
		mockForge.On("Status", mock.Anything, user, repo, mock.Anything, mock.Anything).Return(nil)

		assert.NoError(t, Cancel(context.Background(), mockForge, mockStore, repo, user, pipeline, nil))
		mockForge.AssertNumberOfCalls(t, "Status", 1)
	})

	t.Run("closes a running workflow the queue has no task for", func(t *testing.T) {
		// The agent that held this workflow is gone, so the report that normally
		// comes back over gRPC never will. Left alone the workflow stays
		// `running` forever and, because the forge status is per workflow, so
		// does the commit status.
		step := &model.Step{ID: 100, State: model.StatusRunning, Started: 1}
		workflow := &model.Workflow{
			ID: 4329, PipelineID: 10, Name: "ci", State: model.StatusRunning,
			Started: 1, Children: []*model.Step{step},
		}
		pipeline := &model.Pipeline{ID: 10, Number: 259, Status: model.StatusRunning}

		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("WorkflowGetTree", mock.Anything).Return([]*model.Workflow{workflow}, nil)
		mockStore.On("WorkflowUpdate", mock.MatchedBy(func(w *model.Workflow) bool {
			return w.State == model.StatusKilled
		})).Return(nil)
		mockStore.On("StepUpdate", mock.Anything).Return(nil)
		mockStore.On("UpdatePipeline", mock.Anything).Return(nil)

		mockScheduler := scheduler_mocks.NewMockScheduler(t)
		mockScheduler.On("CancelWorkflows", mock.Anything, []string{"4329"}).
			Return(&queue.ErrTasksNotFound{IDs: []string{"4329"}})
		mockScheduler.On("PublishPipelineEvent", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		server.Config.Services.Scheduler = mockScheduler

		mockForge := forge_mocks.NewMockForge(t)
		mockForge.On("Status", mock.Anything, user, repo, mock.Anything, mock.Anything).Return(nil)

		assert.NoError(t, Cancel(context.Background(), mockForge, mockStore, repo, user, pipeline, nil))
		// the workflow is persisted through the store (the Update* helpers take
		// a value, as everywhere else in this package), so the matcher above is
		// what proves the state; the step is a pointer and is closed in place
		mockStore.AssertCalled(t, "WorkflowUpdate", mock.Anything)
		mockStore.AssertCalled(t, "StepUpdate", mock.Anything)
		assert.Equal(t, model.StatusKilled, step.State)
	})

	t.Run("leaves a running workflow alone while an agent still holds it", func(t *testing.T) {
		step := &model.Step{ID: 100, State: model.StatusRunning, Started: 1}
		workflow := &model.Workflow{
			ID: 4617, PipelineID: 11, Name: "ci", State: model.StatusRunning,
			Started: 1, Children: []*model.Step{step},
		}
		pipeline := &model.Pipeline{ID: 11, Number: 495, Status: model.StatusRunning}

		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("WorkflowGetTree", mock.Anything).Return([]*model.Workflow{workflow}, nil)
		mockStore.On("UpdatePipeline", mock.Anything).Return(nil)

		mockScheduler := scheduler_mocks.NewMockScheduler(t)
		mockScheduler.On("CancelWorkflows", mock.Anything, []string{"4617"}).Return(nil)
		mockScheduler.On("PublishPipelineEvent", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		server.Config.Services.Scheduler = mockScheduler

		mockForge := forge_mocks.NewMockForge(t)
		mockForge.On("Status", mock.Anything, user, repo, mock.Anything, mock.Anything).Return(nil)

		assert.NoError(t, Cancel(context.Background(), mockForge, mockStore, repo, user, pipeline, nil))
		// the agent will report it; the server must not write its state for it
		assert.Equal(t, model.StatusRunning, workflow.State)
		assert.Equal(t, model.StatusRunning, step.State)
		mockStore.AssertNotCalled(t, "WorkflowUpdate", mock.Anything)
		mockStore.AssertNotCalled(t, "StepUpdate", mock.Anything)
	})
}
