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

package gitea

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/franela/goblin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge/gitea/fixtures"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	mocks_store "go.woodpecker-ci.org/woodpecker/v2/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v2/shared/utils"
)

func Test_gitea(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := httptest.NewServer(fixtures.Handler())
	c, _ := New(Opts{
		URL:        s.URL,
		SkipVerify: true,
	})

	mockStore := mocks_store.NewStore(t)
	ctx := store.InjectToContext(context.Background(), mockStore)

	g := goblin.Goblin(t)
	g.Describe("Gitea", func() {
		g.After(func() {
			s.Close()
		})

		g.Describe("Creating a forge", func() {
			g.It("Should return client with specified options", func() {
				forge, _ := New(Opts{
					URL:        "http://localhost:8080",
					SkipVerify: true,
				})

				f, _ := forge.(*Gitea)
				g.Assert(f.url).Equal("http://localhost:8080")
				g.Assert(f.SkipVerify).Equal(true)
			})
		})

		g.Describe("Generating a netrc file", func() {
			g.It("Should return a netrc with the user token", func() {
				forge, _ := New(Opts{})
				netrc, _ := forge.Netrc(fakeUser, fakeRepo)
				g.Assert(netrc.Machine).Equal("gitea.com")
				g.Assert(netrc.Login).Equal(fakeUser.Login)
				g.Assert(netrc.Password).Equal(fakeUser.AccessToken)
			})
			g.It("Should return a netrc with the machine account", func() {
				forge, _ := New(Opts{})
				netrc, _ := forge.Netrc(nil, fakeRepo)
				g.Assert(netrc.Machine).Equal("gitea.com")
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
				g.Assert(repo.ForgeURL).Equal("http://localhost/test_name/repo_name")
			})
			g.It("Should handle a not found error", func() {
				_, err := c.Repo(ctx, fakeUser, "0", fakeRepoNotFound.Owner, fakeRepoNotFound.Name)
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
			err := c.Status(ctx, fakeUser, fakeRepo, fakePipeline, fakeWorkflow)
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
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = http.Header{}
			req.Header.Set(hookEvent, hookPullRequest)
			mockStore.On("GetRepoNameFallback", mock.Anything, mock.Anything).Return(fakeRepo, nil)
			mockStore.On("GetUser", mock.Anything).Return(fakeUser, nil)
			r, b, err := c.Hook(ctx, req)
			g.Assert(r).IsNotNil()
			g.Assert(b).IsNotNil()
			g.Assert(err).IsNil()
			g.Assert(b.Event).Equal(model.EventPull)
			g.Assert(utils.EqualSliceValues(b.ChangedFiles, []string{"README.md"})).IsTrue()
		})
	})
}

var (
	fakeUser = &model.User{
		Login:       "someuser",
		AccessToken: "cfcd2084",
	}

	fakeUserNoRepos = &model.User{
		Login:       "someuser",
		AccessToken: "repos_not_found",
	}

	fakeRepo = &model.Repo{
		Clone:         "http://gitea.com/test_name/repo_name.git",
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

	fakeWorkflow = &model.Workflow{
		Name:  "test",
		State: model.StatusSuccess,
	}
)
