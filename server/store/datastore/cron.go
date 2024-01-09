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
	"go.woodpecker-ci.org/woodpecker/v2/server/model"

	"xorm.io/builder"
)

func (s storage) CronCreate(cron *model.Cron) error {
	if err := cron.Validate(); err != nil {
		return err
	}
	_, err := s.engine.Insert(cron)
	return err
}

func (s storage) CronFind(repo *model.Repo, id int64) (*model.Cron, error) {
	cron := new(model.Cron)
	return cron, wrapGet(s.engine.ID(id).Where("repo_id = ?", repo.ID).Get(cron))
}

func (s storage) CronList(repo *model.Repo, p *model.ListOptions) ([]*model.Cron, error) {
	var crons []*model.Cron
	return crons, s.paginate(p).Where("repo_id = ?", repo.ID).Find(&crons)
}

func (s storage) CronUpdate(_ *model.Repo, cron *model.Cron) error {
	_, err := s.engine.ID(cron.ID).AllCols().Update(cron)
	return err
}

func (s storage) CronDelete(repo *model.Repo, id int64) error {
	return wrapDelete(s.engine.ID(id).Where("repo_id = ?", repo.ID).Delete(new(model.Cron)))
}

// CronListNextExecute returns limited number of jobs with NextExec being less or equal to the provided unix timestamp
func (s storage) CronListNextExecute(nextExec, limit int64) ([]*model.Cron, error) {
	crons := make([]*model.Cron, 0, limit)
	return crons, s.engine.Where(builder.Lte{"next_exec": nextExec}).Limit(int(limit)).Find(&crons)
}

// CronGetLock try to get a lock by updating NextExec
func (s storage) CronGetLock(cron *model.Cron, newNextExec int64) (bool, error) {
	cols, err := s.engine.ID(cron.ID).Where(builder.Eq{"next_exec": cron.NextExec}).
		Cols("next_exec").Update(&model.Cron{NextExec: newNextExec})
	gotLock := cols != 0

	if err == nil && gotLock {
		cron.NextExec = newNextExec
	}

	return gotLock, err
}
