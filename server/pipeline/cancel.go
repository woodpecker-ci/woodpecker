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
	"github.com/woodpecker-ci/woodpecker/server/shared"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// Cancel the pipeline and returns the status.
func Cancel(ctx context.Context, store store.Store, repo *model.Repo, pipeline *model.Pipeline) error {
	if pipeline.Status != model.StatusRunning && pipeline.Status != model.StatusPending && pipeline.Status != model.StatusBlocked {
		return ErrBadRequest{Msg: "Cannot cancel a non-running or non-pending or non-blocked pipeline"}
	}

	procs, err := store.ProcList(pipeline)
	if err != nil {
		return ErrNotFound{Msg: err.Error()}
	}

	// First cancel/evict procs in the queue in one go
	var (
		procsToCancel []string
		procsToEvict  []string
	)
	for _, proc := range procs {
		if proc.PPID != 0 {
			continue
		}
		if proc.State == model.StatusRunning {
			procsToCancel = append(procsToCancel, fmt.Sprint(proc.ID))
		}
		if proc.State == model.StatusPending {
			procsToEvict = append(procsToEvict, fmt.Sprint(proc.ID))
		}
	}

	if len(procsToEvict) != 0 {
		if err := server.Config.Services.Queue.EvictAtOnce(ctx, procsToEvict); err != nil {
			log.Error().Err(err).Msgf("queue: evict_at_once: %v", procsToEvict)
		}
		if err := server.Config.Services.Queue.ErrorAtOnce(ctx, procsToEvict, queue.ErrCancel); err != nil {
			log.Error().Err(err).Msgf("queue: evict_at_once: %v", procsToEvict)
		}
	}
	if len(procsToCancel) != 0 {
		if err := server.Config.Services.Queue.ErrorAtOnce(ctx, procsToCancel, queue.ErrCancel); err != nil {
			log.Error().Err(err).Msgf("queue: evict_at_once: %v", procsToCancel)
		}
	}

	// Then update the DB status for pending pipelines
	// Running ones will be set when the agents stop on the cancel signal
	for _, proc := range procs {
		if proc.State == model.StatusPending {
			if proc.PPID != 0 {
				if _, err = shared.UpdateProcToStatusSkipped(store, *proc, 0); err != nil {
					log.Error().Msgf("error: done: cannot update proc_id %d state: %s", proc.ID, err)
				}
			} else {
				if _, err = shared.UpdateProcToStatusKilled(store, *proc); err != nil {
					log.Error().Msgf("error: done: cannot update proc_id %d state: %s", proc.ID, err)
				}
			}
		}
	}

	killedBuild, err := shared.UpdateToStatusKilled(store, *pipeline)
	if err != nil {
		log.Error().Err(err).Msgf("UpdateToStatusKilled: %v", pipeline)
		return err
	}

	procs, err = store.ProcList(killedBuild)
	if err != nil {
		return ErrNotFound{Msg: err.Error()}
	}
	if killedBuild.Procs, err = model.Tree(procs); err != nil {
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
