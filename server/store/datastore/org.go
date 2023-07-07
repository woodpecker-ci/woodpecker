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

func (s storage) OrgFind(id int64) (*model.Org, error) {
	org := &model.Org{
		ID: id,
	}
	return org, wrapGet(s.engine.Get(org))
}

func (s storage) OrgUpdate(org *model.Org) error {
	_, err := s.engine.ID(org.ID).AllCols().Update(org)
	return err
}

func (s storage) OrgDelete(id int64) error {
	return wrapDelete(s.engine.ID(id).Delete(new(model.Org)))
}

func (s storage) OrgFindByName(name string, p *model.ListOptions) (*model.Org, error) {
	org := &model.Org{
		Name: name,
	}
	return org, wrapGet(s.engine.Get(org))
}

func (s storage) OrgRepoList(org *model.Org, p *model.ListOptions) ([]*model.Repo, error) {
	var repos []*model.Repo
	return repos, s.paginate(p).OrderBy("repo_id").Where("org_id = ?", org.ID).Find(&repos)
}
