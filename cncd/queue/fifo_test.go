package queue

import (
	"context"
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
