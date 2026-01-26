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

func TestBasicQueueOperations(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	dummyTask := genDummyTask()

	// Test push -> poll -> done lifecycle
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

	assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))

	waitForProcess()
	info = q.Info(ctx)
	assert.Len(t, info.Running, 0)
}

func TestTaskDependencies(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	task1 := genDummyTask()
	task2 := &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
	}
	task3 := &model.Task{
		ID:           "3",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
		RunOn:        []string{"success", "failure"},
	}

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))
	waitForProcess()

	// Verify waiting on deps stat
	info := q.Info(ctx)
	assert.Equal(t, 2, info.Stats.WaitingOnDeps)

	// Poll and fail task1
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task1, got)
	assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("exit code 1")))

	// task2 should be polled but not run (no failure in RunOn)
	waitForProcess()
	got, err = q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task2, got)
	assert.False(t, got.ShouldRun())
	assert.Equal(t, model.StatusFailure, got.DepStatus["1"])

	// task3 should run (has failure in RunOn)
	waitForProcess()
	got, err = q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task3, got)
	assert.True(t, got.ShouldRun())
	assert.Equal(t, model.StatusFailure, got.DepStatus["1"])

	waitForProcess()
	info = q.Info(ctx)
	assert.Equal(t, 0, info.Stats.WaitingOnDeps)
}

func TestMultipleDependencies(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	task1 := genDummyTask()
	task2 := &model.Task{ID: "2"}
	task3 := &model.Task{
		ID:           "3",
		Dependencies: []string{"1", "2"},
		DepStatus:    make(map[string]model.StatusValue),
	}

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))
	waitForProcess()

	// Poll both independent tasks
	got1, _ := q.Poll(ctx, 1, filterFnTrue)
	got2, _ := q.Poll(ctx, 2, filterFnTrue)

	// Ensure we got both task1 and task2 (order may vary)
	gotIDs := map[string]bool{got1.ID: true, got2.ID: true}
	assert.True(t, gotIDs["1"] && gotIDs["2"], "Should get both task1 and task2")

	// Complete them with different statuses
	if got1.ID == "1" {
		assert.NoError(t, q.Done(ctx, got1.ID, model.StatusSuccess))
		assert.NoError(t, q.Error(ctx, got2.ID, fmt.Errorf("failed")))
	} else {
		assert.NoError(t, q.Done(ctx, got2.ID, model.StatusSuccess))
		assert.NoError(t, q.Error(ctx, got1.ID, fmt.Errorf("failed")))
	}

	// task3 should have both statuses propagated
	waitForProcess()
	got3, err := q.Poll(ctx, 3, filterFnTrue)
	assert.NoError(t, err)

	// Verify both dependency statuses are set correctly
	assert.Contains(t, got3.DepStatus, "1")
	assert.Contains(t, got3.DepStatus, "2")
	assert.True(t,
		(got3.DepStatus["1"] == model.StatusSuccess && got3.DepStatus["2"] == model.StatusFailure) ||
			(got3.DepStatus["1"] == model.StatusFailure && got3.DepStatus["2"] == model.StatusSuccess),
		"One dependency should succeed and one should fail")
	assert.False(t, got3.ShouldRun())
}

func TestTransitiveDependencies(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	task1 := genDummyTask()
	task2 := &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
	}
	task3 := &model.Task{
		ID:           "3",
		Dependencies: []string{"2"},
		DepStatus:    make(map[string]model.StatusValue),
	}

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))
	waitForProcess()

	// Fail task1
	got, _ := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("exit code 1")))

	// task2 should skip
	waitForProcess()
	got, _ = q.Poll(ctx, 2, filterFnTrue)
	assert.False(t, got.ShouldRun())
	assert.NoError(t, q.Done(ctx, got.ID, model.StatusSkipped))

	// task3 should also skip (transitive)
	waitForProcess()
	got, _ = q.Poll(ctx, 3, filterFnTrue)
	assert.Equal(t, model.StatusSkipped, got.DepStatus["2"])
	assert.False(t, got.ShouldRun())
}

func TestLeaseExpiration(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	q.extension = 0 // Immediate expiration
	dummyTask := genDummyTask()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

	waitForProcess()
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)

	// Test Wait returns error on expiration
	errCh := make(chan error, 1)
	go func() { errCh <- q.Wait(ctx, got.ID) }()

	waitForProcess()
	select {
	case werr := <-errCh:
		assert.Error(t, werr)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for Wait to return")
	}

	// Task should be back in pending
	info := q.Info(ctx)
	assert.Len(t, info.Pending, 1)
}

func TestExtend(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	q.extension = 50 * time.Millisecond
	dummyTask := genDummyTask()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

	waitForProcess()
	got, _ := q.Poll(ctx, 5, filterFnTrue)

	// Correct agent can extend
	assert.NoError(t, q.Extend(ctx, 5, got.ID))

	// Wrong agent cannot extend
	assert.ErrorIs(t, q.Extend(ctx, 999, got.ID), ErrAgentMissMatch)
	assert.ErrorIs(t, q.Extend(ctx, 1, got.ID), ErrAgentMissMatch)

	// Non-existent task
	assert.ErrorIs(t, q.Extend(ctx, 1, "non-existent"), ErrNotFound)

	// Extend prevents expiration
	for i := 0; i < 3; i++ {
		time.Sleep(30 * time.Millisecond)
		assert.NoError(t, q.Extend(ctx, 5, got.ID))
	}

	info := q.Info(ctx)
	assert.Len(t, info.Running, 1)
	assert.Len(t, info.Pending, 0)
}

func TestWait(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	dummyTask := genDummyTask()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

	waitForProcess()
	got, _ := q.Poll(ctx, 1, filterFnTrue)

	// Wait completes on Done
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		assert.NoError(t, q.Wait(ctx, got.ID))
		wg.Done()
	}()

	time.Sleep(time.Millisecond)
	assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))
	wg.Wait()

	// Wait on non-existent task
	assert.NoError(t, q.Wait(ctx, "non-existent"))

	// Wait with context cancellation
	dummyTask2 := &model.Task{ID: "2"}
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask2}))
	waitForProcess()
	got2, _ := q.Poll(ctx, 1, filterFnTrue)

	waitCtx, waitCancel := context.WithCancel(ctx)
	errCh := make(chan error, 1)
	go func() { errCh <- q.Wait(waitCtx, got2.ID) }()

	time.Sleep(50 * time.Millisecond)
	waitCancel()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("Wait should return when context is canceled")
	}
}

func TestError(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	task1 := genDummyTask()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1}))

	waitForProcess()
	got, _ := q.Poll(ctx, 1, filterFnTrue)

	// Error on running task
	assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("test error")))
	waitForProcess()
	info := q.Info(ctx)
	assert.Len(t, info.Running, 0)

	// Error on non-existent task
	assert.Error(t, q.Error(ctx, "totally-fake-id", fmt.Errorf("test error")))
}

func TestErrorAtOnce(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	// Test batch error on running tasks
	task1 := genDummyTask()
	task2 := &model.Task{ID: "2"}
	task3 := &model.Task{ID: "3"}

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2, task3}))
	waitForProcess()

	got1, _ := q.Poll(ctx, 1, filterFnTrue)
	got2, _ := q.Poll(ctx, 2, filterFnTrue)

	assert.NoError(t, q.ErrorAtOnce(ctx, []string{got1.ID, got2.ID}, fmt.Errorf("batch error")))
	waitForProcess()
	info := q.Info(ctx)
	assert.Len(t, info.Running, 0)
	assert.Len(t, info.Pending, 1) // task3 should still be pending

	// Clean up task3
	got3, _ := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, q.Done(ctx, got3.ID, model.StatusSuccess))
	waitForProcess()

	// Test error with non-existent tasks
	task4 := &model.Task{ID: "4"}
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task4}))
	waitForProcess()
	got4, _ := q.Poll(ctx, 1, filterFnTrue)

	err := q.ErrorAtOnce(ctx, []string{got4.ID, "fake-1", "fake-2"}, fmt.Errorf("test"))
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)

	// Verify task4 was still removed despite the error
	waitForProcess()
	info = q.Info(ctx)
	assert.Len(t, info.Running, 0)

	// Test ErrorAtOnce on tasks in waitingOnDeps (to cover removeFromPendingAndWaiting)
	task5 := &model.Task{ID: "5"}
	task6 := &model.Task{
		ID:           "6",
		Dependencies: []string{"5"},
		DepStatus:    make(map[string]model.StatusValue),
	}

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task5, task6}))
	waitForProcess()

	info = q.Info(ctx)
	assert.Equal(t, 1, info.Stats.WaitingOnDeps, "task6 should be waiting on deps")

	// Cancel both tasks - this should remove task6 from waitingOnDeps
	assert.NoError(t, q.ErrorAtOnce(ctx, []string{"5", "6"}, fmt.Errorf("canceled")))

	waitForProcess()
	info = q.Info(ctx)
	assert.Equal(t, 0, info.Stats.WaitingOnDeps, "task6 should be removed from waitingOnDeps")
	assert.Len(t, info.Pending, 0, "no tasks should be pending")
}

func TestErrorAtOnceCancellation(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	// Test ErrCancel with dependency propagation
	task1 := &model.Task{ID: "1"}
	task2 := &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
		RunOn:        []string{"success", "failure"}, // Ensures task runs on kill/cancel
	}

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2}))
	waitForProcess()
	got1, _ := q.Poll(ctx, 1, filterFnTrue)

	assert.NoError(t, q.ErrorAtOnce(ctx, []string{got1.ID}, ErrCancel))

	// Wait for cancellation to be processed and dependency to be updated
	waitForProcess()
	waitForProcess()

	got2, _ := q.Poll(ctx, 2, filterFnTrue)
	assert.Equal(t, model.StatusKilled, got2.DepStatus["1"])
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
}

func TestWorkerManagement(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	// Poll with context cancellation
	pollCtx, pollCancel := context.WithCancel(ctx)
	errCh := make(chan error, 1)
	go func() {
		_, err := q.Poll(pollCtx, 1, filterFnTrue)
		errCh <- err
	}()

	time.Sleep(50 * time.Millisecond)
	pollCancel()

	select {
	case err := <-errCh:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("Poll should return when context is canceled")
	}

	// Kick multiple workers for same agent
	pollResults := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func() {
			_, err := q.Poll(ctx, 42, filterFnTrue)
			pollResults <- err
		}()
	}

	time.Sleep(50 * time.Millisecond)
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
}

func TestLabelBasedScoring(t *testing.T) {
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

	// Agent 1 should get task with org-id 123
	assert.Contains(t, []string{"1", "3"}, findTaskByAgent(receivedTasks, 1))
	// Agent 2 should get task with org-id 456
	assert.Equal(t, "2", findTaskByAgent(receivedTasks, 2))
}

func TestPauseResume(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	dummyTask := genDummyTask()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_, _ = q.Poll(ctx, 1, filterFnTrue)
		wg.Done()
	}()

	q.Pause()
	t0 := time.Now()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))
	waitForProcess()
	q.Resume()

	wg.Wait()
	assert.Greater(t, time.Since(t0), 20*time.Millisecond)
}

func findTaskByAgent(tasks map[string]int64, agentID int64) string {
	for taskID, aid := range tasks {
		if aid == agentID {
			return taskID
		}
	}
	return ""
}

func TestDependencyStatusPropagation(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	// Test basic dependency status propagation from completed task to waiting tasks
	task1 := genDummyTask()
	task2 := &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
	}
	task3 := &model.Task{
		ID:           "3",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
		RunOn:        []string{"success", "failure"},
	}

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2, task3}))
	waitForProcess()

	// Both task2 and task3 should be waiting on task1
	info := q.Info(ctx)
	assert.Equal(t, 2, info.Stats.WaitingOnDeps)

	// Complete task1 - this triggers updateDepStatusInQueue for waitingOnDeps tasks
	got1, _ := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, q.Done(ctx, got1.ID, model.StatusSuccess))

	waitForProcess()

	// Poll task2 and task3 - both should have task1's success status
	got2, _ := q.Poll(ctx, 2, filterFnTrue)
	got3, _ := q.Poll(ctx, 3, filterFnTrue)

	assert.Equal(t, model.StatusSuccess, got2.DepStatus["1"])
	assert.Equal(t, model.StatusSuccess, got3.DepStatus["1"])

	// Now test updateDepStatusInQueue for tasks in pending and running queues
	task4 := &model.Task{ID: "4"}
	task5 := &model.Task{
		ID:           "5",
		Dependencies: []string{"4"},
		DepStatus:    make(map[string]model.StatusValue),
	}
	task6 := &model.Task{
		ID:           "6",
		Dependencies: []string{"4"},
		DepStatus:    make(map[string]model.StatusValue),
		RunOn:        []string{"success", "failure"},
	}

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task4, task5, task6}))
	waitForProcess()

	// Poll task4 and complete it while task5 and task6 are waiting
	got4, _ := q.Poll(ctx, 4, filterFnTrue)
	assert.NoError(t, q.Error(ctx, got4.ID, fmt.Errorf("failed")))

	waitForProcess()

	// task5 should not run (default behavior on failure)
	got5, _ := q.Poll(ctx, 5, filterFnTrue)
	assert.Equal(t, model.StatusFailure, got5.DepStatus["4"])
	assert.False(t, got5.ShouldRun())

	// task6 should run (has failure in RunOn)
	got6, _ := q.Poll(ctx, 6, filterFnTrue)
	assert.Equal(t, model.StatusFailure, got6.DepStatus["4"])
	assert.True(t, got6.ShouldRun())
}
