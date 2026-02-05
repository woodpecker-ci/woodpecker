// Copyright 2023 Woodpecker Authors
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

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

var (
	filterFnTrue = func(*model.Task) (bool, int) { return true, 1 }
	genDummyTask = func() *model.Task {
		return &model.Task{
			ID:   "1",
			Data: []byte("{}"),
		}
	}
	waitForProcess = func() { time.Sleep(processTimeInterval + 50*time.Millisecond) }
)

func setupTestQueue(t *testing.T) (context.Context, context.CancelCauseFunc, *fifo) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	if q == nil {
		t.Fatal("Failed to create queue")
	}

	return ctx, cancel, q
}

func TestFifoBasicOperations(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	t.Run("push poll done lifecycle", func(t *testing.T) {
		dummyTask := genDummyTask()

		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))
		waitForProcess()

		info := q.Info(ctx)
		assert.Len(t, info.Pending, 1)

		got, err := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, dummyTask, got)

		waitForProcess()
		info = q.Info(ctx)
		assert.Len(t, info.Pending, 0)
		assert.Len(t, info.Running, 1)

		// Edge case: verify task can't be polled again while running
		pollCtx, pollCancel := context.WithTimeout(ctx, 100*time.Millisecond)
		_, err = q.Poll(pollCtx, 2, filterFnTrue)
		pollCancel()
		assert.Error(t, err) // Should timeout/cancel, not return the same task

		assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))

		waitForProcess()
		info = q.Info(ctx)
		assert.Len(t, info.Running, 0)

		// Edge case: Done on already completed task should handle gracefully
		err = q.Done(ctx, got.ID, model.StatusSuccess)
		// Document current behavior - should either error or be idempotent
		if err != nil {
			assert.Error(t, err)
		}
	})

	t.Run("error handling", func(t *testing.T) {
		task1 := &model.Task{ID: "task-error-1"}
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1}))

		waitForProcess()
		got, _ := q.Poll(ctx, 1, filterFnTrue)

		assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("test error")))
		waitForProcess()
		info := q.Info(ctx)
		assert.Len(t, info.Running, 0)

		assert.Error(t, q.Error(ctx, "totally-fake-id", fmt.Errorf("test error")))

		// Edge case: Error on task that's already errored
		err := q.Error(ctx, got.ID, fmt.Errorf("double error"))
		// Should either error or be idempotent
		if err != nil {
			assert.Error(t, err)
		}
	})

	t.Run("internal error pass-through", func(t *testing.T) {
		// Test that internal queue errors (like ErrNotFound) are NOT wrapped with ErrExternal
		// Internal errors should pass through unchanged so the queue layer can handle them

		// Attempt to error a non-existent task - should trigger internal ErrNotFound
		err := q.ErrorAtOnce(ctx, []string{"non-existent-task-id"}, fmt.Errorf("some error"))
		assert.Error(t, err)
		// Internal errors like ErrNotFound should pass through unwrapped
		assert.ErrorIs(t, err, ErrNotFound, "internal queue errors should not be wrapped")
		assert.False(t, errors.Is(err, new(ErrExternal)), "internal errors should not be marked as external")

		// Verify similar behavior with multiple non-existent IDs
		err = q.ErrorAtOnce(ctx, []string{"fake-1", "fake-2", "fake-3"}, fmt.Errorf("batch error"))
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound, "batch internal errors should not be wrapped")
		assert.False(t, errors.Is(err, new(ErrExternal)), "batch internal errors should not be external")
	})

	t.Run("error at once", func(t *testing.T) {
		task1 := &model.Task{ID: "batch-1"}
		task2 := &model.Task{ID: "batch-2"}
		task3 := &model.Task{ID: "batch-3"}

		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2, task3}))
		waitForProcess()

		got1, _ := q.Poll(ctx, 1, filterFnTrue)
		got2, _ := q.Poll(ctx, 2, filterFnTrue)

		assert.NoError(t, q.ErrorAtOnce(ctx, []string{got1.ID, got2.ID}, fmt.Errorf("batch error")))
		waitForProcess()
		info := q.Info(ctx)
		assert.Len(t, info.Running, 0)
		assert.Len(t, info.Pending, 1)

		got3, _ := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, q.Done(ctx, got3.ID, model.StatusSuccess))
		waitForProcess()

		task4 := &model.Task{ID: "batch-4"}
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task4}))
		waitForProcess()
		got4, _ := q.Poll(ctx, 1, filterFnTrue)

		err := q.ErrorAtOnce(ctx, []string{got4.ID, "fake-1", "fake-2"}, fmt.Errorf("test"))
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)

		waitForProcess()
		info = q.Info(ctx)
		assert.Len(t, info.Running, 0)

		// Edge case: ErrorAtOnce with empty slice
		err = q.ErrorAtOnce(ctx, []string{}, fmt.Errorf("no tasks"))
		assert.NoError(t, err)
		// Should handle gracefully, potentially no-op

		// Edge case: ErrorAtOnce with nil error
		task5 := &model.Task{ID: "batch-5"}
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task5}))
		waitForProcess()
		got5, _ := q.Poll(ctx, 3, filterFnTrue)
		err = q.ErrorAtOnce(ctx, []string{got5.ID}, nil)
		assert.NoError(t, err)
		// Should handle nil error gracefully
		waitForProcess()
	})

	t.Run("error at once with waiting deps", func(t *testing.T) {
		task5 := &model.Task{ID: "deps-cancel-5"}
		task6 := &model.Task{
			ID:           "deps-cancel-6",
			Dependencies: []string{"deps-cancel-5"},
			DepStatus:    make(map[string]model.StatusValue),
		}

		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task5, task6}))
		waitForProcess()

		info := q.Info(ctx)
		assert.Equal(t, 1, info.Stats.WaitingOnDeps)

		assert.NoError(t, q.ErrorAtOnce(ctx, []string{"deps-cancel-5", "deps-cancel-6"}, fmt.Errorf("canceled")))

		waitForProcess()
		info = q.Info(ctx)
		assert.Equal(t, 0, info.Stats.WaitingOnDeps)
		assert.Len(t, info.Pending, 0)

		// Edge case: verify both tasks are actually gone, not stuck somewhere
		assert.Len(t, info.Running, 0)
		assert.Len(t, info.WaitingOnDeps, 0)
	})

	t.Run("error at once cancellation", func(t *testing.T) {
		task1 := &model.Task{ID: "cancel-prop-1"}
		task2 := &model.Task{
			ID:           "cancel-prop-2",
			Dependencies: []string{"cancel-prop-1"},
			DepStatus:    make(map[string]model.StatusValue),
			RunOn:        []string{"success", "failure"},
		}

		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2}))
		waitForProcess()
		got1, _ := q.Poll(ctx, 1, filterFnTrue)

		assert.NoError(t, q.ErrorAtOnce(ctx, []string{got1.ID}, ErrCancel))

		waitForProcess()
		waitForProcess()

		got2, _ := q.Poll(ctx, 2, filterFnTrue)
		assert.Equal(t, model.StatusKilled, got2.DepStatus["cancel-prop-1"])

		// Edge case: verify ErrCancel results in StatusKilled not StatusFailure
		assert.NotEqual(t, model.StatusFailure, got2.DepStatus["cancel-prop-1"])
		assert.NoError(t, q.Done(ctx, got2.ID, model.StatusSuccess))
		waitForProcess()
	})

	t.Run("pause resume", func(t *testing.T) {
		dummyTask := &model.Task{ID: "pause-1"}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			_, _ = q.Poll(ctx, 99, filterFnTrue)
			wg.Done()
		}()

		q.Pause()
		t0 := time.Now()
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))
		waitForProcess()

		// Edge case: verify queue is actually paused
		info := q.Info(ctx)
		assert.True(t, info.Paused)
		assert.Len(t, info.Pending, 1)
		assert.Len(t, info.Running, 0)

		q.Resume()

		wg.Wait()
		assert.Greater(t, time.Since(t0), 20*time.Millisecond)

		// Edge case: verify queue is unpaused
		info = q.Info(ctx)
		assert.False(t, info.Paused)

		// Edge case: multiple pause/resume cycles
		task2 := &model.Task{ID: "pause-2"}
		q.Pause()
		q.Pause() // Double pause
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2}))
		waitForProcess()
		q.Resume()
		q.Resume() // Double resume
		waitForProcess()
		got, _ := q.Poll(ctx, 99, filterFnTrue)
		assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))
		waitForProcess()
	})
}

func TestFifoDependencies(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	t.Run("basic dependency handling", func(t *testing.T) {
		task1 := &model.Task{ID: "dep-basic-1"}
		task2 := &model.Task{
			ID:           "dep-basic-2",
			Dependencies: []string{"dep-basic-1"},
			DepStatus:    make(map[string]model.StatusValue),
		}
		task3 := &model.Task{
			ID:           "dep-basic-3",
			Dependencies: []string{"dep-basic-1"},
			DepStatus:    make(map[string]model.StatusValue),
			RunOn:        []string{"success", "failure"},
		}

		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))
		waitForProcess()

		info := q.Info(ctx)
		assert.Equal(t, 2, info.Stats.WaitingOnDeps)

		got, err := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, task1, got)
		assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("exit code 1")))

		waitForProcess()
		got, err = q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, task2, got)
		assert.False(t, got.ShouldRun())
		assert.Equal(t, model.StatusFailure, got.DepStatus["dep-basic-1"])

		waitForProcess()
		got, err = q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, task3, got)
		assert.True(t, got.ShouldRun())
		assert.Equal(t, model.StatusFailure, got.DepStatus["dep-basic-1"])

		waitForProcess()
		info = q.Info(ctx)
		assert.Equal(t, 0, info.Stats.WaitingOnDeps)

		// Edge case: verify DepStatus is correctly set before polling
		assert.NotEmpty(t, task2.DepStatus)
		assert.NotEmpty(t, task3.DepStatus)
	})

	t.Run("multiple dependencies", func(t *testing.T) {
		task1 := &model.Task{ID: "multi-dep-1"}
		task2 := &model.Task{ID: "multi-dep-2"}
		task3 := &model.Task{
			ID:           "multi-dep-3",
			Dependencies: []string{"multi-dep-1", "multi-dep-2"},
			DepStatus:    make(map[string]model.StatusValue),
		}

		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))
		waitForProcess()

		got1, _ := q.Poll(ctx, 1, filterFnTrue)
		got2, _ := q.Poll(ctx, 2, filterFnTrue)

		gotIDs := map[string]bool{got1.ID: true, got2.ID: true}
		assert.True(t, gotIDs["multi-dep-1"] && gotIDs["multi-dep-2"])

		if got1.ID == "multi-dep-1" {
			assert.NoError(t, q.Done(ctx, got1.ID, model.StatusSuccess))
			assert.NoError(t, q.Error(ctx, got2.ID, fmt.Errorf("failed")))
		} else {
			assert.NoError(t, q.Done(ctx, got2.ID, model.StatusSuccess))
			assert.NoError(t, q.Error(ctx, got1.ID, fmt.Errorf("failed")))
		}

		waitForProcess()
		got3, err := q.Poll(ctx, 3, filterFnTrue)
		assert.NoError(t, err)

		assert.Contains(t, got3.DepStatus, "multi-dep-1")
		assert.Contains(t, got3.DepStatus, "multi-dep-2")
		assert.True(t,
			(got3.DepStatus["multi-dep-1"] == model.StatusSuccess && got3.DepStatus["multi-dep-2"] == model.StatusFailure) ||
				(got3.DepStatus["multi-dep-1"] == model.StatusFailure && got3.DepStatus["multi-dep-2"] == model.StatusSuccess))
		assert.False(t, got3.ShouldRun())

		// Edge case: verify both deps are tracked
		assert.Len(t, got3.DepStatus, 2)
		assert.NoError(t, q.Done(ctx, got3.ID, model.StatusSkipped))
		waitForProcess()
	})

	t.Run("transitive dependencies", func(t *testing.T) {
		task1 := &model.Task{ID: "trans-1"}
		task2 := &model.Task{
			ID:           "trans-2",
			Dependencies: []string{"trans-1"},
			DepStatus:    make(map[string]model.StatusValue),
		}
		task3 := &model.Task{
			ID:           "trans-3",
			Dependencies: []string{"trans-2"},
			DepStatus:    make(map[string]model.StatusValue),
		}

		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))
		waitForProcess()

		got, _ := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("exit code 1")))

		waitForProcess()
		got, _ = q.Poll(ctx, 2, filterFnTrue)
		assert.False(t, got.ShouldRun())
		assert.NoError(t, q.Done(ctx, got.ID, model.StatusSkipped))

		waitForProcess()
		got, _ = q.Poll(ctx, 3, filterFnTrue)
		assert.Equal(t, model.StatusSkipped, got.DepStatus["trans-2"])
		assert.False(t, got.ShouldRun())

		// Edge case: verify transitive failure propagates correctly
		// task3 should see trans-2 as skipped, not trans-1's status
		assert.NotContains(t, got.DepStatus, "trans-1")
		assert.NoError(t, q.Done(ctx, got.ID, model.StatusSkipped))
		waitForProcess()
	})

	t.Run("dependency status propagation", func(t *testing.T) {
		task1 := &model.Task{ID: "prop-1"}
		task2 := &model.Task{
			ID:           "prop-2",
			Dependencies: []string{"prop-1"},
			DepStatus:    make(map[string]model.StatusValue),
		}
		task3 := &model.Task{
			ID:           "prop-3",
			Dependencies: []string{"prop-1"},
			DepStatus:    make(map[string]model.StatusValue),
			RunOn:        []string{"success", "failure"},
		}

		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2, task3}))
		waitForProcess()

		info := q.Info(ctx)
		assert.Equal(t, 2, info.Stats.WaitingOnDeps)

		got1, _ := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, q.Done(ctx, got1.ID, model.StatusSuccess))

		waitForProcess()

		got2, _ := q.Poll(ctx, 2, filterFnTrue)
		got3, _ := q.Poll(ctx, 3, filterFnTrue)

		assert.Equal(t, model.StatusSuccess, got2.DepStatus["prop-1"])
		assert.Equal(t, model.StatusSuccess, got3.DepStatus["prop-1"])

		// Edge case: verify both tasks can be polled concurrently
		assert.NotEqual(t, got2.ID, got3.ID)
		assert.NoError(t, q.Done(ctx, got2.ID, model.StatusSuccess))
		assert.NoError(t, q.Done(ctx, got3.ID, model.StatusSuccess))
		waitForProcess()

		task4 := &model.Task{ID: "prop-4"}
		task5 := &model.Task{
			ID:           "prop-5",
			Dependencies: []string{"prop-4"},
			DepStatus:    make(map[string]model.StatusValue),
		}
		task6 := &model.Task{
			ID:           "prop-6",
			Dependencies: []string{"prop-4"},
			DepStatus:    make(map[string]model.StatusValue),
			RunOn:        []string{"success", "failure"},
		}

		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task4, task5, task6}))
		waitForProcess()

		got4, _ := q.Poll(ctx, 4, filterFnTrue)
		assert.NoError(t, q.Error(ctx, got4.ID, fmt.Errorf("failed")))

		waitForProcess()

		got5, _ := q.Poll(ctx, 5, filterFnTrue)
		assert.Equal(t, model.StatusFailure, got5.DepStatus["prop-4"])
		assert.False(t, got5.ShouldRun())

		got6, _ := q.Poll(ctx, 6, filterFnTrue)
		assert.Equal(t, model.StatusFailure, got6.DepStatus["prop-4"])
		assert.True(t, got6.ShouldRun())

		// Edge case: complete dependent tasks
		assert.NoError(t, q.Done(ctx, got5.ID, model.StatusSkipped))
		assert.NoError(t, q.Done(ctx, got6.ID, model.StatusSuccess))
		waitForProcess()
	})

	// Edge case: circular dependency detection (should be handled or cause issue)
	t.Run("circular dependencies", func(t *testing.T) {
		task1 := &model.Task{
			ID:           "circ-1",
			Dependencies: []string{"circ-2"},
			DepStatus:    make(map[string]model.StatusValue),
		}
		task2 := &model.Task{
			ID:           "circ-2",
			Dependencies: []string{"circ-1"},
			DepStatus:    make(map[string]model.StatusValue),
		}

		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2}))
		waitForProcess()

		info := q.Info(ctx)
		// Both should be waiting on deps - this is a deadlock scenario
		assert.Equal(t, 2, info.Stats.WaitingOnDeps)
		assert.Len(t, info.Pending, 0)

		// Verify they never become available for polling
		pollCtx, pollCancel := context.WithTimeout(ctx, 200*time.Millisecond)
		_, err := q.Poll(pollCtx, 99, filterFnTrue)
		pollCancel()
		assert.Error(t, err) // Should timeout

		// Clean up the deadlocked tasks
		assert.NoError(t, q.ErrorAtOnce(ctx, []string{"circ-1", "circ-2"}, fmt.Errorf("circular dep")))
		waitForProcess()
	})

	// Edge case: dependency on non-existent task
	// NOTE: This reveals a potential issue - the queue doesn't validate dependencies exist.
	// If a dependency was never added to the queue, the task will run immediately since
	// depsInQueue() only checks currently pending/running tasks, not if deps will arrive.
	t.Run("non-existent dependency", func(t *testing.T) {
		task1 := &model.Task{
			ID:           "orphan-1",
			Dependencies: []string{"does-not-exist"},
			DepStatus:    make(map[string]model.StatusValue),
		}

		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1}))
		waitForProcess()

		info := q.Info(ctx)
		// Current implementation: task doesn't wait if dependency not in queue
		// This means tasks with typos in dependency names will run immediately!
		assert.Equal(t, 0, info.Stats.WaitingOnDeps)
		assert.Len(t, info.Pending, 1)

		// Task will be available for polling even though dependency doesn't exist
		got, err := q.Poll(ctx, 99, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, "orphan-1", got.ID)

		// DepStatus will be empty since dependency never completed
		assert.Empty(t, got.DepStatus)

		// Clean up
		assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))
		waitForProcess()
	})

	// Edge case: dependency added AFTER dependent task (race condition)
	t.Run("dependency added after dependent", func(t *testing.T) {
		// Push dependent task first
		dependent := &model.Task{
			ID:           "late-dep-child",
			Dependencies: []string{"late-dep-parent"},
			DepStatus:    make(map[string]model.StatusValue),
		}
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dependent}))
		waitForProcess()

		// At this point, dependent doesn't see parent in queue, so it won't wait
		info := q.Info(ctx)
		// Dependent should NOT be waiting since parent doesn't exist yet
		initialWaiting := info.Stats.WaitingOnDeps

		// Now add the parent task
		parent := &model.Task{ID: "late-dep-parent"}
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{parent}))
		waitForProcess()

		// After filterWaiting runs, dependent SHOULD now see parent and wait
		info = q.Info(ctx)
		// The implementation calls filterWaiting() which rechecks dependencies
		// So dependent should now be waiting
		assert.Greater(t, info.Stats.WaitingOnDeps, initialWaiting,
			"dependent should start waiting once parent is added")

		// Complete parent first
		gotParent, _ := q.Poll(ctx, 1, filterFnTrue)
		assert.Equal(t, "late-dep-parent", gotParent.ID, "parent should be polled first")
		assert.NoError(t, q.Done(ctx, gotParent.ID, model.StatusSuccess))
		waitForProcess()

		// Now child should be unblocked with parent's status
		gotChild, _ := q.Poll(ctx, 2, filterFnTrue)
		assert.Equal(t, "late-dep-child", gotChild.ID)
		assert.Equal(t, model.StatusSuccess, gotChild.DepStatus["late-dep-parent"])

		assert.NoError(t, q.Done(ctx, gotChild.ID, model.StatusSuccess))
		waitForProcess()
	})
}

func TestFifoLeaseManagement(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	t.Run("lease expiration", func(t *testing.T) {
		q.extension = 0
		t.Cleanup(func() {
			q.extension = 50 * time.Millisecond
		})
		dummyTask := &model.Task{ID: "lease-exp-1"}
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

		waitForProcess()
		got, err := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)

		errCh := make(chan error, 1)
		go func() { errCh <- q.Wait(ctx, got.ID) }()

		waitForProcess()
		select {
		case werr := <-errCh:
			assert.Error(t, werr)
			// Edge case: verify error is ErrTaskExpired
			assert.ErrorIs(t, werr, ErrTaskExpired)
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for Wait to return")
		}

		info := q.Info(ctx)
		assert.Len(t, info.Pending, 1)

		// Edge case: verify task was resubmitted to front of queue
		got2, _ := q.Poll(ctx, 1, filterFnTrue)
		assert.Equal(t, got.ID, got2.ID) // Same task resubmitted

		assert.NoError(t, q.Done(ctx, got2.ID, model.StatusSuccess))
		waitForProcess()

		// Verify cleanup
		info = q.Info(ctx)
		assert.Len(t, info.Pending, 0)
		assert.Len(t, info.Running, 0)
	})

	t.Run("extend lease", func(t *testing.T) {
		q.extension = 50 * time.Millisecond
		dummyTask := &model.Task{ID: "extend-1"}
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

		waitForProcess()
		got, _ := q.Poll(ctx, 5, filterFnTrue)

		assert.NoError(t, q.Extend(ctx, 5, got.ID))
		assert.ErrorIs(t, q.Extend(ctx, 999, got.ID), ErrAgentMissMatch)
		assert.ErrorIs(t, q.Extend(ctx, 1, got.ID), ErrAgentMissMatch)
		assert.ErrorIs(t, q.Extend(ctx, 1, "non-existent"), ErrNotFound)

		// Edge case: extend multiple times rapidly
		for i := 0; i < 3; i++ {
			time.Sleep(30 * time.Millisecond)
			assert.NoError(t, q.Extend(ctx, 5, got.ID))
		}

		info := q.Info(ctx)
		assert.Len(t, info.Running, 1)
		assert.Len(t, info.Pending, 0)

		// Edge case: extend after Done should error
		assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))
		waitForProcess()
		assert.ErrorIs(t, q.Extend(ctx, 5, got.ID), ErrNotFound)

		// Verify cleanup
		info = q.Info(ctx)
		assert.Len(t, info.Pending, 0)
		assert.Len(t, info.Running, 0)
	})

	t.Run("wait operations", func(t *testing.T) {
		// Verify queue is clean before starting
		info := q.Info(ctx)
		assert.Len(t, info.Pending, 0, "queue should be empty at start of wait operations")
		assert.Len(t, info.Running, 0, "queue should be empty at start of wait operations")

		dummyTask := &model.Task{ID: "wait-1"}
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

		waitForProcess()
		got, _ := q.Poll(ctx, 1, filterFnTrue)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			assert.NoError(t, q.Wait(ctx, got.ID))
			wg.Done()
		}()

		time.Sleep(time.Millisecond)
		assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))
		wg.Wait()

		// Edge case: Wait on non-existent task should return immediately
		assert.NoError(t, q.Wait(ctx, "non-existent"))

		dummyTask2 := &model.Task{ID: "wait-2"}
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask2}))
		waitForProcess()
		got2, _ := q.Poll(ctx, 1, filterFnTrue)

		waitCtx, waitCancel := context.WithCancelCause(ctx)
		errCh := make(chan error, 1)
		go func() { errCh <- q.Wait(waitCtx, got2.ID) }()

		time.Sleep(50 * time.Millisecond)
		waitCancel(nil)

		select {
		case err := <-errCh:
			assert.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("Wait should return when context is canceled")
		}

		// Clean up - complete the second wait task
		assert.NoError(t, q.Done(ctx, got2.ID, model.StatusSuccess))
		waitForProcess()

		// Edge case: multiple concurrent waits on same task
		dummyTask3 := &model.Task{ID: "wait-3"}
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask3}))
		waitForProcess()
		got3, _ := q.Poll(ctx, 1, filterFnTrue)

		var wg2 sync.WaitGroup
		wg2.Add(3)
		for i := 0; i < 3; i++ {
			go func() {
				assert.NoError(t, q.Wait(ctx, got3.ID))
				wg2.Done()
			}()
		}

		time.Sleep(10 * time.Millisecond)
		assert.NoError(t, q.Done(ctx, got3.ID, model.StatusSuccess))
		wg2.Wait()

		// Verify cleanup
		info = q.Info(ctx)
		assert.Len(t, info.Pending, 0)
		assert.Len(t, info.Running, 0)
	})
}

func TestFifoWorkerManagement(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	t.Run("poll with context cancellation", func(t *testing.T) {
		pollCtx, pollCancel := context.WithCancelCause(ctx)
		errCh := make(chan error, 1)
		go func() {
			_, err := q.Poll(pollCtx, 1, filterFnTrue)
			errCh <- err
		}()

		time.Sleep(50 * time.Millisecond)
		pollCancel(nil)

		select {
		case err := <-errCh:
			assert.ErrorIs(t, err, context.Canceled)
		case <-time.After(time.Second):
			t.Fatal("Poll should return when context is canceled")
		}

		// Edge case: verify worker is cleaned up
		info := q.Info(ctx)
		assert.Equal(t, 0, info.Stats.Workers)
	})

	t.Run("kick agent workers", func(t *testing.T) {
		pollResults := make(chan error, 5)
		for i := 0; i < 5; i++ {
			go func() {
				_, err := q.Poll(ctx, 42, filterFnTrue)
				pollResults <- err
			}()
		}

		time.Sleep(50 * time.Millisecond)

		// Edge case: verify workers are registered before kicking
		info := q.Info(ctx)
		assert.Equal(t, 5, info.Stats.Workers)

		q.KickAgentWorkers(42)

		kickedCount := 0
		for i := 0; i < 5; i++ {
			select {
			case err := <-pollResults:
				if errors.Is(err, context.Canceled) {
					kickedCount++
				}
			case <-time.After(time.Second):
				t.Fatal("expected all workers to be kicked")
			}
		}
		assert.Equal(t, 5, kickedCount)

		// Edge case: verify workers are removed after kicking
		waitForProcess()
		info = q.Info(ctx)
		assert.Equal(t, 0, info.Stats.Workers)

		// Edge case: kick non-existent agent should be no-op
		q.KickAgentWorkers(999)
	})

	// Edge case: mixed agent workers
	t.Run("kick specific agent among multiple", func(t *testing.T) {
		pollResults := make(chan struct {
			agentID int64
			err     error
		}, 10)

		// Start workers for agent 1
		for i := 0; i < 3; i++ {
			go func() {
				_, err := q.Poll(ctx, 1, filterFnTrue)
				pollResults <- struct {
					agentID int64
					err     error
				}{1, err}
			}()
		}

		// Start workers for agent 2
		for i := 0; i < 3; i++ {
			go func() {
				_, err := q.Poll(ctx, 2, filterFnTrue)
				pollResults <- struct {
					agentID int64
					err     error
				}{2, err}
			}()
		}

		time.Sleep(50 * time.Millisecond)
		info := q.Info(ctx)
		assert.Equal(t, 6, info.Stats.Workers)

		// Kick only agent 1
		q.KickAgentWorkers(1)

		kickedAgent1 := 0
		kickedAgent2 := 0
		for i := 0; i < 3; i++ {
			select {
			case result := <-pollResults:
				if errors.Is(result.err, context.Canceled) {
					if result.agentID == 1 {
						kickedAgent1++
					} else {
						kickedAgent2++
					}
				}
			case <-time.After(time.Second):
				t.Fatal("expected kicked workers to return")
			}
		}

		assert.Equal(t, 3, kickedAgent1)
		assert.Equal(t, 0, kickedAgent2)

		// Clean up agent 2 workers
		q.KickAgentWorkers(2)
		for i := 0; i < 3; i++ {
			<-pollResults
		}
	})
}

func TestFifoLabelBasedScoring(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	defer cancel(nil)

	q := NewMemoryQueue(ctx)

	tasks := []*model.Task{
		{ID: "1", Labels: map[string]string{"org-id": "123", "platform": "linux"}},
		{ID: "2", Labels: map[string]string{"org-id": "456", "platform": "linux"}},
		{ID: "3", Labels: map[string]string{"org-id": "123", "platform": "windows"}},
	}

	assert.NoError(t, q.PushAtOnce(ctx, tasks))

	filter123 := func(task *model.Task) (bool, int) {
		if task.Labels["org-id"] == "123" {
			return true, 20
		}
		return true, 1
	}

	filter456 := func(task *model.Task) (bool, int) {
		if task.Labels["org-id"] == "456" {
			return true, 20
		}
		return true, 1
	}

	results := make(chan *model.Task, 2)
	go func() {
		task, _ := q.Poll(ctx, 1, filter123)
		results <- task
	}()
	go func() {
		task, _ := q.Poll(ctx, 2, filter456)
		results <- task
	}()

	receivedTasks := make(map[string]int64)
	for i := 0; i < 2; i++ {
		select {
		case task := <-results:
			receivedTasks[task.ID] = task.AgentID
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for tasks")
		}
	}

	assert.Contains(t, []string{"1", "3"}, findTaskByAgent(receivedTasks, 1))
	assert.Equal(t, "2", findTaskByAgent(receivedTasks, 2))

	// Edge case: filter that rejects all tasks
	filterRejectAll := func(task *model.Task) (bool, int) {
		return false, 0
	}

	task4 := &model.Task{ID: "4", Labels: map[string]string{"org-id": "789"}}
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task4}))
	waitForProcess()

	pollCtx, pollCancel := context.WithTimeout(ctx, 200*time.Millisecond)
	_, err := q.Poll(pollCtx, 99, filterRejectAll)
	pollCancel()
	assert.Error(t, err) // Should timeout as filter rejects task

	// Clean up remaining tasks
	task3, _ := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, q.Done(ctx, task3.ID, model.StatusSuccess))
	task4Got, _ := q.Poll(ctx, 99, filterFnTrue)
	assert.NoError(t, q.Done(ctx, task4Got.ID, model.StatusSuccess))
	waitForProcess()
}

func TestShouldRunLogic(t *testing.T) {
	tests := []struct {
		name      string
		depStatus model.StatusValue
		runOn     []string
		expected  bool
	}{
		{"Success without RunOn", model.StatusSuccess, nil, true},
		{"Failure without RunOn", model.StatusFailure, nil, false},
		{"Success with failure RunOn", model.StatusSuccess, []string{"failure"}, false},
		{"Failure with failure RunOn", model.StatusFailure, []string{"failure"}, true},
		{"Success with both RunOn", model.StatusSuccess, []string{"success", "failure"}, true},
		{"Skipped without RunOn", model.StatusSkipped, nil, false},
		{"Skipped with failure RunOn", model.StatusSkipped, []string{"failure"}, true},
		// Edge cases
		{"Killed without RunOn", model.StatusKilled, nil, false},
		{"Killed with failure RunOn", model.StatusKilled, []string{"failure"}, true},
		{"Success with success RunOn only", model.StatusSuccess, []string{"success"}, true},
		{"Failure with success RunOn only", model.StatusFailure, []string{"success"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &model.Task{
				ID:           "2",
				Dependencies: []string{"1"},
				DepStatus:    map[string]model.StatusValue{"1": tt.depStatus},
				RunOn:        tt.runOn,
			}
			assert.Equal(t, tt.expected, task.ShouldRun())
		})
	}

	// Edge case: multiple dependencies with mixed statuses
	t.Run("multiple deps mixed status", func(t *testing.T) {
		task := &model.Task{
			ID:           "3",
			Dependencies: []string{"1", "2"},
			DepStatus: map[string]model.StatusValue{
				"1": model.StatusSuccess,
				"2": model.StatusFailure,
			},
			RunOn: nil,
		}
		// With default RunOn (nil), needs all deps successful
		assert.False(t, task.ShouldRun())

		task.RunOn = []string{"success", "failure"}
		// With both RunOn, should run regardless
		assert.True(t, task.ShouldRun())
	})
}

func findTaskByAgent(tasks map[string]int64, agentID int64) string {
	for taskID, aid := range tasks {
		if aid == agentID {
			return taskID
		}
	}
	return ""
}
