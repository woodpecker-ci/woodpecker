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

package datastore

import (
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/model"

	"xorm.io/builder"
)

func (s storage) CronCreate(job *model.CronJob) error {
	if job.RepoID == 0 || job.Title == "" {
		return fmt.Errorf("repoID and Title required")
	}
	_, err := s.engine.Insert(job)
	return err
}

func (s storage) CronFind(repo *model.Repo, id int64) (*model.CronJob, error) {
	cronJob := &model.CronJob{
		RepoID: repo.ID,
		ID:     id,
	}
	return cronJob, wrapGet(s.engine.Get(cronJob))
}

func (s storage) CronList(repo *model.Repo) ([]*model.CronJob, error) {
	cronJobs := make([]*model.CronJob, 0, perPage)
	return cronJobs, s.engine.Where("repo_id = ?", repo.ID).Find(&cronJobs)
}

func (s storage) CronUpdate(repo *model.Repo, cronJob *model.CronJob) error {
	_, err := s.engine.ID(cronJob.ID).AllCols().Update(cronJob)
	return err
}

func (s storage) CronDelete(repo *model.Repo, id int64) error {
	_, err := s.engine.ID(id).Where("repo_id = ?", repo.ID).Delete(new(model.CronJob))
	return err
}

// CronList return limited number of jobs based on NextExec
// is less or equal than unitx timestamp
func (s storage) CronListNextExecute(nextExec, limit int64) ([]*model.CronJob, error) {
	jobs := make([]*model.CronJob, 0, limit)
	return jobs, s.engine.Where(builder.Lte{"next_exec": nextExec}).Limit(int(limit)).Find(&jobs)
}

// CronGetLock try to get a lock by updating NextExec
func (s storage) CronGetLock(job *model.CronJob, newNextExec int64) (bool, error) {
	cols, err := s.engine.ID(job.ID).Where(builder.Eq{"next_exec": job.NextExec}).
		Cols("next_exec").Update(&model.CronJob{NextExec: newNextExec})
	gotLock := cols != 0

	if err == nil && gotLock {
		job.NextExec = newNextExec
	}

	return gotLock, err
}
