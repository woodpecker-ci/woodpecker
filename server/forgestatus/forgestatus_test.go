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

package forgestatus

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_types "go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// fastPolicy is the default in milliseconds.
var fastPolicy = Policy{
	Attempt:   time.Second,
	FirstWait: 5 * time.Millisecond,
	MaxWait:   20 * time.Millisecond,
	Window:    time.Second,
	Senders:   2,
}

// fakeStore holds one repository, its user and its pipelines.
type fakeStore struct {
	mu        sync.Mutex
	pipelines map[int64]*model.Pipeline
	workflows map[int64][]*model.Workflow // pipeline ID -> workflows
	fail      int                         // fail this many next reads of a pipeline
	noRepo    bool                        // the repository is not in the store
}

func newFakeStore() *fakeStore {
	return &fakeStore{pipelines: make(map[int64]*model.Pipeline), workflows: make(map[int64][]*model.Workflow)}
}

func (s *fakeStore) GetPipeline(id int64) (*model.Pipeline, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fail > 0 {
		s.fail--
		return nil, errors.New("database unavailable")
	}
	p, ok := s.pipelines[id]
	if !ok {
		return nil, store_types.ErrRecordNotExist
	}
	c := *p
	return &c, nil
}

func (s *fakeStore) GetRepo(id int64) (*model.Repo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.noRepo {
		return nil, store_types.ErrRecordNotExist
	}
	return &model.Repo{ID: id, UserID: 1, FullName: "octocat/hello"}, nil
}

func (*fakeStore) GetUser(id int64) (*model.User, error) { return &model.User{ID: id}, nil }

func (s *fakeStore) WorkflowList(p *model.Pipeline) ([]*model.Workflow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*model.Workflow
	for _, w := range s.workflows[p.ID] {
		c := *w
		out = append(out, &c)
	}
	return out, nil
}

// set puts the pipeline in the store with the workflows in the states given.
func (s *fakeStore) set(pipelineID int64, status model.StatusValue, states ...model.StatusValue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pipelines[pipelineID] = &model.Pipeline{ID: pipelineID, RepoID: 1, Number: pipelineID, Status: status}
	s.workflows[pipelineID] = nil
	for i, state := range states {
		s.workflows[pipelineID] = append(s.workflows[pipelineID], &model.Workflow{
			ID: pipelineID*100 + int64(i), PipelineID: pipelineID, Name: "w", State: state,
		})
	}
}

// fakeForge records the statuses it took, per workflow, and answers with
// the errors queued.
type fakeForge struct {
	mu      sync.Mutex
	taken   map[int64][]model.StatusValue
	calls   int
	errs    []error
	block   chan struct{} // when set, every request waits for it to close
	reqTime []time.Time
}

func (f *fakeForge) status(ctx context.Context, _ *model.User, _ *model.Repo, _ *model.Pipeline, w *model.Workflow) error {
	f.mu.Lock()
	block := f.block
	f.mu.Unlock()
	if block != nil {
		select {
		case <-block:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.reqTime = append(f.reqTime, time.Now())
	if len(f.errs) > 0 {
		err := f.errs[0]
		f.errs = f.errs[1:]
		if err != nil {
			return err
		}
	}
	if f.taken == nil {
		f.taken = make(map[int64][]model.StatusValue)
	}
	f.taken[w.ID] = append(f.taken[w.ID], w.State)
	return nil
}

func (f *fakeForge) Taken(workflowID int64) []model.StatusValue {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]model.StatusValue(nil), f.taken[workflowID]...)
}

func (f *fakeForge) Calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func newTestSender(t *testing.T, policy Policy, errs ...error) (*Sender, *fakeStore, *fakeForge) {
	t.Helper()
	st := newFakeStore()
	ff := &fakeForge{errs: errs}
	m := forge_mocks.NewMockForge(t)
	m.On("Status", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(ff.status).Maybe()
	ctx, cancel := context.WithCancelCause(t.Context())
	s := newSender(ctx, st, func(*model.Repo) (forge.Forge, error) { return m, nil },
		func(context.Context, forge.Forge, *model.User) {}, policy)
	t.Cleanup(func() {
		cancel(nil)
		s.wg.Wait()
	})
	return s, st, ff
}

// idle waits for every sender to finish.
func idle(t *testing.T, s *Sender) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("senders still running")
	}
}

var errDown = errors.New("502 bad gateway")

func TestPipelineSendsEveryWorkflow(t *testing.T) {
	s, st, ff := newTestSender(t, fastPolicy)
	st.set(1, model.StatusPending, model.StatusPending, model.StatusPending)
	s.Pipeline(1)
	idle(t, s)
	assert.Equal(t, []model.StatusValue{model.StatusPending}, ff.Taken(100))
	assert.Equal(t, []model.StatusValue{model.StatusPending}, ff.Taken(101))
}

func TestWorkflowSendsOnlyThatWorkflow(t *testing.T) {
	s, st, ff := newTestSender(t, fastPolicy)
	st.set(1, model.StatusRunning, model.StatusRunning, model.StatusPending)
	s.Workflow(1, 100)
	idle(t, s)
	assert.Equal(t, []model.StatusValue{model.StatusRunning}, ff.Taken(100))
	assert.Empty(t, ff.Taken(101))
}

// TestSameStatusNotSentTwice: the server marks a workflow several times
// while its state stays the same.
func TestSameStatusNotSentTwice(t *testing.T) {
	s, st, ff := newTestSender(t, fastPolicy)
	st.set(1, model.StatusPending, model.StatusPending)
	for range 3 {
		s.Workflow(1, 100)
		idle(t, s)
	}
	s.Pipeline(1)
	idle(t, s)
	assert.Equal(t, 1, ff.Calls())

	st.set(1, model.StatusRunning, model.StatusRunning)
	s.Workflow(1, 100)
	idle(t, s)
	assert.Equal(t, []model.StatusValue{model.StatusPending, model.StatusRunning}, ff.Taken(100))
}

// TestRetriedUntilTaken: a status that fails is sent again.
func TestRetriedUntilTaken(t *testing.T) {
	s, st, ff := newTestSender(t, fastPolicy, errDown, errDown, errDown)
	st.set(1, model.StatusSuccess, model.StatusSuccess)
	s.Workflow(1, 100)
	idle(t, s)
	assert.Equal(t, []model.StatusValue{model.StatusSuccess}, ff.Taken(100))
	assert.Equal(t, 4, ff.Calls())
}

// TestNewestSentAfterOutage: changes made while the forge fails collapse
// into the newest, and that is what the forge ends with.
func TestNewestSentAfterOutage(t *testing.T) {
	policy := fastPolicy
	policy.FirstWait, policy.MaxWait = 50*time.Millisecond, 50*time.Millisecond
	s, st, ff := newTestSender(t, policy, errDown)
	st.set(1, model.StatusRunning, model.StatusRunning)
	s.Workflow(1, 100) // fails, waits
	require.Eventually(t, func() bool { return ff.Calls() == 1 }, time.Second, time.Millisecond)

	st.set(1, model.StatusSuccess, model.StatusSuccess)
	s.Workflow(1, 100)
	idle(t, s)
	assert.Equal(t, []model.StatusValue{model.StatusSuccess}, ff.Taken(100), "running is stale by then")
}

// TestNewestLastWhileSending: a change during a slow request is sent after
// it, never before.
func TestNewestLastWhileSending(t *testing.T) {
	s, st, ff := newTestSender(t, fastPolicy)
	ff.block = make(chan struct{})
	st.set(1, model.StatusRunning, model.StatusRunning)
	s.Workflow(1, 100)
	time.Sleep(20 * time.Millisecond) // the sender is in the request

	st.set(1, model.StatusSuccess, model.StatusSuccess)
	s.Workflow(1, 100) // does not wait for the forge
	ff.mu.Lock()
	close(ff.block)
	ff.block = nil
	ff.mu.Unlock()
	idle(t, s)
	assert.Equal(t, []model.StatusValue{model.StatusRunning, model.StatusSuccess}, ff.Taken(100))
}

// TestFinalErrorNotRetried: a status the forge refused is not sent again.
func TestFinalErrorNotRetried(t *testing.T) {
	s, st, ff := newTestSender(t, fastPolicy, &types.StatusError{Err: errors.New("422"), Final: true})
	st.set(1, model.StatusSuccess, model.StatusSuccess)
	s.Workflow(1, 100)
	idle(t, s)
	assert.Equal(t, 1, ff.Calls())
	assert.Empty(t, ff.Taken(100))

	// nor remembered as taken: marking it again sends it
	s.Workflow(1, 100)
	idle(t, s)
	assert.Equal(t, []model.StatusValue{model.StatusSuccess}, ff.Taken(100))
}

func TestNotImplementedNotRetried(t *testing.T) {
	s, st, ff := newTestSender(t, fastPolicy, types.ErrNotImplemented)
	st.set(1, model.StatusSuccess, model.StatusSuccess)
	s.Workflow(1, 100)
	idle(t, s)
	assert.Equal(t, 1, ff.Calls())
}

// TestRetryAfterHonoured: the next attempt waits as long as the forge asked.
func TestRetryAfterHonoured(t *testing.T) {
	const after = 150 * time.Millisecond
	s, st, ff := newTestSender(t, fastPolicy, &types.StatusError{Err: errors.New("rate limited"), RetryAfter: after})
	st.set(1, model.StatusSuccess, model.StatusSuccess)
	s.Workflow(1, 100)
	idle(t, s)
	require.Len(t, ff.reqTime, 2)
	assert.GreaterOrEqual(t, ff.reqTime[1].Sub(ff.reqTime[0]), after)
}

// TestGivesUpAfterWindow: a forge that keeps failing is not tried forever.
func TestGivesUpAfterWindow(t *testing.T) {
	policy := fastPolicy
	policy.Window = 100 * time.Millisecond
	errs := make([]error, 1000)
	for i := range errs {
		errs[i] = errDown
	}
	s, st, ff := newTestSender(t, policy, errs...)
	st.set(1, model.StatusSuccess, model.StatusSuccess)
	start := time.Now()
	s.Workflow(1, 100)
	idle(t, s)
	assert.Less(t, time.Since(start), 2*time.Second)
	assert.Greater(t, ff.Calls(), 1)
	assert.Empty(t, ff.Taken(100))
}

// TestStoreFailureRetried: a pipeline that could not be read is read again.
func TestStoreFailureRetried(t *testing.T) {
	s, st, ff := newTestSender(t, fastPolicy)
	st.set(1, model.StatusSuccess, model.StatusSuccess)
	st.fail = 2
	s.Pipeline(1)
	idle(t, s)
	assert.Equal(t, []model.StatusValue{model.StatusSuccess}, ff.Taken(100))
}

// TestStoreFailureGivesUp: a store that keeps failing is not tried forever.
func TestStoreFailureGivesUp(t *testing.T) {
	policy := fastPolicy
	policy.Window = 100 * time.Millisecond
	s, st, ff := newTestSender(t, policy)
	st.set(1, model.StatusSuccess, model.StatusSuccess)
	st.fail = 1 << 30
	s.Pipeline(1)
	idle(t, s)
	assert.Zero(t, ff.Calls())
}

func TestDeletedPipelineDropped(t *testing.T) {
	s, _, ff := newTestSender(t, fastPolicy)
	s.Pipeline(42)
	idle(t, s)
	assert.Zero(t, ff.Calls())
}

// TestStopsWithTheServer: senders end when the server stops, even with a
// status failing.
func TestStopsWithTheServer(t *testing.T) {
	st := newFakeStore()
	ff := &fakeForge{block: make(chan struct{})}
	m := forge_mocks.NewMockForge(t)
	m.On("Status", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(ff.status).Maybe()
	ctx, cancel := context.WithCancelCause(t.Context())
	s := newSender(ctx, st, func(*model.Repo) (forge.Forge, error) { return m, nil },
		func(context.Context, forge.Forge, *model.User) {}, fastPolicy)
	st.set(1, model.StatusRunning, model.StatusRunning)
	st.set(2, model.StatusRunning, model.StatusRunning)
	st.set(3, model.StatusRunning, model.StatusRunning)
	s.Pipeline(1)
	s.Pipeline(2)
	s.Pipeline(3) // waits for a slot: Senders is 2
	time.Sleep(20 * time.Millisecond)
	cancel(nil)
	idle(t, s)
	assert.Zero(t, ff.Calls())
}

// TestSendersBounded: no more pipelines are synced at once than Senders.
func TestSendersBounded(t *testing.T) {
	s, st, ff := newTestSender(t, fastPolicy)
	ff.block = make(chan struct{})
	for id := int64(1); id <= 5; id++ {
		st.set(id, model.StatusRunning, model.StatusRunning)
		s.Pipeline(id)
	}
	time.Sleep(20 * time.Millisecond)
	assert.Len(t, s.slots, fastPolicy.Senders)
	ff.mu.Lock()
	close(ff.block)
	ff.block = nil
	ff.mu.Unlock()
	idle(t, s)
	assert.Equal(t, 5, ff.Calls())
}

// TestPanicFreesThePipeline: a sender that panics does not leave its
// pipeline without one.
func TestPanicFreesThePipeline(t *testing.T) {
	st := newFakeStore()
	ff := &fakeForge{}
	m := forge_mocks.NewMockForge(t)
	first := true
	m.On("Status", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return( //nolint:forbidigo // tests a panic
		func(ctx context.Context, u *model.User, r *model.Repo, p *model.Pipeline, w *model.Workflow) error {
			if first {
				first = false
				panic("boom") //nolint:forbidigo // tests a panic
			}
			return ff.status(ctx, u, r, p, w)
		},
	).Maybe()
	s := newSender(t.Context(), st, func(*model.Repo) (forge.Forge, error) { return m, nil },
		func(context.Context, forge.Forge, *model.User) {}, fastPolicy)
	st.set(1, model.StatusRunning, model.StatusRunning)
	s.Pipeline(1)
	idle(t, s)
	s.Pipeline(1)
	idle(t, s)
	assert.Equal(t, []model.StatusValue{model.StatusRunning}, ff.Taken(100))
	assert.Empty(t, s.slots, "the slot was given back")
}

// TestMarkDuringPanicSent: a mark that came while the sender panicked is
// sent without another mark.
func TestMarkDuringPanicSent(t *testing.T) {
	st := newFakeStore()
	ff := &fakeForge{}
	m := forge_mocks.NewMockForge(t)
	inside, release := make(chan struct{}), make(chan struct{})
	first := true
	m.On("Status", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return( //nolint:forbidigo // tests a panic
		func(ctx context.Context, u *model.User, r *model.Repo, p *model.Pipeline, w *model.Workflow) error {
			if first {
				first = false
				close(inside)
				<-release
				panic("boom") //nolint:forbidigo // tests a panic
			}
			return ff.status(ctx, u, r, p, w)
		},
	).Maybe()
	s := newSender(t.Context(), st, func(*model.Repo) (forge.Forge, error) { return m, nil },
		func(context.Context, forge.Forge, *model.User) {}, fastPolicy)
	st.set(1, model.StatusRunning, model.StatusRunning)
	s.Workflow(1, 100)
	<-inside
	st.set(1, model.StatusSuccess, model.StatusSuccess)
	s.Workflow(1, 100) // the agent's done, while the sender is in the request
	close(release)
	idle(t, s)
	assert.Equal(t, []model.StatusValue{model.StatusSuccess}, ff.Taken(100))
}

// TestNewStatusGetsItsOwnWindow: a status that replaces one that failed for
// most of the window is still tried for a whole window.
func TestNewStatusGetsItsOwnWindow(t *testing.T) {
	policy := fastPolicy
	policy.Window = 200 * time.Millisecond
	errs := make([]error, 0, 1000)
	for range cap(errs) {
		errs = append(errs, errDown)
	}
	s, st, ff := newTestSender(t, policy, errs...)
	st.set(1, model.StatusRunning, model.StatusRunning)
	s.Workflow(1, 100)
	time.Sleep(150 * time.Millisecond) // running has failed for most of its window

	st.set(1, model.StatusSuccess, model.StatusSuccess)
	s.Workflow(1, 100)
	time.Sleep(100 * time.Millisecond) // past the window of running
	ff.mu.Lock()
	ff.errs = nil // the forge is back
	ff.mu.Unlock()
	idle(t, s)
	assert.Equal(t, []model.StatusValue{model.StatusSuccess}, ff.Taken(100))
}

// TestRetryAfterCapped: a forge asking to wait longer than the window does
// not hold the pipeline's statuses that long.
func TestRetryAfterCapped(t *testing.T) {
	policy := fastPolicy
	policy.Window = 100 * time.Millisecond
	s, st, ff := newTestSender(t, policy, &types.StatusError{Err: errors.New("rate limited"), RetryAfter: time.Hour})
	st.set(1, model.StatusSuccess, model.StatusSuccess)
	start := time.Now()
	s.Workflow(1, 100)
	idle(t, s)
	assert.Less(t, time.Since(start), 2*time.Second)
	assert.Equal(t, 2, ff.Calls())
}

// TestZeroPolicy: fields left zero take the defaults.
func TestZeroPolicy(t *testing.T) {
	s, st, ff := newTestSender(t, Policy{})
	assert.Equal(t, DefaultPolicy, s.policy)
	st.set(1, model.StatusSuccess, model.StatusSuccess)
	s.Workflow(1, 100)
	idle(t, s)
	assert.Equal(t, []model.StatusValue{model.StatusSuccess}, ff.Taken(100))
}

// TestRepoGone: statuses of a pipeline whose repository is not in the store
// are dropped, not tried for the window.
func TestRepoGone(t *testing.T) {
	s, st, ff := newTestSender(t, fastPolicy)
	st.set(1, model.StatusSuccess, model.StatusSuccess)
	st.noRepo = true
	s.Pipeline(1)
	idle(t, s)
	assert.Zero(t, ff.Calls())
}

func TestNilSender(t *testing.T) {
	var s *Sender
	s.Pipeline(1)
	s.Workflow(1, 2)
}
