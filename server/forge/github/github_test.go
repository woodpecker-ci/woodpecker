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

package github

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/franela/goblin"
	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge/github/fixtures"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
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

		g.Describe("Creating a forge", func() {
			g.It("Should return client with specified options", func() {
				forge, _ := New(Opts{
					URL:        "http://localhost:8080/",
					Client:     "0ZXh0IjoiI",
					Secret:     "I1NiIsInR5",
					SkipVerify: true,
				})
				f, _ := forge.(*client)
				g.Assert(f.url).Equal("http://localhost:8080")
				g.Assert(f.API).Equal("http://localhost:8080/api/v3/")
				g.Assert(f.Client).Equal("0ZXh0IjoiI")
				g.Assert(f.Secret).Equal("I1NiIsInR5")
				g.Assert(f.SkipVerify).Equal(true)
			})
		})

		g.Describe("Generating a netrc file", func() {
			g.It("Should return a netrc with the user token", func() {
				forge, _ := New(Opts{})
				netrc, _ := forge.Netrc(fakeUser, fakeRepo)
				g.Assert(netrc.Machine).Equal("github.com")
				g.Assert(netrc.Login).Equal(fakeUser.AccessToken)
				g.Assert(netrc.Password).Equal("x-oauth-basic")
			})
			g.It("Should return a netrc with the machine account", func() {
				forge, _ := New(Opts{})
				netrc, _ := forge.Netrc(nil, fakeRepo)
				g.Assert(netrc.Machine).Equal("github.com")
				g.Assert(netrc.Login).Equal("")
				g.Assert(netrc.Password).Equal("")
			})
		})

		g.Describe("Requesting a repository", func() {
			g.It("Should return the repository details", func() {
				repo, err := c.Repo(ctx, fakeUser, fakeRepo.ForgeRemoteID, fakeRepo.Owner, fakeRepo.Name)
				g.Assert(err).IsNil()
				g.Assert(repo.ForgeRemoteID).Equal(fakeRepo.ForgeRemoteID)
				g.Assert(repo.Owner).Equal(fakeRepo.Owner)
				g.Assert(repo.Name).Equal(fakeRepo.Name)
				g.Assert(repo.FullName).Equal(fakeRepo.FullName)
				g.Assert(repo.IsSCMPrivate).IsTrue()
				g.Assert(repo.Clone).Equal(fakeRepo.Clone)
				g.Assert(repo.ForgeURL).Equal(fakeRepo.ForgeURL)
			})
			g.It("Should handle a not found error", func() {
				_, err := c.Repo(ctx, fakeUser, "0", fakeRepoNotFound.Owner, fakeRepoNotFound.Name)
				g.Assert(err).IsNotNil()
			})
		})

		g.It("Should return a user repository list")

		g.It("Should return a user team list")

		g.It("Should register repository hooks")

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
		Login:       "octocat",
		AccessToken: "cfcd2084",
	}

	fakeRepo = &model.Repo{
		ForgeRemoteID: "5",
		Owner:         "octocat",
		Name:          "Hello-World",
		FullName:      "octocat/Hello-World",
		Avatar:        "https://github.com/images/error/octocat_happy.gif",
		ForgeURL:      "https://github.com/octocat/Hello-World",
		Clone:         "https://github.com/octocat/Hello-World.git",
		IsSCMPrivate:  true,
	}

	fakeRepoNotFound = &model.Repo{
		Owner:    "test_name",
		Name:     "repo_not_found",
		FullName: "test_name/repo_not_found",
	}
)
