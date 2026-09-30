// Copyright 2026 Woodpecker Authors
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

package pipeline

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

// persisted is the store's view of one pipeline's workflow states, changed by
// the test while posts are in flight.
type persisted struct {
	mu        sync.Mutex
	pipeline  model.StatusValue
	workflows map[int64]model.StatusValue
}

func (p *persisted) set(workflowID int64, state model.StatusValue) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.workflows[workflowID] = state
}

// statusStore is a store whose WorkflowLoad and GetPipeline answer from p.
func statusStore(t *testing.T, p *persisted) *store_mocks.MockStore {
	s := store_mocks.NewMockStore(t)
	s.On("WorkflowLoad", mock.Anything).Return(func(id int64) (*model.Workflow, error) {
		p.mu.Lock()
		defer p.mu.Unlock()
		return &model.Workflow{ID: id, State: p.workflows[id]}, nil
	}).Maybe()
	s.On("GetPipeline", mock.Anything).Return(func(id int64) (*model.Pipeline, error) {
		p.mu.Lock()
		defer p.mu.Unlock()
		return &model.Pipeline{ID: id, Status: p.pipeline}, nil
	}).Maybe()
	return s
}

// posted records the states a forge received per workflow, in arrival order.
type posted struct {
	mu     sync.Mutex
	states map[int64][]model.StatusValue
}

func (p *posted) add(w *model.Workflow) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.states[w.ID] = append(p.states[w.ID], w.State)
}

func (p *posted) of(id int64) []model.StatusValue {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]model.StatusValue(nil), p.states[id]...)
}

// postedWorkflow is the workflow a forge.Status call was made with.
func postedWorkflow(t *testing.T, args mock.Arguments) *model.Workflow {
	w, ok := args.Get(4).(*model.Workflow)
	if !ok {
		t.Errorf("Status got a %T for its workflow", args.Get(4))
		return &model.Workflow{}
	}
	return w
}

func statusPipeline(n int) *model.Pipeline {
	pl := &model.Pipeline{ID: 1, Number: 7, Commit: "c0ffee", Event: model.EventPull, Status: model.StatusRunning}
	for i := 1; i <= n; i++ {
		pl.Workflows = append(pl.Workflows, &model.Workflow{ID: int64(i), PID: i, Name: "wf", State: model.StatusPending})
	}
	return pl
}

// A pipeline's statuses are posted after its workflows are queued, so an agent
// can run a workflow to the end, and its end be posted, before the pipeline's
// own post of that workflow goes out. That post must not carry the state the
// pipeline was created with: the forge keeps the newest status, and a pending
// one posted last stays pending forever.
func TestUpdatePipelineStatusPostsThePersistedState(t *testing.T) {
	pl := statusPipeline(3)
	db := &persisted{pipeline: model.StatusRunning, workflows: map[int64]model.StatusValue{
		1: model.StatusPending, 2: model.StatusSuccess, 3: model.StatusRunning,
	}}
	got := &posted{states: map[int64][]model.StatusValue{}}
	f := forge_mocks.NewMockForge(t)
	f.On("Status", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { got.add(postedWorkflow(t, args)) }).Return(nil)

	updatePipelineStatus(t.Context(), statusStore(t, db), f, pl, &model.Repo{}, &model.User{})

	assert.Equal(t, []model.StatusValue{model.StatusPending}, got.of(1))
	assert.Equal(t, []model.StatusValue{model.StatusSuccess}, got.of(2), "a finished workflow was posted with the state it was created with")
	assert.Equal(t, []model.StatusValue{model.StatusRunning}, got.of(3), "a running workflow was posted with the state it was created with")
	// the caller's snapshot is not rewritten
	for _, w := range pl.Workflows {
		assert.Equal(t, model.StatusPending, w.State)
	}
}

// The race the first test cannot show: the pipeline's post of a workflow has
// read its state and is still on its way to the forge when the agent finishes
// that workflow and posts its end. The end must reach the forge last.
func TestPostWorkflowStatusKeepsOneWorkflowsPostsInOrder(t *testing.T) {
	pl := statusPipeline(1)
	wf := pl.Workflows[0]
	db := &persisted{pipeline: model.StatusRunning, workflows: map[int64]model.StatusValue{1: model.StatusPending}}
	s := statusStore(t, db)
	got := &posted{states: map[int64][]model.StatusValue{}}

	inFlight := make(chan struct{})
	release := make(chan struct{})
	// release even when the test fails early, or the held post keeps its key
	// locked in the package-global statusLocks and the package hangs
	releaseOnce := sync.OnceFunc(func() { close(release) })
	t.Cleanup(releaseOnce)
	var first sync.Once
	f := forge_mocks.NewMockForge(t)
	f.On("Status", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			w := postedWorkflow(t, args)
			held := false
			first.Do(func() { held = true })
			if held {
				// the pipeline's post: hold it on the wire
				close(inFlight)
				<-release
			}
			got.add(w)
		}).Return(nil)

	bulk := make(chan struct{})
	go func() {
		defer close(bulk)
		updatePipelineStatus(context.Background(), s, f, pl, &model.Repo{}, &model.User{})
	}()
	<-inFlight

	// the agent finishes the workflow and posts it
	db.set(wf.ID, model.StatusSuccess)
	done := &model.Workflow{ID: wf.ID, PID: wf.PID, Name: wf.Name, State: model.StatusSuccess}
	agent := make(chan error, 1)
	go func() {
		agent <- PostWorkflowStatus(context.Background(), s, f, &model.User{}, &model.Repo{}, pl, done)
	}()

	// give the agent's post every chance to overtake the held one
	time.Sleep(100 * time.Millisecond)
	releaseOnce()
	<-bulk
	assert.NoError(t, <-agent)

	states := got.of(wf.ID)
	if assert.NotEmpty(t, states) {
		assert.Equal(t, model.StatusSuccess, states[len(states)-1], "posts reached the forge as %v: the finished workflow is left %s", states, states[len(states)-1])
	}
}

// Posts of different workflows do not wait for each other: the pipeline's
// posts stay concurrent.
func TestPostWorkflowStatusDoesNotSerializeOtherWorkflows(t *testing.T) {
	pl := statusPipeline(2)
	db := &persisted{pipeline: model.StatusRunning, workflows: map[int64]model.StatusValue{1: model.StatusRunning, 2: model.StatusRunning}}
	s := statusStore(t, db)

	inFlight := make(chan struct{})
	release := make(chan struct{})
	// release even when the test fails early, or the held post keeps its key
	// locked in the package-global statusLocks and the package hangs
	releaseOnce := sync.OnceFunc(func() { close(release) })
	t.Cleanup(releaseOnce)
	f := forge_mocks.NewMockForge(t)
	f.On("Status", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(w *model.Workflow) bool { return w.ID == 1 })).
		Run(func(mock.Arguments) { close(inFlight); <-release }).Return(nil)
	f.On("Status", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(w *model.Workflow) bool { return w.ID == 2 })).
		Return(nil)

	held := make(chan error, 1)
	go func() {
		held <- PostWorkflowStatus(context.Background(), s, f, &model.User{}, &model.Repo{}, pl, pl.Workflows[0])
	}()
	<-inFlight
	other := make(chan error, 1)
	go func() {
		other <- PostWorkflowStatus(context.Background(), s, f, &model.User{}, &model.Repo{}, pl, pl.Workflows[1])
	}()
	select {
	case err := <-other:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("a post of one workflow waited for another workflow's")
	}
	releaseOnce()
	assert.NoError(t, <-held)
}

// A deploy pipeline's statuses all land on its one deployment, and carry the
// pipeline's state: posts through different workflows are ordered too, and
// carry the persisted pipeline state.
func TestPostWorkflowStatusOrdersADeploymentsPosts(t *testing.T) {
	pl := statusPipeline(2)
	pl.Event = model.EventDeploy
	pl.Status = model.StatusPending
	db := &persisted{pipeline: model.StatusPending, workflows: map[int64]model.StatusValue{1: model.StatusPending, 2: model.StatusPending}}
	s := statusStore(t, db)

	var mu sync.Mutex
	var order []model.StatusValue
	inFlight := make(chan struct{})
	release := make(chan struct{})
	// release even when the test fails early, or the held post keeps its key
	// locked in the package-global statusLocks and the package hangs
	releaseOnce := sync.OnceFunc(func() { close(release) })
	t.Cleanup(releaseOnce)
	var first sync.Once
	f := forge_mocks.NewMockForge(t)
	f.On("Status", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			p, ok := args.Get(3).(*model.Pipeline)
			if !ok {
				t.Errorf("Status got a %T for its pipeline", args.Get(3))
				return
			}
			held := false
			first.Do(func() { held = true })
			if held {
				close(inFlight)
				<-release
			}
			mu.Lock()
			order = append(order, p.Status)
			mu.Unlock()
		}).Return(nil)

	held := make(chan error, 1)
	go func() {
		held <- PostWorkflowStatus(context.Background(), s, f, &model.User{}, &model.Repo{}, pl, pl.Workflows[0])
	}()
	<-inFlight
	db.mu.Lock()
	db.pipeline = model.StatusSuccess
	db.mu.Unlock()
	finished := *pl
	finished.Status = model.StatusSuccess
	other := make(chan error, 1)
	go func() {
		other <- PostWorkflowStatus(context.Background(), s, f, &model.User{}, &model.Repo{}, &finished, pl.Workflows[1])
	}()
	time.Sleep(100 * time.Millisecond)
	releaseOnce()
	assert.NoError(t, <-held)
	assert.NoError(t, <-other)

	mu.Lock()
	defer mu.Unlock()
	if assert.Len(t, order, 2) {
		assert.Equal(t, model.StatusSuccess, order[1], "the deployment's posts reached the forge as %v", order)
	}
}

// The status lock of a key is dropped once nobody holds or waits for it.
func TestStatusLocksForgetIdleKeys(t *testing.T) {
	var k keyedMutex
	unlock, err := k.lock(t.Context(), statusKey{id: 1})
	assert.NoError(t, err)
	unlock2, err := k.lock(t.Context(), statusKey{id: 2})
	assert.NoError(t, err)
	unlock()
	unlock2()
	k.mu.Lock()
	defer k.mu.Unlock()
	assert.Empty(t, k.locks)
}

// TestStatusLocksExcludeUnderContention drives the refcounted path: waiters
// queue on a key while its holder unlocks, which is exactly when an entry could
// be dropped from the map while still in use and a second mutex minted for the
// same key.
func TestStatusLocksExcludeUnderContention(t *testing.T) {
	var (
		k      keyedMutex
		inside atomic.Int32
		broken atomic.Bool
		wg     sync.WaitGroup
	)
	for range 32 {
		wg.Go(func() {
			for range 500 {
				unlock, err := k.lock(context.Background(), statusKey{id: 1})
				if err != nil {
					broken.Store(true)
					return
				}
				if inside.Add(1) > 1 {
					broken.Store(true)
				}
				runtime.Gosched()
				inside.Add(-1)
				unlock()
			}
		})
	}
	wg.Wait()

	assert.False(t, broken.Load(), "two goroutines held the lock of one key at once")
	k.mu.Lock()
	defer k.mu.Unlock()
	assert.Empty(t, k.locks, "an idle key was not forgotten")
}

// TestPostWorkflowStatusBoundsThePostAndRefreshesTimestamps covers what rides
// along with the persisted state: the timestamps some forges derive a duration
// from, and a deadline on the forge call, which runs while holding the lock.
func TestPostWorkflowStatusBoundsThePostAndRefreshesTimestamps(t *testing.T) {
	handed := &model.Workflow{ID: 7, State: model.StatusPending}
	pipeline := &model.Pipeline{ID: 70, Status: model.StatusPending}

	s := store_mocks.NewMockStore(t)
	s.On("WorkflowLoad", int64(7)).Return(&model.Workflow{
		ID: 7, State: model.StatusFailure, Error: "exit code 1", Started: 100, Finished: 250,
	}, nil)
	s.On("GetPipeline", int64(70)).Return(&model.Pipeline{
		ID: 70, Status: model.StatusSuccess, Started: 90, Finished: 260,
	}, nil)

	f := forge_mocks.NewMockForge(t)
	f.On("Status", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ctx, ok := args.Get(0).(context.Context)
			if !assert.True(t, ok, "Status got a %T for its context", args.Get(0)) {
				return
			}
			_, hasDeadline := ctx.Deadline()
			assert.True(t, hasDeadline, "the forge post must be bounded: it runs while holding the lock")

			p, ok := args.Get(3).(*model.Pipeline)
			if !assert.True(t, ok, "Status got a %T for its pipeline", args.Get(3)) {
				return
			}
			w := postedWorkflow(t, args)
			assert.Equal(t, model.StatusFailure, w.State)
			assert.Equal(t, "exit code 1", w.Error)
			assert.Equal(t, int64(100), w.Started)
			assert.Equal(t, int64(250), w.Finished)
			assert.Equal(t, int64(90), p.Started)
			assert.Equal(t, int64(260), p.Finished)
		}).Return(nil)

	// a caller context with no deadline of its own, as on the RPC path
	assert.NoError(t, PostWorkflowStatus(context.Background(), s, f, &model.User{}, &model.Repo{}, pipeline, handed))
	assert.Equal(t, model.StatusPending, handed.State, "the caller's workflow must not be rewritten")
}

// TestPostWorkflowStatusDeadlineStartsAtThePost: a post that queued behind
// another still gets its whole budget. Had the deadline been taken before the
// lock, the wait would have spent it and the post would then fail at once.
func TestPostWorkflowStatusDeadlineStartsAtThePost(t *testing.T) {
	const queued = 2 * time.Second
	unlock, err := statusLocks.lock(t.Context(), statusKey{id: 8})
	if !assert.NoError(t, err) {
		return
	}
	unlockOnce := sync.OnceFunc(unlock)
	t.Cleanup(unlockOnce)

	s := store_mocks.NewMockStore(t)
	s.On("WorkflowLoad", int64(8)).Return(&model.Workflow{ID: 8}, nil)
	s.On("GetPipeline", int64(80)).Return(&model.Pipeline{ID: 80}, nil)
	f := forge_mocks.NewMockForge(t)
	f.On("Status", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ctx, ok := args.Get(0).(context.Context)
			if !assert.True(t, ok, "Status got a %T for its context", args.Get(0)) {
				return
			}
			deadline, ok := ctx.Deadline()
			if assert.True(t, ok, "the forge post must be bounded") {
				assert.Greater(t, time.Until(deadline), statusPostTimeout-queued/2,
					"the time spent queued for the lock was taken out of the post's budget")
			}
		}).Return(nil)

	done := make(chan error, 1)
	go func() {
		done <- PostWorkflowStatus(context.Background(), s, f, &model.User{}, &model.Repo{}, &model.Pipeline{ID: 80}, &model.Workflow{ID: 8})
	}()
	time.Sleep(queued)
	unlockOnce()
	assert.NoError(t, <-done)
}

// TestPostWorkflowStatusGivesUpWaitingWithItsContext: waiting for the lock is
// abandoned when the caller's context ends, and nothing is posted - otherwise a
// hung forge would hold up every queued post for as long as it hangs.
func TestPostWorkflowStatusGivesUpWaitingWithItsContext(t *testing.T) {
	unlock, err := statusLocks.lock(t.Context(), statusKey{id: 9})
	if !assert.NoError(t, err) {
		return
	}
	defer unlock()

	// strict mocks: the post must fail before reading the store or posting
	s := store_mocks.NewMockStore(t)
	f := forge_mocks.NewMockForge(t)

	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	err = within(t, 5*time.Second, func() error {
		return PostWorkflowStatus(ctx, s, f, &model.User{}, &model.Repo{}, &model.Pipeline{ID: 90}, &model.Workflow{ID: 9})
	})
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

// within runs fn and fails the test if it has not returned by limit, so a wait
// that stopped honoring its context fails instead of hanging the package.
func within(t *testing.T, limit time.Duration, fn func() error) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- fn() }()
	select {
	case err := <-done:
		return err
	case <-time.After(limit):
		t.Fatalf("still waiting after %s", limit)
		return nil
	}
}

// TestPostWorkflowStatusBoundsTheWaitWithoutACallerDeadline covers the RPC
// path, where an agent's Init and Done run on a context with no deadline: the
// wait for the lock must still end, or a hung forge would stall the agent.
func TestPostWorkflowStatusBoundsTheWaitWithoutACallerDeadline(t *testing.T) {
	previous := statusLockTimeout
	statusLockTimeout = 100 * time.Millisecond
	t.Cleanup(func() { statusLockTimeout = previous })

	unlock, err := statusLocks.lock(t.Context(), statusKey{id: 10})
	if !assert.NoError(t, err) {
		return
	}
	defer unlock()

	// strict mocks: nothing may be read or posted
	s := store_mocks.NewMockStore(t)
	f := forge_mocks.NewMockForge(t)

	err = within(t, 5*time.Second, func() error {
		return PostWorkflowStatus(context.Background(), s, f, &model.User{}, &model.Repo{}, &model.Pipeline{ID: 100}, &model.Workflow{ID: 10})
	})
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

// TestStatusLocksForgetAbandonedWaits: a wait given up on must drop its hold on
// the key, or every abandoned wait would keep a map entry alive for good.
func TestStatusLocksForgetAbandonedWaits(t *testing.T) {
	var k keyedMutex
	unlock, err := k.lock(t.Context(), statusKey{id: 1})
	if !assert.NoError(t, err) {
		return
	}

	ctx, cancel := context.WithCancelCause(t.Context())
	cancel(nil)
	err = within(t, 5*time.Second, func() error {
		_, err := k.lock(ctx, statusKey{id: 1})
		return err
	})
	assert.ErrorIs(t, err, context.Canceled)

	unlock()
	k.mu.Lock()
	defer k.mu.Unlock()
	assert.Empty(t, k.locks, "an abandoned wait kept its key alive")
}
