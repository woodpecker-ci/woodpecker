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

// Scheduler combines the Queue & PubSub providers and adds higher-level
// operations that consolidate the queue and pubsub calls belonging to a
// single logical action behind one method.
type Scheduler interface {
	queue.Queue
	pubsub.PubSub

	// PublishPipelineEvent builds a pipeline state-change event and publishes
	// it to all UI subscribers of the repo (and the public topic if public).
	PublishPipelineEvent(c context.Context, repo *model.Repo, pipeline *model.Pipeline) error

	// StartPipeline announces a new pipeline to UI subscribers and enqueues
	// its workflow tasks. Publishing is best-effort; enqueuing is critical and
	// its error is returned.
	StartPipeline(c context.Context, repo *model.Repo, pipeline *model.Pipeline, tasks []*model.Task) error
}

func NewScheduler(q queue.Queue, ps pubsub.PubSub) Scheduler {
	return &proxy{
		q:  q,
		ps: ps,
	}
}
