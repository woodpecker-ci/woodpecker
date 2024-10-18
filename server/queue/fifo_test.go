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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

var filterFnTrue = func(*model.Task) (bool, int) { return true, 1 }

func TestFifo(t *testing.T) {
	want := &model.Task{ID: "1"}
	ctx := context.Background()

	q := New(ctx)
	assert.NoError(t, q.Push(ctx, want))
	info := q.Info(ctx)
	assert.Len(t, info.Pending, 1, "expect task in pending queue")

	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, want, got)

	info = q.Info(ctx)
	assert.Len(t, info.Pending, 0, "expect task removed from pending queue")
	assert.Len(t, info.Running, 1, "expect task in running queue")

	assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))
	info = q.Info(ctx)
	assert.Len(t, info.Pending, 0, "expect task removed from pending queue")
	assert.Len(t, info.Running, 0, "expect task removed from running queue")
}

func TestFifoExpire(t *testing.T) {
	want := &model.Task{ID: "1"}
	ctx, cancel := context.WithCancelCause(context.Background())

	q, _ := New(ctx).(*fifo)
	q.extension = 0
	assert.NoError(t, q.Push(ctx, want))
	info := q.Info(ctx)
	assert.Len(t, info.Pending, 1, "expect task in pending queue")

	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, want, got)

	// cancel the context to let the process func end
	go func() {
		time.Sleep(time.Millisecond)
		cancel(nil)
	}()
	q.process()
	assert.Len(t, info.Pending, 1, "expect task re-added to pending queue")
}

func TestFifoWait(t *testing.T) {
	want := &model.Task{ID: "1"}
	ctx := context.Background()

	q, _ := New(ctx).(*fifo)
	assert.NoError(t, q.Push(ctx, want))

	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, want, got)

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

func TestFifoEvict(t *testing.T) {
	t1 := &model.Task{ID: "1"}
	ctx := context.Background()

	q := New(ctx)
	assert.NoError(t, q.Push(ctx, t1))
	info := q.Info(ctx)
	assert.Len(t, info.Pending, 1, "expect task in pending queue")
	err := q.Evict(ctx, t1.ID)
	assert.NoError(t, err)
	info = q.Info(ctx)
	assert.Len(t, info.Pending, 0)
	err = q.Evict(ctx, t1.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestFifoDependencies(t *testing.T) {
	ctx := context.Background()
	task1 := &model.Task{
		ID: "1",
	}

	task2 := &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
	}

	q, _ := New(ctx).(*fifo)
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task1}))

	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task1, got)

	assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))

	got, err = q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task2, got)
}

func TestFifoErrors(t *testing.T) {
	ctx := context.Background()
	task1 := &model.Task{
		ID: "1",
	}

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

	q, _ := New(ctx).(*fifo)
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))

	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task1, got)

	assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("exit code 1, there was an error")))

	got, err = q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task2, got)
	assert.False(t, got.ShouldRun())

	got, err = q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task3, got)
	assert.True(t, got.ShouldRun())
}

func TestFifoErrors2(t *testing.T) {
	ctx := context.Background()
	task1 := &model.Task{
		ID: "1",
	}

	task2 := &model.Task{
		ID: "2",
	}

	task3 := &model.Task{
		ID:           "3",
		Dependencies: []string{"1", "2"},
		DepStatus:    make(map[string]model.StatusValue),
	}

	q, _ := New(ctx).(*fifo)
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))

	for i := 0; i < 2; i++ {
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

	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task3, got)
	assert.False(t, got.ShouldRun())
}

func TestFifoErrorsMultiThread(t *testing.T) {
	ctx := context.Background()
	task1 := &model.Task{
		ID: "1",
	}

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

	q, _ := New(ctx).(*fifo)
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))

	obtainedWorkCh := make(chan *model.Task)

	for i := 0; i < 10; i++ {
		go func(i int) {
			for {
				fmt.Printf("Worker %d started\n", i)
				got, err := q.Poll(ctx, 1, filterFnTrue)
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
						got, _ := q.Poll(ctx, 1, filterFnTrue)
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
						got, _ := q.Poll(ctx, 1, filterFnTrue)
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
	ctx := context.Background()
	task1 := &model.Task{
		ID: "1",
	}

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

	q, _ := New(ctx).(*fifo)
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))

	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task1, got)
	assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("exit code 1, there was an error")))

	got, err = q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task2, got)
	assert.False(t, got.ShouldRun(), "expect task2 should not run, since task1 failed")
	assert.NoError(t, q.Done(ctx, got.ID, model.StatusSkipped))

	got, err = q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.Equal(t, task3, got)
	assert.False(t, got.ShouldRun(), "expect task3 should not run, task1 failed, thus task2 was skipped, task3 should be skipped too")
}

func TestFifoCancel(t *testing.T) {
	ctx := context.Background()
	task1 := &model.Task{
		ID: "1",
	}

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

	q, _ := New(ctx).(*fifo)
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))

	_, _ = q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, q.Error(ctx, task1.ID, fmt.Errorf("canceled")))
	assert.NoError(t, q.Error(ctx, task2.ID, fmt.Errorf("canceled")))
	assert.NoError(t, q.Error(ctx, task3.ID, fmt.Errorf("canceled")))

	info := q.Info(ctx)
	assert.Len(t, info.Pending, 0, "all pipelines should be canceled")
}

func TestFifoPause(t *testing.T) {
	ctx := context.Background()
	task1 := &model.Task{
		ID: "1",
	}

	q, _ := New(ctx).(*fifo)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_, _ = q.Poll(ctx, 1, filterFnTrue)
		wg.Done()
	}()

	q.Pause()
	t0 := time.Now()
	assert.NoError(t, q.Push(ctx, task1))
	time.Sleep(20 * time.Millisecond)
	q.Resume()

	wg.Wait()
	t1 := time.Now()

	assert.Greater(t, t1.Sub(t0), 20*time.Millisecond, "should have waited til resume")

	q.Pause()
	assert.NoError(t, q.Push(ctx, task1))
	q.Resume()
	_, _ = q.Poll(ctx, 1, filterFnTrue)
}

func TestFifoPauseResume(t *testing.T) {
	ctx := context.Background()
	task1 := &model.Task{
		ID: "1",
	}

	q, _ := New(ctx).(*fifo)
	q.Pause()
	assert.NoError(t, q.Push(ctx, task1))
	q.Resume()

	_, _ = q.Poll(ctx, 1, filterFnTrue)
}

func TestWaitingVsPending(t *testing.T) {
	ctx := context.Background()
	task1 := &model.Task{
		ID: "1",
	}

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

	q, _ := New(ctx).(*fifo)
	assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task2, task3, task1}))

	got, _ := q.Poll(ctx, 1, filterFnTrue)

	info := q.Info(ctx)
	assert.Equal(t, 2, info.Stats.WaitingOnDeps)

	assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("exit code 1, there was an error")))
	got, err := q.Poll(ctx, 1, filterFnTrue)
	assert.NoError(t, err)
	assert.EqualValues(t, task2, got)

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
	ctx := context.Background()
	q := New(ctx)

	// Create tasks with different labels
	tasks := []*model.Task{
		{ID: "1", Labels: map[string]string{"org-id": "123", "platform": "linux"}},
		{ID: "2", Labels: map[string]string{"org-id": "456", "platform": "linux"}},
		{ID: "3", Labels: map[string]string{"org-id": "789", "platform": "windows"}},
		{ID: "4", Labels: map[string]string{"org-id": "123", "platform": "linux"}},
		{ID: "5", Labels: map[string]string{"org-id": "*", "platform": "linux"}},
	}

	for _, task := range tasks {
		assert.NoError(t, q.Push(ctx, task))
	}

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
