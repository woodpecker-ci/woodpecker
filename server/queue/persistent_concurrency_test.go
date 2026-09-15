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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

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
