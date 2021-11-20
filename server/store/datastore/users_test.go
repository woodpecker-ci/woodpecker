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

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestUsers(t *testing.T) {
	store, closer := newTestStore(t, new(model.User), new(model.Repo), new(model.Build), new(model.Proc), new(model.Perm))
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
			_, err = store.engine.Exec("DELETE FROM builds")
			g.Assert(err).IsNil()
			_, err = store.engine.Exec("DELETE FROM procs")
			g.Assert(err).IsNil()
		})

		g.It("Should Update a User", func() {
			user := model.User{
				Login: "joe",
				Email: "foo@bar.com",
				Token: "e42080dddf012c718e476da161d21ad5",
			}
			err1 := store.CreateUser(&user)
			err2 := store.UpdateUser(&user)
			getuser, err3 := store.GetUser(user.ID)
			g.Assert(err1).IsNil()
			g.Assert(err2).IsNil()
			g.Assert(err3).IsNil()
			g.Assert(user.ID).Equal(getuser.ID)
		})

		g.It("Should Add a new User", func() {
			user := model.User{
				Login: "joe",
				Email: "foo@bar.com",
				Token: "e42080dddf012c718e476da161d21ad5",
			}
			err := store.CreateUser(&user)
			g.Assert(err).IsNil()
			g.Assert(user.ID != 0).IsTrue()
		})

		g.It("Should Get a User", func() {
			user := &model.User{
				Login:  "joe",
				Token:  "f0b461ca586c27872b43a0685cbc2847",
				Secret: "976f22a5eef7caacb7e678d6c52f49b1",
				Email:  "foo@bar.com",
				Avatar: "b9015b0857e16ac4d94a0ffd9a0b79c8",
				Active: true,
			}

			g.Assert(store.CreateUser(user)).IsNil()
			getuser, err := store.GetUser(user.ID)
			g.Assert(err).IsNil()
			g.Assert(user.ID).Equal(getuser.ID)
			g.Assert(user.Login).Equal(getuser.Login)
			g.Assert(user.Token).Equal(getuser.Token)
			g.Assert(user.Secret).Equal(getuser.Secret)
			g.Assert(user.Email).Equal(getuser.Email)
			g.Assert(user.Avatar).Equal(getuser.Avatar)
			g.Assert(user.Active).Equal(getuser.Active)
		})

		g.It("Should Get a User By Login", func() {
			user := &model.User{
				Login: "joe",
				Email: "foo@bar.com",
				Token: "e42080dddf012c718e476da161d21ad5",
			}
			g.Assert(store.CreateUser(user))
			getuser, err := store.GetUserLogin(user.Login)
			g.Assert(err).IsNil()
			g.Assert(user.ID).Equal(getuser.ID)
			g.Assert(user.Login).Equal(getuser.Login)
		})

		g.It("Should Enforce Unique User Login", func() {
			user1 := model.User{
				Login: "joe",
				Email: "foo@bar.com",
				Token: "e42080dddf012c718e476da161d21ad5",
			}
			user2 := model.User{
				Login: "joe",
				Email: "foo@bar.com",
				Token: "ab20g0ddaf012c744e136da16aa21ad9",
			}
			err1 := store.CreateUser(&user1)
			err2 := store.CreateUser(&user2)
			g.Assert(err1).IsNil()
			g.Assert(err2 == nil).IsFalse()
		})

		g.It("Should Get a User List", func() {
			user1 := model.User{
				Login: "jane",
				Email: "foo@bar.com",
				Token: "ab20g0ddaf012c744e136da16aa21ad9",
				Hash:  "A",
			}
			user2 := model.User{
				Login: "joe",
				Email: "foo@bar.com",
				Token: "e42080dddf012c718e476da161d21ad5",
			}
			g.Assert(store.CreateUser(&user1)).IsNil()
			g.Assert(store.CreateUser(&user2)).IsNil()
			users, err := store.GetUserList()
			g.Assert(err).IsNil()
			g.Assert(len(users)).Equal(2)
			g.Assert(users[0].Login).Equal(user1.Login)
			g.Assert(users[0].Email).Equal(user1.Email)
			g.Assert(users[0].Token).Equal(user1.Token)
		})

		g.It("Should Get a User Count", func() {
			user1 := model.User{
				Login: "jane",
				Email: "foo@bar.com",
				Token: "ab20g0ddaf012c744e136da16aa21ad9",
				Hash:  "A",
			}
			user2 := model.User{
				Login: "joe",
				Email: "foo@bar.com",
				Token: "e42080dddf012c718e476da161d21ad5",
				Hash:  "B",
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
				Login: "joe",
				Email: "foo@bar.com",
				Token: "e42080dddf012c718e476da161d21ad5",
			}
			g.Assert(store.CreateUser(user)).IsNil()
			user, err1 := store.GetUser(user.ID)
			g.Assert(err1).IsNil()
			err2 := store.DeleteUser(user)
			g.Assert(err2).IsNil()
			_, err3 := store.GetUser(user.ID)
			g.Assert(err3).IsNotNil()
		})

		g.It("Should get the Build feed for a User", func() {
			user := &model.User{
				Login: "joe",
				Email: "foo@bar.com",
				Token: "e42080dddf012c718e476da161d21ad5",
			}
			g.Assert(store.CreateUser(user)).IsNil()

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
			g.Assert(store.CreateRepo(repo1)).IsNil()
			g.Assert(store.CreateRepo(repo2)).IsNil()
			g.Assert(store.CreateRepo(repo3)).IsNil()

			for _, perm := range []*model.Perm{
				{UserID: user.ID, Repo: repo1.FullName, Push: true, Admin: false},
				{UserID: user.ID, Repo: repo2.FullName, Push: false, Admin: true},
			} {
				g.Assert(store.PermUpsert(perm)).IsNil()
			}

			build1 := &model.Build{
				RepoID: repo1.ID,
				Status: model.StatusFailure,
			}
			build2 := &model.Build{
				RepoID: repo1.ID,
				Status: model.StatusSuccess,
			}
			build3 := &model.Build{
				RepoID: repo2.ID,
				Status: model.StatusSuccess,
			}
			build4 := &model.Build{
				RepoID: repo3.ID,
				Status: model.StatusSuccess,
			}
			g.Assert(store.CreateBuild(build1)).IsNil()
			g.Assert(store.CreateBuild(build2)).IsNil()
			g.Assert(store.CreateBuild(build3)).IsNil()
			g.Assert(store.CreateBuild(build4)).IsNil()

			builds, err := store.UserFeed(user)
			g.Assert(err).IsNil()
			g.Assert(len(builds)).Equal(3)
			g.Assert(builds[0].FullName).Equal(repo2.FullName)
			g.Assert(builds[1].FullName).Equal(repo1.FullName)
			g.Assert(builds[2].FullName).Equal(repo1.FullName)
		})
	})
}
