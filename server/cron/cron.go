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

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/pipeline"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

const (
	// checkTime specifies the interval woodpecker checks for new cron jobs to exec
	checkTime = 10 * time.Second

	// checkItems specifies the batch size of jobs to retrieve per check from database
	checkItems = 10
)

// Start starts the cron scheduler loop
func Start(ctx context.Context, store store.Store) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(checkTime):
			go func() {
				now := time.Now()
				log.Trace().Msg("Cron: fetch next jobs")

				jobs, err := store.CronListNextExecute(now.Unix(), checkItems)
				if err != nil {
					log.Error().Err(err).Int64("now", now.Unix()).Msg("obtain cron job list")
					return
				}

				for _, job := range jobs {
					if err := runJob(job, store, now); err != nil {
						log.Error().Err(err).Int64("jobID", job.ID).Msg("run cron job failed")
					}
				}
			}()
		}
	}
}

// CalcNewNext parses a cron string and calculates the next exec time based on it
func CalcNewNext(schedule string, now time.Time) (time.Time, error) {
	c, err := cron.Parse(schedule)
	if err != nil {
		return time.Time{}, fmt.Errorf("cron parse schedule: %v", err)
	}
	return c.Next(now), nil
}

func runJob(job *model.CronJob, store store.Store, now time.Time) error {
	log.Trace().Msgf("Cron: run job [%d]", job.ID)
	ctx := context.Background()

	newNext, err := CalcNewNext(job.Schedule, now)
	if err != nil {
		return err
	}

	// try to get lock on cron job
	gotLock, err := store.CronGetLock(job, newNext.Unix())
	if err != nil {
		return err
	}
	if !gotLock {
		// an other go routine caught it
		return nil
	}

	repo, newBuild, err := createBuild(ctx, job, store)
	if err != nil {
		return err
	}

	_, err = pipeline.Create(ctx, store, repo, newBuild)
	return err
}

func createBuild(ctx context.Context, job *model.CronJob, store store.Store) (*model.Repo, *model.Build, error) {
	remote := server.Config.Services.Remote

	repo, err := store.GetRepo(job.RepoID)
	if err != nil {
		return nil, nil, err
	}

	if job.Branch == "" {
		// fallback to the repos default branch
		job.Branch = repo.Branch
	}

	creator, err := store.GetUser(job.CreatorID)
	if err != nil {
		return nil, nil, err
	}

	commit, err := remote.BranchCommit(ctx, creator, repo, job.Branch)
	if err != nil {
		return nil, nil, err
	}

	return repo, &model.Build{
		Event:     model.EventCron,
		Commit:    commit,
		Ref:       "refs/heads/" + job.Branch,
		Branch:    job.Branch,
		Message:   job.Title,
		Timestamp: job.NextExec,
		Sender:    job.Title,
		Link:      repo.Link,
	}, nil
}
