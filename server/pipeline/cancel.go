// Copyright 2022 Woodpecker Authors
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
	"context"
	"fmt"
	"slices"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// Cancel the pipeline and returns the status.
func Cancel(ctx context.Context, _forge forge.Forge, store store.Store, repo *model.Repo, user *model.User, pipeline *model.Pipeline, cancelInfo *model.CancelInfo) error {
	if pipeline.Status != model.StatusRunning && pipeline.Status != model.StatusPending && pipeline.Status != model.StatusBlocked {
		return &ErrBadRequest{Msg: "Cannot cancel a non-running or non-pending or non-blocked pipeline"}
	}

	workflows, err := store.WorkflowGetTree(pipeline)
	if err != nil {
		return &ErrNotFound{Msg: err.Error()}
	}

	// First cancel/evict workflows in the queue in one go
	var workflowsToCancel []string
	for _, w := range workflows {
		if w.State == model.StatusRunning || w.State == model.StatusPending {
			workflowsToCancel = append(workflowsToCancel, fmt.Sprint(w.ID))
		}
	}

	if len(workflowsToCancel) != 0 {
		if err := server.Config.Services.Queue.ErrorAtOnce(ctx, workflowsToCancel, queue.ErrCancel); err != nil {
			log.Error().Err(err).Msgf("queue: evict_at_once: %v", workflowsToCancel)
		}
	}

	// Then update the DB status for pending pipelines
	// Running ones will be set when the agents stop on the cancel signal
	for _, workflow := range workflows {
		if workflow.State == model.StatusPending {
			if _, err = UpdateWorkflowToStatusSkipped(store, *workflow); err != nil {
				log.Error().Err(err).Msgf("cannot update workflow with id %d state", workflow.ID)
			}
		}
		for _, step := range workflow.Children {
			if step.State == model.StatusPending {
				if _, err = UpdateStepToStatusSkipped(store, *step, 0); err != nil {
					log.Error().Err(err).Msgf("cannot update workflow with id %d state", workflow.ID)
				}
			}
		}
	}

	killedPipeline, err := UpdateToStatusKilled(store, *pipeline, cancelInfo)
	if err != nil {
		log.Error().Err(err).Msgf("UpdateToStatusKilled: %v", pipeline)
		return err
	}

	updatePipelineStatus(ctx, _forge, killedPipeline, repo, user)

	if killedPipeline.Workflows, err = store.WorkflowGetTree(killedPipeline); err != nil {
		return err
	}
	publishToTopic(killedPipeline, repo)

	return nil
}

func cancelPreviousPipelines(
	ctx context.Context,
	_forge forge.Forge,
	_store store.Store,
	pipeline *model.Pipeline,
	repo *model.Repo,
	user *model.User,
) error {
	// check this event should cancel previous pipelines
	eventIncluded := slices.Contains(repo.CancelPreviousPipelineEvents, pipeline.Event)
	if !eventIncluded {
		return nil
	}

	// get all active activeBuilds
	activeBuilds, err := _store.GetActivePipelineList(repo)
	if err != nil {
		return err
	}

	pipelineNeedsCancel := func(active *model.Pipeline) bool {
		// always filter on same event
		if active.Event != pipeline.Event {
			return false
		}

		// find events for the same context
		switch pipeline.Event {
		case model.EventPush:
			return pipeline.Branch == active.Branch
		default:
			return pipeline.Refspec == active.Refspec
		}
	}

	for _, active := range activeBuilds {
		if active.ID == pipeline.ID {
			// same pipeline. e.g. self
			continue
		}

		cancel := pipelineNeedsCancel(active)

		if !cancel {
			continue
		}

		if err = Cancel(ctx, _forge, _store, repo, user, active, &model.CancelInfo{
			Reason: model.CancelReasonSuperseded,
			Data: map[string]string{
				"pipeline_number": fmt.Sprint(pipeline.Number),
			},
		}); err != nil {
			log.Error().
				Err(err).
				Str("ref", active.Ref).
				Int64("id", active.ID).
				Msg("failed to cancel pipeline")
		}
	}

	return nil
}
