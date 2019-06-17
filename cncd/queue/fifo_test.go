package queue

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

var noContext = context.Background()

func TestFifo(t *testing.T) {
	want := &Task{ID: "1"}

	q := New()
	q.Push(noContext, want)
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

	q.Done(noContext, got.ID)
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
	q.Push(noContext, want)
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
	q.Push(noContext, want)

	got, _ := q.Poll(noContext, func(*Task) bool { return true })
	if got != want {
		t.Errorf("expect task returned form queue")
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		q.Wait(noContext, got.ID)
		wg.Done()
	}()

	<-time.After(time.Millisecond)
	q.Done(noContext, got.ID)
	wg.Wait()
}

func TestFifoEvict(t *testing.T) {
	t1 := &Task{ID: "1"}

	q := New()
	q.Push(noContext, t1)
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
		DepStatus:    make(map[string]bool),
	}

	q := New().(*fifo)
	q.Push(noContext, task2)
	q.Push(noContext, task1)

	got, _ := q.Poll(noContext, func(*Task) bool { return true })
	if got != task1 {
		t.Errorf("expect task1 returned from queue as task2 depends on it")
		return
	}

	q.Done(noContext, got.ID)

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
		DepStatus:    make(map[string]bool),
	}

	task3 := &Task{
		ID:           "3",
		Dependencies: []string{"1"},
		DepStatus:    make(map[string]bool),
		RunOn:        []string{"success", "failure"},
	}

	q := New().(*fifo)
	q.Push(noContext, task2)
	q.Push(noContext, task3)
	q.Push(noContext, task1)

	got, _ := q.Poll(noContext, func(*Task) bool { return true })
	if got != task1 {
		t.Errorf("expect task1 returned from queue as task2 depends on it")
		return
	}

	q.Error(noContext, got.ID, fmt.Errorf("exitcode 1, there was an error"))

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

func TestShouldRun(t *testing.T) {
	task := &Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus: map[string]bool{
			"1": true,
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
		DepStatus: map[string]bool{
			"1": true,
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
		DepStatus: map[string]bool{
			"1": false,
		},
	}
	if task.ShouldRun() {
		t.Errorf("expect task to not run")
		return
	}

	task = &Task{
		ID:           "2",
		Dependencies: []string{"1"},
		DepStatus: map[string]bool{
			"1": true,
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
		DepStatus: map[string]bool{
			"1": false,
		},
		RunOn: []string{"failure"},
	}
	if !task.ShouldRun() {
		t.Errorf("expect task to run")
		return
	}
}
