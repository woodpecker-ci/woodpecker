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

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/pipeline"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

const (
	// checkTime specify the interval woodpecker look for new cron jobs to exec
	checkTime = time.Second

	// checkItems specify the jobs to retrieve per check interval from database
	checkItems = 10
)

// Start starts the cron functionality
func Start(ctx context.Context, store store.Store) error {
	select {
	case <-ctx.Done():
		return nil
	case <-time.After(checkTime):
		go func() {
			now := time.Now().Unix()
			jobs, err := store.CronListNextExecute(now, checkItems)
			if err != nil {
				log.Error().Err(err).Int64("now", now).Msg("obtain cron job list")
				return
			}
			for _, job := range jobs {
				if err := runJob(job, store); err != nil {
					log.Error().Err(err).Int64("jobID", job.ID).Msg("run cron job failed")
				}
			}
		}()
	}
	return nil
}

func runJob(job *model.CronJob, store store.Store) error {
	schedule, err := cron.Parse(job.Schedule)
	if err != nil {
		return fmt.Errorf("cron parse schedule: %v", err)
	}
	oldNext := time.Unix(job.NextExec, 0)
	if job.NextExec == 0 {
		oldNext = time.Now()
	}
	newNext := schedule.Next(oldNext) // TODO: can we always use time.Now() here?

	// try to get lock on cron job
	gotLock, err := store.CronGetLock(job, newNext.Unix())
	if err != nil {
		return err
	}
	if !gotLock {
		// an other go routine caught it
		return nil
	}

	repo, newBuild, err := createBuild(job, store)
	if err != nil {
		return err
	}

	_, err = pipeline.Create(context.Background(), store, repo, newBuild)
	return err
}

func createBuild(job *model.CronJob, store store.Store) (*model.Repo, *model.Build, error) {
	repo, err := store.GetRepo(job.RepoID)
	if err != nil {
		return nil, nil, err
	}

	commit, err := "", nil // remote.GetLatestCommit(repo, branch)

	if job.Branch == "" {
		job.Branch = repo.Branch
	}

	return repo, &model.Build{
		Event:   model.EventCron,
		Commit:  commit,
		Ref:     "refs/heads/" + job.Branch,
		Branch:  job.Branch,
		Message: job.Title,
		// Avatar:    avatarFromUser(cj.AuthorID),
		// Author:    getUser(cj.AuthorID),
		// Email:     getMail(cj.AuthorID),
		Timestamp: job.NextExec,
		Sender:    "TODO: Cron?",
	}, nil
}
