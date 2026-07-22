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

// SkippedWorkflowFunc is invoked after the scheduler has finalized a workflow
// as skipped. It lets the caller run the follow-up that does not belong to the
// scheduler, namely syncing the workflow's status to the forge. The scheduler
// has already persisted the skipped state and notified subscribers.
type SkippedWorkflowFunc func(repo *model.Repo, pipeline *model.Pipeline, workflow *model.Workflow)

// Scheduler coordinates the queue and pubsub providers behind a single
// surface. The low-level enqueue (queue.PushAtOnce) and publish
// (pubsub.Publish) calls are intentionally not exposed: callers use the
// consolidated StartPipeline and PublishPipelineEvent methods instead, which
// pair the related queue and pubsub calls of a single logical action.
type Scheduler interface {
	// Queue operations.
	//
	// Poll blocks until the next runnable workflow for the given agent is
	// available. It applies the agent's label filter and transparently
	// finalizes any task whose dependencies preclude it from running as
	// skipped, invoking onSkipped afterwards so the caller can perform the
	// non-scheduling follow-up (e.g. forge status) that the scheduler cannot
	// do itself. It then returns the next runnable workflow.
	Poll(c context.Context, agentID int64, agentFilter rpc.Filter, onSkipped SkippedWorkflowFunc) (*rpc.Workflow, error)
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

	// FinishWorkflow owns the completion of a single workflow: it finalizes
	// the still-running children, persists the workflow's final state,
	// acknowledges it on the queue, rolls the pipeline up to its done state
	// once nothing is left running and publishes the change. It returns the
	// updated pipeline (with its refreshed tree) and workflow so the caller
	// can sync the forge status, close log streams and record metrics.
	FinishWorkflow(c context.Context, repo *model.Repo, pipeline *model.Pipeline, workflow *model.Workflow, state rpc.WorkflowState) (*model.Pipeline, *model.Workflow, error)

	// CancelWorkflows owns the full cancellation of a pipeline's workflows:
	// it evicts the running/pending workflows from the queue (signaling the
	// cancellation to any agents waiting on them), marks the still-pending
	// workflows and steps as skipped, transitions the pipeline to its killed
	// state, and publishes the resulting state change to subscribers. It
	// returns the updated (killed) pipeline so the caller can sync the forge
	// status, which is the only cancellation concern left to the caller.
	CancelWorkflows(c context.Context, repo *model.Repo, pipeline *model.Pipeline, workflows []*model.Workflow, cancelInfo *model.CancelInfo) (*model.Pipeline, error)
}

func NewScheduler(ctx context.Context, store store.Store, q queue.Queue, ps pubsub.PubSub) Scheduler {
	return &impl{
		ctx:   ctx,
		store: store,
		q:     q,
		ps:    ps,
	}
}
