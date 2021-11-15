package queue

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var noContext = context.Background()

func TestFifo(t *testing.T) {
	want := &Task{ID: "1"}

	q := New()
	assert.NoError(t, q.Push(noContext, want))
	info := q.Info(noContext)
	if len(info.Pending) != 1 {
		t.Errorf("expect task in pending queue")
		return
	}

	got, _ := q.Poll(noContext, func(*Task) bool { return true })
	if got != want {
		t.Errorf("expect task returned form queue")
		return
	}

	info = q.Info(noContext)
	if len(info.Pending) != 0 {
		t.Errorf("expect task removed from pending queue")
		return
	}
	if len(info.Running) != 1 {
		t.Errorf("expect task in running queue")
		return
	}

	assert.NoError(t, q.Done(noContext, got.ID, StatusSuccess))
	info = q.Info(noContext)
	if len(info.Pending) != 0 {
		t.Errorf("expect task removed from pending queue")
		return
	}
	if len(info.Running) != 0 {
		t.Errorf("expect task removed from running queue")
		return
	}
}

func TestFifoExpire(t *testing.T) {
	want := &Task{ID: "1"}

	q := New().(*fifo)
	q.extension = 0
	assert.NoError(t, q.Push(noContext, want))
	info := q.Info(noContext)
	if len(info.Pending) != 1 {
		t.Errorf("expect task in pending queue")
		return
	}

	got, _ := q.Poll(noContext, func(*Task) bool { return true })
	if got != want {
		t.Errorf("expect task returned form queue")
		return
	}

	q.process()
	if len(info.Pending) != 1 {
		t.Errorf("expect task re-added to pending queue")
		return
	}
}

func TestFifoWait(t *testing.T) {
	want := &Task{ID: "1"}

	q := New().(*fifo)
	assert.NoError(t, q.Push(noContext, want))

	got, _ := q.Poll(noContext, func(*Task) bool { return true })
	if got != want {
		t.Errorf("expect task returned form queue")
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		assert.NoError(t, q.Wait(noContext, got.ID))
		wg.Done()
	}()

	<-time.After(time.Millisecond)
	assert.NoError(t, q.Done(noContext, got.ID, StatusSuccess))
	wg.Wait()
}

func TestFifoEvict(t *testing.T) {
	t1 := &Task{ID: "1"}

	q := New()
	assert.NoError(t, q.Push(noContext, t1))
	info := q.Info(noContext)
	if len(info.Pending) != 1 {
		t.Errorf("expect task in pending queue")
	}
	if err := q.Evict(noContext, t1.ID); err != nil {
		t.Errorf("expect task evicted from queue")
	}
	info = q.Info(noContext)
	if len(info.Pending) != 0 {
		t.Errorf("expect pending queue has zero items")
	}
	if err := q.Evict(noContext, t1.ID); err != ErrNotFound {
		t.Errorf("expect not found error when evicting item not in queue, got %s", err)
	}
}

func TestFifoDependencies(t *testing.T) {
	task1 := &Task{
		ID: "1",
	}

	task2 := &Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]string),
	}

	q := New().(*fifo)
	assert.NoError(t, q.PushAtOnce(noContext, []*Task{task2, task1}))

	got, _ := q.Poll(noContext, func(*Task) bool { return true })
	if got != task1 {
		t.Errorf("expect task1 returned from queue as task2 depends on it")
		return
	}

	assert.NoError(t, q.Done(noContext, got.ID, StatusSuccess))

	got, _ = q.Poll(noContext, func(*Task) bool { return true })
	if got != task2 {
		t.Errorf("expect task2 returned from queue")
		return
	}
}

func TestFifoErrors(t *testing.T) {
	task1 := &Task{
		ID: "1",
	}

	task2 := &Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]string),
	}

	task3 := &Task{
		ID:           "3",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]string),
		RunOn:        []string{"success", "failure"},
	}

	q := New().(*fifo)
	assert.NoError(t, q.PushAtOnce(noContext, []*Task{task2, task3, task1}))

	got, _ := q.Poll(noContext, func(*Task) bool { return true })
	if got != task1 {
		t.Errorf("expect task1 returned from queue as task2 depends on it")
		return
	}

	assert.NoError(t, q.Error(noContext, got.ID, fmt.Errorf("exitcode 1, there was an error")))

	got, _ = q.Poll(noContext, func(*Task) bool { return true })
	if got != task2 {
		t.Errorf("expect task2 returned from queue")
		return
	}

	if got.ShouldRun() {
		t.Errorf("expect task2 should not run, since task1 failed")
		return
	}

	got, _ = q.Poll(noContext, func(*Task) bool { return true })
	if got != task3 {
		t.Errorf("expect task3 returned from queue")
		return
	}

	if !got.ShouldRun() {
		t.Errorf("expect task3 should run, task1 failed, but task3 runs on failure too")
		return
	}
}

func TestFifoErrors2(t *testing.T) {
	task1 := &Task{
		ID: "1",
	}

	task2 := &Task{
		ID: "2",
	}

	task3 := &Task{
		ID:           "3",
		Dependencies: []string{"1", "2"},
		DepStatus:    make(map[string]string),
	}

	q := New().(*fifo)
	assert.NoError(t, q.PushAtOnce(noContext, []*Task{task2, task3, task1}))

	for i := 0; i < 2; i++ {
		got, _ := q.Poll(noContext, func(*Task) bool { return true })
		if got != task1 && got != task2 {
			t.Errorf("expect task1 or task2 returned from queue as task3 depends on them")
			return
		}

		if got != task1 {
			assert.NoError(t, q.Done(noContext, got.ID, StatusSuccess))
		}
		if got != task2 {
			assert.NoError(t, q.Error(noContext, got.ID, fmt.Errorf("exitcode 1, there was an error")))
		}
	}

	got, _ := q.Poll(noContext, func(*Task) bool { return true })
	if got != task3 {
		t.Errorf("expect task3 returned from queue")
		return
	}

	if got.ShouldRun() {
		t.Errorf("expect task3 should not run, task1 succeeded but task2 failed")
		return
	}
}

func TestFifoErrorsMultiThread(t *testing.T) {
	task1 := &Task{
		ID: "1",
	}

	task2 := &Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]string),
	}

	task3 := &Task{
		ID:           "3",
		Dependencies: []string{"1", "2"},
		DepStatus:    make(map[string]string),
	}

	q := New().(*fifo)
	assert.NoError(t, q.PushAtOnce(noContext, []*Task{task2, task3, task1}))

	obtainedWorkCh := make(chan *Task)

	for i := 0; i < 10; i++ {
		go func(i int) {
			for {
				fmt.Printf("Worker %d started\n", i)
				got, _ := q.Poll(noContext, func(*Task) bool { return true })
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

			if !task1Processed {
				if got != task1 {
					t.Errorf("expect task1 returned from queue as task2 and task3 depends on it")
					return
				} else {
					task1Processed = true
					assert.NoError(t, q.Error(noContext, got.ID, fmt.Errorf("exitcode 1, there was an error")))
					go func() {
						for {
							fmt.Printf("Worker spawned\n")
							got, _ := q.Poll(noContext, func(*Task) bool { return true })
							obtainedWorkCh <- got
						}
					}()
				}
			} else if !task2Processed {
				if got != task2 {
					t.Errorf("expect task2 returned from queue")
					return
				} else {
					task2Processed = true
					assert.NoError(t, q.Done(noContext, got.ID, StatusSuccess))
					go func() {
						for {
							fmt.Printf("Worker spawned\n")
							got, _ := q.Poll(noContext, func(*Task) bool { return true })
							obtainedWorkCh <- got
						}
					}()
				}
			} else {
				if got != task3 {
					t.Errorf("expect task3 returned from queue")
					return
				}

				if got.ShouldRun() {
					t.Errorf("expect task3 should not run, task1 succeeded but task2 failed")
					return
				} else {
					return
				}
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
	task1 := &Task{
		ID: "1",
	}

	task2 := &Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]string),
	}

	task3 := &Task{
		ID:           "3",
		Dependencies: []string{"2"},
		DepStatus:    make(map[string]string),
	}

	q := New().(*fifo)
	assert.NoError(t, q.PushAtOnce(noContext, []*Task{task2, task3, task1}))

	got, _ := q.Poll(noContext, func(*Task) bool { return true })
	if got != task1 {
		t.Errorf("expect task1 returned from queue as task2 depends on it")
		return
	}
	assert.NoError(t, q.Error(noContext, got.ID, fmt.Errorf("exitcode 1, there was an error")))

	got, _ = q.Poll(noContext, func(*Task) bool { return true })
	if got != task2 {
		t.Errorf("expect task2 returned from queue")
		return
	}
	if got.ShouldRun() {
		t.Errorf("expect task2 should not run, since task1 failed")
		return
	}
	assert.NoError(t, q.Done(noContext, got.ID, StatusSkipped))

	got, _ = q.Poll(noContext, func(*Task) bool { return true })
	if got != task3 {
		t.Errorf("expect task3 returned from queue")
		return
	}
	if got.ShouldRun() {
		t.Errorf("expect task3 should not run, task1 failed, thus task2 was skipped, task3 should be skipped too")
		return
	}
}

func TestFifoCancel(t *testing.T) {
	task1 := &Task{
		ID: "1",
	}

	task2 := &Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]string),
	}

	task3 := &Task{
		ID:           "3",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]string),
		RunOn:        []string{"success", "failure"},
	}

	q := New().(*fifo)
	assert.NoError(t, q.PushAtOnce(noContext, []*Task{task2, task3, task1}))

	_, _ = q.Poll(noContext, func(*Task) bool { return true })
	assert.NoError(t, q.Error(noContext, task1.ID, fmt.Errorf("cancelled")))
	assert.NoError(t, q.Error(noContext, task2.ID, fmt.Errorf("cancelled")))
	assert.NoError(t, q.Error(noContext, task3.ID, fmt.Errorf("cancelled")))

	info := q.Info(noContext)
	if len(info.Pending) != 0 {
		t.Errorf("All pipelines should be cancelled")
		return
	}
}

func TestFifoPause(t *testing.T) {
	task1 := &Task{
		ID: "1",
	}

	q := New().(*fifo)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_, _ = q.Poll(noContext, func(*Task) bool { return true })
		wg.Done()
	}()

	q.Pause()
	t0 := time.Now()
	assert.NoError(t, q.Push(noContext, task1))
	time.Sleep(20 * time.Millisecond)
	q.Resume()

	wg.Wait()
	t1 := time.Now()

	if t1.Sub(t0) < 20*time.Millisecond {
		t.Errorf("Should have waited til resume")
	}

	q.Pause()
	assert.NoError(t, q.Push(noContext, task1))
	q.Resume()
	_, _ = q.Poll(noContext, func(*Task) bool { return true })
}

func TestFifoPauseResume(t *testing.T) {
	task1 := &Task{
		ID: "1",
	}

	q := New().(*fifo)
	q.Pause()
	assert.NoError(t, q.Push(noContext, task1))
	q.Resume()

	_, _ = q.Poll(noContext, func(*Task) bool { return true })
}

func TestWaitingVsPending(t *testing.T) {
	task1 := &Task{
		ID: "1",
	}

	task2 := &Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]string),
	}

	task3 := &Task{
		ID:           "3",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]string),
		RunOn:        []string{"success", "failure"},
	}

	q := New().(*fifo)
	assert.NoError(t, q.PushAtOnce(noContext, []*Task{task2, task3, task1}))

	got, _ := q.Poll(noContext, func(*Task) bool { return true })

	info := q.Info(noContext)
	if info.Stats.WaitingOnDeps != 2 {
		t.Errorf("2 should wait on deps")
	}

	assert.NoError(t, q.Error(noContext, got.ID, fmt.Errorf("exitcode 1, there was an error")))
	got, _ = q.Poll(noContext, func(*Task) bool { return true })

	info = q.Info(noContext)
	if info.Stats.WaitingOnDeps != 0 {
		t.Errorf("0 should wait on deps")
	}
	if info.Stats.Pending != 1 {
		t.Errorf("1 should wait for worker")
	}
}

func TestShouldRun(t *testing.T) {
	task := &Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus: map[string]string{
			"1": StatusSuccess,
		},
		RunOn: []string{"failure"},
	}
	if task.ShouldRun() {
		t.Errorf("expect task to not run, it runs on failure only")
		return
	}

	task = &Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus: map[string]string{
			"1": StatusSuccess,
		},
		RunOn: []string{"failure", "success"},
	}
	if !task.ShouldRun() {
		t.Errorf("expect task to run")
		return
	}

	task = &Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus: map[string]string{
			"1": StatusFailure,
		},
	}
	if task.ShouldRun() {
		t.Errorf("expect task to not run")
		return
	}

	task = &Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus: map[string]string{
			"1": StatusSuccess,
		},
		RunOn: []string{"success"},
	}
	if !task.ShouldRun() {
		t.Errorf("expect task to run")
		return
	}

	task = &Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus: map[string]string{
			"1": StatusFailure,
		},
		RunOn: []string{"failure"},
	}
	if !task.ShouldRun() {
		t.Errorf("expect task to run")
		return
	}

	task = &Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus: map[string]string{
			"1": StatusSkipped,
		},
	}
	if task.ShouldRun() {
		t.Errorf("Tasked should not run if dependency is skipped")
		return
	}

	task = &Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus: map[string]string{
			"1": StatusSkipped,
		},
		RunOn: []string{"failure"},
	}
	if !task.ShouldRun() {
		t.Errorf("On Failure tasks should run on skipped deps, something failed higher up the chain")
		return
	}
}
