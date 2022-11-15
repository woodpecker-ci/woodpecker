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

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestGetPipelineQueue(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo), new(model.User), new(model.Perm), new(model.Pipeline))
	defer closer()

	user := &model.User{
		Login: "joe",
		Email: "foo@bar.com",
		Token: "e42080dddf012c718e476da161d21ad5",
	}
	assert.NoError(t, store.CreateUser(user))

	repo1 := &model.Repo{
		Owner:    "bradrydzewski",
		Name:     "test",
		FullName: "bradrydzewski/test",
		ForgeID:  "1",
		IsActive: true,
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
	if err != nil {
		t.Errorf("Unexpected error: repository list with latest pipeline: %s", err)
		return
	}
	if got, want := len(feed), 1; got != want {
		t.Errorf("Want %d repositories, got %d", want, got)
	}
}

func TestUserFeed(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo), new(model.User), new(model.Perm), new(model.Pipeline))
	defer closer()

	user := &model.User{
		Login: "joe",
		Email: "foo@bar.com",
		Token: "e42080dddf012c718e476da161d21ad5",
	}
	assert.NoError(t, store.CreateUser(user))

	repo1 := &model.Repo{
		Owner:    "bradrydzewski",
		Name:     "test1",
		FullName: "bradrydzewski/test1",
		ForgeID:  "1",
		IsActive: true,
	}
	repo2 := &model.Repo{
		Owner:    "johndoe",
		Name:     "test",
		FullName: "johndoe/test2",
		ForgeID:  "2",
		IsActive: true,
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
	if err != nil {
		t.Errorf("Unexpected error: repository list with latest pipeline: %s", err)
		return
	}
	if got, want := len(feed), 1; got != want {
		t.Errorf("Want %d repositories, got %d", want, got)
	}
}

func TestRepoListLatest(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo), new(model.User), new(model.Perm), new(model.Pipeline))
	defer closer()

	user := &model.User{
		Login: "joe",
		Email: "foo@bar.com",
		Token: "e42080dddf012c718e476da161d21ad5",
	}
	assert.NoError(t, store.CreateUser(user))

	repo1 := &model.Repo{
		Owner:    "bradrydzewski",
		Name:     "test",
		FullName: "bradrydzewski/test",
		ForgeID:  "1",
		IsActive: true,
	}
	repo2 := &model.Repo{
		Owner:    "test",
		Name:     "test",
		FullName: "test/test",
		ForgeID:  "2",
		IsActive: true,
	}
	repo3 := &model.Repo{
		Owner:    "octocat",
		Name:     "hello-world",
		FullName: "octocat/hello-world",
		ForgeID:  "3",
		IsActive: true,
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
	if err != nil {
		t.Errorf("Unexpected error: repository list with latest pipeline: %s", err)
		return
	}
	if got, want := len(pipelines), 2; got != want {
		t.Errorf("Want %d repositories, got %d", want, got)
	}
	if got, want := pipelines[0].Status, string(model.StatusRunning); want != got {
		t.Errorf("Want repository status %s, got %s", want, got)
	}
	if got, want := pipelines[0].FullName, repo1.FullName; want != got {
		t.Errorf("Want repository name %s, got %s", want, got)
	}
	if got, want := pipelines[1].Status, string(model.StatusKilled); want != got {
		t.Errorf("Want repository status %s, got %s", want, got)
	}
	if got, want := pipelines[1].FullName, repo2.FullName; want != got {
		t.Errorf("Want repository name %s, got %s", want, got)
	}
}
