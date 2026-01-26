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

func TestFifoBasicOperations(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	t.Run("basic push poll done flow", func(t *testing.T) {
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
	})

	t.Run("task expiration", func(t *testing.T) {
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
	})

	t.Run("pause and resume", func(t *testing.T) {
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
	})

	t.Run("pause then resume", func(t *testing.T) {
		q, _ := NewMemoryQueue(ctx).(*fifo)
		assert.NotNil(t, q)
		dummyTask := genDummyTask()

		q.Pause()
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))
		q.Resume()

		_, _ = q.Poll(ctx, 1, filterFnTrue)
	})
}

func TestFifoWait(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	t.Run("wait on task completion", func(t *testing.T) {
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
	})

	t.Run("wait returns error on expiration", func(t *testing.T) {
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
	})

	t.Run("wait with context canceled", func(t *testing.T) {
		q, _ := NewMemoryQueue(ctx).(*fifo)
		assert.NotNil(t, q)
		dummyTask := genDummyTask()

		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

		waitForProcess()
		got, err := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)

		waitCtx, waitCancel := context.WithCancel(ctx)

		errCh := make(chan error, 1)
		go func() {
			errCh <- q.Wait(waitCtx, got.ID)
		}()

		time.Sleep(50 * time.Millisecond)
		waitCancel()

		select {
		case err := <-errCh:
			assert.NoError(t, err, "Wait should return nil when context is canceled before task completes")
		case <-time.After(time.Second):
			t.Fatal("Wait should return when context is canceled")
		}
	})

	t.Run("wait on non-existent task", func(t *testing.T) {
		q, _ := NewMemoryQueue(ctx).(*fifo)
		assert.NotNil(t, q)

		err := q.Wait(ctx, "non-existent-task")
		assert.NoError(t, err, "Wait should return nil for non-existent task")
	})
}

func TestFifoDependencies(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	t.Run("basic dependency ordering", func(t *testing.T) {
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
	})

	t.Run("waiting vs pending stats", func(t *testing.T) {
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
	})
}

func TestFifoErrors(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	t.Run("error handling with run_on conditions", func(t *testing.T) {
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
	})

	t.Run("multiple dependency failures", func(t *testing.T) {
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
	})

	t.Run("transitive error propagation", func(t *testing.T) {
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
	})

	t.Run("multithreaded error handling", func(t *testing.T) {
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
	})
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

func TestFifoShouldRun(t *testing.T) {
	tests := []struct {
		name     string
		task     *model.Task
		expected bool
		reason   string
	}{
		{
			name: "failure only - success status",
			task: &model.Task{
				ID:           "2",
				Dependencies: []string{"1"},
				DepStatus:    map[string]model.StatusValue{"1": model.StatusSuccess},
				RunOn:        []string{"failure"},
			},
			expected: false,
			reason:   "expect task to not run, it runs on failure only",
		},
		{
			name: "failure and success - success status",
			task: &model.Task{
				ID:           "2",
				Dependencies: []string{"1"},
				DepStatus:    map[string]model.StatusValue{"1": model.StatusSuccess},
				RunOn:        []string{"failure", "success"},
			},
			expected: true,
		},
		{
			name: "no run_on - failure status",
			task: &model.Task{
				ID:           "2",
				Dependencies: []string{"1"},
				DepStatus:    map[string]model.StatusValue{"1": model.StatusFailure},
			},
			expected: false,
		},
		{
			name: "success only - success status",
			task: &model.Task{
				ID:           "2",
				Dependencies: []string{"1"},
				DepStatus:    map[string]model.StatusValue{"1": model.StatusSuccess},
				RunOn:        []string{"success"},
			},
			expected: true,
		},
		{
			name: "failure only - failure status",
			task: &model.Task{
				ID:           "2",
				Dependencies: []string{"1"},
				DepStatus:    map[string]model.StatusValue{"1": model.StatusFailure},
				RunOn:        []string{"failure"},
			},
			expected: true,
		},
		{
			name: "no run_on - skipped status",
			task: &model.Task{
				ID:           "2",
				Dependencies: []string{"1"},
				DepStatus:    map[string]model.StatusValue{"1": model.StatusSkipped},
			},
			expected: false,
			reason:   "task should not run if dependency is skipped",
		},
		{
			name: "failure only - skipped status",
			task: &model.Task{
				ID:           "2",
				Dependencies: []string{"1"},
				DepStatus:    map[string]model.StatusValue{"1": model.StatusSkipped},
				RunOn:        []string{"failure"},
			},
			expected: true,
			reason:   "on failure, tasks should run on skipped deps, something failed higher up the chain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.task.ShouldRun()
			if tt.reason != "" {
				assert.Equal(t, tt.expected, result, tt.reason)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFifoScoring(t *testing.T) {
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

	t.Run("batch error on running tasks", func(t *testing.T) {
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

		assert.NoError(t, q.ErrorAtOnce(ctx, []string{got1.ID, got2.ID}, fmt.Errorf("batch error")))

		waitForProcess()
		info := q.Info(ctx)
		assert.Len(t, info.Running, 0, "expect tasks removed from running queue")
		assert.Len(t, info.Pending, 1, "expect remaining task in pending")
	})

	t.Run("batch error on pending tasks", func(t *testing.T) {
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

		assert.NoError(t, q.ErrorAtOnce(ctx, []string{"2", "3"}, fmt.Errorf("pending error")))

		waitForProcess()
		info := q.Info(ctx)
		assert.Len(t, info.Pending, 1, "only task1 should remain")
		assert.Len(t, info.WaitingOnDeps, 0, "task3 should be removed from waiting")
	})

	t.Run("error at once with cancel error", func(t *testing.T) {
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

		assert.NoError(t, q.ErrorAtOnce(ctx, []string{got1.ID}, ErrCancel))

		waitForProcess()
		info := q.Info(ctx)
		assert.Len(t, info.Running, 0, "task should be removed from running")
	})

	t.Run("error at once with non-existent tasks", func(t *testing.T) {
		task1 := genDummyTask()
		task2 := &model.Task{
			ID: "2",
		}

		q, _ := NewMemoryQueue(ctx).(*fifo)
		assert.NotNil(t, q)

		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2}))

		waitForProcess()
		got, err := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)

		err = q.ErrorAtOnce(ctx, []string{got.ID, "non-existent-id"}, fmt.Errorf("batch error"))
		assert.Error(t, err, "should return error when one of the tasks doesn't exist")
		assert.ErrorIs(t, err, ErrNotFound, "error should contain ErrNotFound")

		waitForProcess()
		info := q.Info(ctx)
		assert.Len(t, info.Running, 0, "running task should be removed")
	})

	t.Run("error at once with multiple non-existent tasks", func(t *testing.T) {
		q, _ := NewMemoryQueue(ctx).(*fifo)
		assert.NotNil(t, q)

		task1 := genDummyTask()
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1}))

		waitForProcess()
		got, err := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)

		err = q.ErrorAtOnce(ctx, []string{got.ID, "fake-1", "fake-2", "fake-3"}, fmt.Errorf("multiple errors"))
		assert.Error(t, err, "should return joined errors")

		errStr := err.Error()
		assert.Contains(t, errStr, "not found", "error should mention not found")
	})

	t.Run("error at once with ErrCancel and dependency update", func(t *testing.T) {
		task1 := genDummyTask()
		task2 := &model.Task{
			ID:           "2",
			Dependencies: []string{"1"},
			DepStatus:    make(map[string]model.StatusValue),
		}

		q, _ := NewMemoryQueue(ctx).(*fifo)
		assert.NotNil(t, q)

		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2}))

		waitForProcess()
		got, err := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)

		assert.NoError(t, q.ErrorAtOnce(ctx, []string{got.ID}, ErrCancel))

		waitForProcess()

		got2, err := q.Poll(ctx, 2, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, model.StatusKilled, got2.DepStatus["1"], "dependency should be marked as killed")
	})

	t.Run("finished with multiple errors", func(t *testing.T) {
		task1 := genDummyTask()
		task2 := &model.Task{
			ID:           "2",
			Dependencies: []string{"1"},
			DepStatus:    make(map[string]model.StatusValue),
		}

		q, _ := NewMemoryQueue(ctx).(*fifo)
		assert.NotNil(t, q)

		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1, task2}))

		waitForProcess()

		got, err := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, "1", got.ID)

		err = q.ErrorAtOnce(ctx, []string{got.ID, "2", "fake-1", "fake-2"}, fmt.Errorf("test error"))

		assert.Error(t, err, "should return error for non-existent tasks")
		assert.ErrorIs(t, err, ErrNotFound, "error should contain ErrNotFound")

		waitForProcess()
		info := q.Info(ctx)
		assert.Len(t, info.Running, 0, "task1 should be removed from running")
		assert.Len(t, info.WaitingOnDeps, 0, "task2 should be removed from waiting")
		assert.Len(t, info.Pending, 0, "no tasks should be pending")
	})
}

func TestFifoExtend(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	t.Run("basic extend operations", func(t *testing.T) {
		q, _ := NewMemoryQueue(ctx).(*fifo)
		assert.NotNil(t, q)

		dummyTask := genDummyTask()
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

		waitForProcess()
		got, err := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)

		assert.NoError(t, q.Extend(ctx, 1, got.ID))

		err = q.Extend(ctx, 999, got.ID)
		assert.ErrorIs(t, err, ErrAgentMissMatch)

		assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))

		err = q.Extend(ctx, 1, "non-existent")
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("extend prevents expiration", func(t *testing.T) {
		q, _ := NewMemoryQueue(ctx).(*fifo)
		assert.NotNil(t, q)

		q.extension = 50 * time.Millisecond

		dummyTask := genDummyTask()
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

		waitForProcess()
		got, err := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)

		for i := 0; i < 3; i++ {
			time.Sleep(30 * time.Millisecond)
			assert.NoError(t, q.Extend(ctx, 1, got.ID))
		}

		info := q.Info(ctx)
		assert.Len(t, info.Running, 1, "task should still be running after extensions")
		assert.Len(t, info.Pending, 0, "task should not be resubmitted")
	})

	t.Run("extend with wrong agent", func(t *testing.T) {
		q, _ := NewMemoryQueue(ctx).(*fifo)
		assert.NotNil(t, q)

		dummyTask := genDummyTask()
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

		waitForProcess()
		got, err := q.Poll(ctx, 5, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), got.AgentID, "task should be assigned to agent 5")

		assert.NoError(t, q.Extend(ctx, 5, got.ID))

		err = q.Extend(ctx, 999, got.ID)
		assert.ErrorIs(t, err, ErrAgentMissMatch, "should return ErrAgentMissMatch when wrong agent tries to extend")

		err = q.Extend(ctx, 1, got.ID)
		assert.ErrorIs(t, err, ErrAgentMissMatch, "should return ErrAgentMissMatch for any wrong agent")

		assert.NoError(t, q.Extend(ctx, 5, got.ID), "correct agent should still be able to extend")

		assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))
	})
}

func TestFifoKickAgentWorkers(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	t.Run("kick single agent workers", func(t *testing.T) {
		q, _ := NewMemoryQueue(ctx).(*fifo)
		assert.NotNil(t, q)

		dummyTask := genDummyTask()
		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{dummyTask}))

		pollResults := make(chan error, 3)

		for agentID := int64(1); agentID <= 3; agentID++ {
			go func(id int64) {
				_, err := q.Poll(ctx, id, filterFnTrue)
				pollResults <- err
			}(agentID)
		}

		time.Sleep(50 * time.Millisecond)

		q.KickAgentWorkers(2)

		select {
		case err := <-pollResults:
			assert.Error(t, err)
			if errors.Is(err, context.Canceled) {
				assert.True(t, errors.Is(err, context.Canceled), "expected context.Canceled")
			}
		case <-time.After(time.Second):
			t.Fatal("expected worker to be kicked")
		}

		info := q.Info(ctx)
		assert.GreaterOrEqual(t, info.Stats.Workers, 2, "other workers should still be registered")
	})

	t.Run("kick multiple workers for same agent", func(t *testing.T) {
		q, _ := NewMemoryQueue(ctx).(*fifo)
		assert.NotNil(t, q)

		pollResults := make(chan error, 5)

		for i := 0; i < 5; i++ {
			go func() {
				_, err := q.Poll(ctx, 42, filterFnTrue)
				pollResults <- err
			}()
		}

		time.Sleep(50 * time.Millisecond)

		info := q.Info(ctx)
		initialWorkers := info.Stats.Workers
		assert.Equal(t, 5, initialWorkers, "all workers should be registered")

		q.KickAgentWorkers(42)

		kickedCount := 0
		for i := 0; i < 5; i++ {
			select {
			case err := <-pollResults:
				assert.Error(t, err)
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
	})
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

	time.Sleep(50 * time.Millisecond)

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

func TestFifoDepStatusUpdates(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	t.Run("update dep status in queue", func(t *testing.T) {
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

		assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))

		waitForProcess()

		got2, err := q.Poll(ctx, 2, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, model.StatusSuccess, got2.DepStatus["1"], "task2 should have task1's status updated")

		got3, err := q.Poll(ctx, 3, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, model.StatusSuccess, got3.DepStatus["1"], "task3 should have task1's status updated")
	})

	t.Run("multiple dep status updates", func(t *testing.T) {
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

		got1, _ := q.Poll(ctx, 1, filterFnTrue)
		got2, _ := q.Poll(ctx, 2, filterFnTrue)

		assert.NoError(t, q.Done(ctx, got1.ID, model.StatusSuccess))
		assert.NoError(t, q.Error(ctx, got2.ID, fmt.Errorf("failed")))

		waitForProcess()

		got3, err := q.Poll(ctx, 3, filterFnTrue)
		assert.NoError(t, err)

		assert.Equal(t, model.StatusSuccess, got3.DepStatus["1"])
		assert.Equal(t, model.StatusFailure, got3.DepStatus["2"])
		assert.False(t, got3.ShouldRun(), "task3 should not run since one dependency failed")
	})

	t.Run("update dep status in pending queue", func(t *testing.T) {
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

		got, err := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, task1, got)

		info := q.Info(ctx)
		assert.Equal(t, 2, info.Stats.WaitingOnDeps, "task2 and task3 should be waiting")

		assert.NoError(t, q.Done(ctx, got.ID, model.StatusSuccess))

		waitForProcess()

		got2, err := q.Poll(ctx, 2, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, model.StatusSuccess, got2.DepStatus["1"],
			"task2's DepStatus should be updated in pending queue")

		got3, err := q.Poll(ctx, 3, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, model.StatusSuccess, got3.DepStatus["1"],
			"task3's DepStatus should be updated in pending queue")
	})

	t.Run("update dep status in running queue", func(t *testing.T) {
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

		got1, err := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)

		got2, err := q.Poll(ctx, 2, filterFnTrue)
		assert.NoError(t, err)

		waitForProcess()

		assert.NoError(t, q.Done(ctx, got1.ID, model.StatusSuccess))

		waitForProcess()

		got4, err := q.Poll(ctx, 4, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, model.StatusSuccess, got4.DepStatus["1"],
			"task4's DepStatus should show task1 succeeded")

		assert.NoError(t, q.Done(ctx, got2.ID, model.StatusSuccess))

		waitForProcess()

		got3, err := q.Poll(ctx, 3, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, model.StatusSuccess, got3.DepStatus["1"],
			"task3's DepStatus should show task1 succeeded")
		assert.Equal(t, model.StatusSuccess, got3.DepStatus["2"],
			"task3's DepStatus should show task2 succeeded")
		assert.True(t, got3.ShouldRun(), "task3 should run since both deps succeeded")

		assert.NoError(t, q.Done(ctx, got3.ID, model.StatusSuccess))
		assert.NoError(t, q.Done(ctx, got4.ID, model.StatusSuccess))
	})

	t.Run("update dep status in waiting queue", func(t *testing.T) {
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

		info := q.Info(ctx)
		assert.Equal(t, 3, info.Stats.WaitingOnDeps, "tasks 2, 3, 4 should be waiting")

		got, err := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, task1, got)

		assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("task1 failed")))

		waitForProcess()

		for i := 2; i <= 4; i++ {
			gotTask, err := q.Poll(ctx, int64(i), filterFnTrue)
			assert.NoError(t, err)
			assert.Equal(t, model.StatusFailure, gotTask.DepStatus["1"],
				"task %d's DepStatus should be updated in waitingOnDeps", i)
			assert.False(t, gotTask.ShouldRun(), "task %d should not run due to failed dependency", i)
		}
	})

	t.Run("update dep status with skipped", func(t *testing.T) {
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

		got1, err := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)
		assert.NoError(t, q.Error(ctx, got1.ID, fmt.Errorf("failed")))

		waitForProcess()

		got2, err := q.Poll(ctx, 2, filterFnTrue)
		assert.NoError(t, err)
		assert.False(t, got2.ShouldRun())
		assert.NoError(t, q.Done(ctx, got2.ID, model.StatusSkipped))

		waitForProcess()

		got3, err := q.Poll(ctx, 3, filterFnTrue)
		assert.NoError(t, err)
		assert.Equal(t, model.StatusSkipped, got3.DepStatus["2"],
			"task3 should have task2's skipped status")
		assert.False(t, got3.ShouldRun(), "task3 should not run due to skipped dependency")
	})
}

func TestFifoRemoveOperations(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	t.Run("remove from pending and waiting not found", func(t *testing.T) {
		task1 := genDummyTask()

		q, _ := NewMemoryQueue(ctx).(*fifo)
		assert.NotNil(t, q)

		assert.NoError(t, q.PushAtOnce(ctx, []*model.Task{task1}))

		waitForProcess()
		got, err := q.Poll(ctx, 1, filterFnTrue)
		assert.NoError(t, err)

		assert.NoError(t, q.Error(ctx, got.ID, fmt.Errorf("test error")))

		err = q.Error(ctx, "totally-fake-id", fmt.Errorf("test error"))
		assert.Error(t, err, "should return error for non-existent task")
	})

	t.Run("remove from waiting on deps", func(t *testing.T) {
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

		assert.NoError(t, q.Error(ctx, "2", fmt.Errorf("cancel task")))

		waitForProcess()
		info = q.Info(ctx)
		assert.Equal(t, 1, info.Stats.WaitingOnDeps, "only task3 should be waiting now")
	})
}
