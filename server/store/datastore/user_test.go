// Copyright 2022 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestUsers(t *testing.T) {
	store, closer := newTestStore(t, new(model.User), new(model.Org), new(model.Secret), new(model.Repo), new(model.Perm))
	defer closer()

	count, err := store.GetUserCount()
	assert.NoError(t, err)
	assert.Zero(t, count)

	user := model.User{
		Login:        "joe",
		AccessToken:  "f0b461ca586c27872b43a0685cbc2847",
		RefreshToken: "976f22a5eef7caacb7e678d6c52f49b1",
		Email:        "foo@bar.com",
		Avatar:       "b9015b0857e16ac4d94a0ffd9a0b79c8",
	}
	err = store.CreateUser(&user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)

	err2 := store.UpdateUser(&user)
	assert.NoError(t, err2)

	getUser, err := store.GetUser(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, getUser.ID)
	assert.Equal(t, user.Login, getUser.Login)
	assert.Equal(t, user.AccessToken, getUser.AccessToken)
	assert.Equal(t, user.RefreshToken, getUser.RefreshToken)
	assert.Equal(t, user.Email, getUser.Email)
	assert.Equal(t, user.Avatar, getUser.Avatar)

	getUser, err = store.GetUserLogin(user.Login)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, getUser.ID)
	assert.Equal(t, user.Login, getUser.Login)

	// check unique login
	user2 := model.User{
		Login:       "Joe",
		Email:       "foo2@bar.com",
		AccessToken: "ab20g0ddaf012c744e136da16aa21ad9",
	}
	err2 = store.CreateUser(&user2)
	assert.Error(t, err2)

	user2 = model.User{
		Login:       "jane",
		Email:       "foo@bar.com",
		AccessToken: "ab20g0ddaf012c744e136da16aa21ad9",
		Hash:        "A",
	}
	assert.NoError(t, store.CreateUser(&user2))
	users, err := store.GetUserList(&model.ListOptions{Page: 1, PerPage: 50})
	assert.NoError(t, err)
	assert.Len(t, users, 2)
	// "jane" user is first due to alphabetic sorting
	assert.Equal(t, user2.Login, users[0].Login)
	assert.Equal(t, user2.Email, users[0].Email)
	assert.Equal(t, user2.AccessToken, users[0].AccessToken)

	count, err = store.GetUserCount()
	assert.NoError(t, err)
	assert.EqualValues(t, 2, count)

	getUser, err1 := store.GetUser(user.ID)
	assert.NoError(t, err1)
	err2 = store.DeleteUser(getUser)
	assert.NoError(t, err2)
	_, err3 := store.GetUser(getUser.ID)
	assert.Error(t, err3)
}

func TestCreateUserWithExistingOrg(t *testing.T) {
	store, closer := newTestStore(t, new(model.User), new(model.Org), new(model.Perm))
	defer closer()

	existingOrg := &model.Org{
		ForgeID: 1,
		IsUser:  true,
		Name:    "existingOrg",
		Private: false,
	}

	err := store.OrgCreate(existingOrg)
	assert.NoError(t, err)
	assert.EqualValues(t, "existingOrg", existingOrg.Name)

	// Create a new user with the same name as the existing organization
	newUser := &model.User{
		Login: "existingOrg",
		Hash:  "A",
	}
	err = store.CreateUser(newUser)
	assert.NoError(t, err)

	updatedOrg, err := store.OrgGet(existingOrg.ID)
	assert.NoError(t, err)
	assert.Equal(t, "existingOrg", updatedOrg.Name)

	newUser2 := &model.User{
		Login: "new-user",
		Hash:  "B",
	}
	err = store.CreateUser(newUser2)
	assert.NoError(t, err)

	newOrg, err := store.OrgFindByName("new-user")
	assert.NoError(t, err)
	assert.Equal(t, "new-user", newOrg.Name)
}
