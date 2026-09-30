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
	"sync"
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
	close(release)
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
	close(release)
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
	close(release)
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
	unlock := k.lock(statusKey{id: 1})
	unlock2 := k.lock(statusKey{id: 2})
	unlock()
	unlock2()
	k.mu.Lock()
	defer k.mu.Unlock()
	assert.Empty(t, k.locks)
}
