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
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func TestGetPipelineQueue(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo), new(model.User), new(model.Perm), new(model.Pipeline), new(model.Org))
	defer closer()

	user := &model.User{
		Login:       "joe",
		Email:       "foo@bar.com",
		AccessToken: "e42080dddf012c718e476da161d21ad5",
	}
	assert.NoError(t, store.CreateUser(user))

	repo1 := &model.Repo{
		Owner:         "bradrydzewski",
		Name:          "test",
		FullName:      "bradrydzewski/test",
		ForgeRemoteID: "1",
		IsActive:      true,
	}

	assert.NoError(t, store.CreateRepo(repo1))
	for _, perm := range []*model.Perm{
		{UserID: user.ID, Repo: repo1, Push: true, Admin: false},
	} {
		assert.NoError(t, store.PermUpsert(perm))
	}
	pipeline1 := &model.Pipeline{
		RepoID: repo1.ID,
		Status: model.StatusPending,
	}
	assert.NoError(t, store.CreatePipeline(pipeline1))

	feed, err := store.GetPipelineQueue()
	assert.NoError(t, err)
	assert.Len(t, feed, 1)
}

func TestUserFeed(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo), new(model.User), new(model.Perm), new(model.Pipeline), new(model.Org))
	defer closer()

	user := &model.User{
		Login:       "joe",
		Email:       "foo@bar.com",
		AccessToken: "e42080dddf012c718e476da161d21ad5",
	}
	assert.NoError(t, store.CreateUser(user))

	repo1 := &model.Repo{
		Owner:         "bradrydzewski",
		Name:          "test1",
		FullName:      "bradrydzewski/test1",
		ForgeRemoteID: "1",
		IsActive:      true,
	}
	repo2 := &model.Repo{
		Owner:         "johndoe",
		Name:          "test",
		FullName:      "johndoe/test2",
		ForgeRemoteID: "2",
		IsActive:      true,
	}

	assert.NoError(t, store.CreateRepo(repo1))
	assert.NoError(t, store.CreateRepo(repo2))

	for _, perm := range []*model.Perm{
		{UserID: user.ID, Repo: repo1, Push: true, Admin: false},
	} {
		assert.NoError(t, store.PermUpsert(perm))
	}

	pipeline1 := &model.Pipeline{
		RepoID: repo1.ID,
		Status: model.StatusFailure,
	}

	assert.NoError(t, store.CreatePipeline(pipeline1))
	feed, err := store.UserFeed(user)
	assert.NoError(t, err)
	assert.Len(t, feed, 1)
}

func TestRepoListLatest(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo), new(model.User), new(model.Perm), new(model.Pipeline), new(model.Org))
	defer closer()

	user := &model.User{
		Login:       "joe",
		Email:       "foo@bar.com",
		AccessToken: "e42080dddf012c718e476da161d21ad5",
	}
	assert.NoError(t, store.CreateUser(user))

	repo1 := &model.Repo{
		ID:            1,
		Owner:         "bradrydzewski",
		Name:          "test",
		FullName:      "bradrydzewski/test",
		ForgeRemoteID: "1",
		IsActive:      true,
	}
	repo2 := &model.Repo{
		ID:            2,
		Owner:         "test",
		Name:          "test",
		FullName:      "test/test",
		ForgeRemoteID: "2",
		IsActive:      true,
	}
	repo3 := &model.Repo{
		ID:            3,
		Owner:         "octocat",
		Name:          "hello-world",
		FullName:      "octocat/hello-world",
		ForgeRemoteID: "3",
		IsActive:      true,
	}
	assert.NoError(t, store.CreateRepo(repo1))
	assert.NoError(t, store.CreateRepo(repo2))
	assert.NoError(t, store.CreateRepo(repo3))

	for _, perm := range []*model.Perm{
		{UserID: user.ID, Repo: repo1, Push: true, Admin: false},
		{UserID: user.ID, Repo: repo2, Push: true, Admin: true},
	} {
		assert.NoError(t, store.PermUpsert(perm))
	}

	pipeline1 := &model.Pipeline{
		RepoID: repo1.ID,
		Status: model.StatusFailure,
	}
	pipeline2 := &model.Pipeline{
		RepoID: repo1.ID,
		Status: model.StatusRunning,
	}
	pipeline3 := &model.Pipeline{
		RepoID: repo2.ID,
		Status: model.StatusKilled,
	}
	pipeline4 := &model.Pipeline{
		RepoID: repo3.ID,
		Status: model.StatusError,
	}
	assert.NoError(t, store.CreatePipeline(pipeline1))
	assert.NoError(t, store.CreatePipeline(pipeline2))
	assert.NoError(t, store.CreatePipeline(pipeline3))
	assert.NoError(t, store.CreatePipeline(pipeline4))

	pipelines, err := store.RepoListLatest(user)
	assert.NoError(t, err)
	assert.Len(t, pipelines, 2)
	assert.EqualValues(t, model.StatusRunning, pipelines[0].Status)
	assert.Equal(t, repo1.ID, pipelines[0].RepoID)
	assert.EqualValues(t, model.StatusKilled, pipelines[1].Status)
	assert.Equal(t, repo2.ID, pipelines[1].RepoID)
}
