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

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func (s storage) GetUser(id int64) (*model.User, error) {
	user := new(model.User)
	return user, wrapGet(s.engine.ID(id).Get(user))
}

func (s storage) GetUserByRemoteID(forgeID int64, userRemoteID model.ForgeRemoteID) (*model.User, error) {
	sess := s.engine.NewSession()
	user := new(model.User)
	return user, wrapGet(sess.Where("forge_id = ? AND forge_remote_id = ?", forgeID, userRemoteID).Get(user))
}

func (s storage) GetUserByLogin(forgeID int64, login string) (*model.User, error) {
	sess := s.engine.NewSession()
	user := new(model.User)
	return user, wrapGet(sess.Where("forge_id = ? AND login=?", forgeID, login).Get(user))
}

func (s storage) GetUserList(p *model.ListOptionsWithAll) ([]*model.User, error) {
	var users []*model.User
	return users, s.paginate(p).OrderBy(conflictsFirst + ", login").Find(&users)
}

func (s storage) GetUserCount() (int64, error) {
	return s.engine.Count(new(model.User))
}

// CreateUser creates the user. Its org is linked on login, by the forge remote id of the user.
func (s storage) CreateUser(user *model.User) error {
	// only Insert set auto created ID back to object
	return wrapInsert(s.engine.Insert(user))
}

func (s storage) UpdateUser(user *model.User) error {
	return retryWrite(func() error {
		return wrapUpdate(s.engine.ID(user.ID).AllCols().Update(user))
	})
}

func (s storage) DeleteUser(user *model.User) error {
	return retryWrite(func() error {
		sess := s.engine.NewSession()
		defer sess.Close()
		if err := sess.Begin(); err != nil {
			return err
		}

		if user.OrgID != 0 {
			if err := s.orgDelete(sess, user.OrgID); err != nil {
				return fmt.Errorf("failed to delete org: %w", err)
			}
		}

		if err := wrapDelete(sess.ID(user.ID).Delete(new(model.User))); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		if _, err := sess.Where("user_id = ?", user.ID).Delete(new(model.Perm)); err != nil {
			return fmt.Errorf("failed to delete perms: %w", err)
		}

		return sess.Commit()
	})
}
