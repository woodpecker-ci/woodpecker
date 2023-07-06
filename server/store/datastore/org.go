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
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) OrgCreate(org *model.Org) error {
	_, err := s.engine.Insert(org)
	return err
}

func (s storage) OrgFind(repo *model.Repo, id int64) (*model.Cron, error) {
	cron := &model.Cron{
		RepoID: repo.ID,
		ID:     id,
	}
	return cron, wrapGet(s.engine.Get(cron))
}

func (s storage) OrgUpdate(org *model.Cron) error {
	_, err := s.engine.ID(org.ID).AllCols().Update(org)
	return err
}

func (s storage) OrgDelete(id int64) error {
	return wrapDelete(s.engine.ID(id).Delete(new(model.Org)))
}
