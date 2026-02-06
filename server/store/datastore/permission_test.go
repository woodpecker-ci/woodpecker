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
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func TestPermFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo), new(model.Perm), new(model.User))
	defer closer()

	user := &model.User{ID: 1}
	repo := &model.Repo{
		UserID:        1,
		FullName:      "bradrydzewski/test",
		Owner:         "bradrydzewski",
		Name:          "test",
		ForgeRemoteID: "1",
	}
	assert.NoError(t, store.CreateRepo(repo))

	err := store.PermUpsert(
		&model.Perm{
			UserID: user.ID,
			RepoID: repo.ID,
			Pull:   true,
			Push:   false,
			Admin:  false,
		},
	)
	assert.NoError(t, err)

	perm, err := store.PermFind(user, repo)
	assert.NoError(t, err)
	assert.True(t, perm.Pull)
	assert.False(t, perm.Push)
	assert.False(t, perm.Admin)
}

func TestPermUpsert(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo), new(model.Perm), new(model.User))
	defer closer()

	user := &model.User{ID: 1}
	repo := &model.Repo{
		UserID:        1,
		FullName:      "bradrydzewski/test",
		Owner:         "bradrydzewski",
		Name:          "test",
		ForgeRemoteID: "1",
	}
	assert.NoError(t, store.CreateRepo(repo))

	err := store.PermUpsert(
		&model.Perm{
			UserID: user.ID,
			RepoID: repo.ID,
			Pull:   true,
			Push:   false,
			Admin:  false,
		},
	)
	assert.NoError(t, err)

	perm, err := store.PermFind(user, repo)
	assert.NoError(t, err)
	assert.True(t, perm.Pull)
	assert.False(t, perm.Push)
	assert.False(t, perm.Admin)

	//
	// this will attempt to replace the existing permissions
	// using the insert or replace logic.
	//

	err = store.PermUpsert(
		&model.Perm{
			UserID: user.ID,
			RepoID: repo.ID,
			Pull:   true,
			Push:   true,
			Admin:  true,
		},
	)
	assert.NoError(t, err)

	perm, err = store.PermFind(user, repo)
	assert.NoError(t, err)
	assert.True(t, perm.Pull)
	assert.True(t, perm.Push)
	assert.True(t, perm.Admin)
}

func TestPermPruneDeleteAll(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo), new(model.Perm), new(model.User))
	defer closer()

	user := &model.User{ID: 1}

	repo1 := &model.Repo{
		UserID:        1,
		FullName:      "woodpecker-ci/woodpecker1",
		Owner:         "woodpecker-ci",
		Name:          "repo1",
		ForgeRemoteID: "101",
	}
	repo2 := &model.Repo{
		UserID:        1,
		FullName:      "woodpecker-ci/woodpecker2",
		Owner:         "woodpecker",
		Name:          "repo2",
		ForgeRemoteID: "102",
	}
	assert.NoError(t, store.CreateRepo(repo1))
	assert.NoError(t, store.CreateRepo(repo2))

	assert.NoError(t, store.PermUpsert(&model.Perm{UserID: user.ID, RepoID: repo1.ID, Pull: true}))
	assert.NoError(t, store.PermUpsert(&model.Perm{UserID: user.ID, RepoID: repo2.ID, Pull: true}))

	_, err := store.PermFind(user, repo1)
	assert.NoError(t, err)
	_, err = store.PermFind(user, repo2)
	assert.NoError(t, err)

	assert.NoError(t, store.PermPrune(user.ID, []int64{}))

	_, err = store.PermFind(user, repo1)
	assert.ErrorIs(t, err, types.RecordNotExist)
	_, err = store.PermFind(user, repo2)
	assert.ErrorIs(t, err, types.RecordNotExist)
}

func TestPermPruneKeepOne(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo), new(model.Perm), new(model.User))
	defer closer()

	user := &model.User{ID: 1}

	repo1 := &model.Repo{
		UserID:        1,
		FullName:      "woodpecker-ci/woodpecker1",
		Owner:         "woodpecker-ci",
		Name:          "repo1",
		ForgeRemoteID: "101",
	}
	repo2 := &model.Repo{
		UserID:        1,
		FullName:      "woodpecker-ci/woodpecker2",
		Owner:         "woodpecker",
		Name:          "repo2",
		ForgeRemoteID: "102",
	}
	assert.NoError(t, store.CreateRepo(repo1))
	assert.NoError(t, store.CreateRepo(repo2))

	assert.NoError(t, store.PermUpsert(&model.Perm{UserID: user.ID, RepoID: repo1.ID, Pull: true}))
	assert.NoError(t, store.PermUpsert(&model.Perm{UserID: user.ID, RepoID: repo2.ID, Pull: true}))

	_, err := store.PermFind(user, repo1)
	assert.NoError(t, err)
	_, err = store.PermFind(user, repo2)
	assert.NoError(t, err)

	// Prune everything EXCEPT repo2
	assert.NoError(t, store.PermPrune(user.ID, []int64{repo2.ID}))

	_, err = store.PermFind(user, repo1)
	assert.ErrorIs(t, err, types.RecordNotExist)

	_, err = store.PermFind(user, repo2)
	assert.NoError(t, err)
}
