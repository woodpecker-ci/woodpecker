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
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	scheduler_mocks "go.woodpecker-ci.org/woodpecker/v3/server/scheduler/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

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
				s.On("GetActivePipelineList", repo).Return([]*model.Pipeline{older, newer, &otherContext, &otherEvent}, nil).Twice()
				s.On("WorkflowGetTree", mock.Anything).Return([]*model.Workflow{}, nil)
				scheduler.On("CancelWorkflows", mock.Anything, mock.Anything).Return(nil)
				scheduler.On("PublishPipelineEvent", mock.Anything, repo, mock.Anything).Return(nil)
				var canceled []*model.Pipeline
				s.On("UpdatePipeline", mock.Anything).Run(func(args mock.Arguments) {
					canceled = append(canceled, args.Get(0).(*model.Pipeline))
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
