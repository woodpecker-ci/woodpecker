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

package github

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/franela/goblin"
	"github.com/gin-gonic/gin"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote/github/fixtures"
)

func Test_github(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := httptest.NewServer(fixtures.Handler())
	c, _ := New(Opts{
		URL:        s.URL,
		SkipVerify: true,
	})

	ctx := context.Background()
	g := goblin.Goblin(t)
	g.Describe("GitHub", func() {
		g.After(func() {
			s.Close()
		})

		g.Describe("Creating a remote", func() {
			g.It("Should return client with specified options", func() {
				remote, _ := New(Opts{
					URL:        "http://localhost:8080/",
					Client:     "0ZXh0IjoiI",
					Secret:     "I1NiIsInR5",
					SkipVerify: true,
				})
				c, _ := remote.(*client)
				g.Assert(c.URL).Equal("http://localhost:8080")
				g.Assert(c.API).Equal("http://localhost:8080/api/v3/")
				g.Assert(c.Client).Equal("0ZXh0IjoiI")
				g.Assert(c.Secret).Equal("I1NiIsInR5")
				g.Assert(c.SkipVerify).Equal(true)
			})
		})

		g.Describe("Generating a netrc file", func() {
			g.It("Should return a netrc with the user token", func() {
				remote, _ := New(Opts{})
				netrc, _ := remote.Netrc(fakeUser, fakeRepo)
				g.Assert(netrc.Machine).Equal("github.com")
				g.Assert(netrc.Login).Equal(fakeUser.Token)
				g.Assert(netrc.Password).Equal("x-oauth-basic")
			})
			g.It("Should return a netrc with the machine account", func() {
				remote, _ := New(Opts{})
				netrc, _ := remote.Netrc(nil, fakeRepo)
				g.Assert(netrc.Machine).Equal("github.com")
				g.Assert(netrc.Login).Equal("")
				g.Assert(netrc.Password).Equal("")
			})
		})

		g.Describe("Requesting a repository", func() {
			g.It("Should return the repository details", func() {
				repo, err := c.Repo(ctx, fakeUser, fakeRepo.Owner, fakeRepo.Name)
				g.Assert(err).IsNil()
				g.Assert(repo.Owner).Equal(fakeRepo.Owner)
				g.Assert(repo.Name).Equal(fakeRepo.Name)
				g.Assert(repo.FullName).Equal(fakeRepo.FullName)
				g.Assert(repo.IsSCMPrivate).IsTrue()
				g.Assert(repo.Clone).Equal(fakeRepo.Clone)
				g.Assert(repo.Link).Equal(fakeRepo.Link)
			})
			g.It("Should handle a not found error", func() {
				_, err := c.Repo(ctx, fakeUser, fakeRepoNotFound.Owner, fakeRepoNotFound.Name)
				g.Assert(err).IsNotNil()
			})
		})

		g.Describe("Requesting repository permissions", func() {
			g.It("Should return the permission details", func() {
				perm, err := c.Perm(ctx, fakeUser, fakeRepo)
				g.Assert(err).IsNil()
				g.Assert(perm.Admin).IsTrue()
				g.Assert(perm.Push).IsTrue()
				g.Assert(perm.Pull).IsTrue()
			})
			g.It("Should handle a not found error", func() {
				_, err := c.Perm(ctx, fakeUser, fakeRepoNotFound)
				g.Assert(err).IsNotNil()
			})
		})

		g.It("Should return a user repository list")

		g.It("Should return a user team list")

		g.It("Should register repositroy hooks")

		g.It("Should return a repository file")

		g.Describe("Given an authentication request", func() {
			g.It("Should redirect to the GitHub login page")
			g.It("Should create an access token")
			g.It("Should handle an access token error")
			g.It("Should return the authenticated user")
			g.It("Should handle authentication errors")
		})
	})
}

var (
	fakeUser = &model.User{
		Login: "octocat",
		Token: "cfcd2084",
	}

	fakeRepo = &model.Repo{
		Owner:        "octocat",
		Name:         "Hello-World",
		FullName:     "octocat/Hello-World",
		Avatar:       "https://github.com/images/error/octocat_happy.gif",
		Link:         "https://github.com/octocat/Hello-World",
		Clone:        "https://github.com/octocat/Hello-World.git",
		IsSCMPrivate: true,
	}

	fakeRepoNotFound = &model.Repo{
		Owner:    "test_name",
		Name:     "repo_not_found",
		FullName: "test_name/repo_not_found",
	}
)
