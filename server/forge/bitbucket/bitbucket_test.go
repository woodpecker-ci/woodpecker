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

package bitbucket

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/franela/goblin"
	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge/bitbucket/fixtures"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/bitbucket/internal"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func Test_bitbucket(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := httptest.NewServer(fixtures.Handler())
	c := &config{url: s.URL, API: s.URL}

	g := goblin.Goblin(t)
	ctx := context.Background()
	g.Describe("Bitbucket client", func() {
		g.After(func() {
			s.Close()
		})

		g.It("Should return client with default endpoint", func() {
			forge, _ := New(&Opts{Client: "4vyW6b49Z", Secret: "a5012f6c6"})

			f, _ := forge.(*config)
			g.Assert(f.url).Equal(DefaultURL)
			g.Assert(f.API).Equal(DefaultAPI)
			g.Assert(f.Client).Equal("4vyW6b49Z")
			g.Assert(f.Secret).Equal("a5012f6c6")
		})

		g.It("Should return the netrc file", func() {
			forge, _ := New(&Opts{})
			netrc, _ := forge.Netrc(fakeUser, fakeRepo)
			g.Assert(netrc.Machine).Equal("bitbucket.org")
			g.Assert(netrc.Login).Equal("x-token-auth")
			g.Assert(netrc.Password).Equal(fakeUser.AccessToken)
		})

		g.Describe("Given an authorization request", func() {
			g.It("Should redirect to authorize", func() {
				user, _, err := c.Login(ctx, &types.OAuthRequest{})
				g.Assert(err).IsNil()
				g.Assert(user).IsNil()
			})
			g.It("Should return authenticated user", func() {
				u, _, err := c.Login(ctx, &types.OAuthRequest{
					Code: "code",
				})
				g.Assert(err).IsNil()
				g.Assert(u.Login).Equal(fakeUser.Login)
				g.Assert(u.AccessToken).Equal("2YotnFZFEjr1zCsicMWpAA")
				g.Assert(u.RefreshToken).Equal("tGzv3JOkF0XG5Qx2TlKWIA")
			})
			g.It("Should handle failure to exchange code", func() {
				_, _, err := c.Login(ctx, &types.OAuthRequest{
					Code: "code_bad_request",
				})
				g.Assert(err).IsNotNil()
			})
			g.It("Should handle failure to resolve user", func() {
				_, _, err := c.Login(ctx, &types.OAuthRequest{
					Code: "code_user_not_found",
				})
				g.Assert(err).IsNotNil()
			})
		})

		g.Describe("Given an access token", func() {
			g.It("Should return the authenticated user", func() {
				login, err := c.Auth(ctx, fakeUser.AccessToken, fakeUser.RefreshToken)
				g.Assert(err).IsNil()
				g.Assert(login).Equal(fakeUser.Login)
			})
			g.It("Should handle a failure to resolve user", func() {
				_, err := c.Auth(ctx, fakeUserNotFound.AccessToken, fakeUserNotFound.RefreshToken)
				g.Assert(err).IsNotNil()
			})
		})

		g.Describe("Given a refresh token", func() {
			g.It("Should return a refresh access token", func() {
				ok, err := c.Refresh(ctx, fakeUserRefresh)
				g.Assert(err).IsNil()
				g.Assert(ok).IsTrue()
				g.Assert(fakeUserRefresh.AccessToken).Equal("2YotnFZFEjr1zCsicMWpAA")
				g.Assert(fakeUserRefresh.RefreshToken).Equal("tGzv3JOkF0XG5Qx2TlKWIA")
			})
			g.It("Should handle an empty access token", func() {
				ok, err := c.Refresh(ctx, fakeUserRefreshEmpty)
				g.Assert(err).IsNotNil()
				g.Assert(ok).IsFalse()
			})
			g.It("Should handle a failure to refresh", func() {
				ok, err := c.Refresh(ctx, fakeUserRefreshFail)
				g.Assert(err).IsNotNil()
				g.Assert(ok).IsFalse()
			})
		})

		g.Describe("When requesting a repository", func() {
			g.It("Should return the details", func() {
				repo, err := c.Repo(ctx, fakeUser, "", fakeRepo.Owner, fakeRepo.Name)
				g.Assert(err).IsNil()
				g.Assert(repo.FullName).Equal(fakeRepo.FullName)
			})
			g.It("Should handle not found errors", func() {
				_, err := c.Repo(ctx, fakeUser, "", fakeRepoNotFound.Owner, fakeRepoNotFound.Name)
				g.Assert(err).IsNotNil()
			})
		})

		g.Describe("When requesting user repositories", func() {
			g.It("Should return the details", func() {
				repos, err := c.Repos(ctx, fakeUser)
				g.Assert(err).IsNil()
				g.Assert(repos[0].FullName).Equal(fakeRepo.FullName)
			})
			g.It("Should handle organization not found errors", func() {
				_, err := c.Repos(ctx, fakeUserNoTeams)
				g.Assert(err).IsNotNil()
			})
			g.It("Should handle not found errors", func() {
				_, err := c.Repos(ctx, fakeUserNoRepos)
				g.Assert(err).IsNotNil()
			})
		})

		g.Describe("When requesting user teams", func() {
			g.It("Should return the details", func() {
				teams, err := c.Teams(ctx, fakeUser)
				g.Assert(err).IsNil()
				g.Assert(teams[0].Login).Equal("ueberdev42")
				g.Assert(teams[0].Avatar).Equal("https://bitbucket.org/workspaces/ueberdev42/avatar/?ts=1658761964")
			})
			g.It("Should handle not found error", func() {
				_, err := c.Teams(ctx, fakeUserNoTeams)
				g.Assert(err).IsNotNil()
			})
		})

		g.Describe("When downloading a file", func() {
			g.It("Should return the bytes", func() {
				raw, err := c.File(ctx, fakeUser, fakeRepo, fakePipeline, "file")
				g.Assert(err).IsNil()
				g.Assert(len(raw) != 0).IsTrue()
			})
			g.It("Should handle not found error", func() {
				_, err := c.File(ctx, fakeUser, fakeRepo, fakePipeline, "file_not_found")
				g.Assert(err).IsNotNil()
				g.Assert(errors.Is(err, &types.ErrConfigNotFound{})).IsTrue()
			})
		})

		g.Describe("When requesting repo branch HEAD", func() {
			g.It("Should return the details", func() {
				branchHead, err := c.BranchHead(ctx, fakeUser, fakeRepo, "branch_name")
				g.Assert(err).IsNil()
				g.Assert(branchHead.SHA).Equal("branch_head_name")
				g.Assert(branchHead.ForgeURL).Equal("https://bitbucket.org/commitlink")
			})
			g.It("Should handle not found errors", func() {
				_, err := c.BranchHead(ctx, fakeUser, fakeRepo, "branch_not_found")
				g.Assert(err).IsNotNil()
			})
		})

		g.Describe("When requesting repo pull requests", func() {
			listOpts := model.ListOptions{
				All:     false,
				Page:    1,
				PerPage: 10,
			}
			g.It("Should return the details", func() {
				repoPRs, err := c.PullRequests(ctx, fakeUser, fakeRepo, &listOpts)
				g.Assert(err).IsNil()
				g.Assert(repoPRs[0].Title).Equal("PRs title")
				g.Assert(repoPRs[0].Index).Equal(model.ForgeRemoteID("123"))
			})
			g.It("Should handle not found errors", func() {
				_, err := c.PullRequests(ctx, fakeUser, fakeRepoNotFound, &listOpts)
				g.Assert(err).IsNotNil()
			})
		})

		g.Describe("When requesting repo directory contents", func() {
			g.It("Should return the details", func() {
				files, err := c.Dir(ctx, fakeUser, fakeRepo, fakePipeline, "dir")
				g.Assert(err).IsNil()
				g.Assert(len(files)).Equal(3)
				g.Assert(files[0].Name).Equal("README.md")
				g.Assert(string(files[0].Data)).Equal("dummy payload")
			})
			g.It("Should handle not found errors", func() {
				_, err := c.Dir(ctx, fakeUser, fakeRepo, fakePipeline, "dir_not_found")
				g.Assert(err).IsNotNil()
				g.Assert(errors.Is(err, &types.ErrConfigNotFound{})).IsTrue()
			})
		})

		g.Describe("When activating a repository", func() {
			g.It("Should error when malformed hook", func() {
				err := c.Activate(ctx, fakeUser, fakeRepo, "%gh&%ij")
				g.Assert(err).IsNotNil()
			})
			g.It("Should create the hook", func() {
				err := c.Activate(ctx, fakeUser, fakeRepo, "http://127.0.0.1")
				g.Assert(err).IsNil()
			})
		})

		g.Describe("When deactivating a repository", func() {
			g.It("Should error when listing hooks fails", func() {
				err := c.Deactivate(ctx, fakeUser, fakeRepoNoHooks, "http://127.0.0.1")
				g.Assert(err).IsNotNil()
			})
			g.It("Should successfully remove hooks", func() {
				err := c.Deactivate(ctx, fakeUser, fakeRepo, "http://127.0.0.1")
				g.Assert(err).IsNil()
			})
			g.It("Should successfully deactivate when hook already removed", func() {
				err := c.Deactivate(ctx, fakeUser, fakeRepoEmptyHook, "http://127.0.0.1")
				g.Assert(err).IsNil()
			})
		})

		g.Describe("Given a list of hooks", func() {
			g.It("Should return the matching hook", func() {
				hooks := []*internal.Hook{
					{URL: "http://127.0.0.1/hook"},
				}
				hook := matchingHooks(hooks, "http://127.0.0.1/")
				g.Assert(hook).Equal(hooks[0])
			})
			g.It("Should handle no matches", func() {
				hooks := []*internal.Hook{
					{URL: "http://localhost/hook"},
				}
				hook := matchingHooks(hooks, "http://127.0.0.1/")
				g.Assert(hook).IsNil()
			})
			g.It("Should handle malformed hook urls", func() {
				var hooks []*internal.Hook
				hook := matchingHooks(hooks, "%gh&%ij")
				g.Assert(hook).IsNil()
			})
		})

		g.It("Should update the status", func() {
			err := c.Status(ctx, fakeUser, fakeRepo, fakePipeline, fakeWorkflow)
			g.Assert(err).IsNil()
		})

		g.It("Should parse the hook", func() {
			buf := bytes.NewBufferString(fixtures.HookPush)
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = http.Header{}
			req.Header.Set(hookEvent, hookPush)

			r, b, err := c.Hook(ctx, req)
			g.Assert(err).IsNil()
			g.Assert(r.FullName).Equal("martinherren1984/publictestrepo")
			g.Assert(b.Commit).Equal("c14c1bb05dfb1fdcdf06b31485fff61b0ea44277")
		})
	})
}

var (
	fakeUser = &model.User{
		Login:       "superman",
		AccessToken: "cfcd2084",
	}

	fakeUserRefresh = &model.User{
		Login:        "superman",
		RefreshToken: "cfcd2084",
	}

	fakeUserRefreshFail = &model.User{
		Login:        "superman",
		RefreshToken: "refresh_token_not_found",
	}

	fakeUserRefreshEmpty = &model.User{
		Login:        "superman",
		RefreshToken: "refresh_token_is_empty",
	}

	fakeUserNotFound = &model.User{
		Login:       "superman",
		AccessToken: "user_not_found",
	}

	fakeUserNoTeams = &model.User{
		Login:       "superman",
		AccessToken: "teams_not_found",
	}

	fakeUserNoRepos = &model.User{
		Login:       "superman",
		AccessToken: "repos_not_found",
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

	fakeRepoNoHooks = &model.Repo{
		Owner:    "test_name",
		Name:     "hooks_not_found",
		FullName: "test_name/hooks_not_found",
	}

	fakeRepoEmptyHook = &model.Repo{
		Owner:    "test_name",
		Name:     "hook_empty",
		FullName: "test_name/hook_empty",
	}

	fakePipeline = &model.Pipeline{
		Commit: "9ecad50",
	}

	fakeWorkflow = &model.Workflow{
		Name:  "test",
		State: model.StatusSuccess,
	}
)
