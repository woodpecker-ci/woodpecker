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

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
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
	//
	// Poll blocks until the next runnable workflow for the given agent is
	// available. It applies the agent's label filter, transparently skips
	// tasks whose dependencies preclude running (invoking markSkipped so the
	// caller can finalize them), and returns the runnable workflow.
	// TODO: markSkipped is a callback helper that is only needed as we use the rpc.Done to mark skipped workflows as done
	// this is a hack for another refactor later.
	Poll(c context.Context, agentID int64, agentFilter rpc.Filter, markSkipped func(taskID string) error) (*rpc.Workflow, error)
	Extend(c context.Context, agentID int64, workflowID string) error
	Done(c context.Context, id string, exitStatus model.StatusValue) error
	Error(c context.Context, id string, err error) error
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

	// CancelWorkflows cancels the given workflows: it evicts them from the
	// queue and signals the cancellation to any agents waiting on them. This is
	// the entry point for the scheduler to later own the full cancel cleanup.
	CancelWorkflows(c context.Context, workflowIDs []string) error
}

func NewScheduler(ctx context.Context, store store.Store, q queue.Queue, ps pubsub.PubSub) Scheduler {
	return &impl{
		ctx:   ctx,
		store: store,
		q:     q,
		ps:    ps,
	}
}
