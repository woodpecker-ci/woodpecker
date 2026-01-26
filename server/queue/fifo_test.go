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

		assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))

		waitForProcess()
		info = q.Info(ctx)
		assert.Len(t, info.Running, 0)
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
		q.Resume()

		wg.Wait()
		assert.Greater(t, time.Since(t0), 20*time.Millisecond)
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
	})
}

func TestFifoLeaseManagement(t *testing.T) {
	ctx, cancel, q := setupTestQueue(t)
	defer cancel(nil)

	t.Run("lease expiration", func(t *testing.T) {
		q.extension = 0
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
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for Wait to return")
		}

		info := q.Info(ctx)
		assert.Len(t, info.Pending, 1)

		// Clean up - poll and complete the task
		got, _ = q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))
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

		for i := 0; i < 3; i++ {
			time.Sleep(30 * time.Millisecond)
			assert.NoError(t, q.Extend(ctx, 5, got.ID))
		}

		info := q.Info(ctx)
		assert.Len(t, info.Running, 1)
		assert.Len(t, info.Pending, 0)

		// Clean up - complete the extended task
		assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))
		waitForProcess()

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

		assert.NoError(t, q.Wait(ctx, "non-existent"))

		dummyTask2 := &model.Task{ID: "wait-2"}
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

		// Clean up - complete the second wait task
		assert.NoError(t, q.Done(ctx, got2.ID, model.StatusSuccess))
		waitForProcess()

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

func findTaskByAgent(tasks map[string]int64, agentID int64) string {
	for taskID, aid := range tasks {
		if aid == agentID {
			return taskID
		}
	}
	return ""
}
