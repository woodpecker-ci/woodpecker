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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

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

func expirePersistentLease(t *testing.T, ctx context.Context, q *fifo, id string) <-chan error {
	t.Helper()
	waitCtx := registeredWaitContext{Context: ctx, registered: make(chan struct{})}
	waitResult := make(chan error, 1)
	go func() { waitResult <- q.Wait(waitCtx, id) }()
	select {
	case <-waitCtx.registered:
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
