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

	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/queue"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// Cancel the pipeline and returns the status.
func Cancel(ctx context.Context, store store.Store, repo *model.Repo, pipeline *model.Pipeline) error {
	if pipeline.Status != model.StatusRunning && pipeline.Status != model.StatusPending && pipeline.Status != model.StatusBlocked {
		return &ErrBadRequest{Msg: "Cannot cancel a non-running or non-pending or non-blocked pipeline"}
	}

	steps, err := store.StepList(pipeline)
	if err != nil {
		return &ErrNotFound{Msg: err.Error()}
	}

	// First cancel/evict steps in the queue in one go
	var (
		stepsToCancel []string
		stepsToEvict  []string
	)
	for _, step := range steps {
		if step.PPID != 0 {
			continue
		}
		if step.State == model.StatusRunning {
			stepsToCancel = append(stepsToCancel, fmt.Sprint(step.ID))
		}
		if step.State == model.StatusPending {
			stepsToEvict = append(stepsToEvict, fmt.Sprint(step.ID))
		}
	}

	if len(stepsToEvict) != 0 {
		if err := server.Config.Services.Queue.EvictAtOnce(ctx, stepsToEvict); err != nil {
			log.Error().Err(err).Msgf("queue: evict_at_once: %v", stepsToEvict)
		}
		if err := server.Config.Services.Queue.ErrorAtOnce(ctx, stepsToEvict, queue.ErrCancel); err != nil {
			log.Error().Err(err).Msgf("queue: evict_at_once: %v", stepsToEvict)
		}
	}
	if len(stepsToCancel) != 0 {
		if err := server.Config.Services.Queue.ErrorAtOnce(ctx, stepsToCancel, queue.ErrCancel); err != nil {
			log.Error().Err(err).Msgf("queue: evict_at_once: %v", stepsToCancel)
		}
	}

	// Then update the DB status for pending pipelines
	// Running ones will be set when the agents stop on the cancel signal
	for _, step := range steps {
		if step.State == model.StatusPending {
			if step.PPID != 0 {
				if _, err = UpdateStepToStatusSkipped(store, *step, 0); err != nil {
					log.Error().Msgf("error: done: cannot update step_id %d state: %s", step.ID, err)
				}
			} else {
				if _, err = UpdateStepToStatusKilled(store, *step); err != nil {
					log.Error().Msgf("error: done: cannot update step_id %d state: %s", step.ID, err)
				}
			}
		}
	}

	killedBuild, err := UpdateToStatusKilled(store, *pipeline)
	if err != nil {
		log.Error().Err(err).Msgf("UpdateToStatusKilled: %v", pipeline)
		return err
	}

	steps, err = store.StepList(killedBuild)
	if err != nil {
		return &ErrNotFound{Msg: err.Error()}
	}
	if killedBuild.Steps, err = model.Tree(steps); err != nil {
		return err
	}
	if err := publishToTopic(ctx, killedBuild, repo); err != nil {
		log.Error().Err(err).Msg("publishToTopic")
	}

	return nil
}

func cancelPreviousPipelines(
	ctx context.Context,
	_store store.Store,
	pipeline *model.Pipeline,
	repo *model.Repo,
) error {
	// check this event should cancel previous pipelines
	eventIncluded := false
	for _, ev := range repo.CancelPreviousPipelineEvents {
		if ev == pipeline.Event {
			eventIncluded = true
			break
		}
	}
	if !eventIncluded {
		return nil
	}

	// get all active activeBuilds
	activeBuilds, err := _store.GetActivePipelineList(repo, -1)
	if err != nil {
		return err
	}

	pipelineNeedsCancel := func(active *model.Pipeline) (bool, error) {
		// always filter on same event
		if active.Event != pipeline.Event {
			return false, nil
		}

		// find events for the same context
		switch pipeline.Event {
		case model.EventPush:
			return pipeline.Branch == active.Branch, nil
		default:
			return pipeline.Refspec == active.Refspec, nil
		}
	}

	for _, active := range activeBuilds {
		if active.ID == pipeline.ID {
			// same pipeline. e.g. self
			continue
		}

		cancel, err := pipelineNeedsCancel(active)
		if err != nil {
			log.Error().
				Err(err).
				Str("Ref", active.Ref).
				Msg("Error while trying to cancel pipeline, skipping")
			continue
		}

		if !cancel {
			continue
		}

		if err = Cancel(ctx, _store, repo, active); err != nil {
			log.Error().
				Err(err).
				Str("Ref", active.Ref).
				Int64("ID", active.ID).
				Msg("Failed to cancel pipeline")
		}
	}

	return nil
}
