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

package gitea

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/franela/goblin"
	"github.com/gin-gonic/gin"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote/gitea/fixtures"
)

func Test_gitea(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := httptest.NewServer(fixtures.Handler())
	c, _ := New(Opts{
		URL:        s.URL,
		SkipVerify: true,
	})

	ctx := context.Background()
	g := goblin.Goblin(t)
	g.Describe("Gitea", func() {
		g.After(func() {
			s.Close()
		})

		g.Describe("Creating a remote", func() {
			g.It("Should return client with specified options", func() {
				remote, _ := New(Opts{
					URL:         "http://localhost:8080",
					Context:     "continuous-integration/test",
					Username:    "someuser",
					Password:    "password",
					SkipVerify:  true,
					PrivateMode: true,
				})
				g.Assert(remote.(*Gitea).URL).Equal("http://localhost:8080")
				g.Assert(remote.(*Gitea).Context).Equal("continuous-integration/test")
				g.Assert(remote.(*Gitea).Machine).Equal("localhost")
				g.Assert(remote.(*Gitea).Username).Equal("someuser")
				g.Assert(remote.(*Gitea).Password).Equal("password")
				g.Assert(remote.(*Gitea).SkipVerify).Equal(true)
				g.Assert(remote.(*Gitea).PrivateMode).Equal(true)
			})
			g.It("Should handle malformed url", func() {
				_, err := New(Opts{URL: "%gh&%ij"})
				g.Assert(err).IsNotNil()
			})
		})

		g.Describe("Generating a netrc file", func() {
			g.It("Should return a netrc with the user token", func() {
				remote, _ := New(Opts{
					URL: "http://gitea.com",
				})
				netrc, _ := remote.Netrc(fakeUser, nil)
				g.Assert(netrc.Machine).Equal("gitea.com")
				g.Assert(netrc.Login).Equal(fakeUser.Login)
				g.Assert(netrc.Password).Equal(fakeUser.Token)
			})
			g.It("Should return a netrc with the machine account", func() {
				remote, _ := New(Opts{
					URL:      "http://gitea.com",
					Username: "someuser",
					Password: "password",
				})
				netrc, _ := remote.Netrc(nil, nil)
				g.Assert(netrc.Machine).Equal("gitea.com")
				g.Assert(netrc.Login).Equal("someuser")
				g.Assert(netrc.Password).Equal("password")
			})
		})

		g.Describe("Requesting a repository", func() {
			g.It("Should return the repository details", func() {
				repo, err := c.Repo(ctx, fakeUser, fakeRepo.Owner, fakeRepo.Name)
				g.Assert(err).IsNil()
				g.Assert(repo.Owner).Equal(fakeRepo.Owner)
				g.Assert(repo.Name).Equal(fakeRepo.Name)
				g.Assert(repo.FullName).Equal(fakeRepo.Owner + "/" + fakeRepo.Name)
				g.Assert(repo.IsSCMPrivate).IsTrue()
				g.Assert(repo.Clone).Equal("http://localhost/test_name/repo_name.git")
				g.Assert(repo.Link).Equal("http://localhost/test_name/repo_name")
			})
			g.It("Should handle a not found error", func() {
				_, err := c.Repo(ctx, fakeUser, fakeRepoNotFound.Owner, fakeRepoNotFound.Name)
				g.Assert(err).IsNotNil()
			})
		})

		g.Describe("Requesting repository permissions", func() {
			g.It("Should return the permission details", func() {
				perm, err := c.Perm(ctx, fakeUser, fakeRepo.Owner, fakeRepo.Name)
				g.Assert(err).IsNil()
				g.Assert(perm.Admin).IsTrue()
				g.Assert(perm.Push).IsTrue()
				g.Assert(perm.Pull).IsTrue()
			})
			g.It("Should handle a not found error", func() {
				_, err := c.Perm(ctx, fakeUser, fakeRepoNotFound.Owner, fakeRepoNotFound.Name)
				g.Assert(err).IsNotNil()
			})
		})

		g.Describe("Requesting a repository list", func() {
			g.It("Should return the repository list", func() {
				repos, err := c.Repos(ctx, fakeUser)
				g.Assert(err).IsNil()
				g.Assert(repos[0].Owner).Equal(fakeRepo.Owner)
				g.Assert(repos[0].Name).Equal(fakeRepo.Name)
				g.Assert(repos[0].FullName).Equal(fakeRepo.Owner + "/" + fakeRepo.Name)
			})
			g.It("Should handle a not found error", func() {
				_, err := c.Repos(ctx, fakeUserNoRepos)
				g.Assert(err).IsNotNil()
			})
		})

		g.It("Should register repository hooks", func() {
			err := c.Activate(ctx, fakeUser, fakeRepo, "http://localhost")
			g.Assert(err).IsNil()
		})

		g.It("Should remove repository hooks", func() {
			err := c.Deactivate(ctx, fakeUser, fakeRepo, "http://localhost")
			g.Assert(err).IsNil()
		})

		g.It("Should return a repository file", func() {
			raw, err := c.File(ctx, fakeUser, fakeRepo, fakeBuild, ".woodpecker.yml")
			g.Assert(err).IsNil()
			g.Assert(string(raw)).Equal("{ platform: linux/amd64 }")
		})

		g.It("Should return nil from send build status", func() {
			err := c.Status(ctx, fakeUser, fakeRepo, fakeBuild, "http://gitea.io", nil)
			g.Assert(err).IsNil()
		})

		g.Describe("Given an authentication request", func() {
			g.It("Should redirect to login form")
			g.It("Should create an access token")
			g.It("Should handle an access token error")
			g.It("Should return the authenticated user")
		})

		g.Describe("Given a repository hook", func() {
			g.It("Should skip non-push events")
			g.It("Should return push details")
			g.It("Should handle a parsing error")
		})
	})
}

var (
	fakeUser = &model.User{
		Login: "someuser",
		Token: "cfcd2084",
	}

	fakeUserNoRepos = &model.User{
		Login: "someuser",
		Token: "repos_not_found",
	}

	fakeRepo = &model.Repo{
		Owner:    "test_name",
		Name:     "repo_name",
		FullName: "test_name/repo_name",
	}

	fakeRepoNotFound = &model.Repo{
		Owner:    "test_name",
		Name:     "repo_not_found",
		FullName: "test_name/repo_not_found",
	}

	fakeBuild = &model.Build{
		Commit: "9ecad50",
	}
)
