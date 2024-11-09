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
	"strings"

	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func (s storage) OrgCreate(org *model.Org) error {
	return s.orgCreate(org, s.engine.NewSession())
}

func (s storage) orgCreate(org *model.Org, sess *xorm.Session) error {
	// sanitize
	org.Name = strings.ToLower(org.Name)
	if org.Name == "" {
		return fmt.Errorf("org name is empty")
	}
	// insert
	_, err := sess.Insert(org)
	return err
}

func (s storage) OrgGet(id int64) (*model.Org, error) {
	org := new(model.Org)
	return org, wrapGet(s.engine.ID(id).Get(org))
}

func (s storage) OrgUpdate(org *model.Org) error {
	// sanitize
	org.Name = strings.ToLower(org.Name)
	// update
	_, err := s.engine.ID(org.ID).AllCols().Update(org)
	return err
}

func (s storage) OrgDelete(id int64) error {
	return s.orgDelete(s.engine.NewSession(), id)
}

func (s storage) orgDelete(sess *xorm.Session, id int64) error {
	if _, err := sess.Where("org_id = ?", id).Delete(new(model.Secret)); err != nil {
		return err
	}

	var repos []*model.Repo
	if err := sess.Where("org_id = ?", id).Find(&repos); err != nil {
		return err
	}

	for _, repo := range repos {
		if err := s.deleteRepo(sess, repo); err != nil {
			return err
		}
	}

	return wrapDelete(sess.ID(id).Delete(new(model.Org)))
}

func (s storage) OrgFindByName(name string) (*model.Org, error) {
	// sanitize
	name = strings.ToLower(name)
	// find
	org := new(model.Org)
	return org, wrapGet(s.engine.Where("name = ?", name).Get(org))
}

func (s storage) OrgRepoList(org *model.Org, p *model.ListOptions) ([]*model.Repo, error) {
	var repos []*model.Repo
	return repos, s.paginate(p).OrderBy("id").Where("org_id = ?", org.ID).Find(&repos)
}

func (s storage) OrgList(p *model.ListOptions) ([]*model.Org, error) {
	var orgs []*model.Org
	return orgs, s.paginate(p).Where("is_user = ?", false).OrderBy("id").Find(&orgs)
}
