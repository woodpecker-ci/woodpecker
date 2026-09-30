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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
	"go.woodpecker-ci.org/woodpecker/v3/server/scheduler/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestCancelWorkflow(t *testing.T) {
	for _, state := range []model.StatusValue{model.StatusPending, model.StatusRunning} {
		for _, sibling := range []bool{false, true} {
			t.Run(string(state)+map[bool]string{true: "/siblings", false: "/last"}[sibling], func(t *testing.T) {
				s := store_mocks.NewMockStore(t)
				sched := mocks.NewMockScheduler(t)
				old := server.Config.Services.Scheduler
				server.Config.Services.Scheduler = sched
				t.Cleanup(func() { server.Config.Services.Scheduler = old })
				f := forge_mocks.NewMockForge(t)
				repo := &model.Repo{ID: 1, UserID: 2}
				user := &model.User{ID: 2}
				pl := &model.Pipeline{ID: 3, Status: state}
				wf := &model.Workflow{ID: 9, PipelineID: 3, State: state}
				s.On("WorkflowLoad", int64(9)).Return(wf, nil)
				s.On("GetPipeline", int64(3)).Return(pl, nil)
				sched.On("CancelWorkflows", mock.Anything, []string{"9"}).Return(nil).Once()
				if state == model.StatusPending {
					s.On("WorkflowCancelPending", int64(9), mock.Anything).Return(true, nil).Once()
					tree := []*model.Workflow{{ID: 9, PipelineID: 3, State: model.StatusCanceled, Finished: 10}}
					if sibling {
						tree = append(tree, &model.Workflow{ID: 10, PipelineID: 3, State: model.StatusPending})
					}
					s.On("WorkflowGetTree", mock.Anything).Return(tree, nil)
					if !sibling {
						s.On("UpdatePipeline", mock.MatchedBy(func(p *model.Pipeline) bool {
							return p.Finished > 0 && p.Status != model.StatusSuccess && p.CancelInfo == nil
						})).Return(nil).Once()
					}
					s.On("GetUser", int64(2)).Return(user, nil)
					f.On("Status", mock.Anything, user, repo, mock.Anything, tree[0]).Return(nil).Once()
					sched.On("PublishPipelineEvent", mock.Anything, repo, mock.MatchedBy(func(p *model.Pipeline) bool {
						return len(p.Workflows) == len(tree) && (!sibling || (p.Status == model.StatusPending && p.Finished == 0))
					})).Return(nil).Once()
				}
				require.NoError(t, CancelWorkflow(t.Context(), f, s, repo, pl, 9))
				if state == model.StatusRunning {
					s.AssertNotCalled(t, "WorkflowCancelPending", mock.Anything, mock.Anything)
					s.AssertNotCalled(t, "UpdatePipeline", mock.Anything)
				}
			})
		}
	}
}

func TestCancelWorkflowFailures(t *testing.T) {
	for _, which := range []string{"ownership", "blocked", "finished", "parent", "scheduler", "storage"} {
		t.Run(which, func(t *testing.T) {
			s := store_mocks.NewMockStore(t)
			sched := mocks.NewMockScheduler(t)
			old := server.Config.Services.Scheduler
			server.Config.Services.Scheduler = sched
			t.Cleanup(func() { server.Config.Services.Scheduler = old })
			wf := &model.Workflow{ID: 9, PipelineID: 3, State: model.StatusPending}
			pl := &model.Pipeline{ID: 3, Status: model.StatusRunning}
			switch which {
			case "ownership":
				wf.PipelineID = 4
			case "blocked":
				wf.State = model.StatusBlocked
			case "finished":
				wf.State = model.StatusSuccess
			case "parent":
				pl.Status = model.StatusKilled
			}
			s.On("WorkflowLoad", int64(9)).Return(wf, nil)
			if which != "ownership" {
				s.On("GetPipeline", int64(3)).Return(pl, nil)
			}
			if which == "scheduler" {
				sched.On("CancelWorkflows", mock.Anything, []string{"9"}).Return(errors.New("queue unavailable"))
			}
			if which == "storage" {
				sched.On("CancelWorkflows", mock.Anything, []string{"9"}).Return(nil)
				s.On("WorkflowCancelPending", int64(9), mock.Anything).Return(false, errors.New("database unavailable"))
			}
			assert.Error(t, CancelWorkflow(t.Context(), nil, s, &model.Repo{ID: 1}, pl, 9))
			s.AssertNotCalled(t, "UpdatePipeline", mock.Anything)
		})
	}
}

func TestCancelWorkflowBeforeQueueInsertion(t *testing.T) {
	s := store_mocks.NewMockStore(t)
	sched := mocks.NewMockScheduler(t)
	old := server.Config.Services.Scheduler
	server.Config.Services.Scheduler = sched
	t.Cleanup(func() { server.Config.Services.Scheduler = old })
	f := forge_mocks.NewMockForge(t)
	repo := &model.Repo{ID: 1, UserID: 2}
	user := &model.User{ID: 2}
	pl := &model.Pipeline{ID: 3, Status: model.StatusPending}
	wf := &model.Workflow{ID: 9, PipelineID: 3, State: model.StatusPending}
	canceled := &model.Workflow{ID: 9, PipelineID: 3, State: model.StatusCanceled, Finished: 10}
	tree := []*model.Workflow{canceled, {ID: 10, PipelineID: 3, State: model.StatusPending}}

	s.On("WorkflowLoad", int64(9)).Return(wf, nil).Twice()
	s.On("GetPipeline", int64(3)).Return(pl, nil).Twice()
	sched.On("CancelWorkflows", mock.Anything, []string{"9"}).Return(queue.ErrNotFound).Once()
	s.On("WorkflowCancelPending", int64(9), mock.Anything).Return(true, nil).Once()
	s.On("WorkflowGetTree", mock.Anything).Return(tree, nil).Once()
	s.On("GetUser", int64(2)).Return(user, nil).Once()
	f.On("Status", mock.Anything, user, repo, mock.Anything, canceled).Return(nil).Once()
	sched.On("PublishPipelineEvent", mock.Anything, repo, mock.Anything).Return(nil).Once()

	require.NoError(t, CancelWorkflow(t.Context(), f, s, repo, pl, 9))
}

func TestPipelineStatusKeepsActiveWork(t *testing.T) {
	for _, active := range []model.StatusValue{model.StatusPending, model.StatusRunning} {
		assert.Equal(t, active, PipelineStatus([]*model.Workflow{{State: model.StatusKilled}, {State: active}}))
	}
	assert.NotEqual(t, model.StatusSuccess, PipelineStatus([]*model.Workflow{{State: model.StatusCanceled}, {State: model.StatusSuccess}}))
}

func TestConcurrentWorkflowCancellation(t *testing.T) {
	s := store_mocks.NewMockStore(t)
	sched := mocks.NewMockScheduler(t)
	old := server.Config.Services.Scheduler
	server.Config.Services.Scheduler = sched
	t.Cleanup(func() { server.Config.Services.Scheduler = old })
	f := forge_mocks.NewMockForge(t)
	repo := &model.Repo{ID: 1, UserID: 2}
	user := &model.User{ID: 2}
	pl := &model.Pipeline{ID: 3, Status: model.StatusRunning}
	w := &model.Workflow{ID: 9, PipelineID: 3, State: model.StatusPending}
	// All accesses happen under the same pipeline lifecycle lock.
	s.On("WorkflowLoad", int64(9)).Return(func(int64) *model.Workflow { copy := *w; return &copy }, nil)
	s.On("GetPipeline", int64(3)).Return(pl, nil)
	s.On("WorkflowCancelPending", int64(9), mock.Anything).Run(func(mock.Arguments) { w.State = model.StatusCanceled }).Return(true, nil).Once()
	s.On("WorkflowGetTree", mock.Anything).Return([]*model.Workflow{w, {ID: 10, State: model.StatusRunning}}, nil)
	s.On("GetUser", int64(2)).Return(user, nil)
	f.On("Status", mock.Anything, user, repo, mock.Anything, mock.Anything).Return(nil).Once()
	sched.On("CancelWorkflows", mock.Anything, []string{"9"}).Return(nil).Once()
	sched.On("PublishPipelineEvent", mock.Anything, repo, mock.Anything).Return(nil).Once()
	start := make(chan struct{})
	results := make(chan error, 2)
	for range 2 {
		go func() { <-start; results <- CancelWorkflow(t.Context(), f, s, repo, pl, 9) }()
	}
	close(start)
	first, second := <-results, <-results
	if first == nil {
		require.ErrorAs(t, second, new(*ErrBadRequest))
	} else {
		require.ErrorAs(t, first, new(*ErrBadRequest))
		require.NoError(t, second)
	}
}
