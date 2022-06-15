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
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/queue"
	"github.com/woodpecker-ci/woodpecker/server/shared"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// Cancel the build and returns the status.
func Cancel(
	ctx context.Context,
	_store store.Store,
	repo *model.Repo,
	build *model.Build,
) (int, error) {
	procs, err := _store.ProcList(build)
	if err != nil {
		return http.StatusNotFound, err
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

	// Then update the DB status for pending builds
	// Running ones will be set when the agents stop on the cancel signal
	for _, proc := range procs {
		if proc.State == model.StatusPending {
			if proc.PPID != 0 {
				if _, err = shared.UpdateProcToStatusSkipped(_store, *proc, 0); err != nil {
					log.Error().Msgf("error: done: cannot update proc_id %d state: %s", proc.ID, err)
				}
			} else {
				if _, err = shared.UpdateProcToStatusKilled(_store, *proc); err != nil {
					log.Error().Msgf("error: done: cannot update proc_id %d state: %s", proc.ID, err)
				}
			}
		}
	}

	killedBuild, err := shared.UpdateToStatusKilled(_store, *build)
	if err != nil {
		log.Error().Err(err).Msgf("UpdateToStatusKilled: %v", build)
		return http.StatusInternalServerError, err
	}

	procs, err = _store.ProcList(killedBuild)
	if err != nil {
		return http.StatusNotFound, err
	}
	if killedBuild.Procs, err = model.Tree(procs); err != nil {
		return http.StatusInternalServerError, err
	}
	if err := publishToTopic(ctx, killedBuild, repo); err != nil {
		log.Error().Err(err).Msg("publishToTopic")
	}

	return http.StatusNoContent, nil
}

func cancelPreviousPipelines(
	ctx context.Context,
	_store store.Store,
	build *model.Build,
	repo *model.Repo,
) error {
	// check this event should cancel previous pipelines
	eventIncluded := false
	for _, ev := range repo.CancelPreviousPipelineEvents {
		if ev == build.Event {
			eventIncluded = true
			break
		}
	}
	if !eventIncluded {
		return nil
	}

	// get all active activeBuilds
	activeBuilds, err := _store.GetActiveBuildList(repo, -1)
	if err != nil {
		return err
	}

	buildNeedsCancel := func(active *model.Build) (bool, error) {
		// always filter on same event
		if active.Event != build.Event {
			return false, nil
		}

		// find events for the same context
		switch build.Event {
		case model.EventPush:
			return build.Branch == active.Branch, nil
		default:
			return build.Refspec == active.Refspec, nil
		}
	}

	for _, active := range activeBuilds {
		if active.ID == build.ID {
			// same build. e.g. self
			continue
		}

		cancel, err := buildNeedsCancel(active)
		if err != nil {
			log.Error().
				Err(err).
				Str("Ref", active.Ref).
				Msg("Error while trying to cancel build, skipping")
			continue
		}
		if !cancel {
			continue
		}
		_, err = Cancel(ctx, _store, repo, active)
		if err != nil {
			log.Error().
				Err(err).
				Str("Ref", active.Ref).
				Int64("ID", active.ID).
				Msg("Failed to cancel build")
		}
	}

	return nil
}
