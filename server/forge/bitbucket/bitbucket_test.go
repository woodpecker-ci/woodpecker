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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucket/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucket/internal"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestNew(t *testing.T) {
	forge, _ := New(&Opts{Client: "4vyW6b49Z", Secret: "a5012f6c6"})

	f, _ := forge.(*config)
	assert.Equal(t, DefaultURL, f.url)
	assert.Equal(t, DefaultAPI, f.API)
	assert.Equal(t, "4vyW6b49Z", f.Client)
	assert.Equal(t, "a5012f6c6", f.Secret)
}

func TestBitbucket(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := httptest.NewServer(fixtures.Handler())
	defer s.Close()
	c := &config{url: s.URL, API: s.URL}

	ctx := t.Context()

	forge, _ := New(&Opts{})
	netrc, _ := forge.Netrc(fakeUser, fakeRepo)
	assert.Equal(t, "bitbucket.org", netrc.Machine)
	assert.Equal(t, "x-token-auth", netrc.Login)
	assert.Equal(t, fakeUser.AccessToken, netrc.Password)
	assert.Equal(t, model.ForgeTypeBitbucket, netrc.Type)

	user, _, err := c.Login(ctx, &types.OAuthRequest{})
	assert.NoError(t, err)
	assert.Nil(t, user)

	u, _, err := c.Login(ctx, &types.OAuthRequest{
		Code: "code",
	})
	assert.NoError(t, err)
	assert.Equal(t, fakeUser.Login, u.Login)
	assert.Equal(t, "2YotnFZFEjr1zCsicMWpAA", u.AccessToken)
	assert.Equal(t, "tGzv3JOkF0XG5Qx2TlKWIA", u.RefreshToken)

	_, _, err = c.Login(ctx, &types.OAuthRequest{
		Code: "code_bad_request",
	})
	assert.Error(t, err)

	_, _, err = c.Login(ctx, &types.OAuthRequest{
		Code: "code_user_not_found",
	})
	assert.Error(t, err)

	login, err := c.Auth(ctx, fakeUser.AccessToken, fakeUser.RefreshToken)
	assert.NoError(t, err)
	assert.Equal(t, fakeUser.Login, login)

	_, err = c.Auth(ctx, fakeUserNotFound.AccessToken, fakeUserNotFound.RefreshToken)
	assert.Error(t, err)

	ok, err := c.Refresh(ctx, fakeUserRefresh)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "2YotnFZFEjr1zCsicMWpAA", fakeUserRefresh.AccessToken)
	assert.Equal(t, "tGzv3JOkF0XG5Qx2TlKWIA", fakeUserRefresh.RefreshToken)

	ok, err = c.Refresh(ctx, fakeUserRefreshEmpty)
	assert.Error(t, err)
	assert.False(t, ok)

	ok, err = c.Refresh(ctx, fakeUserRefreshFail)
	assert.Error(t, err)
	assert.False(t, ok)

	repo, err := c.Repo(ctx, fakeUser, "", fakeRepo.Owner, fakeRepo.Name)
	assert.NoError(t, err)
	assert.Equal(t, fakeRepo.FullName, repo.FullName)

	_, err = c.Repo(ctx, fakeUser, "", fakeRepoNotFound.Owner, fakeRepoNotFound.Name)
	assert.Error(t, err)

	repos, err := c.Repos(ctx, fakeUser)
	assert.NoError(t, err)
	assert.Equal(t, fakeRepo.FullName, repos[0].FullName)

	_, err = c.Repos(ctx, fakeUserNoTeams)
	assert.Error(t, err)

	_, err = c.Repos(ctx, fakeUserNoRepos)
	assert.Error(t, err)

	teams, err := c.Teams(ctx, fakeUser)
	assert.NoError(t, err)
	assert.Equal(t, "ueberdev42", teams[0].Login)
	assert.Equal(t, "https://bitbucket.org/workspaces/ueberdev42/avatar/?ts=1658761964", teams[0].Avatar)

	_, err = c.Teams(ctx, fakeUserNoTeams)
	assert.Error(t, err)

	raw, err := c.File(ctx, fakeUser, fakeRepo, fakePipeline, "file")
	assert.NoError(t, err)
	assert.True(t, len(raw) != 0)

	_, err = c.File(ctx, fakeUser, fakeRepo, fakePipeline, "file_not_found")
	assert.Error(t, err)
	assert.ErrorIs(t, err, &types.ErrConfigNotFound{})

	branchHead, err := c.BranchHead(ctx, fakeUser, fakeRepo, "branch_name")
	assert.NoError(t, err)
	assert.Equal(t, "branch_head_name", branchHead.SHA)
	assert.Equal(t, "https://bitbucket.org/commitlink", branchHead.ForgeURL)

	_, err = c.BranchHead(ctx, fakeUser, fakeRepo, "branch_not_found")
	assert.Error(t, err)

	listOpts := model.ListOptions{
		All:     false,
		Page:    1,
		PerPage: 10,
	}

	repoPRs, err := c.PullRequests(ctx, fakeUser, fakeRepo, &listOpts)
	assert.NoError(t, err)
	assert.Equal(t, "PRs title", repoPRs[0].Title)
	assert.Equal(t, model.ForgeRemoteID("123"), repoPRs[0].Index)

	_, err = c.PullRequests(ctx, fakeUser, fakeRepoNotFound, &listOpts)
	assert.Error(t, err)

	files, err := c.Dir(ctx, fakeUser, fakeRepo, fakePipeline, "dir")
	assert.NoError(t, err)
	assert.Len(t, files, 3)
	assert.Equal(t, "README.md", files[0].Name)
	assert.Equal(t, "dummy payload", string(files[0].Data))

	_, err = c.Dir(ctx, fakeUser, fakeRepo, fakePipeline, "dir_not_found")
	assert.Error(t, err)
	assert.ErrorIs(t, err, &types.ErrConfigNotFound{})

	err = c.Activate(ctx, fakeUser, fakeRepo, "%gh&%ij")
	assert.Error(t, err)

	err = c.Activate(ctx, fakeUser, fakeRepo, "http://127.0.0.1")
	assert.NoError(t, err)

	err = c.Deactivate(ctx, fakeUser, fakeRepoNoHooks, "http://127.0.0.1")
	assert.Error(t, err)

	err = c.Deactivate(ctx, fakeUser, fakeRepo, "http://127.0.0.1")
	assert.NoError(t, err)

	err = c.Deactivate(ctx, fakeUser, fakeRepoEmptyHook, "http://127.0.0.1")
	assert.NoError(t, err)

	hooks := []*internal.Hook{
		{URL: "http://127.0.0.1/hook"},
	}
	hook := matchingHooks(hooks, "http://127.0.0.1/")
	assert.Equal(t, hooks[0], hook)

	hooks = []*internal.Hook{
		{URL: "http://localhost/hook"},
	}
	hook = matchingHooks(hooks, "http://127.0.0.1/")
	assert.Nil(t, hook)

	hooks = nil
	hook = matchingHooks(hooks, "%gh&%ij")
	assert.Nil(t, hook)

	err = c.Status(ctx, fakeUser, fakeRepo, fakePipeline, fakeWorkflow)
	assert.NoError(t, err)

	buf := bytes.NewBufferString(fixtures.HookPush)
	req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
	req.Header = http.Header{}
	req.Header.Set(hookEvent, hookPush)

	r, b, err := c.Hook(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, "martinherren1984/publictestrepo", r.FullName)
	assert.Equal(t, "c14c1bb05dfb1fdcdf06b31485fff61b0ea44277", b.Commit)
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
