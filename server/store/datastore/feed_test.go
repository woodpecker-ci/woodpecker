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

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestRepoListLatest(t *testing.T) {
	store := newTestStore(t, new(model.Repo), new(model.User), new(model.Perm), new(model.Build))
	defer func() {
		store.engine.Exec("delete from repos")
		store.engine.Exec("delete from users")
		store.engine.Exec("delete from perms")
	}()

	user := &model.User{
		Login: "joe",
		Email: "foo@bar.com",
		Token: "e42080dddf012c718e476da161d21ad5",
	}
	store.CreateUser(user)

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
		Owner:    "octocat",
		Name:     "hello-world",
		FullName: "octocat/hello-world",
		IsActive: true,
	}
	store.CreateRepo(repo1)
	store.CreateRepo(repo2)
	store.CreateRepo(repo3)

	store.PermBatch([]*model.Perm{
		{UserID: user.ID, Repo: repo1.FullName, Push: true, Admin: false},
		{UserID: user.ID, Repo: repo2.FullName, Push: true, Admin: true},
	})

	build1 := &model.Build{
		RepoID: repo1.ID,
		Status: model.StatusFailure,
	}
	build2 := &model.Build{
		RepoID: repo1.ID,
		Status: model.StatusRunning,
	}
	build3 := &model.Build{
		RepoID: repo2.ID,
		Status: model.StatusKilled,
	}
	build4 := &model.Build{
		RepoID: repo3.ID,
		Status: model.StatusError,
	}
	store.CreateBuild(build1)
	store.CreateBuild(build2)
	store.CreateBuild(build3)
	store.CreateBuild(build4)

	builds, err := store.RepoListLatest(user)
	if err != nil {
		t.Errorf("Unexpected error: repository list with latest build: %s", err)
		return
	}
	if got, want := len(builds), 2; got != want {
		t.Errorf("Want %d repositories, got %d", want, got)
	}
	if got, want := builds[0].Status, model.StatusRunning; want != got {
		t.Errorf("Want repository status %s, got %s", want, got)
	}
	if got, want := builds[0].FullName, repo1.FullName; want != got {
		t.Errorf("Want repository name %s, got %s", want, got)
	}
	if got, want := builds[1].Status, model.StatusKilled; want != got {
		t.Errorf("Want repository status %s, got %s", want, got)
	}
	if got, want := builds[1].FullName, repo2.FullName; want != got {
		t.Errorf("Want repository name %s, got %s", want, got)
	}
}
