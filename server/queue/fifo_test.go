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

var noContext = context.Background()

func TestFifo(t *testing.T) {
	want := &model.Task{ID: "1"}

	q := New(context.Background())
	assert.NoError(t, q.Push(noContext, want))
	info := q.Info(noContext)
	assert.Len(t, info.Pending, 1, "expect task in pending queue")

	got, err := q.Poll(noContext, 1, func(*model.Task) bool { return true })
	assert.NoError(t, err)
	assert.Equal(t, want, got)

	info = q.Info(noContext)
	assert.Len(t, info.Pending, 0, "expect task removed from pending queue")
	assert.Len(t, info.Running, 1, "expect task in running queue")

	assert.NoError(t, q.Done(noContext, got.ID, model.StatusSuccess))
	info = q.Info(noContext)
	assert.Len(t, info.Pending, 0, "expect task removed from pending queue")
	assert.Len(t, info.Running, 0, "expect task removed from running queue")
}

func TestFifoExpire(t *testing.T) {
	want := &model.Task{ID: "1"}

	q, _ := New(context.Background()).(*fifo)
	q.extension = 0
	assert.NoError(t, q.Push(noContext, want))
	info := q.Info(noContext)
	assert.Len(t, info.Pending, 1, "expect task in pending queue")

	got, err := q.Poll(noContext, 1, func(*model.Task) bool { return true })
	assert.NoError(t, err)
	assert.Equal(t, want, got)

	q.process()
	assert.Len(t, info.Pending, 1, "expect task re-added to pending queue")
}

func TestFifoWait(t *testing.T) {
	want := &model.Task{ID: "1"}

	q, _ := New(context.Background()).(*fifo)
	assert.NoError(t, q.Push(noContext, want))

	got, err := q.Poll(noContext, 1, func(*model.Task) bool { return true })
	assert.NoError(t, err)
	assert.Equal(t, want, got)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		assert.NoError(t, q.Wait(noContext, got.ID))
		wg.Done()
	}()

	<-time.After(time.Millisecond)
	assert.NoError(t, q.Done(noContext, got.ID, model.StatusSuccess))
	wg.Wait()
}

func TestFifoEvict(t *testing.T) {
	t1 := &model.Task{ID: "1"}

	q := New(context.Background())
	assert.NoError(t, q.Push(noContext, t1))
	info := q.Info(noContext)
	assert.Len(t, info.Pending, 1, "expect task in pending queue")
	err := q.Evict(noContext, t1.ID)
	assert.NoError(t, err)
	info = q.Info(noContext)
	assert.Len(t, info.Pending, 0)
	err = q.Evict(noContext, t1.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestFifoDependencies(t *testing.T) {
	task1 := &model.Task{
		ID: "1",
	}

	task2 := &model.Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]model.StatusValue),
	}

	q, _ := New(context.Background()).(*fifo)
	assert.NoError(t, q.PushAtOnce(noContext, []*model.Task{task2, task1}))

	got, err := q.Poll(noContext, 1, func(*model.Task) bool { return true })
	assert.NoError(t, err)
	assert.Equal(t, task1, got)

	assert.NoError(t, q.Done(noContext, got.ID, model.StatusSuccess))

	got, err = q.Poll(noContext, 1, func(*model.Task) bool { return true })
	assert.NoError(t, err)
	assert.Equal(t, task2, got)
}

func TestFifoErrors(t *testing.T) {
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

	q, _ := New(context.Background()).(*fifo)
	assert.NoError(t, q.PushAtOnce(noContext, []*model.Task{task2, task3, task1}))

	got, err := q.Poll(noContext, 1, func(*model.Task) bool { return true })
	assert.NoError(t, err)
	assert.Equal(t, task1, got)

	assert.NoError(t, q.Error(noContext, got.ID, fmt.Errorf("exit code 1, there was an error")))

	got, err = q.Poll(noContext, 1, func(*model.Task) bool { return true })
	assert.NoError(t, err)
	assert.Equal(t, task2, got)
	assert.False(t, got.ShouldRun())

	got, err = q.Poll(noContext, 1, func(*model.Task) bool { return true })
	assert.NoError(t, err)
	assert.Equal(t, task3, got)
	assert.True(t, got.ShouldRun())
}

func TestFifoErrors2(t *testing.T) {
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

	q, _ := New(context.Background()).(*fifo)
	assert.NoError(t, q.PushAtOnce(noContext, []*model.Task{task2, task3, task1}))

	for i := 0; i < 2; i++ {
		got, err := q.Poll(noContext, 1, func(*model.Task) bool { return true })
		assert.NoError(t, err)
		assert.False(t, got != task1 && got != task2, "expect task1 or task2 returned from queue as task3 depends on them")

		if got != task1 {
			assert.NoError(t, q.Done(noContext, got.ID, model.StatusSuccess))
		}
		if got != task2 {
			assert.NoError(t, q.Error(noContext, got.ID, fmt.Errorf("exit code 1, there was an error")))
		}
	}

	got, err := q.Poll(noContext, 1, func(*model.Task) bool { return true })
	assert.NoError(t, err)
	assert.Equal(t, task3, got)
	assert.False(t, got.ShouldRun())
}

func TestFifoErrorsMultiThread(t *testing.T) {
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

	q, _ := New(context.Background()).(*fifo)
	assert.NoError(t, q.PushAtOnce(noContext, []*model.Task{task2, task3, task1}))

	obtainedWorkCh := make(chan *model.Task)

	for i := 0; i < 10; i++ {
		go func(i int) {
			for {
				fmt.Printf("Worker %d started\n", i)
				got, err := q.Poll(noContext, 1, func(*model.Task) bool { return true })
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
				assert.NoError(t, q.Error(noContext, got.ID, fmt.Errorf("exit code 1, there was an error")))
				go func() {
					for {
						fmt.Printf("Worker spawned\n")
						got, _ := q.Poll(noContext, 1, func(*model.Task) bool { return true })
						obtainedWorkCh <- got
					}
				}()
			case !task2Processed:
				assert.Equal(t, task2, got)
				task2Processed = true
				assert.NoError(t, q.Done(noContext, got.ID, model.StatusSuccess))
				go func() {
					for {
						fmt.Printf("Worker spawned\n")
						got, _ := q.Poll(noContext, 1, func(*model.Task) bool { return true })
						obtainedWorkCh <- got
					}
				}()
			default:
				assert.Equal(t, task3, got)
				assert.False(t, got.ShouldRun(), "expect task3 should not run, task1 succeeded but task2 failed")
				return
			}

		case <-time.After(5 * time.Second):
			info := q.Info(noContext)
			fmt.Println(info.String())
			t.Errorf("test timed out")
			return
		}
	}
}

func TestFifoTransitiveErrors(t *testing.T) {
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

	q, _ := New(context.Background()).(*fifo)
	assert.NoError(t, q.PushAtOnce(noContext, []*model.Task{task2, task3, task1}))

	got, err := q.Poll(noContext, 1, func(*model.Task) bool { return true })
	assert.NoError(t, err)
	assert.Equal(t, task1, got)
	assert.NoError(t, q.Error(noContext, got.ID, fmt.Errorf("exit code 1, there was an error")))

	got, err = q.Poll(noContext, 1, func(*model.Task) bool { return true })
	assert.NoError(t, err)
	assert.Equal(t, task2, got)
	assert.False(t, got.ShouldRun(), "expect task2 should not run, since task1 failed")
	assert.NoError(t, q.Done(noContext, got.ID, model.StatusSkipped))

	got, err = q.Poll(noContext, 1, func(*model.Task) bool { return true })
	assert.NoError(t, err)
	assert.Equal(t, task3, got)
	assert.False(t, got.ShouldRun(), "expect task3 should not run, task1 failed, thus task2 was skipped, task3 should be skipped too")
}

func TestFifoCancel(t *testing.T) {
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

	q, _ := New(context.Background()).(*fifo)
	assert.NoError(t, q.PushAtOnce(noContext, []*model.Task{task2, task3, task1}))

	_, _ = q.Poll(noContext, 1, func(*model.Task) bool { return true })
	assert.NoError(t, q.Error(noContext, task1.ID, fmt.Errorf("canceled")))
	assert.NoError(t, q.Error(noContext, task2.ID, fmt.Errorf("canceled")))
	assert.NoError(t, q.Error(noContext, task3.ID, fmt.Errorf("canceled")))

	info := q.Info(noContext)
	assert.Len(t, info.Pending, 0, "all pipelines should be canceled")
}

func TestFifoPause(t *testing.T) {
	task1 := &model.Task{
		ID: "1",
	}

	q, _ := New(context.Background()).(*fifo)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_, _ = q.Poll(noContext, 1, func(*model.Task) bool { return true })
		wg.Done()
	}()

	q.Pause()
	t0 := time.Now()
	assert.NoError(t, q.Push(noContext, task1))
	time.Sleep(20 * time.Millisecond)
	q.Resume()

	wg.Wait()
	t1 := time.Now()

	assert.Greater(t, t1.Sub(t0), 20*time.Millisecond, "should have waited til resume")

	q.Pause()
	assert.NoError(t, q.Push(noContext, task1))
	q.Resume()
	_, _ = q.Poll(noContext, 1, func(*model.Task) bool { return true })
}

func TestFifoPauseResume(t *testing.T) {
	task1 := &model.Task{
		ID: "1",
	}

	q, _ := New(context.Background()).(*fifo)
	q.Pause()
	assert.NoError(t, q.Push(noContext, task1))
	q.Resume()

	_, _ = q.Poll(noContext, 1, func(*model.Task) bool { return true })
}

func TestWaitingVsPending(t *testing.T) {
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

	q, _ := New(context.Background()).(*fifo)
	assert.NoError(t, q.PushAtOnce(noContext, []*model.Task{task2, task3, task1}))

	got, _ := q.Poll(noContext, 1, func(*model.Task) bool { return true })

	info := q.Info(noContext)
	assert.Equal(t, 2, info.Stats.WaitingOnDeps)

	assert.NoError(t, q.Error(noContext, got.ID, fmt.Errorf("exit code 1, there was an error")))
	got, err := q.Poll(noContext, 1, func(*model.Task) bool { return true })
	assert.NoError(t, err)
	assert.EqualValues(t, task2, got)

	info = q.Info(noContext)
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
