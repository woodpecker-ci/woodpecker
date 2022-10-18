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
	"time"

	"github.com/franela/goblin"
	"github.com/stretchr/testify/assert"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestRepos(t *testing.T) {
	store, closer := newTestStore(t, new(model.Repo), new(model.User), new(model.Pipeline))
	defer closer()

	g := goblin.Goblin(t)
	g.Describe("Repo", func() {
		// before each test be sure to purge the package
		// table data from the database.
		g.BeforeEach(func() {
			_, err := store.engine.Exec("DELETE FROM pipelines")
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
		RemoteID: "1",
	}
	repo2 := &model.Repo{
		Owner:    "test",
		Name:     "test",
		FullName: "test/test",
		RemoteID: "2",
	}
	repo3 := &model.Repo{
		Owner:    "octocat",
		Name:     "hello-world",
		FullName: "octocat/hello-world",
		RemoteID: "3",
	}
	assert.NoError(t, store.CreateRepo(repo1))
	assert.NoError(t, store.CreateRepo(repo2))
	assert.NoError(t, store.CreateRepo(repo3))

	for _, perm := range []*model.Perm{
		{UserID: user.ID, Repo: repo1},
		{UserID: user.ID, Repo: repo2},
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
		RemoteID: "1",
	}
	repo2 := &model.Repo{
		Owner:    "test",
		Name:     "test",
		FullName: "test/test",
		RemoteID: "2",
	}
	repo3 := &model.Repo{
		Owner:    "octocat",
		Name:     "hello-world",
		FullName: "octocat/hello-world",
		RemoteID: "3",
	}
	repo4 := &model.Repo{
		Owner:    "demo",
		Name:     "demo",
		FullName: "demo/demo",
		RemoteID: "4",
	}
	assert.NoError(t, store.CreateRepo(repo1))
	assert.NoError(t, store.CreateRepo(repo2))
	assert.NoError(t, store.CreateRepo(repo3))
	assert.NoError(t, store.CreateRepo(repo4))

	for _, perm := range []*model.Perm{
		{UserID: user.ID, Repo: repo1, Push: true, Admin: false},
		{UserID: user.ID, Repo: repo2, Push: false, Admin: true},
		{UserID: user.ID, Repo: repo3},
		{UserID: user.ID, Repo: repo4},
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
	store, closer := newTestStore(t, new(model.Repo), new(model.User), new(model.Perm), new(model.Redirection))
	defer closer()

	if !assert.NoError(t, store.CreateRepo(&model.Repo{
		RemoteID: "5",
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
			RemoteID: "5",
			UserID:   1,
			FullName: "foo/bar",
			Owner:    "foo",
			Name:     "bar",
			IsActive: true,
			Perm: &model.Perm{
				UserID: 1,
				Pull:   true,
				Push:   true,
				Admin:  true,
				Synced: time.Now().Unix(),
			},
		},
		{
			RemoteID: "6",
			UserID:   1,
			FullName: "bar/baz",
			Owner:    "bar",
			Name:     "baz",
			IsActive: true,
		},
		{
			RemoteID: "7",
			UserID:   1,
			FullName: "baz/qux",
			Owner:    "baz",
			Name:     "qux",
			IsActive: true,
		},
		{
			RemoteID: "8",
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

	// check DB state
	perm, err := store.PermFind(&model.User{ID: 1}, repos[0])
	assert.NoError(t, err)
	assert.True(t, perm.Admin)

	repo := &model.Repo{
		RemoteID: "5",
		FullName: "foo/bar",
		Owner:    "foo",
		Name:     "bar",
		Perm: &model.Perm{
			UserID: 1,
			Pull:   true,
			Push:   true,
			Admin:  false,
			Synced: time.Now().Unix(),
		},
	}
	assert.NoError(t, store.RepoBatch([]*model.Repo{repo}))
	assert.EqualValues(t, repos[0].ID, repo.ID)

	// check current DB state
	_, err = store.engine.ID(repo.ID).Get(repo)
	assert.NoError(t, err)
	assert.True(t, repo.IsActive)
	perm, err = store.PermFind(&model.User{ID: 1}, repos[0])
	assert.NoError(t, err)
	assert.False(t, perm.Admin)

	allRepos := make([]*model.Repo, 0, 4)
	assert.NoError(t, store.engine.Find(&allRepos))
	assert.Len(t, allRepos, 4)

	count, err := store.GetRepoCount()
	assert.NoError(t, err)
	assert.EqualValues(t, 3, count)
}

func TestRepoCrud(t *testing.T) {
	store, closer := newTestStore(t,
		new(model.Repo),
		new(model.User),
		new(model.Perm),
		new(model.Pipeline),
		new(model.PipelineConfig),
		new(model.Logs),
		new(model.Proc),
		new(model.File),
		new(model.Secret),
		new(model.Registry),
		new(model.Config),
		new(model.Redirection))
	defer closer()

	repo := model.Repo{
		UserID:   1,
		FullName: "bradrydzewski/test",
		Owner:    "bradrydzewski",
		Name:     "test",
	}
	assert.NoError(t, store.CreateRepo(&repo))
	pipeline := model.Pipeline{
		RepoID: repo.ID,
	}
	proc := model.Proc{
		Name: "a proc",
	}
	assert.NoError(t, store.CreatePipeline(&pipeline, &proc))

	// create unrelated
	repoUnrelated := model.Repo{
		UserID:   2,
		FullName: "x/x",
		Owner:    "x",
		Name:     "x",
	}
	assert.NoError(t, store.CreateRepo(&repoUnrelated))
	pipelineUnrelated := model.Pipeline{
		RepoID: repoUnrelated.ID,
	}
	procUnrelated := model.Proc{
		Name: "a unrelated proc",
	}
	assert.NoError(t, store.CreatePipeline(&pipelineUnrelated, &procUnrelated))

	_, err := store.GetRepo(repo.ID)
	assert.NoError(t, err)
	assert.NoError(t, store.DeleteRepo(&repo))
	_, err = store.GetRepo(repo.ID)
	assert.Error(t, err)

	procCount, err := store.engine.Count(new(model.Proc))
	assert.NoError(t, err)
	assert.EqualValues(t, 1, procCount)
	pipelineCount, err := store.engine.Count(new(model.Pipeline))
	assert.NoError(t, err)
	assert.EqualValues(t, 1, pipelineCount)
}

func TestRepoRedirection(t *testing.T) {
	store, closer := newTestStore(t,
		new(model.Repo),
		new(model.Redirection))
	defer closer()

	repo := model.Repo{
		UserID:   1,
		RemoteID: "1",
		FullName: "bradrydzewski/test",
		Owner:    "bradrydzewski",
		Name:     "test",
	}
	assert.NoError(t, store.CreateRepo(&repo))

	repoUpdated := model.Repo{
		RemoteID: "1",
		FullName: "bradrydzewski/test-renamed",
		Owner:    "bradrydzewski",
		Name:     "test-renamed",
	}

	assert.NoError(t, store.RepoBatch([]*model.Repo{&repoUpdated}))

	// test redirection from old repo name
	repoFromStore, err := store.GetRepoNameFallback("1", "bradrydzewski/test")
	assert.NoError(t, err)
	assert.Equal(t, repoFromStore.FullName, repoUpdated.FullName)

	// test getting repo without remote ID (use name fallback)
	repo = model.Repo{
		UserID:   1,
		FullName: "bradrydzewski/test-no-remote-id",
		Owner:    "bradrydzewski",
		Name:     "test-no-remote-id",
	}
	assert.NoError(t, store.CreateRepo(&repo))

	repoFromStore, err = store.GetRepoNameFallback("", "bradrydzewski/test-no-remote-id")
	assert.NoError(t, err)
	assert.Equal(t, repoFromStore.FullName, repo.FullName)
}
