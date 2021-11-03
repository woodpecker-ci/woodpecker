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

package datastore_xorm

import (
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) GetUser(id int64) (*model.User, error) {
	user := new(model.User)
	return user, wrapGet(s.engine.ID(id).Get(user))
}

func (s storage) GetUserLogin(login string) (*model.User, error) {
	user := new(model.User)
	return user, wrapGet(s.engine.Where("user_login=?", login).Get(user))
}

func (s storage) GetUserList() ([]*model.User, error) {
	users := make([]*model.User, 0, 10)
	return users, s.engine.Find(&users)
}

func (s storage) GetUserCount() (int64, error) {
	return s.engine.Count(&model.User{})
}

func (s storage) CreateUser(user *model.User) error {
	_, err := s.engine.InsertOne(user)
	return err
}

func (s storage) UpdateUser(user *model.User) error {
	_, err := s.engine.ID(user.ID).AllCols().Update(user)
	return err
}

func (s storage) DeleteUser(user *model.User) error {
	_, err := s.engine.ID(user.ID).Delete(&user)
	// TODO: delete related content that need this user to work
	return err
}
