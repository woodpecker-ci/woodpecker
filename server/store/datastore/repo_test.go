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

	"github.com/franela/goblin"
	"github.com/stretchr/testify/assert"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestRepos(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo), new(model.User), new(model.Build))
	defer closer()

	g := goblin.Goblin(t)
	g.Describe("Repo", func() {

		// before each test be sure to purge the package
		// table data from the database.
		g.BeforeEach(func() {
			_, err := store.engine.Exec("DELETE FROM builds")
			g.Assert(err).IsNil()
			_, err = store.engine.Exec("DELETE FROM repos")
			g.Assert(err).IsNil()
			_, err = store.engine.Exec("DELETE FROM users")
			g.Assert(err).IsNil()
		})

		g.It("Should Set a Repo", func() {
			repo := model.Repo{
				UserID:   1,
				FullName: "bradrydzewski/test",
				Owner:    "bradrydzewski",
				Name:     "test",
			}
			err1 := store.CreateRepo(&repo)
			err2 := store.UpdateRepo(&repo)
			getRepo, err3 := store.GetRepo(repo.ID)

			g.Assert(err1).IsNil()
			g.Assert(err2).IsNil()
			g.Assert(err3).IsNil()
			g.Assert(repo.ID).Equal(getRepo.ID)
		})

		g.It("Should Add a Repo", func() {
			repo := model.Repo{
				UserID:   1,
				FullName: "bradrydzewski/test",
				Owner:    "bradrydzewski",
				Name:     "test",
			}
			err := store.CreateRepo(&repo)
			g.Assert(err).IsNil()
			g.Assert(repo.ID != 0).IsTrue()
		})

		g.It("Should Get a Repo by ID", func() {
			repo := model.Repo{
				UserID:   1,
				FullName: "bradrydzewski/test",
				Owner:    "bradrydzewski",
				Name:     "test",
			}
			g.Assert(store.CreateRepo(&repo)).IsNil()
			getrepo, err := store.GetRepo(repo.ID)
			g.Assert(err).IsNil()
			g.Assert(repo.ID).Equal(getrepo.ID)
			g.Assert(repo.UserID).Equal(getrepo.UserID)
			g.Assert(repo.Owner).Equal(getrepo.Owner)
			g.Assert(repo.Name).Equal(getrepo.Name)
		})

		g.It("Should Get a Repo by Name", func() {
			repo := model.Repo{
				UserID:   1,
				FullName: "bradrydzewski/test",
				Owner:    "bradrydzewski",
				Name:     "test",
			}
			g.Assert(store.CreateRepo(&repo)).IsNil()
			getrepo, err := store.GetRepoName(repo.FullName)
			g.Assert(err).IsNil()
			g.Assert(repo.ID).Equal(getrepo.ID)
			g.Assert(repo.UserID).Equal(getrepo.UserID)
			g.Assert(repo.Owner).Equal(getrepo.Owner)
			g.Assert(repo.Name).Equal(getrepo.Name)
		})

		g.It("Should Enforce Unique Repo Name", func() {
			repo1 := model.Repo{
				UserID:   1,
				FullName: "bradrydzewski/test",
				Owner:    "bradrydzewski",
				Name:     "test",
			}
			repo2 := model.Repo{
				UserID:   2,
				FullName: "bradrydzewski/test",
				Owner:    "bradrydzewski",
				Name:     "test",
			}
			err1 := store.CreateRepo(&repo1)
			err2 := store.CreateRepo(&repo2)
			g.Assert(err1).IsNil()
			g.Assert(err2 == nil).IsFalse()
		})
	})
}

func TestRepoList(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo), new(model.User), new(model.Perm))
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
	}
	repo2 := &model.Repo{
		Owner:    "test",
		Name:     "test",
		FullName: "test/test",
	}
	repo3 := &model.Repo{
		Owner:    "octocat",
		Name:     "hello-world",
		FullName: "octocat/hello-world",
	}
	assert.NoError(t, store.CreateRepo(repo1))
	assert.NoError(t, store.CreateRepo(repo2))
	assert.NoError(t, store.CreateRepo(repo3))

	for _, perm := range []*model.Perm{
		{UserID: user.ID, Repo: repo1.FullName},
		{UserID: user.ID, Repo: repo2.FullName},
	} {
		assert.NoError(t, store.PermUpsert(perm))
	}

	repos, err := store.RepoList(user, false)
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := len(repos), 2; got != want {
		t.Errorf("Want %d repositories, got %d", want, got)
	}
	if got, want := repos[0].ID, repo1.ID; got != want {
		t.Errorf("Want repository id %d, got %d", want, got)
	}
	if got, want := repos[1].ID, repo2.ID; got != want {
		t.Errorf("Want repository id %d, got %d", want, got)
	}
}

func TestOwnedRepoList(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo), new(model.User), new(model.Perm))
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
	}
	repo2 := &model.Repo{
		Owner:    "test",
		Name:     "test",
		FullName: "test/test",
	}
	repo3 := &model.Repo{
		Owner:    "octocat",
		Name:     "hello-world",
		FullName: "octocat/hello-world",
	}
	repo4 := &model.Repo{
		Owner:    "demo",
		Name:     "demo",
		FullName: "demo/demo",
	}
	assert.NoError(t, store.CreateRepo(repo1))
	assert.NoError(t, store.CreateRepo(repo2))
	assert.NoError(t, store.CreateRepo(repo3))
	assert.NoError(t, store.CreateRepo(repo4))

	for _, perm := range []*model.Perm{
		{UserID: user.ID, Repo: repo1.FullName, Push: true, Admin: false},
		{UserID: user.ID, Repo: repo2.FullName, Push: false, Admin: true},
		{UserID: user.ID, Repo: repo3.FullName},
		{UserID: user.ID, Repo: repo4.FullName},
	} {
		assert.NoError(t, store.PermUpsert(perm))
	}

	repos, err := store.RepoList(user, true)
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := len(repos), 2; got != want {
		t.Errorf("Want %d repositories, got %d", want, got)
	}
	if got, want := repos[0].ID, repo1.ID; got != want {
		t.Errorf("Want repository id %d, got %d", want, got)
	}
	if got, want := repos[1].ID, repo2.ID; got != want {
		t.Errorf("Want repository id %d, got %d", want, got)
	}
}

func TestRepoCount(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo))
	defer closer()

	repo1 := &model.Repo{
		Owner:    "bradrydzewski",
		Name:     "test",
		FullName: "bradrydzewski/test",
		IsActive: true,
	}
	repo2 := &model.Repo{
		Owner:    "test",
		Name:     "test",
		FullName: "test/test",
		IsActive: true,
	}
	repo3 := &model.Repo{
		Owner:    "test",
		Name:     "test-ui",
		FullName: "test/test-ui",
		IsActive: false,
	}
	assert.NoError(t, store.CreateRepo(repo1))
	assert.NoError(t, store.CreateRepo(repo2))
	assert.NoError(t, store.CreateRepo(repo3))

	count, _ := store.GetRepoCount()
	if got, want := count, int64(2); got != want {
		t.Errorf("Want %d repositories, got %d", want, got)
	}
}

func TestRepoBatch(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo), new(model.User), new(model.Perm))
	defer closer()

	if !assert.NoError(t, store.CreateRepo(&model.Repo{
		UserID:   1,
		FullName: "foo/bar",
		Owner:    "foo",
		Name:     "bar",
		IsActive: true,
	})) {
		return
	}

	repos := []*model.Repo{
		{
			UserID:   1,
			FullName: "foo/bar",
			Owner:    "foo",
			Name:     "bar",
			IsActive: true,
		},
		{
			UserID:   1,
			FullName: "bar/baz",
			Owner:    "bar",
			Name:     "baz",
			IsActive: true,
		},
		{
			UserID:   1,
			FullName: "baz/qux",
			Owner:    "baz",
			Name:     "qux",
			IsActive: true,
		},
		{
			UserID:   0, // not activated repos do hot have a user id assigned
			FullName: "baz/notes",
			Owner:    "baz",
			Name:     "notes",
			IsActive: false,
		},
	}
	if !assert.NoError(t, store.RepoBatch(repos)) {
		return
	}

	repo := &model.Repo{
		FullName: "foo/bar",
		Owner:    "foo",
		Name:     "bar",
	}
	assert.NoError(t, store.RepoBatch([]*model.Repo{repo}))
	assert.EqualValues(t, repos[0].ID, repo.ID)
	_, err := store.engine.ID(repo.ID).Get(repo)
	assert.NoError(t, err)
	assert.True(t, repo.IsActive)

	allRepos := make([]*model.Repo, 0, 4)
	assert.NoError(t, store.engine.Find(&allRepos))
	assert.Len(t, allRepos, 4)

	count, err := store.GetRepoCount()
	assert.NoError(t, err)
	assert.EqualValues(t, 3, count)
}

func TestRepoCrud(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo), new(model.User), new(model.Perm))
	defer closer()

	repo := model.Repo{
		UserID:   1,
		FullName: "bradrydzewski/test",
		Owner:    "bradrydzewski",
		Name:     "test",
	}
	assert.NoError(t, store.CreateRepo(&repo))
	_, err1 := store.GetRepo(repo.ID)
	err2 := store.DeleteRepo(&repo)
	_, err3 := store.GetRepo(repo.ID)
	if err1 != nil {
		t.Errorf("Unexpected error: select repository: %s", err1)
	}
	if err2 != nil {
		t.Errorf("Unexpected error: delete repository: %s", err2)
	}
	if err3 == nil {
		t.Errorf("Expected error: sql.ErrNoRows")
	}
}
