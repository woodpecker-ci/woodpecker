// Copyright 2021 Woodpecker Authors
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

	"xorm.io/builder"
	"xorm.io/xorm"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) PermFind(user *model.User, repo *model.Repo) (*model.Perm, error) {
	perm := &model.Perm{
		UserID: user.ID,
		RepoID: repo.ID,
	}
	return perm, wrapGet(s.engine.Get(perm))
}

func (s storage) PermUpsert(perm *model.Perm) error {
	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	if err := s.permUpsert(sess, perm); err != nil {
		return err
	}

	return sess.Commit()
}

func (s storage) permUpsert(sess *xorm.Session, perm *model.Perm) error {
	if perm.RepoID == 0 && perm.Repo == nil {
		return fmt.Errorf("could not determine repo for permission: %v", perm)
	}

	// lookup repo based on name or forge ID if possible
	if perm.RepoID == 0 && perm.Repo != nil {
		r, err := s.getRepoNameFallback(sess, perm.Repo.ForgeRemoteID, perm.Repo.FullName)
		if err != nil {
			return err
		}
		perm.RepoID = r.ID
	}

	exist, err := sess.Where("perm_user_id = ? AND perm_repo_id = ?", perm.UserID, perm.RepoID).
		Exist(new(model.Perm))
	if err != nil {
		return err
	}

	if exist {
		_, err = sess.Where("perm_user_id = ? AND perm_repo_id = ?", perm.UserID, perm.RepoID).
			AllCols().Update(perm)
	} else {
		// only Insert set auto created ID back to object
		_, err = sess.Insert(perm)
	}
	return err
}

func (s storage) PermDelete(perm *model.Perm) error {
	_, err := s.engine.
		Where("perm_user_id = ? AND perm_repo_id = ?", perm.UserID, perm.RepoID).
		Delete(new(model.Perm))
	return err
}

func (s storage) PermFlush(user *model.User, before int64) error {
	_, err := s.engine.
		Where("perm_user_id = ? AND perm_synced < ?", user.ID, before).
		Delete(new(model.Perm))
	return err
}

// userPushOrAdminCondition return condition where user must have push or admin rights
// if used make sure to have permission table ("perms") joined
func userPushOrAdminCondition(userID int64) builder.Cond {
	return builder.Eq{"perms.perm_user_id": userID}.
		And(builder.Eq{"perms.perm_push": true}.
			Or(builder.Eq{"perms.perm_admin": true}))
}
