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

//go:build test

package queue

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/datastore"
)

func openQueueSQLiteStore(t *testing.T, ctx context.Context, path string) (store.Store, func()) {
	t.Helper()

	s, err := datastore.NewEngine(&store.Opts{
		Driver: "sqlite3",
		Config: path,
		XORM: store.XORM{
			MaxOpenConns: 1,
			MaxIdleConns: 1,
		},
	})
	require.NoError(t, err)

	closed := false
	closeStore := func() {
		t.Helper()
		if !closed {
			closed = true
			require.NoError(t, s.Close())
		}
	}
	t.Cleanup(closeStore)
	require.NoError(t, s.Ping())
	require.NoError(t, s.Migrate(ctx, true))
	return s, closeStore
}

func newSQLitePersistentQueue(t *testing.T, ctx context.Context, s store.Store) (Queue, *fifo, context.CancelFunc) {
	t.Helper()

	queueCtx, cancel := context.WithCancelCause(ctx)
	stop := func() { cancel(nil) }
	t.Cleanup(stop)
	memory, ok := NewMemoryQueue(queueCtx).(*fifo)
	require.True(t, ok)
	return WithTaskStore(queueCtx, memory, s), memory, stop
}

func createQueueSQLiteWorkflow(t *testing.T, s store.Store, state model.StatusValue) *model.Task {
	t.Helper()

	require.NoError(t, s.WorkflowsCreate([]*model.Workflow{{
		ID:         1,
		PipelineID: 1,
		PID:        1,
		State:      state,
	}}))
	return genDummyTask()
}

func requireQueueSQLiteBackupCount(t *testing.T, s store.Store, count int) {
	t.Helper()

	tasks, err := s.TaskList()
	require.NoError(t, err)
	require.Len(t, tasks, count)
}

func requireQueueSQLiteEmpty(t *testing.T, ctx context.Context, q Queue, s store.Store) {
	t.Helper()

	info := q.Info(ctx)
	require.Zero(t, info.Stats.Pending)
	require.Zero(t, info.Stats.Running)
	require.Zero(t, info.Stats.WaitingOnDeps)
	requireQueueSQLiteBackupCount(t, s, 0)
}

func expireQueueSQLiteTask(t *testing.T, ctx context.Context, q *fifo, id string) {
	t.Helper()

	q.Lock()
	expired := q.running[id]
	if expired != nil {
		expired.deadline = time.Now().Add(-time.Second)
	}
	q.Unlock()
	require.NotNil(t, expired)

	// Let the normal process loop expire the selected lease. Waiting on this
	// generation's channel avoids changing queue-wide timing or sleeping.
	select {
	case <-expired.done:
		require.ErrorIs(t, expired.error, ErrTaskExpired)
	case <-ctx.Done():
		t.Fatal("queue did not expire the selected task before the context ended")
	}

	info := q.Info(ctx)
	require.Equal(t, 1, info.Stats.Pending)
	require.Zero(t, info.Stats.Running)
}

func TestPersistentQueueSQLiteRestoresQueuedTask(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	path := filepath.Join(t.TempDir(), "queue.sqlite")

	s, closeStore := openQueueSQLiteStore(t, ctx, path)
	q, _, stop := newSQLitePersistentQueue(t, ctx, s)
	task := createQueueSQLiteWorkflow(t, s, model.StatusPending)
	require.NoError(t, q.PushAtOnce(ctx, []*model.Task{task}))
	requireQueueSQLiteBackupCount(t, s, 1)

	stop()
	closeStore()
	s, _ = openQueueSQLiteStore(t, ctx, path)
	q, _, _ = newSQLitePersistentQueue(t, ctx, s)
	require.Equal(t, 1, q.Info(ctx).Stats.Pending)
	requireQueueSQLiteBackupCount(t, s, 1)

	restored, err := q.Poll(ctx, 2, filterFnTrue)
	require.NoError(t, err)
	require.NotNil(t, restored)
	require.Equal(t, task.ID, restored.ID)
	require.Equal(t, task.Data, restored.Data)
	requireQueueSQLiteBackupCount(t, s, 0)
	require.NoError(t, q.Done(ctx, restored.ID, model.StatusSuccess))
	requireQueueSQLiteEmpty(t, ctx, q, s)
}

func TestPersistentQueueSQLiteDoesNotRestoreAssignedTask(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	path := filepath.Join(t.TempDir(), "queue.sqlite")

	s, closeStore := openQueueSQLiteStore(t, ctx, path)
	q, _, stop := newSQLitePersistentQueue(t, ctx, s)
	task := createQueueSQLiteWorkflow(t, s, model.StatusRunning)
	require.NoError(t, q.PushAtOnce(ctx, []*model.Task{task}))
	assigned, err := q.Poll(ctx, 1, filterFnTrue)
	require.NoError(t, err)
	require.NotNil(t, assigned)
	require.Equal(t, 1, q.Info(ctx).Stats.Running)
	requireQueueSQLiteBackupCount(t, s, 0)

	stop()
	closeStore()
	s, _ = openQueueSQLiteStore(t, ctx, path)
	q, _, _ = newSQLitePersistentQueue(t, ctx, s)
	requireQueueSQLiteEmpty(t, ctx, q, s)
	workflow, err := s.WorkflowLoad(1)
	require.NoError(t, err)
	require.Equal(t, model.StatusRunning, workflow.State)
}

func TestPersistentQueueSQLiteDoesNotRestoreExpiredPendingTask(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	path := filepath.Join(t.TempDir(), "queue.sqlite")

	s, closeStore := openQueueSQLiteStore(t, ctx, path)
	q, memory, stop := newSQLitePersistentQueue(t, ctx, s)
	task := createQueueSQLiteWorkflow(t, s, model.StatusRunning)
	require.NoError(t, q.PushAtOnce(ctx, []*model.Task{task}))
	assigned, err := q.Poll(ctx, 1, filterFnTrue)
	require.NoError(t, err)
	require.NotNil(t, assigned)
	expireQueueSQLiteTask(t, ctx, memory, task.ID)
	requireQueueSQLiteBackupCount(t, s, 0)

	// Expiry requeues only in memory. This preserves the existing restart
	// limitation: validating recovery during Poll does not make expired work
	// durable while it is waiting for another agent.
	stop()
	closeStore()
	s, _ = openQueueSQLiteStore(t, ctx, path)
	q, _, _ = newSQLitePersistentQueue(t, ctx, s)
	requireQueueSQLiteEmpty(t, ctx, q, s)
	workflow, err := s.WorkflowLoad(1)
	require.NoError(t, err)
	require.Equal(t, model.StatusRunning, workflow.State)
}

func TestPersistentQueueSQLiteRecoversExpiredTask(t *testing.T) {
	for _, state := range []model.StatusValue{model.StatusPending, model.StatusRunning} {
		t.Run(string(state), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()

			s, _ := openQueueSQLiteStore(t, ctx, filepath.Join(t.TempDir(), "queue.sqlite"))
			q, memory, _ := newSQLitePersistentQueue(t, ctx, s)
			task := createQueueSQLiteWorkflow(t, s, state)
			require.NoError(t, q.PushAtOnce(ctx, []*model.Task{task}))
			requireQueueSQLiteBackupCount(t, s, 1)

			assigned, err := q.Poll(ctx, 1, filterFnTrue)
			require.NoError(t, err)
			require.NotNil(t, assigned)
			requireQueueSQLiteBackupCount(t, s, 0)
			expireQueueSQLiteTask(t, ctx, memory, task.ID)
			requireQueueSQLiteBackupCount(t, s, 0)

			recovered, err := q.Poll(ctx, 2, filterFnTrue)
			require.NoError(t, err)
			require.NotNil(t, recovered, "an expired nonterminal task must survive its missing backup")
			require.Equal(t, task.ID, recovered.ID)
			require.EqualValues(t, 2, recovered.AgentID)
			require.Equal(t, 1, q.Info(ctx).Stats.Running)
			require.NoError(t, q.Done(ctx, recovered.ID, model.StatusSuccess))
			requireQueueSQLiteEmpty(t, ctx, q, s)
		})
	}
}
