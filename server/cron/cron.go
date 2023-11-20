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

package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/server/forge"
	"go.woodpecker-ci.org/woodpecker/server/model"
	"go.woodpecker-ci.org/woodpecker/server/pipeline"
	"go.woodpecker-ci.org/woodpecker/server/store"
)

const (
	// checkTime specifies the interval woodpecker checks for new crons to exec
	checkTime = 10 * time.Second

	// checkItems specifies the batch size of crons to retrieve per check from database
	checkItems = 10
)

// Start starts the cron scheduler loop
func Start(ctx context.Context, store store.Store, forge forge.Forge) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(checkTime):
			go func() {
				now := time.Now()
				log.Trace().Msg("Cron: fetch next crons")

				crons, err := store.CronListNextExecute(now.Unix(), checkItems)
				if err != nil {
					log.Error().Err(err).Int64("now", now.Unix()).Msg("obtain cron list")
					return
				}

				for _, cron := range crons {
					if err := runCron(store, forge, cron, now); err != nil {
						log.Error().Err(err).Int64("cronID", cron.ID).Msg("run cron failed")
					}
				}
			}()
		}
	}
}

// CalcNewNext parses a cron string and calculates the next exec time based on it
func CalcNewNext(schedule string, now time.Time) (time.Time, error) {
	// remove local timezone
	now = now.UTC()

	// TODO: allow the users / the admin to set a specific timezone

	c, err := cron.Parse(schedule)
	if err != nil {
		return time.Time{}, fmt.Errorf("cron parse schedule: %w", err)
	}
	return c.Next(now), nil
}

func runCron(store store.Store, forge forge.Forge, cron *model.Cron, now time.Time) error {
	log.Trace().Msgf("Cron: run id[%d]", cron.ID)
	ctx := context.Background()

	newNext, err := CalcNewNext(cron.Schedule, now)
	if err != nil {
		return err
	}

	// try to get lock on cron
	gotLock, err := store.CronGetLock(cron, newNext.Unix())
	if err != nil {
		return err
	}
	if !gotLock {
		// another go routine caught it
		return nil
	}

	repo, newPipeline, err := CreatePipeline(ctx, store, forge, cron)
	if err != nil {
		return err
	}

	_, err = pipeline.Create(ctx, store, repo, newPipeline)
	return err
}

func CreatePipeline(ctx context.Context, store store.Store, f forge.Forge, cron *model.Cron) (*model.Repo, *model.Pipeline, error) {
	repo, err := store.GetRepo(cron.RepoID)
	if err != nil {
		return nil, nil, err
	}

	if cron.Branch == "" {
		// fallback to the repos default branch
		cron.Branch = repo.Branch
	}

	creator, err := store.GetUser(cron.CreatorID)
	if err != nil {
		return nil, nil, err
	}

	// If the forge has a refresh token, the current access token
	// may be stale. Therefore, we should refresh prior to dispatching
	// the pipeline.
	forge.Refresh(ctx, f, store, creator)

	commit, err := f.BranchHead(ctx, creator, repo, cron.Branch)
	if err != nil {
		return nil, nil, err
	}

	return repo, &model.Pipeline{
		Event:     model.EventCron,
		Commit:    commit,
		Ref:       "refs/heads/" + cron.Branch,
		Branch:    cron.Branch,
		Message:   cron.Name,
		Timestamp: cron.NextExec,
		Sender:    cron.Name,
		ForgeURL:  repo.ForgeURL,
	}, nil
}
