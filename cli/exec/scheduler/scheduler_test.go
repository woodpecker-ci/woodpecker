// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scheduler_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"

	"go.woodpecker-ci.org/woodpecker/v3/cli/exec/scheduler"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
)

// item is a test helper for assembling a builder.Item with only the
// fields the scheduler inspects.
func item(name string, deps ...string) *builder.Item {
	return &builder.Item{
		Workflow:  &builder.Workflow{Name: name},
		DependsOn: append([]string(nil), deps...),
	}
}

// collectEvents drains all events from a scheduler into a slice. The
// returned done channel is closed when the input channel closes, so
// tests can synchronize before reading the slice.
func collectEvents(ch <-chan scheduler.Event) (*[]scheduler.Event, <-chan struct{}) {
	var out []scheduler.Event
	done := make(chan struct{})
	go func() {
		defer close(done)
		for ev := range ch {
			out = append(out, ev)
		}
	}()
	return &out, done
}

func TestLinearChainRunsInOrder(t *testing.T) {
	var order []string
	var mu sync.Mutex

	run := func(_ context.Context, it *builder.Item) error {
		mu.Lock()
		order = append(order, it.Workflow.Name)
		mu.Unlock()
		return nil
	}

	s := scheduler.New(scheduler.Options{
		Items: []*builder.Item{
			item("a"),
			item("b", "a"),
			item("c", "b"),
		},
		Run:      run,
		Parallel: 4, // plenty of slots — ordering must come from deps, not from cap
	})

	err := s.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, order)
}

func TestParallelIndependentWorkflowsRespectCap(t *testing.T) {
	var inFlight int32
	var maxInFlight int32
	// Block each worker until we release them, so we can observe the
	// actual concurrency instead of racing through trivially-fast run
	// functions.
	start := make(chan struct{})

	run := func(_ context.Context, _ *builder.Item) error {
		cur := atomic.AddInt32(&inFlight, 1)
		for {
			prev := atomic.LoadInt32(&maxInFlight)
			if cur <= prev || atomic.CompareAndSwapInt32(&maxInFlight, prev, cur) {
				break
			}
		}
		<-start
		atomic.AddInt32(&inFlight, -1)
		return nil
	}

	const n = 10
	const cap = 3
	items := make([]*builder.Item, n)
	for i := 0; i < n; i++ {
		items[i] = item(fmt.Sprintf("wf%d", i))
	}

	s := scheduler.New(scheduler.Options{
		Items:    items,
		Run:      run,
		Parallel: cap,
	})

	errCh := make(chan error, 1)
	go func() { errCh <- s.Run(context.Background()) }()

	// Give the scheduler time to saturate the semaphore before
	// releasing the workers. If the scheduler respects the cap, exactly
	// `cap` workers will be running; if it doesn't, we'll see more.
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&inFlight) == cap
	}, 2*time.Second, 5*time.Millisecond, "scheduler did not reach capacity")

	// Double-check by waiting a moment — if the cap is broken,
	// additional workers will have piled on by now.
	time.Sleep(50 * time.Millisecond)
	assert.LessOrEqual(t, atomic.LoadInt32(&maxInFlight), int32(cap),
		"more than %d workers ran concurrently", cap)

	close(start)
	require.NoError(t, <-errCh)
}

func TestFailurePropagatesAsBlocked(t *testing.T) {
	//   root (fails)
	//    ├── a (should be blocked)
	//    │    └── c (should be blocked, transitive)
	//    └── b (should be blocked)
	//   sibling (unrelated, should still succeed)
	var ranSibling atomic.Bool
	run := func(_ context.Context, it *builder.Item) error {
		switch it.Workflow.Name {
		case "root":
			return errors.New("root failed")
		case "sibling":
			ranSibling.Store(true)
			return nil
		}
		t.Errorf("unexpected run of %q; should have been blocked", it.Workflow.Name)
		return nil
	}

	evCh := make(chan scheduler.Event, 64)
	events, done := collectEvents(evCh)

	s := scheduler.New(scheduler.Options{
		Items: []*builder.Item{
			item("root"),
			item("a", "root"),
			item("b", "root"),
			item("c", "a"),
			item("sibling"),
		},
		Run:    run,
		Events: evCh,
	})

	err := s.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "root failed")
	assert.True(t, ranSibling.Load(), "sibling workflow should have run (fail-fast is OFF)")

	// Block until the collector has observed channel close; this is
	// both the sync barrier for reading the slice and a latency-free
	// alternative to time.Sleep.
	<-done

	byWf := map[string]scheduler.State{}
	for _, ev := range *events {
		if ev.State.Terminal() {
			byWf[ev.Workflow] = ev.State
		}
	}
	assert.Equal(t, scheduler.StateFailure, byWf["root"])
	assert.Equal(t, scheduler.StateBlocked, byWf["a"])
	assert.Equal(t, scheduler.StateBlocked, byWf["b"])
	assert.Equal(t, scheduler.StateBlocked, byWf["c"])
	assert.Equal(t, scheduler.StateSuccess, byWf["sibling"])
}

func TestContextCancelStopsNewWorkAndWaitsForRunning(t *testing.T) {
	// a is running; cancel ctx mid-run. b (depends on a) should never
	// start. a's run func should receive a canceled ctx.
	started := make(chan struct{})
	run := func(ctx context.Context, it *builder.Item) error {
		switch it.Workflow.Name {
		case "a":
			close(started)
			<-ctx.Done()
			return ctx.Err()
		case "b":
			t.Error("b should never have started")
			return nil
		}
		return nil
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	s := scheduler.New(scheduler.Options{
		Items: []*builder.Item{
			item("a"),
			item("b", "a"),
		},
		Run: run,
	})

	errCh := make(chan error, 1)
	go func() { errCh <- s.Run(ctx) }()

	<-started
	cancel(nil)

	select {
	case err := <-errCh:
		require.Error(t, err, "canceled run func returns ctx.Err which is aggregated")
		assert.True(t, errors.Is(err, context.Canceled))
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not return after ctx cancel")
	}
}

func TestMultipleIndependentFailuresAggregate(t *testing.T) {
	errA := errors.New("a broke")
	errB := errors.New("b broke")
	run := func(_ context.Context, it *builder.Item) error {
		switch it.Workflow.Name {
		case "a":
			return errA
		case "b":
			return errB
		}
		return nil
	}

	s := scheduler.New(scheduler.Options{
		Items: []*builder.Item{item("a"), item("b"), item("c")},
		Run:   run,
	})

	err := s.Run(context.Background())
	require.Error(t, err)

	errs := multierr.Errors(err)
	assert.Len(t, errs, 2)
	assert.Contains(t, errs, errA)
	assert.Contains(t, errs, errB)
}

func TestDuplicateWorkflowName(t *testing.T) {
	s := scheduler.New(scheduler.Options{
		Items: []*builder.Item{item("a"), item("a")},
		Run:   func(context.Context, *builder.Item) error { return nil },
	})
	err := s.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate workflow name")
}

func TestUnknownDependency(t *testing.T) {
	// builder normally strips items with missing deps before the
	// scheduler sees them, so this is a defensive check for programmer
	// error at the scheduler boundary.
	s := scheduler.New(scheduler.Options{
		Items: []*builder.Item{item("a", "missing")},
		Run:   func(context.Context, *builder.Item) error { return nil },
	})
	err := s.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown workflow")
}

func TestEmptyItemsIsNoOp(t *testing.T) {
	evCh := make(chan scheduler.Event, 1)
	s := scheduler.New(scheduler.Options{
		Items:  nil,
		Run:    func(context.Context, *builder.Item) error { return nil },
		Events: evCh,
	})
	err := s.Run(context.Background())
	require.NoError(t, err)
	// Channel must be closed.
	_, ok := <-evCh
	assert.False(t, ok)
}

func TestMissingRunFunc(t *testing.T) {
	s := scheduler.New(scheduler.Options{Items: []*builder.Item{item("a")}})
	err := s.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Run is required")
}

func TestEventsChannelClosedOnReturn(t *testing.T) {
	evCh := make(chan scheduler.Event, 16)
	s := scheduler.New(scheduler.Options{
		Items:  []*builder.Item{item("a")},
		Run:    func(context.Context, *builder.Item) error { return nil },
		Events: evCh,
	})
	require.NoError(t, s.Run(context.Background()))
	// Drain then verify closure.
	for range evCh {
	}
}

func TestBlockedErrorMessage(t *testing.T) {
	e := &scheduler.BlockedError{Dependency: "root"}
	assert.Contains(t, e.Error(), "root")
}

func TestStateStringAndTerminal(t *testing.T) {
	cases := []struct {
		s        scheduler.State
		terminal bool
		str      string
	}{
		{scheduler.StatePending, false, "pending"},
		{scheduler.StateReady, false, "ready"},
		{scheduler.StateRunning, false, "running"},
		{scheduler.StateSuccess, true, "success"},
		{scheduler.StateFailure, true, "failure"},
		{scheduler.StateBlocked, true, "blocked"},
		{scheduler.StateCanceled, true, "canceled"},
	}
	for _, c := range cases {
		assert.Equal(t, c.terminal, c.s.Terminal(), c.str)
		assert.Equal(t, c.str, c.s.String())
	}
}

func TestUnboundedParallel(t *testing.T) {
	// Parallel < 0 means unbounded. Launch more items than any
	// reasonable NumCPU and verify they all run concurrently.
	var inFlight int32
	var maxInFlight int32
	start := make(chan struct{})
	run := func(_ context.Context, _ *builder.Item) error {
		cur := atomic.AddInt32(&inFlight, 1)
		for {
			prev := atomic.LoadInt32(&maxInFlight)
			if cur <= prev || atomic.CompareAndSwapInt32(&maxInFlight, prev, cur) {
				break
			}
		}
		<-start
		return nil
	}
	const n = 64
	items := make([]*builder.Item, n)
	for i := 0; i < n; i++ {
		items[i] = item(fmt.Sprintf("wf%d", i))
	}
	s := scheduler.New(scheduler.Options{Items: items, Run: run, Parallel: -1})
	errCh := make(chan error, 1)
	go func() { errCh <- s.Run(context.Background()) }()
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&inFlight) == n
	}, 2*time.Second, 5*time.Millisecond)
	close(start)
	require.NoError(t, <-errCh)
	assert.Equal(t, int32(n), maxInFlight)
}
