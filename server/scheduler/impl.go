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
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

type impl struct {
	ctx context.Context

	store store.Store
	q     queue.Queue
	ps    pubsub.PubSub

	paused bool
	lock   sync.RWMutex
}

//
// Queue.
//

func (p *impl) Done(c context.Context, id string, exitStatus model.StatusValue) error {
	return p.q.Done(c, id, exitStatus)
}

func (p *impl) Error(c context.Context, id string, err error) error {
	return p.q.Error(c, id, err)
}

func (p *impl) Extend(c context.Context, agentID int64, workflowID string) error {
	return p.q.Extend(c, agentID, workflowID)
}

func (p *impl) Info(c context.Context) queue.InfoT {
	return p.q.Info(c)
}

func (p *impl) KickAgentWorkers(agentID int64) {
	p.q.KickAgentWorkers(agentID)
}

// TODO: markSkipped is a callback helper that is only needed as we use the rpc.Done to mark skipped workflows as done
// this is a hack for another refactor later.
func (p *impl) Poll(c context.Context, agentID int64, agentFilter rpc.Filter, markSkipped func(taskID string) error) (*rpc.Workflow, error) {
	filter := createFilterFunc(agentFilter)

	for {
		// poll blocks until a task is available or the context is canceled / worker is kicked
		task, err := p.q.Poll(c, agentID, filter)
		if err != nil || task == nil {
			return nil, err
		}

		if task.ShouldRun() {
			workflow := new(rpc.Workflow)
			err = json.Unmarshal(task.Data, workflow)
			return workflow, err
		}

		// task should not run, so let the caller mark it as done
		if err := markSkipped(task.ID); err != nil {
			log.Error().Err(err).Msgf("marking workflow task '%s' as done failed", task.ID)
		}
	}
}

func (p *impl) Pause() {
	p.lock.Lock()
	defer p.lock.Unlock()
	if !p.paused {
		p.q.Pause()
	}
}

func (p *impl) Resume() {
	p.lock.Lock()
	defer p.lock.Unlock()
	if !p.paused {
		p.q.Resume()
	}
}

func (p *impl) Wait(c context.Context, id string) error {
	return p.q.Wait(c, id)
}

//
// PubSub.
//

func (p *impl) Subscribe(c context.Context, t pubsub.Topics, r pubsub.Receiver) error {
	return p.ps.Subscribe(c, t, r)
}

//
// Scheduler.
//

// PublishPipelineEvent builds a pipeline state-change event and publishes it
// to the repo topic (and the public topic for public repos).
func (p *impl) PublishPipelineEvent(c context.Context, repo *model.Repo, pipeline *model.Pipeline) error {
	data, err := json.Marshal(model.Event{
		Repo:     *repo,
		Pipeline: *pipeline,
	})
	if err != nil {
		return fmt.Errorf("can't marshal JSON: %w", err)
	}

	message := pubsub.Message{
		ID:   ulid.Make().String(),
		Data: data,
	}

	subTopics := make(pubsub.Topics)
	// if repo is public, push to public topic
	if !repo.IsSCMPrivate {
		subTopics[pubsub.PublicTopic] = struct{}{}
	}
	// publish to repo specific topic
	subTopics[pubsub.GetRepoTopic(repo)] = struct{}{}

	return p.ps.Publish(c, subTopics, message)
}

// StartPipeline announces a new pipeline to UI subscribers and enqueues its
// workflow tasks. The pubsub notification is best-effort and only logged on
// failure, matching the previous behavior where a failed announcement did not
// prevent the pipeline from being queued.
func (p *impl) StartPipeline(c context.Context, repo *model.Repo, pipeline *model.Pipeline, tasks []*model.Task) error {
	if err := p.PublishPipelineEvent(c, repo, pipeline); err != nil {
		log.Error().Err(err).Msg("could not push pipeline status change to pubsub provider")
	}

	return p.q.PushAtOnce(c, tasks)
}

// CancelWorkflows owns the full cancellation of a pipeline's workflows. It
// evicts the running/pending workflows from the queue first (so any waiting
// agents receive the cancellation signal as early as possible), then marks the
// still-pending workflows and steps as skipped, transitions the pipeline to its
// killed state, and publishes the resulting state change. The returned pipeline
// carries its refreshed workflow tree so the caller can sync the forge status.
func (p *impl) CancelWorkflows(c context.Context, repo *model.Repo, pipeline *model.Pipeline, workflows []*model.Workflow, cancelInfo *model.CancelInfo) (*model.Pipeline, error) {
	// First evict the running and pending workflows from the queue, signaling
	// the cancellation (queue.ErrCancel) to any agents currently waiting on
	// them.
	var workflowIDs []string
	for _, w := range workflows {
		if w.State == model.StatusRunning || w.State == model.StatusPending {
			workflowIDs = append(workflowIDs, fmt.Sprint(w.ID))
		}
	}
	if len(workflowIDs) > 0 {
		if err := p.q.ErrorAtOnce(c, workflowIDs, queue.ErrCancel); err != nil {
			log.Error().Err(err).Msgf("cancel workflows: %v", workflowIDs)
		}
	}

	// Mark the still-pending workflows and steps as skipped. Running ones are
	// finalized by their agents once they observe the cancellation signal.
	hasPendingOnly := true
	for _, workflow := range workflows {
		if workflow.State == model.StatusPending {
			workflow.State = model.StatusSkipped
			if err := p.store.WorkflowUpdate(workflow); err != nil {
				log.Error().Err(err).Msgf("cannot update workflow with id %d state", workflow.ID)
			}
		} else {
			hasPendingOnly = false
		}
		for _, step := range workflow.Children {
			if step.State == model.StatusPending {
				step.State = model.StatusCanceled
				if err := p.store.StepUpdate(step); err != nil {
					log.Error().Err(err).Msgf("cannot update step with id %d state", step.ID)
				}
			}
		}
	}

	plState := model.StatusKilled
	if hasPendingOnly {
		plState = model.StatusCanceled
	}
	pipeline.Status = plState
	pipeline.Finished = time.Now().Unix()
	pipeline.CancelInfo = cancelInfo
	if err := p.store.UpdatePipeline(pipeline); err != nil {
		log.Error().Err(err).Msgf("UpdateToStatusKilled: %v", pipeline)
		return nil, err
	}

	var err error
	if pipeline.Workflows, err = p.store.WorkflowGetTree(pipeline); err != nil {
		return nil, err
	}

	if err := p.PublishPipelineEvent(c, repo, pipeline); err != nil {
		log.Error().Err(err).Msg("could not push pipeline status change to pubsub provider")
	}

	return pipeline, nil
}
