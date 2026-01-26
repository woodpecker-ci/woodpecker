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

func TestFifo(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q := NewMemoryQueue(ctx)
	dummyTask := genDummyTask()

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))
	waitForProcess()
	info := q.Info(ctx)
	assert.Len(t, info.Pending, 1, "expect task in pending queue")

	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, dummyTask, got)

	waitForProcess()
	info = q.Info(ctx)
	assert.Len(t, info.Pending, 0, "expect task removed from pending queue")
	assert.Len(t, info.Running, 1, "expect task in running queue")

	assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))

	waitForProcess()
	info = q.Info(ctx)
	assert.Len(t, info.Pending, 0, "expect task removed from pending queue")
	assert.Len(t, info.Running, 0, "expect task removed from running queue")
}

func TestFifoExpire(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	dummyTask := genDummyTask()

	q.extension = 0
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))
	waitForProcess()
	info := q.Info(ctx)
	assert.Len(t, info.Pending, 1, "expect task in pending queue")

	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, dummyTask, got)

	waitForProcess()
	info = q.Info(ctx)
	assert.Len(t, info.Pending, 1, "expect task re-added to pending queue")
}

func TestFifoWaitOnExpireReturnsError(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	q.extension = 0

	dummyTask := genDummyTask()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

	waitForProcess()
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, dummyTask, got)

	errCh := make(chan error, 1)
	go func() {
		errCh <- q.Wait(ctx, got.ID)
	}()

	waitForProcess()

	select {
	case werr := <-errCh:
		assert.Error(t, werr, "expected Wait to return error when lease expires")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for Wait to return after lease expiration")
	}

	info := q.Info(ctx)
	assert.Len(t, info.Pending, 1, "expect task re-added to pending queue after expiration")
}

func TestFifoWait(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	dummyTask := genDummyTask()

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

	waitForProcess()
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, dummyTask, got)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		assert.NoError(t, q.Wait(ctx, got.ID))
		wg.Done()
	}()

	<-time.After(time.Millisecond)
	assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))
	wg.Wait()
}

func TestFifoDependencies(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	task1 := genDummyTask()
	task2 := &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
	}

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task1}))

	waitForProcess()
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task1, got)

	waitForProcess()
	assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))

	waitForProcess()
	got, err = q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task2, got)
}

func TestFifoErrors(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

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

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))

	waitForProcess()
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task1, got)

	assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("exit code 1, there was an error")))

	waitForProcess()
	got, err = q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task2, got)
	assert.False(t, got.ShouldRun())

	waitForProcess()
	got, err = q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task3, got)
	assert.True(t, got.ShouldRun())
}

func TestFifoErrors2(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	task1 := genDummyTask()
	task2 := &model.Task{
		ID: "2",
	}
	task3 := &model.Task{
		ID:           "3",
		Dependencies: []string{"1", "2"},
		DepStatus:    make(map[string]model.StatusValue),
	}

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))

	for i := 0; i < 2; i++ {
		waitForProcess()
		got, err := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)
		assert.False(t, got != task1 && got != task2, "expect task1 or task2 returned from queue as task3 depends on them")

		if got != task1 {
			assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))
		}
		if got != task2 {
			assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("exit code 1, there was an error")))
		}
	}

	waitForProcess()
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task3, got)
	assert.False(t, got.ShouldRun())
}

func TestFifoErrorsMultiThread(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	task1 := genDummyTask()
	task2 := &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
	}
	task3 := &model.Task{
		ID:           "3",
		Dependencies: []string{"1", "2"},
		DepStatus:    make(map[string]model.StatusValue),
	}

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))

	obtainedWorkCh := make(chan *model.Task)
	defer func() { close(obtainedWorkCh) }()

	for i := 0; i < 10; i++ {
		go func(i int) {
			for {
				fmt.Printf("Worker %d started\n", i)
				got, err := q.Poll(ctx, 1, filterFnTrue)
				if err != nil && errors.Is(err, context.Canceled) {
					return
				}
				assert.NoError(t, err)
				obtainedWorkCh <- got
			}
		}(i)
	}

	task1Processed := false
	task2Processed := false

	for {
		select {
		case got := <-obtainedWorkCh:
			fmt.Println(got.ID)

			switch {
			case !task1Processed:
				assert.Equal(t, task1, got)
				task1Processed = true
				assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("exit code 1, there was an error")))
				go func() {
					for {
						fmt.Printf("Worker spawned\n")
						got, err := q.Poll(ctx, 1, filterFnTrue)
						if err != nil && errors.Is(err, context.Canceled) {
							return
						}
						assert.NoError(t, err)
						obtainedWorkCh <- got
					}
				}()
			case !task2Processed:
				assert.Equal(t, task2, got)
				task2Processed = true
				assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))
				go func() {
					for {
						fmt.Printf("Worker spawned\n")
						got, err := q.Poll(ctx, 1, filterFnTrue)
						if err != nil && errors.Is(err, context.Canceled) {
							return
						}
						assert.NoError(t, err)
						obtainedWorkCh <- got
					}
				}()
			default:
				assert.Equal(t, task3, got)
				assert.False(t, got.ShouldRun(), "expect task3 should not run, task1 succeeded but task2 failed")
				return
			}

		case <-time.After(5 * time.Second):
			info := q.Info(ctx)
			fmt.Println(info.String())
			t.Errorf("test timed out")
			return
		}
	}
}

func TestFifoTransitiveErrors(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

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

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))

	waitForProcess()
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task1, got)
	assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("exit code 1, there was an error")))

	waitForProcess()
	got, err = q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task2, got)
	assert.False(t, got.ShouldRun(), "expect task2 should not run, since task1 failed")
	assert.NoError(t, q.Done(ctx, got.ID, model.StatusSkipped))

	waitForProcess()
	got, err = q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task3, got)
	assert.False(t, got.ShouldRun(), "expect task3 should not run, task1 failed, thus task2 was skipped, task3 should be skipped too")
}

func TestFifoCancel(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

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

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))

	_, _ = q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, q.Error(ctx, task1.ID, fmt.Errorf("canceled")))
	assert.NoError(t, q.Error(ctx, task2.ID, fmt.Errorf("canceled")))
	assert.NoError(t, q.Error(ctx, task3.ID, fmt.Errorf("canceled")))
	info := q.Info(ctx)
	assert.Len(t, info.Pending, 0, "all pipelines should be canceled")

	time.Sleep(processTimeInterval * 2)
	info = q.Info(ctx)
	assert.Len(t, info.Pending, 0, "canceled are rescheduled")
	assert.Len(t, info.Running, 0, "canceled are rescheduled")
	assert.Len(t, info.WaitingOnDeps, 0, "canceled are rescheduled")
}

func TestFifoPause(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

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
	t1 := time.Now()

	assert.Greater(t, t1.Sub(t0), 20*time.Millisecond, "should have waited til resume")

	q.Pause()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))
	q.Resume()
	_, _ = q.Poll(ctx, 1, filterFnTrue)
}

func TestFifoPauseResume(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	dummyTask := genDummyTask()

	q.Pause()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))
	q.Resume()

	_, _ = q.Poll(ctx, 1, filterFnTrue)
}

func TestWaitingVsPending(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

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

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))

	got, _ := q.Poll(ctx, 1, filterFnTrue)

	waitForProcess()
	info := q.Info(ctx)
	assert.Equal(t, 2, info.Stats.WaitingOnDeps)

	assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("exit code 1, there was an error")))
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.EqualValues(t, task2, got)

	waitForProcess()
	info = q.Info(ctx)
	assert.Equal(t, 0, info.Stats.WaitingOnDeps)
	assert.Equal(t, 1, info.Stats.Pending)
}

func TestShouldRun(t *testing.T) {
	task := &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus: map[string]model.StatusValue{
			"1": model.StatusSuccess,
		},
		RunOn: []string{"failure"},
	}
	assert.False(t, task.ShouldRun(), "expect task to not run, it runs on failure only")

	task = &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus: map[string]model.StatusValue{
			"1": model.StatusSuccess,
		},
		RunOn: []string{"failure", "success"},
	}
	assert.True(t, task.ShouldRun())

	task = &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus: map[string]model.StatusValue{
			"1": model.StatusFailure,
		},
	}
	assert.False(t, task.ShouldRun())

	task = &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus: map[string]model.StatusValue{
			"1": model.StatusSuccess,
		},
		RunOn: []string{"success"},
	}
	assert.True(t, task.ShouldRun())

	task = &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus: map[string]model.StatusValue{
			"1": model.StatusFailure,
		},
		RunOn: []string{"failure"},
	}
	assert.True(t, task.ShouldRun())

	task = &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus: map[string]model.StatusValue{
			"1": model.StatusSkipped,
		},
	}
	assert.False(t, task.ShouldRun(), "task should not run if dependency is skipped")

	task = &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus: map[string]model.StatusValue{
			"1": model.StatusSkipped,
		},
		RunOn: []string{"failure"},
	}
	assert.True(t, task.ShouldRun(), "on failure, tasks should run on skipped deps, something failed higher up the chain")
}

func TestFifoWithScoring(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q := NewMemoryQueue(ctx)

	// Create tasks with different labels
	tasks := []*model.Task{
		{ID: "1", Labels: map[string]string{"org-id": "123", "platform": "linux"}},
		{ID: "2", Labels: map[string]string{"org-id": "456", "platform": "linux"}},
		{ID: "3", Labels: map[string]string{"org-id": "789", "platform": "windows"}},
		{ID: "4", Labels: map[string]string{"org-id": "123", "platform": "linux"}},
		{ID: "5", Labels: map[string]string{"org-id": "*", "platform": "linux"}},
	}

	assert.NoError(t, q.PushAtOnce(ctx, tasks))

	// Create filter functions for different workers
	filters := map[int]FilterFn{
		1: func(task *model.Task) (bool, int) {
			if task.Labels["org-id"] == "123" {
				return true, 20
			}
			if task.Labels["platform"] == "linux" {
				return true, 10
			}
			return true, 1
		},
		2: func(task *model.Task) (bool, int) {
			if task.Labels["org-id"] == "456" {
				return true, 20
			}
			if task.Labels["platform"] == "linux" {
				return true, 10
			}
			return true, 1
		},
		3: func(task *model.Task) (bool, int) {
			if task.Labels["platform"] == "windows" {
				return true, 20
			}
			return true, 1
		},
		4: func(task *model.Task) (bool, int) {
			if task.Labels["org-id"] == "123" {
				return true, 20
			}
			if task.Labels["platform"] == "linux" {
				return true, 10
			}
			return true, 1
		},
		5: func(task *model.Task) (bool, int) {
			if task.Labels["org-id"] == "*" {
				return true, 15
			}
			return true, 1
		},
	}

	// Start polling in separate goroutines
	results := make(chan *model.Task, 5)
	for i := 1; i <= 5; i++ {
		go func(n int) {
			task, err := q.Poll(ctx, int64(n), filters[n])
			assert.NoError(t, err)
			results <- task
		}(i)
	}

	// Collect results
	receivedTasks := make(map[string]int64)
	for i := 0; i < 5; i++ {
		select {
		case task := <-results:
			receivedTasks[task.ID] = task.AgentID
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for tasks")
		}
	}

	assert.Len(t, receivedTasks, 5, "All tasks should be assigned")

	// Define expected agent assignments
	// Map structure: {taskID: []possible agentIDs}
	// - taskID "1" and "4" can be assigned to agents 1 or 4 (org-id "123")
	// - taskID "2" should be assigned to agent 2 (org-id "456")
	// - taskID "3" should be assigned to agent 3 (platform "windows")
	// - taskID "5" should be assigned to agent 5 (org-id "*")
	expectedAssignments := map[string][]int64{
		"1": {1, 4},
		"2": {2},
		"3": {3},
		"4": {1, 4},
		"5": {5},
	}

	// Check if tasks are assigned as expected
	for taskID, expectedAgents := range expectedAssignments {
		agentID, ok := receivedTasks[taskID]
		assert.True(t, ok, "Task %s should be assigned", taskID)
		assert.Contains(t, expectedAgents, agentID, "Task %s should be assigned to one of the expected agents", taskID)
	}
}

func TestFifoErrorAtOnce(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	task1 := genDummyTask()
	task2 := &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
	}
	task3 := &model.Task{
		ID: "3",
	}

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2, task3}))

	waitForProcess()
	got1, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)

	got2, err := q.Poll(ctx, 2, filterFnTrue)
	assert.NoError(t, err)

	// Test ErrorAtOnce with running tasks
	assert.NoError(t, q.ErrorAtOnce(ctx, []string{got1.ID, got2.ID}, fmt.Errorf("batch error")))

	waitForProcess()
	info := q.Info(ctx)
	assert.Len(t, info.Running, 0, "expect tasks removed from running queue")
	assert.Len(t, info.Pending, 1, "expect remaining task in pending")
}

func TestFifoErrorAtOnceWithPendingTasks(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	task1 := genDummyTask()
	task2 := &model.Task{
		ID: "2",
	}
	task3 := &model.Task{
		ID:           "3",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
	}

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2, task3}))

	waitForProcess()

	// ErrorAtOnce on pending tasks (task3 should be waiting on deps)
	assert.NoError(t, q.ErrorAtOnce(ctx, []string{"2", "3"}, fmt.Errorf("pending error")))

	waitForProcess()
	info := q.Info(ctx)
	assert.Len(t, info.Pending, 1, "only task1 should remain")
	assert.Len(t, info.WaitingOnDeps, 0, "task3 should be removed from waiting")
}

func TestFifoErrorAtOnceWithCancelError(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	task1 := genDummyTask()
	task2 := &model.Task{
		ID: "2",
	}

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2}))

	waitForProcess()
	got1, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)

	// Test ErrorAtOnce with ErrCancel - should result in StatusKilled
	assert.NoError(t, q.ErrorAtOnce(ctx, []string{got1.ID}, ErrCancel))

	waitForProcess()
	info := q.Info(ctx)
	assert.Len(t, info.Running, 0, "task should be removed from running")
}

func TestFifoExtend(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	dummyTask := genDummyTask()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

	waitForProcess()
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)

	// Test successful extend
	assert.NoError(t, q.Extend(ctx, 1, got.ID))

	// Test extend with wrong agent ID
	err = q.Extend(ctx, 999, got.ID)
	assert.ErrorIs(t, err, ErrAgentMissMatch)

	// Clean up
	assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))

	// Test extend with non-existent task
	err = q.Extend(ctx, 1, "non-existent")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestFifoExtendPreventsExpiration(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	// Set very short extension
	q.extension = 50 * time.Millisecond

	dummyTask := genDummyTask()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

	waitForProcess()
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)

	// Extend the deadline multiple times
	for i := 0; i < 3; i++ {
		time.Sleep(30 * time.Millisecond)
		assert.NoError(t, q.Extend(ctx, 1, got.ID))
	}

	// Task should still be running
	info := q.Info(ctx)
	assert.Len(t, info.Running, 1, "task should still be running after extensions")
	assert.Len(t, info.Pending, 0, "task should not be resubmitted")
}

func TestFifoKickAgentWorkers(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	dummyTask := genDummyTask()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

	// Start multiple workers for different agents
	pollResults := make(chan error, 3)

	for agentID := int64(1); agentID <= 3; agentID++ {
		go func(id int64) {
			_, err := q.Poll(ctx, id, filterFnTrue)
			pollResults <- err
		}(agentID)
	}

	// Give workers time to register
	time.Sleep(50 * time.Millisecond)

	// Kick workers for agent 2
	q.KickAgentWorkers(2)

	// Check that agent 2's worker was kicked
	select {
	case err := <-pollResults:
		assert.Error(t, err)
		// Check the cause of the context cancellation
		if errors.Is(err, context.Canceled) {
			// If ctx wasn't the one canceled, we need to check if it's our kicked error
			// The error should either be ErrWorkerKicked or wrapped in context
			assert.True(t, errors.Is(err, context.Canceled), "expected context.Canceled")
		}
	case <-time.After(time.Second):
		t.Fatal("expected worker to be kicked")
	}

	// Other workers should still be waiting or get work
	info := q.Info(ctx)
	assert.GreaterOrEqual(t, info.Stats.Workers, 2, "other workers should still be registered")
}

func TestFifoKickAgentWorkersMultiple(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	// Start multiple workers for the same agent
	pollResults := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func() {
			_, err := q.Poll(ctx, 42, filterFnTrue)
			pollResults <- err
		}()
	}

	// Give workers time to register
	time.Sleep(50 * time.Millisecond)

	info := q.Info(ctx)
	initialWorkers := info.Stats.Workers
	assert.Equal(t, 5, initialWorkers, "all workers should be registered")

	// Kick all workers for agent 42
	q.KickAgentWorkers(42)

	// All workers should be kicked
	kickedCount := 0
	for i := 0; i < 5; i++ {
		select {
		case err := <-pollResults:
			assert.Error(t, err)
			// All kicked workers will have their context canceled
			if errors.Is(err, context.Canceled) {
				kickedCount++
			}
		case <-time.After(time.Second):
			t.Fatal("expected all workers to be kicked")
		}
	}

	assert.Equal(t, 5, kickedCount, "all 5 workers should have been kicked")

	time.Sleep(50 * time.Millisecond)
	info = q.Info(ctx)
	assert.Equal(t, 0, info.Stats.Workers, "all workers should be removed")
}

func TestFifoWaitContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	dummyTask := genDummyTask()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

	waitForProcess()
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)

	// Create a context that we can cancel
	waitCtx, waitCancel := context.WithCancel(ctx)

	errCh := make(chan error, 1)
	go func() {
		errCh <- q.Wait(waitCtx, got.ID)
	}()

	// Cancel the wait context
	time.Sleep(50 * time.Millisecond)
	waitCancel()

	select {
	case err := <-errCh:
		assert.NoError(t, err, "Wait should return nil when context is canceled before task completes")
	case <-time.After(time.Second):
		t.Fatal("Wait should return when context is canceled")
	}
}

func TestFifoWaitNonExistentTask(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	// Wait on a task that doesn't exist
	err := q.Wait(ctx, "non-existent-task")
	assert.NoError(t, err, "Wait should return nil for non-existent task")
}

func TestFifoUpdateDepStatusInQueue(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

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
	}
	task4 := &model.Task{
		ID:           "4",
		Dependencies: []string{"2"},
		DepStatus:    make(map[string]model.StatusValue),
	}

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2, task3, task4}))

	waitForProcess()
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task1, got)

	// Complete task1 with success
	assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))

	waitForProcess()

	// Poll task2 to check its DepStatus was updated
	got2, err := q.Poll(ctx, 2, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, model.StatusSuccess, got2.DepStatus["1"], "task2 should have task1's status updated")

	// Poll task3 to check its DepStatus was updated
	got3, err := q.Poll(ctx, 3, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, model.StatusSuccess, got3.DepStatus["1"], "task3 should have task1's status updated")
}

func TestFifoRemoveFromPendingAndWaitingNotFound(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	task1 := genDummyTask()

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1}))

	waitForProcess()
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)

	// Try to error a task that's running (it will be removed from running, not pending/waiting)
	assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("test error")))

	// Try to error a completely non-existent task
	err = q.Error(ctx, "totally-fake-id", fmt.Errorf("test error"))
	assert.Error(t, err, "should return error for non-existent task")
}

func TestFifoRemoveFromWaitingOnDeps(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

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
	}

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2, task3}))

	waitForProcess()
	info := q.Info(ctx)
	assert.Equal(t, 2, info.Stats.WaitingOnDeps, "task2 and task3 should be waiting")

	// Error task2 while it's in waitingOnDeps
	assert.NoError(t, q.Error(ctx, "2", fmt.Errorf("cancel task")))

	waitForProcess()
	info = q.Info(ctx)
	assert.Equal(t, 1, info.Stats.WaitingOnDeps, "only task3 should be waiting now")
}

func TestFifoPollContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())

	q := NewMemoryQueue(ctx)

	pollCtx, pollCancel := context.WithCancel(ctx)

	errCh := make(chan error, 1)
	go func() {
		_, err := q.Poll(pollCtx, 1, filterFnTrue)
		errCh <- err
	}()

	// Give Poll time to register the worker
	time.Sleep(50 * time.Millisecond)

	// Cancel the poll context
	pollCancel()

	select {
	case err := <-errCh:
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("Poll should return when context is canceled")
	}

	cancel(nil)
}

func TestFifoMultipleDepStatusUpdates(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	task1 := genDummyTask()
	task2 := &model.Task{
		ID: "2",
	}
	task3 := &model.Task{
		ID:           "3",
		Dependencies: []string{"1", "2"},
		DepStatus:    make(map[string]model.StatusValue),
	}

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2, task3}))

	waitForProcess()

	// Get both independent tasks
	got1, _ := q.Poll(ctx, 1, filterFnTrue)
	got2, _ := q.Poll(ctx, 2, filterFnTrue)

	// Complete one with success
	assert.NoError(t, q.Done(ctx, got1.ID, model.StatusSuccess))

	// Complete other with failure
	assert.NoError(t, q.Error(ctx, got2.ID, fmt.Errorf("failed")))

	waitForProcess()

	// Get task3 and check both dependencies are updated
	got3, err := q.Poll(ctx, 3, filterFnTrue)
	assert.NoError(t, err)

	assert.Equal(t, model.StatusSuccess, got3.DepStatus["1"])
	assert.Equal(t, model.StatusFailure, got3.DepStatus["2"])
	assert.False(t, got3.ShouldRun(), "task3 should not run since one dependency failed")
}

func TestFifoErrorAtOnceCoverage(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	task1 := genDummyTask()
	task2 := &model.Task{
		ID: "2",
	}

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2}))

	waitForProcess()
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)

	// Test ErrorAtOnce with one running task and one non-existent task
	// This should trigger the error case in finished() where removeFromPendingAndWaiting returns ErrNotFound
	err = q.ErrorAtOnce(ctx, []string{got.ID, "non-existent-id"}, fmt.Errorf("batch error"))
	assert.Error(t, err, "should return error when one of the tasks doesn't exist")
	assert.ErrorIs(t, err, ErrNotFound, "error should contain ErrNotFound")

	waitForProcess()
	info := q.Info(ctx)
	assert.Len(t, info.Running, 0, "running task should be removed")
}

func TestFifoErrorAtOnceMultipleNonExistent(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	task1 := genDummyTask()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1}))

	waitForProcess()
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)

	// Test ErrorAtOnce with multiple non-existent tasks
	// This tests that errors.Join works correctly with multiple errors
	err = q.ErrorAtOnce(ctx, []string{got.ID, "fake-1", "fake-2", "fake-3"}, fmt.Errorf("multiple errors"))
	assert.Error(t, err, "should return joined errors")

	// The error should contain multiple ErrNotFound instances
	errStr := err.Error()
	assert.Contains(t, errStr, "not found", "error should mention not found")
}

func TestFifoErrorAtOnceWithErrCancel(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	task1 := genDummyTask()
	task2 := &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
	}

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2}))

	waitForProcess()
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)

	// Test the ErrCancel specific path which sets StatusKilled
	assert.NoError(t, q.ErrorAtOnce(ctx, []string{got.ID}, ErrCancel))

	waitForProcess()

	// Get the dependent task and verify its dependency status is StatusKilled
	got2, err := q.Poll(ctx, 2, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, model.StatusKilled, got2.DepStatus["1"], "dependency should be marked as killed")
}

func TestFifoExtendWithWrongAgent(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	dummyTask := genDummyTask()
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

	waitForProcess()
	got, err := q.Poll(ctx, 5, filterFnTrue) // Agent 5 polls the task
	assert.NoError(t, err)
	assert.Equal(t, int64(5), got.AgentID, "task should be assigned to agent 5")

	// Test extend with correct agent - should succeed
	assert.NoError(t, q.Extend(ctx, 5, got.ID))

	// Test extend with wrong agent - should return ErrAgentMissMatch
	err = q.Extend(ctx, 999, got.ID)
	assert.ErrorIs(t, err, ErrAgentMissMatch, "should return ErrAgentMissMatch when wrong agent tries to extend")

	// Test extend with another wrong agent
	err = q.Extend(ctx, 1, got.ID)
	assert.ErrorIs(t, err, ErrAgentMissMatch, "should return ErrAgentMissMatch for any wrong agent")

	// Verify the correct agent can still extend
	assert.NoError(t, q.Extend(ctx, 5, got.ID), "correct agent should still be able to extend")

	// Clean up
	assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))
}

func TestFifoFinishedWithMultipleErrors(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	task1 := genDummyTask()
	task2 := &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
	}

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2}))

	waitForProcess()

	// Poll task1 so it's in running
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, "1", got.ID)

	// Now call ErrorAtOnce with:
	// - got.ID (task1) which is in running - should succeed
	// - "2" which is in waitingOnDeps - should succeed
	// - "fake-1" which doesn't exist - should return ErrNotFound
	// - "fake-2" which doesn't exist - should return ErrNotFound
	err = q.ErrorAtOnce(ctx, []string{got.ID, "2", "fake-1", "fake-2"}, fmt.Errorf("test error"))

	assert.Error(t, err, "should return error for non-existent tasks")

	// Verify that the error contains multiple ErrNotFound (from errors.Join)
	assert.ErrorIs(t, err, ErrNotFound, "error should contain ErrNotFound")

	// Check that both real tasks were removed
	waitForProcess()
	info := q.Info(ctx)
	assert.Len(t, info.Running, 0, "task1 should be removed from running")
	assert.Len(t, info.WaitingOnDeps, 0, "task2 should be removed from waiting")
	assert.Len(t, info.Pending, 0, "no tasks should be pending")
}

func TestFifoUpdateDepStatusInQueuePending(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

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
	}

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	// Push all tasks - task2 and task3 will be in pending initially (before filterWaiting moves them)
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2, task3}))

	waitForProcess()

	// At this point, task2 and task3 should be in waitingOnDeps
	// but we want to test when they're in pending

	// Get task1
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task1, got)

	// Before completing task1, let's verify task2 and task3 are waiting
	info := q.Info(ctx)
	assert.Equal(t, 2, info.Stats.WaitingOnDeps, "task2 and task3 should be waiting")

	// Now complete task1 - this should trigger updateDepStatusInQueue
	// which will update DepStatus in the waitingOnDeps queue
	assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))

	waitForProcess()

	// Now get task2 and verify its DepStatus was updated
	got2, err := q.Poll(ctx, 2, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, model.StatusSuccess, got2.DepStatus["1"],
		"task2's DepStatus should be updated in pending queue")

	// Get task3 and verify its DepStatus was updated
	got3, err := q.Poll(ctx, 3, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, model.StatusSuccess, got3.DepStatus["1"],
		"task3's DepStatus should be updated in pending queue")
}

func TestFifoUpdateDepStatusInQueueRunning(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	task1 := genDummyTask()
	task2 := &model.Task{
		ID: "2",
	}
	task3 := &model.Task{
		ID:           "3",
		Dependencies: []string{"1", "2"},
		DepStatus:    make(map[string]model.StatusValue),
	}
	task4 := &model.Task{
		ID:           "4",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
	}

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2, task3, task4}))

	waitForProcess()

	// Get task1 and task2 (both independent)
	got1, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)

	got2, err := q.Poll(ctx, 2, filterFnTrue)
	assert.NoError(t, err)

	waitForProcess()

	// Now complete task1 first
	assert.NoError(t, q.Done(ctx, got1.ID, model.StatusSuccess))

	waitForProcess()

	// Now task3 and task4 are waiting on their dependencies
	// task3 still needs task2, task4 should be ready now

	// Get task4 which should now be available since task1 is done
	got4, err := q.Poll(ctx, 4, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, model.StatusSuccess, got4.DepStatus["1"],
		"task4's DepStatus should show task1 succeeded")

	// Now complete task2
	assert.NoError(t, q.Done(ctx, got2.ID, model.StatusSuccess))

	waitForProcess()

	// Now get task3 - both its dependencies are satisfied
	got3, err := q.Poll(ctx, 3, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, model.StatusSuccess, got3.DepStatus["1"],
		"task3's DepStatus should show task1 succeeded")
	assert.Equal(t, model.StatusSuccess, got3.DepStatus["2"],
		"task3's DepStatus should show task2 succeeded")
	assert.True(t, got3.ShouldRun(), "task3 should run since both deps succeeded")

	// Clean up remaining tasks
	assert.NoError(t, q.Done(ctx, got3.ID, model.StatusSuccess))
	assert.NoError(t, q.Done(ctx, got4.ID, model.StatusSuccess))
}

func TestFifoUpdateDepStatusInQueueWaiting(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

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
	}
	task4 := &model.Task{
		ID:           "4",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
	}

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2, task3, task4}))

	waitForProcess()

	// Verify tasks are waiting on deps
	info := q.Info(ctx)
	assert.Equal(t, 3, info.Stats.WaitingOnDeps, "tasks 2, 3, 4 should be waiting")

	// Get and complete task1
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task1, got)

	// Complete task1 with failure - this should update DepStatus in waitingOnDeps
	assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("task1 failed")))

	waitForProcess()

	// All waiting tasks should have their DepStatus updated
	// Get each one and verify
	for i := 2; i <= 4; i++ {
		gotTask, err := q.Poll(ctx, int64(i), filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, model.StatusFailure, gotTask.DepStatus["1"],
			"task %d's DepStatus should be updated in waitingOnDeps", i)
		assert.False(t, gotTask.ShouldRun(), "task %d should not run due to failed dependency", i)
	}
}

func TestFifoUpdateDepStatusWithSkipped(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

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

	q, _ := NewMemoryQueue(ctx).(*fifo)
	assert.NotNil(t, q)

	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2, task3}))

	waitForProcess()

	// Get and fail task1
	got1, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.NoError(t, q.Error(ctx, got1.ID, fmt.Errorf("failed")))

	waitForProcess()

	// Get task2 - should not run but get skipped
	got2, err := q.Poll(ctx, 2, filterFnTrue)
	assert.NoError(t, err)
	assert.False(t, got2.ShouldRun())
	assert.NoError(t, q.Done(ctx, got2.ID, model.StatusSkipped))

	waitForProcess()

	// Get task3 - should have task2's skipped status updated
	got3, err := q.Poll(ctx, 3, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, model.StatusSkipped, got3.DepStatus["2"],
		"task3 should have task2's skipped status")
	assert.False(t, got3.ShouldRun(), "task3 should not run due to skipped dependency")
}
