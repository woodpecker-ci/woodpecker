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

//go:build test

package api

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub/memory"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
	queue_mocks "go.woodpecker-ci.org/woodpecker/v3/server/queue/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/scheduler"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

// installScheduler wires a scheduler backed by a mock queue into the global
// server config and returns the mock so tests can stub Info/Pause/Resume.
// GetQueueInfo, PauseQueue and ResumeQueue only call queue methods, so a real
// scheduler proxying to the mock queue is enough.
func installScheduler(t *testing.T, s store.Store) *queue_mocks.MockQueue {
	t.Helper()
	q := queue_mocks.NewMockQueue(t)
	server.Config.Services.Scheduler = scheduler.NewScheduler(t.Context(), s, q, memory.New())
	return q
}

// seedAgent inserts an agent so processQueueTasks can resolve its name.
func seedAgent(t *testing.T, s store.Store, name string) *model.Agent {
	t.Helper()
	agent := &model.Agent{Name: name, Backend: "local", Platform: "windows/amd64"}
	require.NoError(t, s.AgentCreate(agent))
	return agent
}

// seedPipeline inserts a pipeline for the repo so processQueueTasks can resolve
// its number.
func seedPipeline(t *testing.T, s store.Store, repoID int64) *model.Pipeline {
	t.Helper()
	p := &model.Pipeline{RepoID: repoID, Status: model.StatusRunning}
	require.NoError(t, s.CreatePipeline(p))
	return p
}

func TestGetQueueInfo(t *testing.T) {
	s := newTestStore(t)
	repo, _ := cronFixture(t, s)

	t.Run("formats running and waiting tasks with agent and pipeline info", func(t *testing.T) {
		agent := seedAgent(t, s, "BUILDSERVER-restricted_D")
		pipe := seedPipeline(t, s, repo.ID)

		// Mirrors a real queue state: one task waiting on deps (no agent yet)
		// and one running task assigned to an agent.
		waiting := &model.Task{
			ID: "76054", Name: "build", PipelineID: pipe.ID, RepoID: repo.ID,
			Dependencies: []string{"76057"},
		}
		running := &model.Task{
			ID: "76057", Name: "check_preconditions", PipelineID: pipe.ID, RepoID: repo.ID,
			AgentID: agent.ID,
		}

		info := queue.InfoT{
			WaitingOnDeps: []*model.Task{waiting},
			Running:       []*model.Task{running},
			Paused:        false,
		}
		info.Stats.Workers = 13
		info.Stats.WaitingOnDeps = 1
		info.Stats.Running = 1

		q := installScheduler(t, s)
		q.On("Info", mock.Anything).Return(info)

		tc := newTestContext(t, s)
		GetQueueInfo(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		var got model.QueueInfo
		tc.decodeJSON(t, &got)

		assert.False(t, got.Paused)
		assert.Equal(t, 13, got.Stats.WorkerCount)
		assert.Equal(t, 1, got.Stats.WaitingOnDepsCount)
		assert.Equal(t, 1, got.Stats.RunningCount)

		require.Len(t, got.WaitingOnDeps, 1)
		assert.Equal(t, "76054", got.WaitingOnDeps[0].ID)
		assert.Empty(t, got.WaitingOnDeps[0].AgentName) // no agent assigned
		assert.Equal(t, pipe.Number, got.WaitingOnDeps[0].PipelineNumber)

		require.Len(t, got.Running, 1)
		assert.Equal(t, "76057", got.Running[0].ID)
		assert.Equal(t, "BUILDSERVER-restricted_D", got.Running[0].AgentName)
		assert.Equal(t, pipe.Number, got.Running[0].PipelineNumber)
	})

	t.Run("empty queue returns empty lists", func(t *testing.T) {
		q := installScheduler(t, s)
		q.On("Info", mock.Anything).Return(queue.InfoT{})

		tc := newTestContext(t, s)
		GetQueueInfo(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		var got model.QueueInfo
		tc.decodeJSON(t, &got)
		assert.Empty(t, got.Pending)
		assert.Empty(t, got.Running)
		assert.Empty(t, got.WaitingOnDeps)
	})

	t.Run("unknown agent is ignored", func(t *testing.T) {
		pipe := seedPipeline(t, s, repo.ID)
		info := queue.InfoT{Running: []*model.Task{
			{ID: "1", AgentID: 99999, PipelineID: pipe.ID, RepoID: repo.ID},
		}}

		q := installScheduler(t, s)
		q.On("Info", mock.Anything).Return(info)

		tc := newTestContext(t, s)
		GetQueueInfo(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		var got model.QueueInfo
		tc.decodeJSON(t, &got)
		require.Len(t, got.Running, 1)
		assert.Equal(t, "1", got.Running[0].ID)
		assert.Empty(t, got.Running[0].AgentName)
	})

	t.Run("unknown pipeline returns internal error", func(t *testing.T) {
		info := queue.InfoT{Running: []*model.Task{
			{ID: "1", PipelineID: 99999, RepoID: repo.ID},
		}}

		q := installScheduler(t, s)
		q.On("Info", mock.Anything).Return(info)

		tc := newTestContext(t, s)
		GetQueueInfo(tc.Ctx)

		assert.Equal(t, http.StatusInternalServerError, tc.Recorder.Code)
		assert.Contains(t, tc.Recorder.Body.String(), "pipeline not found")
	})
}

func TestPauseResumeQueue(t *testing.T) {
	s := newTestStore(t)

	t.Run("pause returns no content and pauses queue", func(t *testing.T) {
		q := installScheduler(t, s)
		q.On("Pause").Return()

		tc := newTestContext(t, s)
		PauseQueue(tc.Ctx)

		assert.Equal(t, http.StatusNoContent, tc.Ctx.Writer.Status())
		q.AssertCalled(t, "Pause")
	})

	t.Run("resume returns no content and resumes queue", func(t *testing.T) {
		q := installScheduler(t, s)
		q.On("Resume").Return()

		tc := newTestContext(t, s)
		ResumeQueue(tc.Ctx)

		assert.Equal(t, http.StatusNoContent, tc.Ctx.Writer.Status())
		q.AssertCalled(t, "Resume")
	})
}

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
