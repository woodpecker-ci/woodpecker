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

		g.It("Should Get a Repo by Name (case-insensitive)", func() {
			repo := model.Repo{
				UserID:   1,
				FullName: "bradrydzewski/TEST",
				Owner:    "bradrydzewski",
				Name:     "TEST",
			}
			g.Assert(store.CreateRepo(&repo)).IsNil()
			getrepo, err := store.GetRepoName("Bradrydzewski/test")
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
		Owner:         "bradrydzewski",
		Name:          "test",
		FullName:      "bradrydzewski/test",
		ForgeRemoteID: "1",
	}
	repo2 := &model.Repo{
		Owner:         "test",
		Name:          "test",
		FullName:      "test/test",
		ForgeRemoteID: "2",
	}
	repo3 := &model.Repo{
		Owner:         "octocat",
		Name:          "hello-world",
		FullName:      "octocat/hello-world",
		ForgeRemoteID: "3",
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

	repos, err := store.RepoList(user, false, false)
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
		Owner:         "bradrydzewski",
		Name:          "test",
		FullName:      "bradrydzewski/test",
		ForgeRemoteID: "1",
	}
	repo2 := &model.Repo{
		Owner:         "test",
		Name:          "test",
		FullName:      "test/test",
		ForgeRemoteID: "2",
	}
	repo3 := &model.Repo{
		Owner:         "octocat",
		Name:          "hello-world",
		FullName:      "octocat/hello-world",
		ForgeRemoteID: "3",
	}
	repo4 := &model.Repo{
		Owner:         "demo",
		Name:          "demo",
		FullName:      "demo/demo",
		ForgeRemoteID: "4",
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

	repos, err := store.RepoList(user, true, false)
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

func TestRepoCrud(t *testing.T) {
	store, closer := newTestStore(t,
		new(model.Repo),
		new(model.User),
		new(model.Perm),
		new(model.Pipeline),
		new(model.PipelineConfig),
		new(model.LogEntry),
		new(model.Step),
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
	step := model.Step{
		Name: "a step",
	}
	assert.NoError(t, store.CreatePipeline(&pipeline, &step))

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
	stepUnrelated := model.Step{
		Name: "a unrelated step",
	}
	assert.NoError(t, store.CreatePipeline(&pipelineUnrelated, &stepUnrelated))

	_, err := store.GetRepo(repo.ID)
	assert.NoError(t, err)
	assert.NoError(t, store.DeleteRepo(&repo))
	_, err = store.GetRepo(repo.ID)
	assert.Error(t, err)

	stepCount, err := store.engine.Count(new(model.Step))
	assert.NoError(t, err)
	assert.EqualValues(t, 1, stepCount)
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
		UserID:        1,
		ForgeRemoteID: "1",
		FullName:      "bradrydzewski/test",
		Owner:         "bradrydzewski",
		Name:          "test",
	}
	assert.NoError(t, store.CreateRepo(&repo))

	repoUpdated := model.Repo{
		ID:            repo.ID,
		ForgeRemoteID: "1",
		FullName:      "bradrydzewski/test-renamed",
		Owner:         "bradrydzewski",
		Name:          "test-renamed",
	}

	assert.NoError(t, store.UpdateRepo(&repoUpdated))
	assert.NoError(t, store.CreateRedirection(&model.Redirection{
		RepoID:   repo.ID,
		FullName: repo.FullName,
	}))

	// test redirection from old repo name
	repoFromStore, err := store.GetRepoNameFallback("1", "bradrydzewski/test")
	assert.NoError(t, err)
	assert.Equal(t, repoFromStore.FullName, repoUpdated.FullName)

	// test getting repo without forge ID (use name fallback)
	repo = model.Repo{
		UserID:   1,
		FullName: "bradrydzewski/test-no-forge-id",
		Owner:    "bradrydzewski",
		Name:     "test-no-forge-id",
	}
	assert.NoError(t, store.CreateRepo(&repo))

	repoFromStore, err = store.GetRepoNameFallback("", "bradrydzewski/test-no-forge-id")
	assert.NoError(t, err)
	assert.Equal(t, repoFromStore.FullName, repo.FullName)
}
