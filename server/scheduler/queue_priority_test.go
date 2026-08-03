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

package scheduler

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
	queue_mocks "go.woodpecker-ci.org/woodpecker/v3/server/queue/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestReprioritizeRepoUpdatesPendingAndWaitingTasks(t *testing.T) {
	repo := &model.Repo{ID: 7, FullName: "postgis/postgis"}
	otherRepoTask := &model.Task{ID: "other", RepoID: 8, PipelineID: 3}
	pendingTask := &model.Task{ID: "pending", RepoID: 7, PipelineID: 1}
	waitingTask := &model.Task{ID: "waiting", RepoID: 7, PipelineID: 2}
	rules := []model.QueuePriorityRule{
		{Priority: 100, Event: model.EventPush, Branch: "master"},
		{Priority: -20, Event: model.EventPull, EventReason: "synchroniz*"},
	}

	store := store_mocks.NewMockStore(t)
	store.EXPECT().GetPipeline(int64(1)).Return(&model.Pipeline{Event: model.EventPush, Branch: "master"}, nil).Once()
	store.EXPECT().GetPipeline(int64(2)).Return(&model.Pipeline{Event: model.EventPull, EventReason: []string{"synchronized"}}, nil).Once()

	queue := queue_mocks.NewMockQueue(t)
	queue.EXPECT().Info(mock.Anything).Return(queueInfo(pendingTask, waitingTask, otherRepoTask)).Once()
	queue.EXPECT().Reprioritize(mock.Anything, map[string]int{
		"pending": 100,
		"waiting": -20,
	}).Return(nil).Once()

	scheduler := &impl{store: store, q: queue}
	require.NoError(t, scheduler.ReprioritizeRepo(t.Context(), repo, rules))
}

func queueInfo(pending, waiting, other *model.Task) queue.InfoT {
	return queue.InfoT{
		Pending:       []*model.Task{pending, other},
		WaitingOnDeps: []*model.Task{waiting},
	}
}
