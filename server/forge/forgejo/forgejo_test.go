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

package forgejo

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/franela/goblin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/woodpecker-ci/woodpecker/shared/utils"

	"github.com/woodpecker-ci/woodpecker/server/forge/forgejo/fixtures"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
	mocks_store "github.com/woodpecker-ci/woodpecker/server/store/mocks"
)

func Test_forgejo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := httptest.NewServer(fixtures.Handler())
	c, _ := New(Opts{
		URL:        s.URL,
		SkipVerify: true,
	})

	mockStore := mocks_store.NewStore(t)
	ctx := store.InjectToContext(context.Background(), mockStore)

	g := goblin.Goblin(t)
	g.Describe("Forgejo", func() {
		g.After(func() {
			s.Close()
		})

		g.Describe("Creating a forge", func() {
			g.It("Should return client with specified options", func() {
				forge, _ := New(Opts{
					URL:        "http://localhost:8080",
					SkipVerify: true,
				})
				g.Assert(forge.(*Forgejo).URL).Equal("http://localhost:8080")
				g.Assert(forge.(*Forgejo).SkipVerify).Equal(true)
			})
			g.It("Should handle malformed url", func() {
				_, err := New(Opts{URL: "%gh&%ij"})
				g.Assert(err).IsNotNil()
			})
		})

		g.Describe("Generating a netrc file", func() {
			g.It("Should return a netrc with the user token", func() {
				forge, _ := New(Opts{})
				netrc, _ := forge.Netrc(fakeUser, fakeRepo)
				g.Assert(netrc.Machine).Equal("forgejo.com")
				g.Assert(netrc.Login).Equal(fakeUser.Login)
				g.Assert(netrc.Password).Equal(fakeUser.Token)
			})
			g.It("Should return a netrc with the machine account", func() {
				forge, _ := New(Opts{})
				netrc, _ := forge.Netrc(nil, fakeRepo)
				g.Assert(netrc.Machine).Equal("forgejo.com")
				g.Assert(netrc.Login).Equal("")
				g.Assert(netrc.Password).Equal("")
			})
		})

		g.Describe("Requesting a repository", func() {
			g.It("Should return the repository details", func() {
				repo, err := c.Repo(ctx, fakeUser, fakeRepo.ForgeRemoteID, fakeRepo.Owner, fakeRepo.Name)
				g.Assert(err).IsNil()
				g.Assert(repo.Owner).Equal(fakeRepo.Owner)
				g.Assert(repo.Name).Equal(fakeRepo.Name)
				g.Assert(repo.FullName).Equal(fakeRepo.Owner + "/" + fakeRepo.Name)
				g.Assert(repo.IsSCMPrivate).IsTrue()
				g.Assert(repo.Clone).Equal("http://localhost/test_name/repo_name.git")
				g.Assert(repo.Link).Equal("http://localhost/test_name/repo_name")
			})
			g.It("Should handle a not found error", func() {
				_, err := c.Repo(ctx, fakeUser, "0", fakeRepoNotFound.Owner, fakeRepoNotFound.Name)
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

		g.Describe("Requesting a repository list", func() {
			g.It("Should return the repository list", func() {
				repos, err := c.Repos(ctx, fakeUser)
				g.Assert(err).IsNil()
				g.Assert(repos[0].ForgeRemoteID).Equal(fakeRepo.ForgeRemoteID)
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
			fmt.Printf("%v\n", err)
			g.Assert(err).Equal(nil)
			g.Assert(err).IsNil()
		})

		g.It("Should remove repository hooks", func() {
			err := c.Deactivate(ctx, fakeUser, fakeRepo, "http://localhost")
			g.Assert(err).IsNil()
		})

		g.It("Should return a repository file", func() {
			raw, err := c.File(ctx, fakeUser, fakeRepo, fakePipeline, ".woodpecker.yml")
			g.Assert(err).IsNil()
			g.Assert(string(raw)).Equal("{ platform: linux/amd64 }")
		})

		g.It("Should return nil from send pipeline status", func() {
			err := c.Status(ctx, fakeUser, fakeRepo, fakePipeline, fakeStep)
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

		g.It("Given a PR hook", func() {
			buf := bytes.NewBufferString(fixtures.HookPullRequest)
			req, _ := http.NewRequest("POST", "/hook", buf)
			req.Header = http.Header{}
			req.Header.Set(hookEvent, hookPullRequest)
			mockStore.On("GetRepoNameFallback", mock.Anything, mock.Anything).Return(fakeRepo, nil)
			mockStore.On("GetUser", mock.Anything).Return(fakeUser, nil)
			r, b, err := c.Hook(ctx, req)
			g.Assert(r).IsNotNil()
			g.Assert(b).IsNotNil()
			g.Assert(err).IsNil()
			g.Assert(b.Event).Equal(model.EventPull)
			g.Assert(utils.EqualStringSlice(b.ChangedFiles, []string{"README.md"})).IsTrue()
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
		Clone:         "http://forgejo.com/test_name/repo_name.git",
		ForgeRemoteID: "5",
		Owner:         "test_name",
		Name:          "repo_name",
		FullName:      "test_name/repo_name",
	}

	fakeRepoNotFound = &model.Repo{
		Owner:    "test_name",
		Name:     "repo_not_found",
		FullName: "test_name/repo_not_found",
	}

	fakePipeline = &model.Pipeline{
		Commit: "9ecad50",
	}

	fakeStep = &model.Step{
		Name:  "test",
		State: model.StatusSuccess,
	}
)
