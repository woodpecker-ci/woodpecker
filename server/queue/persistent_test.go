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
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
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
	store.EXPECT().TaskDelete("1").Return(types.ErrRecordNotExist).Once()

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
