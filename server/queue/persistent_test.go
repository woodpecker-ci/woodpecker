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

// A task that lingers in memory after its workflow was canceled and its backup
// removed must be dropped on Poll instead of being handed to the agent,
// otherwise it loops forever (re-poll, illegal-instruction, resubmit).
func TestPersistentQueuePollDropsStaleTask(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	store := store_mocks.NewMockStore(t)
	store.EXPECT().TaskDelete("1").Return(types.ErrRecordNotExist).Once()
	store.EXPECT().WorkflowLoad(int64(1)).Return(&model.Workflow{ID: 1, State: model.StatusCanceled}, nil).Once()

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
	store.EXPECT().TaskDelete("1").Return(nil).Once()
	store.EXPECT().WorkflowLoad(int64(1)).Return(nil, types.ErrRecordNotExist).Once()

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
			ctx, cancel, q := setupTestQueue(t)
			defer cancel(nil)

			store := store_mocks.NewMockStore(t)
			store.EXPECT().TaskDelete("1").Return(nil).Once()
			store.EXPECT().WorkflowLoad(int64(1)).Return(&model.Workflow{ID: 1, State: state}, nil).Once()

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

// A task that is still present in the backup store and whose workflow is still
// active is polled normally.
func TestPersistentQueuePollReturnsLiveTask(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	store := store_mocks.NewMockStore(t)
	store.EXPECT().TaskDelete("1").Return(nil).Once()
	store.EXPECT().WorkflowLoad(int64(1)).Return(&model.Workflow{ID: 1, State: model.StatusPending}, nil).Once()

	pq := &persistentQueue{Queue: q, store: store}

	task := genDummyTask()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task}))

	got, err := pq.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "1", got.ID)
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
	store.EXPECT().TaskDelete("1").Return(nil).Once()
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

func TestWithTaskStoreRestoresPausedState(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	store := store_mocks.NewMockStore(t)
	store.EXPECT().TaskList().Return(nil, nil).Once()
	store.EXPECT().ServerConfigGet(serverConfigQueuePaused).Return("true", nil).Once()

	pq := WithTaskStore(ctx, q, store)

	assert.True(t, pq.Info(ctx).Paused)
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

func TestWithTaskStoreRestoresPausedStateAfterRestart(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	store := store_mocks.NewMockStore(t)
	store.EXPECT().ServerConfigSet(serverConfigQueuePaused, "true").Return(nil).Once()

	pq := &persistentQueue{Queue: q, store: store}
	pq.Pause()
	assert.True(t, pq.Info(ctx).Paused)

	// Fresh in-memory queue after process restart; paused flag comes from store.
	ctx2, cancel2, q2 := setupTestQueue(t)
	defer cancel2(nil)

	store.EXPECT().TaskList().Return(nil, nil).Once()
	store.EXPECT().ServerConfigGet(serverConfigQueuePaused).Return("true", nil).Once()

	restored := WithTaskStore(ctx2, q2, store)
	assert.True(t, restored.Info(ctx2).Paused)
}

func newSQLiteStore(t *testing.T) store.Store {
	t.Helper()

	s, err := datastore.NewEngine(&store.Opts{
		Driver: "sqlite3",
		Config: ":memory:",
		// in-memory sqlite loses its schema when the pool drops the connection
		XORM: store.XORM{MaxOpenConns: 1, MaxIdleConns: 1},
	})
	require.NoError(t, err)
	require.NoError(t, s.Migrate(t.Context(), true))
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// A dependency can finish while its dependent task is still queued. The
// queue tracks that outcome only in memory, so a reload (server restart or
// HA leader change) must restore it, otherwise Task.ShouldRun() sees an empty
// DepStatus and vacuously lets the dependent run (#7154).
func TestWithTaskStoreRestoresDepStatusAfterReload(t *testing.T) {
	failed := func(ctx context.Context, q Queue, id string) error {
		return q.Error(ctx, id, errors.New("workflow finished with error"))
	}
	done := func(state model.StatusValue) func(context.Context, Queue, string) error {
		return func(ctx context.Context, q Queue, id string) error {
			return q.Done(ctx, id, state)
		}
	}

	tests := []struct {
		name     string
		depState model.StatusValue
		finish   func(context.Context, Queue, string) error
		runOn    []string
		wantRun  bool
	}{
		{
			name:     "failed dependency blocks success gate",
			depState: model.StatusFailure,
			finish:   failed,
			runOn:    []string{string(model.StatusSuccess)},
			wantRun:  false,
		},
		{
			name:     "succeeded dependency lets success gate run",
			depState: model.StatusSuccess,
			finish:   done(model.StatusSuccess),
			runOn:    []string{string(model.StatusSuccess)},
			wantRun:  true,
		},
		{
			name:     "succeeded dependency blocks failure-only gate",
			depState: model.StatusSuccess,
			finish:   done(model.StatusSuccess),
			runOn:    []string{string(model.StatusFailure)},
			wantRun:  false,
		},
		{
			name:     "skipped dependency blocks success gate",
			depState: model.StatusSkipped,
			finish:   done(model.StatusSkipped),
			runOn:    []string{string(model.StatusSuccess)},
			wantRun:  false,
		},
		{
			name:     "killed dependency blocks success gate",
			depState: model.StatusKilled,
			finish:   done(model.StatusKilled),
			runOn:    []string{string(model.StatusSuccess)},
			wantRun:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newSQLiteStore(t)
			dep, gate := createDepAndGate(t, s, tt.runOn)

			// first server: the dependency runs and finishes, the gate stays queued
			ctx1, cancel1 := context.WithCancelCause(t.Context())
			q1 := WithTaskStore(ctx1, NewMemoryQueue(ctx1), s)
			require.NoError(t, q1.PushAtOnce(ctx1, []*model.Task{dep, gate}))
			polled, err := q1.Poll(ctx1, 1, filterFnTrue)
			require.NoError(t, err)
			require.Equal(t, dep.ID, polled.ID)
			setWorkflowState(t, s, dep.ID, tt.depState)
			require.NoError(t, tt.finish(ctx1, q1, dep.ID))
			cancel1(nil)

			// second server: reload the still queued gate from the store
			ctx2, cancel2 := context.WithCancelCause(t.Context())
			defer cancel2(nil)
			q2 := WithTaskStore(ctx2, NewMemoryQueue(ctx2), s)

			reloaded := queuedTask(t, q2.Info(ctx2), gate.ID)
			assert.Equal(t, tt.depState, reloaded.DepStatus[dep.ID])
			assert.Equal(t, tt.wantRun, reloaded.ShouldRun())
		})
	}
}

// A dependency that has not finished yet is left to the live queue, which
// sets its status once the agent reports back.
func TestWithTaskStoreKeepsUnfinishedDepStatusEmpty(t *testing.T) {
	s := newSQLiteStore(t)
	dep, gate := createDepAndGate(t, s, []string{string(model.StatusSuccess)})
	require.NoError(t, s.TaskInsert(gate))
	setWorkflowState(t, s, dep.ID, model.StatusRunning)

	ctx, cancel := context.WithCancelCause(t.Context())
	defer cancel(nil)
	q := WithTaskStore(ctx, NewMemoryQueue(ctx), s)

	reloaded := queuedTask(t, q.Info(ctx), gate.ID)
	assert.NotContains(t, reloaded.DepStatus, dep.ID)
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

// createDepAndGate stores a pending dependency workflow and a pending gate
// workflow depending on it, and returns their queue tasks.
func createDepAndGate(t *testing.T, s store.Store, gateRunOn []string) (dep, gate *model.Task) {
	t.Helper()

	workflows := []*model.Workflow{
		{PipelineID: 1, PID: 1, Name: "dep", State: model.StatusPending},
		{PipelineID: 1, PID: 2, Name: "gate", State: model.StatusPending},
	}
	require.NoError(t, s.WorkflowsCreate(workflows))

	dep = &model.Task{
		ID:        fmt.Sprint(workflows[0].ID),
		RunOn:     []string{string(model.StatusSuccess)},
		DepStatus: make(map[string]model.StatusValue),
	}
	gate = &model.Task{
		ID:           fmt.Sprint(workflows[1].ID),
		Dependencies: []string{dep.ID},
		RunOn:        gateRunOn,
		DepStatus:    make(map[string]model.StatusValue),
	}
	return dep, gate
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

func queuedTask(t *testing.T, info InfoT, id string) *model.Task {
	t.Helper()

	for _, task := range append(info.Pending, info.WaitingOnDeps...) {
		if task.ID == id {
			return task
		}
	}
	require.FailNowf(t, "task not queued", "task %s not in %s", id, info.String())
	return nil
}

type persistentValidationResult struct {
	task *model.Task
	err  error
}

// Wait evaluates Done after capturing its running entry. This handshake lets
// the test expire that lease only after the waiter has registered for it.
type registeredWaitContext struct {
	context.Context
	registered chan struct{}
}

func (c registeredWaitContext) Done() <-chan struct{} {
	close(c.registered)
	return c.Context.Done()
}

func withWaitRegistration(ctx context.Context) (context.Context, <-chan struct{}) {
	registered := make(chan struct{})
	return registeredWaitContext{Context: ctx, registered: registered}, registered
}

func expirePersistentLease(t *testing.T, ctx context.Context, q *fifo, id string) <-chan error {
	t.Helper()
	waitCtx, registered := withWaitRegistration(ctx)
	waitResult := make(chan error, 1)
	q.Lock()
	agentID := q.running[id].item.AgentID
	q.Unlock()
	go func(waitCtx context.Context) { waitResult <- q.Wait(waitCtx, agentID, id) }(waitCtx)
	select {
	case <-registered:
	case <-ctx.Done():
		t.Fatal("waiter did not register")
	}

	q.Lock()
	old := q.running[id]
	old.deadline = time.Now().Add(-time.Second)
	q.Unlock()
	select {
	case <-old.done:
	case <-ctx.Done():
		t.Fatal("lease did not expire")
	}
	return waitResult
}

func TestPersistentQueueRecoversExpiredLease(t *testing.T) {
	for _, state := range []model.StatusValue{model.StatusPending, model.StatusRunning} {
		t.Run(string(state), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			q, ok := NewMemoryQueue(ctx).(*fifo)
			require.True(t, ok)
			store := store_mocks.NewMockStore(t)
			task := genDummyTask()
			store.EXPECT().TaskInsert(task).Return(nil).Once()
			store.EXPECT().TaskDelete("1").Return(nil).Once()
			store.EXPECT().TaskDelete("1").Return(types.ErrRecordNotExist).Times(3)
			store.EXPECT().WorkflowLoad(int64(1)).Return(&model.Workflow{ID: 1, State: state}, nil).Times(3)
			pq := &persistentQueue{Queue: q, store: store}
			require.NoError(t, pq.PushAtOnce(ctx, []*model.Task{task}))
			got, err := pq.Poll(ctx, 1, filterFnTrue)
			require.NoError(t, err)
			require.NotNil(t, got)

			for agentID := int64(2); agentID <= 3; agentID++ {
				oldWaiter := expirePersistentLease(t, ctx, q, "1")
				assert.Equal(t, 1, q.Info(ctx).Stats.Pending)
				got, err = pq.Poll(ctx, agentID, filterFnTrue)
				require.NoError(t, err)
				require.NotNil(t, got, "missing backup must not drop an expired live workflow")
				assert.Equal(t, agentID, got.AgentID)
				select {
				case waitErr := <-oldWaiter:
					assert.ErrorIs(t, waitErr, ErrTaskExpired)
				case <-ctx.Done():
					t.Fatal("old lease waiter did not return")
				}
				assert.Equal(t, 1, q.Info(ctx).Stats.Running)
			}
			require.NoError(t, pq.Done(ctx, "1", model.StatusSuccess))
			assert.Zero(t, q.Info(ctx).Stats.Pending)
			assert.Zero(t, q.Info(ctx).Stats.Running)
		})
	}
}

func TestPersistentQueueMissingBackupFiltersStaleWorkflows(t *testing.T) {
	states := []model.StatusValue{
		model.StatusSuccess, model.StatusFailure, model.StatusKilled,
		model.StatusCanceled, model.StatusSkipped, model.StatusError,
		model.StatusDeclined, "missing",
	}
	for _, state := range states {
		t.Run(string(state), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()
			q, ok := NewMemoryQueue(ctx).(*fifo)
			require.True(t, ok)
			store := store_mocks.NewMockStore(t)
			store.EXPECT().TaskDelete("1").Return(types.ErrRecordNotExist).Once()
			if state == "missing" {
				store.EXPECT().WorkflowLoad(int64(1)).Return(nil, types.ErrRecordNotExist).Once()
			} else {
				store.EXPECT().WorkflowLoad(int64(1)).Return(&model.Workflow{ID: 1, State: state}, nil).Once()
			}
			pq := &persistentQueue{Queue: q, store: store}
			require.NoError(t, q.PushAtOnce(ctx, []*model.Task{genDummyTask()}))
			got, err := pq.Poll(ctx, 1, filterFnTrue)
			require.NoError(t, err)
			assert.Nil(t, got)
			assert.Zero(t, q.Info(ctx).Stats.Pending)
			assert.Zero(t, q.Info(ctx).Stats.Running)
		})
	}
}

func TestPersistentQueuePollRejectsCanceledAssignment(t *testing.T) {
	for _, backupMissing := range []bool{false, true} {
		name := "present"
		if backupMissing {
			name = "missing"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()
			q, ok := NewMemoryQueue(ctx).(*fifo)
			require.True(t, ok)
			store := store_mocks.NewMockStore(t)
			validating, release := make(chan struct{}), make(chan struct{})
			unblock := sync.OnceFunc(func() { close(release) })
			defer unblock()
			var deleteErr error
			if backupMissing {
				deleteErr = types.ErrRecordNotExist
			}
			store.EXPECT().TaskDelete("1").Run(func(string) {
				close(validating)
				select {
				case <-release:
				case <-ctx.Done():
				}
			}).Return(deleteErr).Once()
			store.EXPECT().TaskDelete("1").Return(types.ErrRecordNotExist).Once()
			// CancelWorkflows evicts first; a running workflow is not marked
			// terminal until the agent stops. Its DB state can still be Running.
			store.EXPECT().WorkflowLoad(int64(1)).Return(&model.Workflow{ID: 1, State: model.StatusRunning}, nil).Maybe()
			pq := &persistentQueue{Queue: q, store: store}
			require.NoError(t, q.PushAtOnce(ctx, []*model.Task{genDummyTask()}))
			result := make(chan persistentValidationResult, 1)
			pollDone := make(chan struct{})
			go func() {
				defer close(pollDone)
				got, err := pq.Poll(ctx, 1, filterFnTrue)
				result <- persistentValidationResult{got, err}
			}()
			defer func() {
				unblock()
				select {
				case <-pollDone:
				case <-ctx.Done():
				}
			}()
			select {
			case <-validating:
			case <-ctx.Done():
				t.Fatal("Poll did not reach storage validation")
			}
			// This must complete while storage is blocked: validation cannot
			// hold the FIFO mutex or resurrect this canceled assignment.
			canceled := make(chan error, 1)
			go func() { canceled <- pq.ErrorAtOnce(ctx, []string{"1"}, ErrCancel) }()
			select {
			case cancelErr := <-canceled:
				require.NoError(t, cancelErr)
				require.NoError(t, ctx.Err(), "cancellation must finish before storage is released")
			case <-ctx.Done():
				t.Fatal("cancellation blocked on storage validation")
			}
			unblock()
			select {
			case got := <-result:
				require.NoError(t, got.err)
				assert.Nil(t, got.task)
			case <-ctx.Done():
				t.Fatal("Poll did not return")
			}
			assert.Zero(t, q.Info(ctx).Stats.Running)
		})
	}
}

func TestPersistentQueuePollRejectsReplacedLease(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	q, ok := NewMemoryQueue(ctx).(*fifo)
	require.True(t, ok)
	store := store_mocks.NewMockStore(t)
	validating, release := make(chan struct{}), make(chan struct{})
	unblock := sync.OnceFunc(func() { close(release) })
	defer unblock()
	store.EXPECT().TaskDelete("1").Run(func(string) {
		close(validating)
		select {
		case <-release:
		case <-ctx.Done():
		}
	}).Return(nil).Once()
	store.EXPECT().TaskDelete("1").Return(types.ErrRecordNotExist).Times(2)
	store.EXPECT().WorkflowLoad(int64(1)).Return(&model.Workflow{ID: 1, State: model.StatusRunning}, nil).Maybe()
	pq := &persistentQueue{Queue: q, store: store}
	require.NoError(t, q.PushAtOnce(ctx, []*model.Task{genDummyTask()}))
	oldResult := make(chan persistentValidationResult, 1)
	pollDone := make(chan struct{})
	go func() {
		defer close(pollDone)
		got, err := pq.Poll(ctx, 1, filterFnTrue)
		oldResult <- persistentValidationResult{got, err}
	}()
	defer func() {
		unblock()
		select {
		case <-pollDone:
		case <-ctx.Done():
		}
	}()
	select {
	case <-validating:
	case <-ctx.Done():
		t.Fatal("Poll did not reach storage validation")
	}
	oldWaiter := expirePersistentLease(t, ctx, q, "1")
	// The same agent may acquire a new lease while the previous Poll is
	// blocked. Checking only task ID and agent ID would accept both results.
	got, err := pq.Poll(ctx, 1, filterFnTrue)
	require.NoError(t, err)
	require.NotNil(t, got)
	unblock()
	select {
	case old := <-oldResult:
		require.NoError(t, old.err)
		assert.Nil(t, old.task)
	case <-ctx.Done():
		t.Fatal("old Poll did not return")
	}
	select {
	case waitErr := <-oldWaiter:
		assert.ErrorIs(t, waitErr, ErrTaskExpired)
	case <-ctx.Done():
		t.Fatal("old lease waiter did not return")
	}
	assert.Equal(t, 1, q.Info(ctx).Stats.Running)
	require.NoError(t, pq.Done(ctx, "1", model.StatusSuccess))
}

func TestPersistentQueueTransientStoreErrorsDoNotDropTask(t *testing.T) {
	for _, stage := range []string{"delete", "load"} {
		t.Run(stage, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()
			q, ok := NewMemoryQueue(ctx).(*fifo)
			require.True(t, ok)
			store := store_mocks.NewMockStore(t)
			transient := errors.New("temporary database failure")
			if stage == "delete" {
				store.EXPECT().TaskDelete("1").Return(transient).Once()
				store.EXPECT().WorkflowLoad(int64(1)).Return(&model.Workflow{ID: 1, State: model.StatusRunning}, nil).Maybe()
			} else {
				store.EXPECT().TaskDelete("1").Return(types.ErrRecordNotExist).Once()
				store.EXPECT().WorkflowLoad(int64(1)).Return(nil, transient).Once()
			}
			pq := &persistentQueue{Queue: q, store: store}
			require.NoError(t, q.PushAtOnce(ctx, []*model.Task{genDummyTask()}))
			got, err := pq.Poll(ctx, 1, filterFnTrue)
			require.NoError(t, err)
			require.NotNil(t, got, "transient store failure must not complete the lease as stale")
			assert.Equal(t, 1, q.Info(ctx).Stats.Running)
		})
	}
}

func TestPersistentQueueMissingBackupRejectsInvalidWorkflowID(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	q, ok := NewMemoryQueue(ctx).(*fifo)
	require.True(t, ok)
	store := store_mocks.NewMockStore(t)
	store.EXPECT().TaskDelete("invalid").Return(types.ErrRecordNotExist).Once()
	pq := &persistentQueue{Queue: q, store: store}
	require.NoError(t, q.PushAtOnce(ctx, []*model.Task{{ID: "invalid"}}))
	got, err := pq.Poll(ctx, 1, filterFnTrue)
	require.NoError(t, err)
	assert.Nil(t, got)
	assert.Zero(t, q.Info(ctx).Stats.Running)
}

func TestPersistentQueueCanceledPollReturns(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	q, ok := NewMemoryQueue(ctx).(*fifo)
	require.True(t, ok)
	store := store_mocks.NewMockStore(t)
	pq := &persistentQueue{Queue: q, store: store}
	cancel(nil)
	got, err := pq.Poll(ctx, 1, filterFnTrue)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, got)
	assert.Zero(t, q.Info(ctx).Stats.Workers)
}

func TestPersistentQueueDeleteErrorStillFiltersStaleWorkflow(t *testing.T) {
	for _, state := range []model.StatusValue{model.StatusCanceled, "missing"} {
		t.Run(string(state), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()
			q, ok := NewMemoryQueue(ctx).(*fifo)
			require.True(t, ok)
			store := store_mocks.NewMockStore(t)
			store.EXPECT().TaskDelete("1").Return(errors.New("temporary delete failure")).Once()
			if state == "missing" {
				store.EXPECT().WorkflowLoad(int64(1)).Return(nil, types.ErrRecordNotExist).Once()
			} else {
				store.EXPECT().WorkflowLoad(int64(1)).Return(&model.Workflow{ID: 1, State: state}, nil).Once()
			}
			pq := &persistentQueue{Queue: q, store: store}
			require.NoError(t, q.PushAtOnce(ctx, []*model.Task{genDummyTask()}))
			got, err := pq.Poll(ctx, 1, filterFnTrue)
			require.NoError(t, err)
			assert.Nil(t, got, "a failed backup delete must not bypass known stale workflow state")
			assert.Zero(t, q.Info(ctx).Stats.Running)
		})
	}
}

// Poll creates its cancellation context while holding the FIFO mutex, before
// registering its worker. Once both contexts have been observed, Info acquires
// that mutex to confirm registration completed before dispatch resumes.
type persistentPollRegistrationContext struct {
	context.Context
	registered chan struct{}
	once       sync.Once
}

func (c *persistentPollRegistrationContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.registered) })
	return c.Context.Done()
}

func TestPersistentQueueExpiryAssignsOnlyOneRecoveryPoller(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	q, ok := NewMemoryQueue(ctx).(*fifo)
	require.True(t, ok)
	store := store_mocks.NewMockStore(t)
	task := genDummyTask()
	store.EXPECT().TaskInsert(task).Return(nil).Once()
	store.EXPECT().TaskDelete("1").Return(nil).Once()
	store.EXPECT().TaskDelete("1").Return(types.ErrRecordNotExist).Times(2)
	store.EXPECT().WorkflowLoad(int64(1)).Return(&model.Workflow{ID: 1, State: model.StatusRunning}, nil).Twice()
	pq := &persistentQueue{Queue: q, store: store}
	require.NoError(t, pq.PushAtOnce(ctx, []*model.Task{task}))
	assigned, err := pq.Poll(ctx, 1, filterFnTrue)
	require.NoError(t, err)
	require.NotNil(t, assigned)
	oldWaiter := expirePersistentLease(t, ctx, q, "1")
	require.Equal(t, 1, q.Info(ctx).Stats.Pending)

	q.Pause()
	pollCtx, stopPollers := context.WithCancelCause(ctx)
	defer stopPollers(nil)
	type pollResult struct {
		agentID int64
		task    *model.Task
		err     error
	}
	results := make(chan pollResult, 2)
	registrations := make([]*persistentPollRegistrationContext, 0, 2)
	for _, agentID := range []int64{2, 3} {
		registration := &persistentPollRegistrationContext{
			Context:    pollCtx,
			registered: make(chan struct{}),
		}
		registrations = append(registrations, registration)
		go func() {
			got, pollErr := pq.Poll(registration, agentID, filterFnTrue)
			results <- pollResult{agentID: agentID, task: got, err: pollErr}
		}()
	}
	for _, registration := range registrations {
		select {
		case <-registration.registered:
		case <-ctx.Done():
			t.Fatal("recovery poller did not register")
		}
	}
	require.Equal(t, 2, q.Info(ctx).Stats.Workers)
	q.Resume()

	var winner pollResult
	select {
	case winner = <-results:
	case <-ctx.Done():
		t.Fatal("no poller recovered the expired task")
	}
	require.NoError(t, winner.err)
	require.NotNil(t, winner.task)
	require.Equal(t, task.ID, winner.task.ID)
	require.Equal(t, winner.agentID, winner.task.AgentID)
	info := q.Info(ctx)
	require.Equal(t, 1, info.Stats.Running)
	require.Equal(t, 1, info.Stats.Workers)
	require.Zero(t, info.Stats.Pending)

	stopPollers(nil)
	select {
	case loser := <-results:
		require.NotEqual(t, winner.agentID, loser.agentID)
		require.Nil(t, loser.task, "one expiry must create only one new assignment")
		require.ErrorIs(t, loser.err, context.Canceled)
	case <-ctx.Done():
		t.Fatal("unassigned poller did not return after cancellation")
	}
	require.NoError(t, pq.Done(ctx, task.ID, model.StatusSuccess))
	select {
	case waitErr := <-oldWaiter:
		require.ErrorIs(t, waitErr, ErrTaskExpired)
	case <-ctx.Done():
		t.Fatal("expired generation waiter did not return")
	}
	info = q.Info(ctx)
	require.Zero(t, info.Stats.Workers)
	require.Zero(t, info.Stats.Pending)
	require.Zero(t, info.Stats.Running)
	require.Zero(t, info.Stats.WaitingOnDeps)
}

func TestPersistentQueueCancelExpiredPendingPreservesOtherTasks(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	q, ok := NewMemoryQueue(ctx).(*fifo)
	require.True(t, ok)
	store := store_mocks.NewMockStore(t)
	expired := &model.Task{
		ID: "1", PipelineID: 1, Created: 1,
		ConcurrencyGroup: "repo:deploy", ConcurrencyLimit: 1,
	}
	dependent := &model.Task{
		ID: "2", PipelineID: 1,
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
		RunOn:        []string{"success", "failure"},
	}
	unrelated := &model.Task{
		ID: "3", PipelineID: 2, Created: 2,
		ConcurrencyGroup: "repo:deploy", ConcurrencyLimit: 1,
	}
	for _, task := range []*model.Task{expired, dependent, unrelated} {
		store.EXPECT().TaskInsert(task).Return(nil).Once()
		store.EXPECT().TaskDelete(task.ID).Return(nil).Once()
		store.EXPECT().TaskDelete(task.ID).Return(types.ErrRecordNotExist).Once()
	}
	store.EXPECT().WorkflowLoad(int64(1)).Return(&model.Workflow{ID: 1, State: model.StatusRunning}, nil).Once()
	store.EXPECT().WorkflowLoad(int64(2)).Return(&model.Workflow{ID: 2, State: model.StatusPending}, nil).Once()
	store.EXPECT().WorkflowLoad(int64(3)).Return(&model.Workflow{ID: 3, State: model.StatusPending}, nil).Once()
	pq := &persistentQueue{Queue: q, store: store}
	require.NoError(t, pq.PushAtOnce(ctx, []*model.Task{expired, dependent, unrelated}))
	assigned, err := pq.Poll(ctx, 1, filterFnTrue)
	require.NoError(t, err)
	require.NotNil(t, assigned)
	require.Equal(t, expired.ID, assigned.ID)
	oldWaiter := expirePersistentLease(t, ctx, q, expired.ID)

	q.Lock()
	reserved := !q.canRunConcurrent(unrelated)
	q.Unlock()
	require.True(t, reserved, "the expired pending task must retain its concurrency reservation")
	require.NoError(t, pq.ErrorAtOnce(ctx, []string{expired.ID}, ErrCancel))
	q.Lock()
	released := q.canRunConcurrent(unrelated)
	dependencyStatus := dependent.DepStatus[expired.ID]
	q.Unlock()
	require.True(t, released, "cancellation must free the expired task's reservation")
	require.Equal(t, model.StatusKilled, dependencyStatus)

	got, err := pq.Poll(ctx, 2, func(task *model.Task) (bool, int) {
		return task.ID == unrelated.ID, 1
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, unrelated.ID, got.ID)
	require.NoError(t, pq.Done(ctx, got.ID, model.StatusSuccess))
	got, err = pq.Poll(ctx, 3, filterFnTrue)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, dependent.ID, got.ID, "the canceled task must not reappear")
	require.Equal(t, model.StatusKilled, got.DepStatus[expired.ID])
	require.NoError(t, pq.Done(ctx, got.ID, model.StatusSuccess))
	select {
	case waitErr := <-oldWaiter:
		require.ErrorIs(t, waitErr, ErrTaskExpired)
	case <-ctx.Done():
		t.Fatal("expired generation waiter did not return")
	}
	info := q.Info(ctx)
	require.Zero(t, info.Stats.Workers)
	require.Zero(t, info.Stats.Pending)
	require.Zero(t, info.Stats.Running)
	require.Zero(t, info.Stats.WaitingOnDeps)
}

type wrappedPollTestQueue struct {
	Queue
	task      *model.Task
	err       error
	doneID    string
	doneState model.StatusValue
}

func (q *wrappedPollTestQueue) Poll(context.Context, int64, func(*model.Task) (bool, int)) (*model.Task, error) {
	return q.task, q.err
}

func (q *wrappedPollTestQueue) Done(_ context.Context, id string, state model.StatusValue) error {
	q.doneID, q.doneState = id, state
	return nil
}

// WithTaskStore can wrap queues other than FIFO. A partial Poll result must
// still be validated, without losing the wrapped queue's error for live work.
func TestPersistentQueueWrappedPollError(t *testing.T) {
	for _, state := range []model.StatusValue{model.StatusPending, model.StatusCanceled, "no-task"} {
		t.Run(string(state), func(t *testing.T) {
			ctx := t.Context()
			store := store_mocks.NewMockStore(t)
			pollErr := errors.New("wrapped queue poll error")
			var task *model.Task
			if state != "no-task" {
				task = genDummyTask()
				store.EXPECT().TaskDelete(task.ID).Return(nil).Once()
				store.EXPECT().WorkflowLoad(int64(1)).Return(&model.Workflow{ID: 1, State: state}, nil).Once()
			}
			q := &wrappedPollTestQueue{task: task, err: pollErr}
			pq := &persistentQueue{Queue: q, store: store}

			got, err := pq.Poll(ctx, 7, filterFnTrue)
			if state == model.StatusCanceled {
				assert.Nil(t, got, "stale partial results must still be removed")
				assert.NoError(t, err)
				assert.Equal(t, task.ID, q.doneID)
				assert.Equal(t, state, q.doneState)
			} else {
				assert.Same(t, task, got)
				assert.ErrorIs(t, err, pollErr)
				assert.Empty(t, q.doneID)
			}
		})
	}
}
