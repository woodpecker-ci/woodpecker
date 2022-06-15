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

	"github.com/woodpecker-ci/woodpecker/server/store"
)

func StartLoop(ctx context.Context, store store.Store) {
	// TODO
}

/*

// ExclusiveUpdateNext only update if next_exec has not changed
// It then calculates next exec time and save it
func ExclusiveUpdateNext(cj *CronJob) {
		oldExec := cj.NextExec
		cj.NextExec = calc(cj.NextExec+cj.Schedule.Unix()))
		updated, err :=engine.Where("id=? AND next_exec=?", cj.IDm oldExec).Cols("next_exec").Update(cj)
		err != nil -> well error handling
		if updated == 0 {
			-> no excluseive lock -> somebody else did the job
		}
}

next_exec -> sql(select * where next_exec <= now())


for each min {
	for each db.getNextExecs(repo.IsActive, time.Now())... {
		branch := cj.Branch
		repo := getRepo(cj.RepoID)
		if branch == "" {
			branch = repo.DefaultBranch
		}

		commit, err := remote.GetLatestCommit(repo, branch)

		return repo, &model.Build{
			Event:        model.EventCron,
			Commit:       commit,
			Ref:          "refs/heads/"+branch,
			Branch:       branch,
			Message:      cj.Title,
			Avatar:       avatarFromUser(cj.AuthorID),
			Author:       getUser(cj.AuthorID),
			Email:        getMail(cj.AuthorID),
			Timestamp:    time.Now().UTC().Unix(),
			Sender:       "TODO: Cron?",
		}
		) -> (*model.Repo, *model.Build, error)
	}
}

*/
