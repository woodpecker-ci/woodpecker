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
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func (s storage) GetUser(id int64) (*model.User, error) {
	user := new(model.User)
	return user, wrapGet(s.engine.ID(id).Get(user))
}

func (s storage) GetUserRemoteID(remoteID model.ForgeRemoteID, login string) (*model.User, error) {
	sess := s.engine.NewSession()
	user := new(model.User)
	err := wrapGet(sess.Where("forge_remote_id = ?", remoteID).Get(user))
	if err != nil {
		return s.getUserLogin(sess, login)
	}
	return user, err
}

func (s storage) GetUserLogin(login string) (*model.User, error) {
	return s.getUserLogin(s.engine.NewSession(), login)
}

func (s storage) getUserLogin(sess *xorm.Session, login string) (*model.User, error) {
	user := new(model.User)
	return user, wrapGet(sess.Where("login=?", login).Get(user))
}

func (s storage) GetUserList(p *model.ListOptions) ([]*model.User, error) {
	var users []*model.User
	return users, s.paginate(p).OrderBy("login").Find(&users)
}

func (s storage) GetUserCount() (int64, error) {
	return s.engine.Count(new(model.User))
}

func (s storage) CreateUser(user *model.User) error {
	sess := s.engine.NewSession()
	org := &model.Org{
		Name:   user.Login,
		IsUser: true,
	}
	err := s.orgCreate(org, sess)
	if err != nil {
		return err
	}
	user.OrgID = org.ID
	// only Insert set auto created ID back to object
	_, err = sess.Insert(user)
	return err
}

func (s storage) UpdateUser(user *model.User) error {
	_, err := s.engine.ID(user.ID).AllCols().Update(user)
	return err
}

func (s storage) DeleteUser(user *model.User) error {
	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	if err := s.orgDelete(sess, user.OrgID); err != nil {
		return err
	}

	if err := wrapDelete(sess.ID(user.ID).Delete(new(model.User))); err != nil {
		return err
	}

	if _, err := sess.Where("user_id = ?", user.ID).Delete(new(model.Perm)); err != nil {
		return err
	}

	return sess.Commit()
}
