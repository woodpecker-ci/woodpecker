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

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

const (
	// Condition for users and orgs with a forge name conflict.
	hasConflict = "conflict_forge_remote_id <> ''"
	// Order with users and orgs with a forge name conflict first.
	conflictsFirst = "CASE WHEN " + hasConflict + " THEN 0 ELSE 1 END"
)

func (s storage) OrgCreate(org *model.Org) error {
	return retryWrite(func() error {
		sess := s.engine.NewSession()
		defer sess.Close()
		return s.orgCreate(org, sess)
	})
}

func (s storage) orgCreate(org *model.Org, sess *xorm.Session) error {
	if org.Name == "" {
		return fmt.Errorf("org name is empty")
	}
	return wrapInsert(omitInvalidForgeRemoteID(sess, org).Insert(org))
}

// omitInvalidForgeRemoteID keeps the forge remote id of orgs without a valid
// one at NULL, so they don't collide in the unique index.
func omitInvalidForgeRemoteID(sess *xorm.Session, org *model.Org) *xorm.Session {
	if !org.ForgeRemoteID.IsValid() {
		return sess.Omit("forge_remote_id")
	}
	return sess
}

func (s storage) OrgGet(id int64) (*model.Org, error) {
	org := new(model.Org)
	return org, wrapGet(s.engine.ID(id).Get(org))
}

func (s storage) OrgUpdate(org *model.Org) error {
	return retryWrite(func() error {
		sess := s.engine.NewSession()
		defer sess.Close()
		return s.orgUpdate(sess, org)
	})
}

func (s storage) orgUpdate(sess *xorm.Session, org *model.Org) error {
	// update
	return wrapUpdate(omitInvalidForgeRemoteID(sess, org).ID(org.ID).AllCols().Update(org))
}

func (s storage) OrgDelete(id int64) error {
	return retryWrite(func() error {
		return s.orgDelete(s.engine.NewSession(), id)
	})
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

func (s storage) OrgFindByName(name string, forgeID int64) (*model.Org, error) {
	return s.orgFindByName(s.engine.NewSession(), name, forgeID)
}

func (s storage) orgFindByName(sess *xorm.Session, name string, forgeID int64) (*model.Org, error) {
	// sanitize
	org := new(model.Org)
	return org, wrapGet(sess.Where("LOWER(name) = ?", strings.ToLower(name)).And("forge_id = ?", forgeID).Get(org))
}

func (s storage) OrgFindByRemoteID(forgeID int64, remoteID model.ForgeRemoteID) (*model.Org, error) {
	org := new(model.Org)
	return org, wrapGet(s.engine.Where("forge_id = ? AND forge_remote_id = ?", forgeID, remoteID).Get(org))
}

func (s storage) OrgLookup(name string) (*model.Org, error) {
	var orgs []*model.Org
	// we limit to 2 orgs, if we have >= 2 we return an error anyways.
	err := s.engine.Where("LOWER(name) = ?", strings.ToLower(name)).Limit(2).Find(&orgs) //nolint:mnd
	if err != nil {
		return nil, err
	}
	if len(orgs) < 1 {
		return nil, types.ErrRecordNotExist
	}
	if len(orgs) > 1 {
		return nil, fmt.Errorf("found more than one org with this name")
	}
	return orgs[0], nil
}

func (s storage) OrgRepoList(org *model.Org, p *model.ListOptionsWithAll) ([]*model.Repo, error) {
	var repos []*model.Repo
	return repos, s.paginate(p).OrderBy("id").Where("org_id = ?", org.ID).Find(&repos)
}

// OrgList lists all orgs, and user orgs with a name conflict. Name conflicts come first.
func (s storage) OrgList(p *model.ListOptionsWithAll) ([]*model.Org, error) {
	var orgs []*model.Org
	return orgs, s.paginate(p).
		Where("is_user = ? OR "+hasConflict, false).
		OrderBy(conflictsFirst + ", id").
		Find(&orgs)
}
