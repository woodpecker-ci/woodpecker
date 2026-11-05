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
