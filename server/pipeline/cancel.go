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
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

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

	// First cancel/evict the running and pending workflows from the queue
	var workflowsToCancel []string
	for _, w := range workflows {
		if w.State == model.StatusRunning || w.State == model.StatusPending {
			workflowsToCancel = append(workflowsToCancel, fmt.Sprint(w.ID))
		}
	}

	// Workflows the queue has no task for are not being executed by anyone: the
	// agent that held them is gone (its container was replaced, it was OOM-killed,
	// the host rebooted). Nothing will ever report them finished, so they are
	// collected here and closed below instead of being left to an agent that no
	// longer exists.
	orphaned := map[int64]bool{}
	if err := server.Config.Services.Scheduler.CancelWorkflows(ctx, workflowsToCancel); err != nil {
		log.Error().Err(err).Msgf("cancel workflows: %v", workflowsToCancel)

		var notFound *queue.ErrTasksNotFound
		if errors.As(err, &notFound) {
			for _, id := range notFound.IDs {
				workflowID, convErr := strconv.ParseInt(id, 10, 64)
				if convErr != nil {
					log.Error().Err(convErr).Msgf("cannot parse workflow id %q", id)
					continue
				}
				orphaned[workflowID] = true
			}
		}
	}

	hasPendingOnly := true
	now := time.Now().Unix()

	// Then update the DB status for pending pipelines
	// Running ones will be set when the agents stop on the cancel signal, unless
	// no agent holds them any more, in which case that signal is never coming.
	for _, workflow := range workflows {
		switch {
		case workflow.State == model.StatusPending:
			if _, err = UpdateWorkflowToStatusSkipped(store, *workflow); err != nil {
				log.Error().Err(err).Msgf("cannot update workflow with id %d state", workflow.ID)
			}
		case orphaned[workflow.ID]:
			hasPendingOnly = false
			if _, err = UpdateWorkflowToStatusKilled(store, *workflow, now); err != nil {
				log.Error().Err(err).Msgf("cannot update orphaned workflow with id %d state", workflow.ID)
			}
			// its steps are closed by the helper, so skip the pending-step pass
			continue
		default:
			hasPendingOnly = false
		}
		for _, step := range workflow.Children {
			if step.State == model.StatusPending {
				if _, err = UpdateStepToStatusSkipped(store, *step, 0, model.StatusCanceled); err != nil {
					log.Error().Err(err).Msgf("cannot update workflow with id %d state", workflow.ID)
				}
			}
		}
	}

	plState := model.StatusKilled
	if hasPendingOnly {
		plState = model.StatusCanceled
	}
	killedPipeline, err := UpdateToStatusKilled(store, *pipeline, cancelInfo, plState)
	if err != nil {
		log.Error().Err(err).Msgf("UpdateToStatusKilled: %v", pipeline)
		return err
	}

	// Load the workflows BEFORE publishing the status. updatePipelineStatus is a
	// loop over pipeline.Workflows, and the pipeline handed to Cancel does not
	// carry them — so with these two statements the other way around the loop ran
	// zero times and cancelling a pipeline never updated the forge status at all.
	// It went unnoticed because the agent sets the status again when it reports
	// the killed steps; for a pipeline with no agent left there is no second path,
	// and the commit stayed `pending` forever.
	if killedPipeline.Workflows, err = store.WorkflowGetTree(killedPipeline); err != nil {
		return err
	}

	updatePipelineStatus(ctx, _forge, killedPipeline, repo, user)

	if err := server.Config.Services.Scheduler.PublishPipelineEvent(ctx, repo, killedPipeline); err != nil {
		log.Error().Err(err).Msg("could not push pipeline status change to pubsub provider")
	}

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
			SupersededBy: pipeline.Number,
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
