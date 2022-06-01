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
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// checkTime specify the interfall woodpecker look for new cron jobs to exec
const checkTime = time.Minute

// Start starts the cron functionality
func Start(ctx context.Context, store store.Store) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(checkTime):
		go func() {
			now := time.Now().Unix()
			jobs, err := store.CronList(now, 10)
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
}

func runJob(job *model.CronJob, store store.Store) error {
	schedule, err := cron.Parse(job.Schedule)
	if err != nil {
		return fmt.Errorf("cron parse schedule: %v", err)
	}
	newNext := schedule.Next(time.Unix(job.NextExec, 0))

	// try to get lock on cron job
	gotLock, err := store.CronGetLock(job, newNext.Unix())
	if err != nil {
		return err
	}
	if !gotLock {
		// a other go routine catched it
		return nil
	}

	repo, build, err := createBuild(job, store)
	if err != nil {
		return err
	}
	fmt.Printf("\n%#v\n%#v\n", repo, build)
	// TODO -> build start

	return nil
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
		Event:     model.EventCron,
		Commit:    commit,
		Ref:       "refs/heads/" + branch,
		Branch:    job.Branch,
		Message:   job.Title,
		Avatar:    avatarFromUser(cj.AuthorID),
		Author:    getUser(cj.AuthorID),
		Email:     getMail(cj.AuthorID),
		Timestamp: time.Now().UTC().Unix(),
		Sender:    "TODO: Cron?",
	}, nil
}
