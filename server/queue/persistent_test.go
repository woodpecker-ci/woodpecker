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

package queue

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/datastore"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// A task that lingers in the in-memory queue but is already gone from the
// backup store must be dropped on Poll instead of being handed to the agent,
// otherwise it loops forever (re-poll, illegal-instruction, resubmit).
func TestPersistentQueuePollDropsStaleTask(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	store := store_mocks.NewMockStore(t)
	store.EXPECT().TaskSetAgent("1", int64(1)).Return(types.ErrRecordNotExist).Once()

	pq := &persistentQueue{Queue: q, store: store}

	task := genDummyTask()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task}))

	got, err := pq.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Nil(t, got, "stale task must not be returned to the agent")

	info := q.Info(ctx)
	assert.Equal(t, 0, info.Stats.Pending, "stale task must be removed from pending")
	assert.Equal(t, 0, info.Stats.Running, "stale task must be removed from running")
}

func TestPersistentQueuePollDropsTaskWithMissingWorkflow(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	store := store_mocks.NewMockStore(t)
	store.EXPECT().TaskSetAgent("1", int64(1)).Return(nil).Once()
	store.EXPECT().WorkflowLoad(int64(1)).Return(nil, types.ErrRecordNotExist).Once()
	store.EXPECT().TaskDelete("1").Return(nil).Once()

	pq := &persistentQueue{Queue: q, store: store}

	task := genDummyTask()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task}))

	got, err := pq.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Nil(t, got, "task without a workflow must not be returned to the agent")

	info := q.Info(ctx)
	assert.Equal(t, 0, info.Stats.Pending)
	assert.Equal(t, 0, info.Stats.Running)
}

func TestPersistentQueuePollDropsTerminalWorkflowTask(t *testing.T) {
	terminalStates := []model.StatusValue{
		model.StatusSuccess,
		model.StatusFailure,
		model.StatusKilled,
		model.StatusCanceled,
		model.StatusSkipped,
		model.StatusError,
		model.StatusDeclined,
	}

	for _, state := range terminalStates {
		t.Run(string(state), func(t *testing.T) {
			t.Parallel()
			ctx, cancel, q := setupTestQueue(t)
			defer cancel(nil)

			store := store_mocks.NewMockStore(t)
			store.EXPECT().TaskSetAgent("1", int64(1)).Return(nil).Once()
			store.EXPECT().WorkflowLoad(int64(1)).Return(&model.Workflow{ID: 1, State: state}, nil).Once()
			store.EXPECT().TaskDelete("1").Return(nil).Once()

			pq := &persistentQueue{Queue: q, store: store}

			task := genDummyTask()
			assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task}))

			got, err := pq.Poll(ctx, 1, filterFnTrue)
			assert.NoError(t, err)
			assert.Nil(t, got, "terminal workflow task must not be returned to the agent")

			info := q.Info(ctx)
			assert.Equal(t, 0, info.Stats.Pending)
			assert.Equal(t, 0, info.Stats.Running)
		})
	}
}

func TestPersistentQueueDoneRemovesPendingTaskFromBackup(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	store := store_mocks.NewMockStore(t)
	store.EXPECT().TaskDelete("1").Return(nil).Once()

	pq := &persistentQueue{Queue: q, store: store}

	task := genDummyTask()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task}))
	assert.NoError(t, pq.Done(ctx, task.ID, model.StatusSuccess))

	info := q.Info(ctx)
	assert.Equal(t, 0, info.Stats.Pending)
	assert.Equal(t, 0, info.Stats.Running)
}

func TestPersistentQueueDoneIgnoresAlreadyRemovedBackupTask(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	store := store_mocks.NewMockStore(t)
	store.EXPECT().TaskSetAgent("1", int64(1)).Return(nil).Once()
	store.EXPECT().WorkflowLoad(int64(1)).Return(&model.Workflow{ID: 1, State: model.StatusPending}, nil).Once()
	store.EXPECT().TaskDelete("1").Return(types.ErrRecordNotExist).Once()

	pq := &persistentQueue{Queue: q, store: store}

	task := genDummyTask()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task}))

	got, err := pq.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.NoError(t, pq.Done(ctx, task.ID, model.StatusSuccess))

	info := q.Info(ctx)
	assert.Equal(t, 0, info.Stats.Pending)
	assert.Equal(t, 0, info.Stats.Running)
}

func TestPersistentQueuePausePersistsState(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	store := store_mocks.NewMockStore(t)
	store.EXPECT().ServerConfigSet(serverConfigQueuePaused, "true").Return(nil).Once()

	pq := &persistentQueue{Queue: q, store: store}
	pq.Pause()

	assert.True(t, pq.Info(ctx).Paused)
}

func TestPersistentQueueResumePersistsState(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	store := store_mocks.NewMockStore(t)
	store.EXPECT().ServerConfigSet(serverConfigQueuePaused, "true").Return(nil).Once()
	store.EXPECT().ServerConfigSet(serverConfigQueuePaused, "false").Return(nil).Once()

	pq := &persistentQueue{Queue: q, store: store}
	pq.Pause()
	pq.Resume()

	assert.False(t, pq.Info(ctx).Paused)
}

func TestWithTaskStoreLeavesQueueRunningWhenNotPaused(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	store := store_mocks.NewMockStore(t)
	store.EXPECT().TaskList().Return(nil, nil).Once()
	store.EXPECT().ServerConfigGet(serverConfigQueuePaused).Return("false", nil).Once()

	pq := WithTaskStore(ctx, q, store)

	assert.False(t, pq.Info(ctx).Paused)
}

func TestWithTaskStoreLeavesQueueRunningWhenNoPausedConfig(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	store := store_mocks.NewMockStore(t)
	store.EXPECT().TaskList().Return(nil, nil).Once()
	store.EXPECT().ServerConfigGet(serverConfigQueuePaused).Return("", types.ErrRecordNotExist).Once()

	pq := WithTaskStore(ctx, q, store)

	assert.False(t, pq.Info(ctx).Paused)
}

// A dependency that cannot be loaded must not stop the other dependencies of
// the task from being restored.
func TestRestoreDepStatusSkipsUnloadableDependency(t *testing.T) {
	s := store_mocks.NewMockStore(t)
	s.EXPECT().WorkflowLoad(int64(1)).Return(nil, errors.New("db down")).Once()
	s.EXPECT().WorkflowLoad(int64(2)).Return(&model.Workflow{State: model.StatusFailure}, nil).Once()

	task := &model.Task{ID: "3", Dependencies: []string{"invalid", "1", "2"}}
	restoreDepStatus(s, task)

	assert.Equal(t, map[string]model.StatusValue{"2": model.StatusFailure}, task.DepStatus)
}

func setWorkflowState(t *testing.T, s store.Store, id string, state model.StatusValue) {
	t.Helper()

	var workflowID int64
	_, err := fmt.Sscan(id, &workflowID)
	require.NoError(t, err)
	workflow, err := s.WorkflowLoad(workflowID)
	require.NoError(t, err)
	workflow.State = state
	require.NoError(t, s.WorkflowUpdate(workflow))
}

// A server restart must bring the queue back as it was: queued, waiting and
// running tasks, finished dependencies (#7154) and the paused flag. Running
// tasks only get a new lease, so their agents can reconnect and go on.
func TestWithTaskStoreRestart(t *testing.T) {
	s := newTestStore(t)
	tasks := createTasks(t, s, "running", "canceled", "limited", "dependent", "done", "after-done", "failed", "after-failed", "queued")
	for _, name := range []string{"running", "limited"} {
		tasks[name].ConcurrencyGroup, tasks[name].ConcurrencyLimit = "group", 1
	}
	tasks["dependent"].Dependencies = []string{tasks["running"].ID}
	tasks["after-done"].Dependencies = []string{tasks["done"].ID}
	tasks["after-failed"].Dependencies = []string{tasks["failed"].ID}

	// first server: fill the queue, then pause and stop it
	ctx1, cancel1 := context.WithCancelCause(t.Context())
	fifo1, _ := NewMemoryQueue(ctx1, 0).(*fifo)
	q1 := WithTaskStore(ctx1, fifo1, s)
	require.False(t, q1.Info(ctx1).Paused)
	require.NoError(t, q1.PushAtOnce(ctx1, slices.Collect(maps.Values(tasks))))
	polls := map[string]int64{"running": 1, "canceled": 2, "done": 2, "failed": 2}
	var wg sync.WaitGroup
	for name, agentID := range polls {
		wg.Go(func() {
			got, err := q1.Poll(ctx1, agentID, onlyTask(tasks[name]))
			assert.NoError(t, err)
			assert.Equal(t, tasks[name], got)
		})
	}
	wg.Wait()
	for name := range polls {
		setWorkflowState(t, s, tasks[name].ID, model.StatusRunning)
	}
	setWorkflowState(t, s, tasks["done"].ID, model.StatusSuccess)
	require.NoError(t, q1.Done(ctx1, tasks["done"].ID, model.StatusSuccess))
	setWorkflowState(t, s, tasks["failed"].ID, model.StatusFailure)
	require.NoError(t, q1.Error(ctx1, tasks["failed"].ID, errors.New("workflow failed")))
	waitForProcess()
	before := q1.Info(ctx1)
	oldLease := leaseOf(fifo1, tasks["running"])
	q1.Pause()
	cancel1(nil)

	// second server: same queue, new lease
	ctx2, cancel2 := context.WithCancelCause(t.Context())
	defer cancel2(nil)
	fifo2, _ := NewMemoryQueue(ctx2, 0).(*fifo)
	q2 := WithTaskStore(ctx2, fifo2, s)
	require.True(t, q2.Info(ctx2).Paused)
	canceled := make(chan error, 1)
	go func() { canceled <- q2.Wait(ctx2, tasks["canceled"].ID) }()
	q2.Resume()
	waitForProcess()
	after := q2.Info(ctx2)
	assert.Equal(t, queueState(before), queueState(after))
	assert.True(t, leaseOf(fifo2, tasks["running"]).After(oldLease), "restored task must get a new lease")
	assert.True(t, queuedTask(after, tasks["after-done"]).ShouldRun())
	assert.False(t, queuedTask(after, tasks["after-failed"]).ShouldRun())

	// the agents of the running tasks can go on
	assert.NoError(t, q2.Extend(ctx2, 1, tasks["running"].ID))
	pollCtx, pollCancel := context.WithTimeout(ctx2, 2*processTimeInterval)
	got, _ := q2.Poll(pollCtx, 3, onlyTask(tasks["limited"]))
	pollCancel()
	assert.Nil(t, got, "restored running task must still use its concurrency slot")
	require.NoError(t, q2.ErrorAtOnce(ctx2, []string{tasks["canceled"].ID}, ErrCancel))
	select {
	case err := <-canceled:
		assert.ErrorIs(t, err, ErrCancel)
	case <-time.After(time.Second):
		t.Fatal("cancel did not reach Wait")
	}

	// finishing the running task frees its slot and its dependent
	require.NoError(t, q2.Done(ctx2, tasks["running"].ID, model.StatusSuccess))
	pollCtx, pollCancel = context.WithTimeout(ctx2, time.Second)
	got, err := q2.Poll(pollCtx, 3, onlyTask(tasks["limited"]))
	pollCancel()
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Eventually(t, func() bool {
		dependent := queuedTask(q2.Info(ctx2), tasks["dependent"])
		return dependent != nil && dependent.DepStatus[tasks["running"].ID] == model.StatusSuccess
	}, time.Second, 10*time.Millisecond, "dependent must be released")

	stored, err := s.TaskList()
	require.NoError(t, err)
	var names []string
	for _, task := range stored {
		names = append(names, task.Name)
	}
	assert.ElementsMatch(t, []string{"limited", "dependent", "after-done", "after-failed", "queued"}, names)
	cancel2(nil)

	// third server: the agent of the restored task does not come back
	ctx3, cancel3 := context.WithCancelCause(t.Context())
	defer cancel3(nil)
	q3 := WithTaskStore(ctx3, NewMemoryQueue(ctx3, 200*time.Millisecond), s)
	running := q3.Info(ctx3).Running
	require.Len(t, running, 1)
	assert.Equal(t, int64(3), running[0].AgentID)
	require.Eventually(t, func() bool {
		return queuedTask(q3.Info(ctx3), tasks["limited"]) != nil
	}, 2*time.Second, 10*time.Millisecond, "expired task must be queued again")
}

// newTestStore returns a migrated in-memory SQLite store, or the database set
// by WOODPECKER_DATABASE_DRIVER and WOODPECKER_DATABASE_DATASOURCE.
func newTestStore(t *testing.T) store.Store {
	t.Helper()

	opts := &store.Opts{
		Driver: "sqlite3",
		Config: ":memory:",
		// in-memory sqlite loses its schema when the pool drops the connection
		XORM: store.XORM{MaxOpenConns: 1, MaxIdleConns: 1},
	}
	if driver := os.Getenv("WOODPECKER_DATABASE_DRIVER"); driver != "" {
		opts = &store.Opts{Driver: driver, Config: os.Getenv("WOODPECKER_DATABASE_DATASOURCE")}
	}
	s, err := datastore.NewEngine(opts)
	require.NoError(t, err)
	require.NoError(t, s.Migrate(t.Context(), true))
	t.Cleanup(func() { _ = s.Close() })

	// a shared database must not keep the queue between tests
	clearQueue := func() {
		tasks, err := s.TaskList()
		require.NoError(t, err)
		for _, task := range tasks {
			require.NoError(t, s.TaskDelete(task.ID))
		}
		_ = s.ServerConfigDelete(serverConfigQueuePaused)
	}
	clearQueue()
	t.Cleanup(clearQueue)
	return s
}

// createTasks stores a pending workflow per name and returns their queue
// tasks by name.
func createTasks(t *testing.T, s store.Store, names ...string) map[string]*model.Task {
	t.Helper()

	// a new pipeline id per call, as a shared database keeps old workflows
	pipelineID := time.Now().UnixNano()
	workflows := make([]*model.Workflow, len(names))
	for i, name := range names {
		workflows[i] = &model.Workflow{PipelineID: pipelineID, PID: i + 1, Name: name, State: model.StatusPending}
	}
	require.NoError(t, s.WorkflowsCreate(workflows))

	tasks := make(map[string]*model.Task, len(names))
	for _, workflow := range workflows {
		tasks[workflow.Name] = &model.Task{
			ID:        fmt.Sprint(workflow.ID),
			Name:      workflow.Name,
			DepStatus: make(map[string]model.StatusValue),
		}
	}
	return tasks
}

// queueState lists the tasks of each queue list in a comparable form.
func queueState(info InfoT) map[string][]string {
	state := make(map[string][]string)
	for list, tasks := range map[string][]*model.Task{"pending": info.Pending, "waiting": info.WaitingOnDeps, "running": info.Running} {
		for _, task := range tasks {
			state[list] = append(state[list], fmt.Sprintf("%s agent=%d deps=%v status=%v group=%s/%d",
				task.Name, task.AgentID, task.Dependencies, task.DepStatus, task.ConcurrencyGroup, task.ConcurrencyLimit))
		}
		slices.Sort(state[list])
	}
	return state
}

func queuedTask(info InfoT, task *model.Task) *model.Task {
	for _, queued := range append(info.Pending, info.WaitingOnDeps...) {
		if queued.ID == task.ID {
			return queued
		}
	}
	return nil
}

func leaseOf(q *fifo, task *model.Task) time.Time {
	q.Lock()
	defer q.Unlock()
	if running, ok := q.running[task.ID]; ok {
		return running.deadline
	}
	return time.Time{}
}

func onlyTask(task *model.Task) func(*model.Task) (bool, int) {
	return func(candidate *model.Task) (bool, int) { return candidate.ID == task.ID, 1 }
}
