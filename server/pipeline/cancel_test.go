// Copyright 2026 Woodpecker Authors
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

package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	scheduler_mocks "go.woodpecker-ci.org/woodpecker/v3/server/scheduler/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

// Evicting a running workflow signals its agent, which then reports the
// workflow done and the server kills every step still pending. Cancel must
// mark pending steps canceled before that signal, otherwise its write from
// the stale workflow tree races with the agent and can leave them killed.
func TestCancelMarksPendingStepsBeforeEvictingWorkflows(t *testing.T) {
	repo := &model.Repo{ID: 1}
	user := &model.User{ID: 1}
	running := &model.Pipeline{ID: 1, Status: model.StatusRunning}
	workflow := &model.Workflow{ID: 2, State: model.StatusRunning, Children: []*model.Step{
		{ID: 3, State: model.StatusRunning, Started: 1},
		{ID: 4, State: model.StatusPending},
	}}

	s := store_mocks.NewMockStore(t)
	s.EXPECT().WorkflowGetTree(mock.Anything).Return([]*model.Workflow{workflow}, nil).Twice()
	pendingCanceled := s.EXPECT().StepUpdate(mock.MatchedBy(func(step *model.Step) bool {
		return step.ID == 4 && step.State == model.StatusCanceled
	})).Return(nil).Once()
	s.EXPECT().UpdatePipeline(mock.MatchedBy(func(p *model.Pipeline) bool {
		return p.Status == model.StatusKilled
	})).Return(nil).Once()

	scheduler := scheduler_mocks.NewMockScheduler(t)
	scheduler.EXPECT().CancelWorkflows(mock.Anything, []string{"2"}).Return(nil).Once().NotBefore(pendingCanceled)
	scheduler.EXPECT().PublishPipelineEvent(mock.Anything, repo, mock.Anything).Return(nil).Once()

	forge := forge_mocks.NewMockForge(t)

	previous := server.Config.Services.Scheduler
	server.Config.Services.Scheduler = scheduler
	t.Cleanup(func() { server.Config.Services.Scheduler = previous })

	assert.NoError(t, Cancel(t.Context(), forge, s, repo, user, running, nil))
}

// Cancel finishes the workflows that did not start and sends their status to
// the forge, even though the pipeline it gets has no workflows loaded. Running
// workflows get their status once their agent reports them done.
func TestCancelSendsStatusOfWorkflowsThatDidNotStart(t *testing.T) {
	tests := []struct {
		name      string
		pipeline  model.StatusValue
		workflows []*model.Workflow
		canceled  []string
		finished  map[int64]model.StatusValue // step ID -> state Cancel sets
		sent      int64                       // workflow ID
		status    model.StatusValue           // of the pipeline after the cancel
	}{
		{
			name:     "pending workflow of a running pipeline",
			pipeline: model.StatusRunning,
			workflows: []*model.Workflow{
				{ID: 2, State: model.StatusRunning, Children: []*model.Step{{ID: 3, State: model.StatusRunning, Started: 1}}},
				{ID: 4, State: model.StatusPending, Children: []*model.Step{{ID: 5, State: model.StatusPending}}},
			},
			canceled: []string{"2", "4"},
			finished: map[int64]model.StatusValue{5: model.StatusCanceled},
			sent:     4,
			status:   model.StatusKilled,
		},
		{
			name:     "blocked pipeline",
			pipeline: model.StatusBlocked,
			workflows: []*model.Workflow{
				{ID: 2, State: model.StatusBlocked, Children: []*model.Step{{ID: 3, State: model.StatusBlocked}}},
			},
			finished: map[int64]model.StatusValue{3: model.StatusSkipped},
			sent:     2,
			status:   model.StatusCanceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &model.Repo{ID: 1}
			user := &model.User{ID: 1}
			pipeline := &model.Pipeline{ID: 1, Status: tt.pipeline}

			s := store_mocks.NewMockStore(t)
			s.EXPECT().WorkflowGetTree(mock.Anything).Return(tt.workflows, nil).Twice()
			for id, state := range tt.finished {
				s.EXPECT().StepUpdate(mock.MatchedBy(func(step *model.Step) bool {
					return step.ID == id && step.State == state
				})).Return(nil).Once()
			}
			s.EXPECT().WorkflowUpdate(mock.MatchedBy(func(w *model.Workflow) bool {
				return w.ID == tt.sent && w.State == model.StatusSkipped
			})).Return(nil).Once()
			s.EXPECT().UpdatePipeline(mock.MatchedBy(func(p *model.Pipeline) bool {
				return p.Status == tt.status
			})).Return(nil).Once()

			scheduler := scheduler_mocks.NewMockScheduler(t)
			scheduler.EXPECT().CancelWorkflows(mock.Anything, tt.canceled).Return(nil).Once()
			scheduler.EXPECT().PublishPipelineEvent(mock.Anything, repo, mock.Anything).Return(nil).Once()

			forge := forge_mocks.NewMockForge(t)
			forge.EXPECT().Status(mock.Anything, user, repo, mock.Anything, mock.MatchedBy(func(w *model.Workflow) bool {
				return w.ID == tt.sent && w.State == model.StatusSkipped
			})).Return(nil).Once()

			previous := server.Config.Services.Scheduler
			server.Config.Services.Scheduler = scheduler
			t.Cleanup(func() { server.Config.Services.Scheduler = previous })

			assert.NoError(t, Cancel(t.Context(), forge, s, repo, user, pipeline, nil))
		})
  }
}

func TestCancelPreviousPipelinesOrdering(t *testing.T) {
	for _, event := range []model.WebhookEvent{model.EventPull, model.EventPush} {
		for _, newestFirst := range []bool{false, true} {
			name := string(event) + "/older handler first"
			if newestFirst {
				name = string(event) + "/newer handler first"
			}
			t.Run(name, func(t *testing.T) {
				repo := &model.Repo{ID: 1, CancelPreviousPipelineEvents: []model.WebhookEvent{event}}
				older := &model.Pipeline{ID: 20, Number: 10, Status: model.StatusPending, Event: event, Branch: "main", Refspec: "feature:main"}
				newer := &model.Pipeline{ID: 21, Number: 11, Status: model.StatusPending, Event: event, Branch: "main", Refspec: "feature:main"}
				otherContext := *older
				otherContext.ID = 18
				otherContext.Number = 8
				otherContext.Branch = "other"
				otherContext.Refspec = "other:main"
				otherEvent := *older
				otherEvent.ID = 19
				otherEvent.Number = 9
				otherEvent.Event = model.EventManual
				s := store_mocks.NewMockStore(t)
				scheduler := scheduler_mocks.NewMockScheduler(t)
				previous := server.Config.Services.Scheduler
				server.Config.Services.Scheduler = scheduler
				t.Cleanup(func() { server.Config.Services.Scheduler = previous })

				// Both handlers can read the same active list before either cancellation is stored.
				s.EXPECT().GetActivePipelineList(repo).Return([]*model.Pipeline{older, newer, &otherContext, &otherEvent}, nil).Twice()
				s.EXPECT().WorkflowGetTree(mock.Anything).Return([]*model.Workflow{}, nil)
				scheduler.EXPECT().CancelWorkflows(mock.Anything, mock.Anything).Return(nil)
				scheduler.EXPECT().PublishPipelineEvent(mock.Anything, repo, mock.Anything).Return(nil)
				var canceled []*model.Pipeline
				s.EXPECT().UpdatePipeline(mock.Anything).Run(func(pipeline *model.Pipeline) {
					canceled = append(canceled, pipeline)
				}).Return(nil)

				order := []*model.Pipeline{older, newer}
				if newestFirst {
					order = []*model.Pipeline{newer, older}
				}
				for _, p := range order {
					require.NoError(t, cancelPreviousPipelines(t.Context(), nil, s, p, repo, nil))
				}
				require.Len(t, canceled, 1)
				assert.Equal(t, older.ID, canceled[0].ID)
				assert.Equal(t, model.StatusCanceled, canceled[0].Status)
				assert.Equal(t, newer.Number, canceled[0].CancelInfo.SupersededBy)
			})
		}
	}
}
