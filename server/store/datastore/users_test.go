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

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func TestUsers(t *testing.T) {
	store, closer := newTestStore(t, new(model.User), new(model.Repo), new(model.Pipeline), new(model.Step), new(model.Perm), new(model.Org), new(model.Secret))
	defer closer()

	g := goblin.Goblin(t)
	g.Describe("User", func() {
		// before each test be sure to purge the package
		// table data from the database.
		g.BeforeEach(func() {
			_, err := store.engine.Exec("DELETE FROM users")
			g.Assert(err).IsNil()
			_, err = store.engine.Exec("DELETE FROM repos")
			g.Assert(err).IsNil()
			_, err = store.engine.Exec("DELETE FROM pipelines")
			g.Assert(err).IsNil()
			_, err = store.engine.Exec("DELETE FROM steps")
			g.Assert(err).IsNil()
			_, err = store.engine.Exec("DELETE FROM orgs")
			g.Assert(err).IsNil()
		})

		g.It("Should Update a User", func() {
			user := model.User{
				Login:       "joe",
				Email:       "foo@bar.com",
				AccessToken: "e42080dddf012c718e476da161d21ad5",
			}
			err1 := store.CreateUser(&user)
			err2 := store.UpdateUser(&user)
			getUser, err3 := store.GetUser(user.ID)
			g.Assert(err1).IsNil()
			g.Assert(err2).IsNil()
			g.Assert(err3).IsNil()
			g.Assert(user.ID).Equal(getUser.ID)
		})

		g.It("Should Add a new User", func() {
			user := model.User{
				Login:       "joe",
				Email:       "foo@bar.com",
				AccessToken: "e42080dddf012c718e476da161d21ad5",
			}
			err := store.CreateUser(&user)
			g.Assert(err).IsNil()
			g.Assert(user.ID != 0).IsTrue()
		})

		g.It("Should Get a User", func() {
			user := &model.User{
				Login:        "joe",
				AccessToken:  "f0b461ca586c27872b43a0685cbc2847",
				RefreshToken: "976f22a5eef7caacb7e678d6c52f49b1",
				Email:        "foo@bar.com",
				Avatar:       "b9015b0857e16ac4d94a0ffd9a0b79c8",
			}

			g.Assert(store.CreateUser(user)).IsNil()
			getUser, err := store.GetUser(user.ID)
			g.Assert(err).IsNil()
			g.Assert(user.ID).Equal(getUser.ID)
			g.Assert(user.Login).Equal(getUser.Login)
			g.Assert(user.AccessToken).Equal(getUser.AccessToken)
			g.Assert(user.RefreshToken).Equal(getUser.RefreshToken)
			g.Assert(user.Email).Equal(getUser.Email)
			g.Assert(user.Avatar).Equal(getUser.Avatar)
		})

		g.It("Should Get a User By Login", func() {
			user := &model.User{
				Login:       "joe",
				Email:       "foo@bar.com",
				AccessToken: "e42080dddf012c718e476da161d21ad5",
			}
			g.Assert(store.CreateUser(user))
			getUser, err := store.GetUserLogin(user.Login)
			g.Assert(err).IsNil()
			g.Assert(user.ID).Equal(getUser.ID)
			g.Assert(user.Login).Equal(getUser.Login)
		})

		g.It("Should Enforce Unique User Login", func() {
			user1 := model.User{
				Login:       "joe",
				Email:       "foo@bar.com",
				AccessToken: "e42080dddf012c718e476da161d21ad5",
			}
			user2 := model.User{
				Login:       "joe",
				Email:       "foo@bar.com",
				AccessToken: "ab20g0ddaf012c744e136da16aa21ad9",
			}
			err1 := store.CreateUser(&user1)
			err2 := store.CreateUser(&user2)
			g.Assert(err1).IsNil()
			g.Assert(err2 == nil).IsFalse()
		})

		g.It("Should Get a User List", func() {
			user1 := model.User{
				Login:       "jane",
				Email:       "foo@bar.com",
				AccessToken: "ab20g0ddaf012c744e136da16aa21ad9",
				Hash:        "A",
			}
			user2 := model.User{
				Login:       "joe",
				Email:       "foo@bar.com",
				AccessToken: "e42080dddf012c718e476da161d21ad5",
			}
			g.Assert(store.CreateUser(&user1)).IsNil()
			g.Assert(store.CreateUser(&user2)).IsNil()
			users, err := store.GetUserList(&model.ListOptions{Page: 1, PerPage: 50})
			g.Assert(err).IsNil()
			g.Assert(len(users)).Equal(2)
			g.Assert(users[0].Login).Equal(user1.Login)
			g.Assert(users[0].Email).Equal(user1.Email)
			g.Assert(users[0].AccessToken).Equal(user1.AccessToken)
		})

		g.It("Should Get a User Count", func() {
			user1 := model.User{
				Login:       "jane",
				Email:       "foo@bar.com",
				AccessToken: "ab20g0ddaf012c744e136da16aa21ad9",
				Hash:        "A",
			}
			user2 := model.User{
				Login:       "joe",
				Email:       "foo@bar.com",
				AccessToken: "e42080dddf012c718e476da161d21ad5",
				Hash:        "B",
			}
			g.Assert(store.CreateUser(&user1)).IsNil()
			g.Assert(store.CreateUser(&user2)).IsNil()
			count, err := store.GetUserCount()
			g.Assert(err).IsNil()
			g.Assert(count).Equal(int64(2))
		})

		g.It("Should Get a User Count Zero", func() {
			count, err := store.GetUserCount()
			g.Assert(err).IsNil()
			g.Assert(count).Equal(int64(0))
		})

		g.It("Should Del a User", func() {
			user := &model.User{
				Login:       "joe",
				Email:       "foo@bar.com",
				AccessToken: "e42080dddf012c718e476da161d21ad5",
			}
			g.Assert(store.CreateUser(user)).IsNil()
			user, err1 := store.GetUser(user.ID)
			g.Assert(err1).IsNil()
			err2 := store.DeleteUser(user)
			g.Assert(err2).IsNil()
			_, err3 := store.GetUser(user.ID)
			g.Assert(err3).IsNotNil()
		})

		g.It("Should get the Pipeline feed for a User", func() {
			user := &model.User{
				Login:       "joe",
				Email:       "foo@bar.com",
				AccessToken: "e42080dddf012c718e476da161d21ad5",
			}
			g.Assert(store.CreateUser(user)).IsNil()

			repo1 := &model.Repo{
				Owner:         "bradrydzewski",
				Name:          "test",
				FullName:      "bradrydzewski/test",
				IsActive:      true,
				ForgeRemoteID: "1",
			}
			repo2 := &model.Repo{
				Owner:         "test",
				Name:          "test",
				FullName:      "test/test",
				IsActive:      true,
				ForgeRemoteID: "2",
			}
			repo3 := &model.Repo{
				Owner:         "octocat",
				Name:          "hello-world",
				FullName:      "octocat/hello-world",
				IsActive:      true,
				ForgeRemoteID: "3",
			}
			g.Assert(store.CreateRepo(repo1)).IsNil()
			g.Assert(store.CreateRepo(repo2)).IsNil()
			g.Assert(store.CreateRepo(repo3)).IsNil()

			for _, perm := range []*model.Perm{
				{UserID: user.ID, Repo: repo1, Push: true, Admin: false},
				{UserID: user.ID, Repo: repo2, Push: false, Admin: true},
			} {
				g.Assert(store.PermUpsert(perm)).IsNil()
			}

			pipeline1 := &model.Pipeline{
				RepoID: repo1.ID,
				Status: model.StatusFailure,
			}
			pipeline2 := &model.Pipeline{
				RepoID: repo1.ID,
				Status: model.StatusSuccess,
			}
			pipeline3 := &model.Pipeline{
				RepoID: repo2.ID,
				Status: model.StatusSuccess,
			}
			pipeline4 := &model.Pipeline{
				RepoID: repo3.ID,
				Status: model.StatusSuccess,
			}
			g.Assert(store.CreatePipeline(pipeline1)).IsNil()
			g.Assert(store.CreatePipeline(pipeline2)).IsNil()
			g.Assert(store.CreatePipeline(pipeline3)).IsNil()
			g.Assert(store.CreatePipeline(pipeline4)).IsNil()

			pipelines, err := store.UserFeed(user)
			g.Assert(err).IsNil()
			g.Assert(len(pipelines)).Equal(3)
			g.Assert(pipelines[0].RepoID).Equal(repo2.ID)
			g.Assert(pipelines[1].RepoID).Equal(repo1.ID)
			g.Assert(pipelines[2].RepoID).Equal(repo1.ID)
		})
	})
}
