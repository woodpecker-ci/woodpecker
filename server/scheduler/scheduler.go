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
	"context"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
)

// FilterFn filters tasks when polling the queue. If it returns false the task
// is skipped; the int is a match score (higher is better).
type FilterFn func(*model.Task) (bool, int)

// Scheduler coordinates the queue and pubsub providers behind a single
// surface. The low-level enqueue (queue.PushAtOnce) and publish
// (pubsub.Publish) calls are intentionally not exposed: callers use the
// consolidated StartPipeline and PublishPipelineEvent methods instead, which
// pair the related queue and pubsub calls of a single logical action.
type Scheduler interface {
	// Queue operations.
	Poll(c context.Context, agentID int64, f FilterFn) (*model.Task, error)
	Extend(c context.Context, agentID int64, workflowID string) error
	Done(c context.Context, id string, exitStatus model.StatusValue) error
	Error(c context.Context, id string, err error) error
	ErrorAtOnce(c context.Context, ids []string, err error) error
	Wait(c context.Context, id string) error
	Info(c context.Context) queue.InfoT
	Pause()
	Resume()
	KickAgentWorkers(agentID int64)

	// PubSub operations.
	Subscribe(c context.Context, t pubsub.Topics, r pubsub.Receiver) error

	// Consolidated operations.
	PublishPipelineEvent(c context.Context, repo *model.Repo, pipeline *model.Pipeline) error
	StartPipeline(c context.Context, repo *model.Repo, pipeline *model.Pipeline, tasks []*model.Task) error
}

func NewScheduler(q queue.Queue, ps pubsub.PubSub) Scheduler {
	return &proxy{
		q:  q,
		ps: ps,
	}
}
