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
	"fmt"
	"testing"

	"github.com/franela/goblin"
	"github.com/stretchr/testify/assert"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestBuilds(t *testing.T) {
	repo := &model.Repo{
		UserID:   1,
		FullName: "bradrydzewski/test",
		Owner:    "bradrydzewski",
		Name:     "test",
	}

	store, closer := newTestStore(t, new(model.Repo), new(model.Proc), new(model.Build))
	defer closer()

	g := goblin.Goblin(t)
	g.Describe("Builds", func() {
		g.Before(func() {
			_, err := store.engine.Exec("DELETE FROM repos")
			g.Assert(err).IsNil()
			g.Assert(store.CreateRepo(repo)).IsNil()
		})
		g.After(func() {
			_, err := store.engine.Exec("DELETE FROM repos")
			g.Assert(err).IsNil()
		})

		// before each test be sure to purge the package
		// table data from the database.
		g.BeforeEach(func() {
			_, err := store.engine.Exec("DELETE FROM builds")
			g.Assert(err).IsNil()
			_, err = store.engine.Exec("DELETE FROM procs")
			g.Assert(err).IsNil()
		})

		g.It("Should Fail early when the repo is not existing", func() {
			build := model.Build{
				RepoID: 100,
				Status: model.StatusSuccess,
			}
			err := store.CreateBuild(&build)
			g.Assert(err).IsNotNil()

			count, err := store.GetBuildCount()
			g.Assert(err).IsNil()
			g.Assert(count == 0).IsTrue()
			fmt.Println("GOT COUNT", count)

		})

		g.It("Should Post a Build", func() {
			build := model.Build{
				RepoID: repo.ID,
				Status: model.StatusSuccess,
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
			}
			err := store.CreateBuild(&build)
			g.Assert(err).IsNil()
			g.Assert(build.ID != 0).IsTrue()
			g.Assert(build.Number).Equal(int64(1))
			g.Assert(build.Commit).Equal("85f8c029b902ed9400bc600bac301a0aadb144ac")

			count, err := store.GetBuildCount()
			g.Assert(err).IsNil()
			g.Assert(count > 0).IsTrue()
			fmt.Println("GOT COUNT", count)
		})

		g.It("Should Put a Build", func() {
			build := model.Build{
				RepoID: repo.ID,
				Number: 5,
				Status: model.StatusSuccess,
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
			}
			err := store.CreateBuild(&build)
			g.Assert(err).IsNil()
			build.Status = model.StatusRunning
			err1 := store.UpdateBuild(&build)
			getbuild, err2 := store.GetBuild(build.ID)
			g.Assert(err1).IsNil()
			g.Assert(err2).IsNil()
			g.Assert(build.ID).Equal(getbuild.ID)
			g.Assert(build.RepoID).Equal(getbuild.RepoID)
			g.Assert(build.Status).Equal(getbuild.Status)
			g.Assert(build.Number).Equal(getbuild.Number)
		})

		g.It("Should Get a Build", func() {
			build := model.Build{
				RepoID: repo.ID,
				Status: model.StatusSuccess,
			}
			err := store.CreateBuild(&build, []*model.Proc{}...)
			g.Assert(err).IsNil()
			getbuild, err := store.GetBuild(build.ID)
			g.Assert(err).IsNil()
			g.Assert(build.ID).Equal(getbuild.ID)
			g.Assert(build.RepoID).Equal(getbuild.RepoID)
			g.Assert(build.Status).Equal(getbuild.Status)
		})

		g.It("Should Get a Build by Number", func() {
			build1 := &model.Build{
				RepoID: repo.ID,
				Status: model.StatusPending,
			}
			build2 := &model.Build{
				RepoID: repo.ID,
				Status: model.StatusPending,
			}
			err1 := store.CreateBuild(build1, []*model.Proc{}...)
			g.Assert(err1).IsNil()
			err2 := store.CreateBuild(build2, []*model.Proc{}...)
			g.Assert(err2).IsNil()
			getbuild, err3 := store.GetBuildNumber(&model.Repo{ID: 1}, build2.Number)
			g.Assert(err3).IsNil()
			g.Assert(build2.ID).Equal(getbuild.ID)
			g.Assert(build2.RepoID).Equal(getbuild.RepoID)
			g.Assert(build2.Number).Equal(getbuild.Number)
		})

		g.It("Should Get a Build by Ref", func() {
			build1 := &model.Build{
				RepoID: repo.ID,
				Status: model.StatusPending,
				Ref:    "refs/pull/5",
			}
			build2 := &model.Build{
				RepoID: repo.ID,
				Status: model.StatusPending,
				Ref:    "refs/pull/6",
			}
			err1 := store.CreateBuild(build1, []*model.Proc{}...)
			g.Assert(err1).IsNil()
			err2 := store.CreateBuild(build2, []*model.Proc{}...)
			g.Assert(err2).IsNil()
			getbuild, err3 := store.GetBuildRef(&model.Repo{ID: 1}, "refs/pull/6")
			g.Assert(err3).IsNil()
			g.Assert(build2.ID).Equal(getbuild.ID)
			g.Assert(build2.RepoID).Equal(getbuild.RepoID)
			g.Assert(build2.Number).Equal(getbuild.Number)
			g.Assert(build2.Ref).Equal(getbuild.Ref)
		})

		g.It("Should Get a Build by Ref", func() {
			build1 := &model.Build{
				RepoID: repo.ID,
				Status: model.StatusPending,
				Ref:    "refs/pull/5",
			}
			build2 := &model.Build{
				RepoID: repo.ID,
				Status: model.StatusPending,
				Ref:    "refs/pull/6",
			}
			err1 := store.CreateBuild(build1, []*model.Proc{}...)
			g.Assert(err1).IsNil()
			err2 := store.CreateBuild(build2, []*model.Proc{}...)
			g.Assert(err2).IsNil()
			getbuild, err3 := store.GetBuildRef(&model.Repo{ID: 1}, "refs/pull/6")
			g.Assert(err3).IsNil()
			g.Assert(build2.ID).Equal(getbuild.ID)
			g.Assert(build2.RepoID).Equal(getbuild.RepoID)
			g.Assert(build2.Number).Equal(getbuild.Number)
			g.Assert(build2.Ref).Equal(getbuild.Ref)
		})

		g.It("Should Get a Build by Commit", func() {
			build1 := &model.Build{
				RepoID: repo.ID,
				Status: model.StatusPending,
				Branch: "master",
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
			}
			build2 := &model.Build{
				RepoID: repo.ID,
				Status: model.StatusPending,
				Branch: "dev",
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144aa",
			}
			err1 := store.CreateBuild(build1, []*model.Proc{}...)
			g.Assert(err1).IsNil()
			err2 := store.CreateBuild(build2, []*model.Proc{}...)
			g.Assert(err2).IsNil()
			getbuild, err3 := store.GetBuildCommit(&model.Repo{ID: 1}, build2.Commit, build2.Branch)
			g.Assert(err3).IsNil()
			g.Assert(build2.ID).Equal(getbuild.ID)
			g.Assert(build2.RepoID).Equal(getbuild.RepoID)
			g.Assert(build2.Number).Equal(getbuild.Number)
			g.Assert(build2.Commit).Equal(getbuild.Commit)
			g.Assert(build2.Branch).Equal(getbuild.Branch)
		})

		g.It("Should Get the last Build", func() {
			build1 := &model.Build{
				RepoID: repo.ID,
				Status: model.StatusFailure,
				Branch: "master",
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
				Event:  model.EventPush,
			}
			build2 := &model.Build{
				RepoID: repo.ID,
				Status: model.StatusSuccess,
				Branch: "master",
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144aa",
				Event:  model.EventPush,
			}
			err1 := store.CreateBuild(build1, []*model.Proc{}...)
			err2 := store.CreateBuild(build2, []*model.Proc{}...)
			getbuild, err3 := store.GetBuildLast(&model.Repo{ID: 1}, build2.Branch)
			g.Assert(err1).IsNil()
			g.Assert(err2).IsNil()
			g.Assert(err3).IsNil()
			g.Assert(build2.ID).Equal(getbuild.ID)
			g.Assert(build2.RepoID).Equal(getbuild.RepoID)
			g.Assert(build2.Number).Equal(getbuild.Number)
			g.Assert(build2.Status).Equal(getbuild.Status)
			g.Assert(build2.Branch).Equal(getbuild.Branch)
			g.Assert(build2.Commit).Equal(getbuild.Commit)
		})

		g.It("Should Get the last Build Before Build N", func() {
			build1 := &model.Build{
				RepoID: repo.ID,
				Status: model.StatusFailure,
				Branch: "master",
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
			}
			build2 := &model.Build{
				RepoID: repo.ID,
				Status: model.StatusSuccess,
				Branch: "master",
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144aa",
			}
			build3 := &model.Build{
				RepoID: repo.ID,
				Status: model.StatusRunning,
				Branch: "master",
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144aa",
			}
			err1 := store.CreateBuild(build1, []*model.Proc{}...)
			g.Assert(err1).IsNil()
			err2 := store.CreateBuild(build2, []*model.Proc{}...)
			g.Assert(err2).IsNil()
			err3 := store.CreateBuild(build3, []*model.Proc{}...)
			g.Assert(err3).IsNil()
			getbuild, err4 := store.GetBuildLastBefore(&model.Repo{ID: 1}, build3.Branch, build3.ID)
			g.Assert(err4).IsNil()
			g.Assert(build2.ID).Equal(getbuild.ID)
			g.Assert(build2.RepoID).Equal(getbuild.RepoID)
			g.Assert(build2.Number).Equal(getbuild.Number)
			g.Assert(build2.Status).Equal(getbuild.Status)
			g.Assert(build2.Branch).Equal(getbuild.Branch)
			g.Assert(build2.Commit).Equal(getbuild.Commit)
		})

		g.It("Should get recent Builds", func() {
			build1 := &model.Build{
				RepoID: repo.ID,
				Status: model.StatusFailure,
			}
			build2 := &model.Build{
				RepoID: repo.ID,
				Status: model.StatusSuccess,
			}
			err1 := store.CreateBuild(build1, []*model.Proc{}...)
			g.Assert(err1).IsNil()
			err2 := store.CreateBuild(build2, []*model.Proc{}...)
			g.Assert(err2).IsNil()
			builds, err3 := store.GetBuildList(&model.Repo{ID: 1}, 1)
			g.Assert(err3).IsNil()
			g.Assert(len(builds)).Equal(2)
			g.Assert(builds[0].ID).Equal(build2.ID)
			g.Assert(builds[0].RepoID).Equal(build2.RepoID)
			g.Assert(builds[0].Status).Equal(build2.Status)
		})
	})
}

func TestBuildIncrement(t *testing.T) {
	store, closer := newTestStore(t, new(model.Build))
	defer closer()

	buildA := &model.Build{RepoID: 1}
	if !assert.NoError(t, store.CreateBuild(buildA)) {
		return
	}
	assert.EqualValues(t, 1, buildA.Number)

	buildB := &model.Build{RepoID: 1}
	assert.NoError(t, store.CreateBuild(buildB))
	assert.EqualValues(t, 2, buildB.Number)

	buildC := &model.Build{RepoID: 2}
	assert.NoError(t, store.CreateBuild(buildC))
	assert.EqualValues(t, 1, buildC.Number)
}
