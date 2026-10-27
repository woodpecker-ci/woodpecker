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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucket/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucket/internal"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestNew(t *testing.T) {
	forge, _ := New(1, &Opts{OAuthClientID: "4vyW6b49Z", OAuthClientSecret: "a5012f6c6"})

	f, _ := forge.(*config)
	assert.Equal(t, DefaultURL, f.url)
	assert.Equal(t, DefaultAPI, f.api)
	assert.Equal(t, "4vyW6b49Z", f.oAuthClientID)
	assert.Equal(t, "a5012f6c6", f.oAuthSecret)
}

func TestBitbucket(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := httptest.NewServer(fixtures.Handler())
	defer s.Close()
	c := &config{url: s.URL, api: s.URL}

	ctx := t.Context()

	forge, _ := New(1, &Opts{})
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

	repos, err := c.Repos(ctx, fakeUser, &model.ListOptions{Page: 1, PerPage: 10})
	assert.NoError(t, err)
	assert.Len(t, repos, 1)
	assert.Equal(t, fakeRepo.FullName, repos[0].FullName)

	_, err = c.Repos(ctx, fakeUserNoTeams, &model.ListOptions{Page: 1, PerPage: 10})
	assert.Error(t, err)

	_, err = c.Repos(ctx, fakeUserNoRepos, &model.ListOptions{Page: 1, PerPage: 10})
	assert.Error(t, err)

	teams, err := c.Teams(ctx, fakeUser, &model.ListOptions{Page: 1, PerPage: 10})
	assert.NoError(t, err)
	assert.Equal(t, "test_name", teams[0].Login)
	assert.Equal(t, "https://bitbucket.org/workspaces/ueberdev42/avatar/?ts=1658761964", teams[0].Avatar)

	_, err = c.Teams(ctx, fakeUserNoTeams, &model.ListOptions{Page: 1, PerPage: 10})
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
		Page:    1,
		PerPage: 10,
	}

	repoPRs, err := c.PullRequests(ctx, fakeUser, fakeRepo, &listOpts)
	assert.NoError(t, err)
	assert.Equal(t, "PRs title", repoPRs[0].Title)
	assert.Equal(t, model.ForgeRemoteID("123"), repoPRs[0].Index)

	_, err = c.PullRequests(ctx, fakeUser, fakeRepoNotFound, &listOpts)
	assert.Error(t, err)

	// sub directories are skipped, files are loaded with their content
	files, err := c.Dir(ctx, fakeUser, fakeRepo, fakePipeline, "dir")
	assert.NoError(t, err)
	if assert.Len(t, files, 2) {
		assert.Equal(t, "README.md", files[0].Name)
		assert.Equal(t, "dummy payload", string(files[0].Data))
		assert.Equal(t, ".gitignore", files[1].Name)
	}

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

	err = c.Deactivate(ctx, fakeUser, fakeRepoHooksNoNext, "http://127.0.0.1")
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

	mockStore := store_mocks.NewMockStore(t)
	ctx = store.InjectToContext(ctx, mockStore)
	mockStore.On("GetUser", mock.Anything).Return(fakeUser, nil)
	mockStore.On("GetRepoForgeID", mock.Anything, mock.Anything).Return(fakeRepoFromHook, nil)

	r, b, err := c.Hook(ctx, signHookRequest(t, req), testHookSecret)
	assert.NoError(t, err)
	assert.Equal(t, "martinherren1984/publictestrepo", r.FullName)
	assert.Equal(t, "master", r.Branch)
	assert.Equal(t, "c14c1bb05dfb1fdcdf06b31485fff61b0ea44277", b.Commit)
	assert.Equal(t, []string{"main.go"}, b.ChangedFiles)
}

func TestActivateSetsWebhookSecret(t *testing.T) {
	// field of POST /repositories/{workspace}/{repo_slug}/hooks in
	// https://dac-static.atlassian.com/cloud/bitbucket/swagger.v3.json
	var created map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&created))
			_, _ = w.Write([]byte(`{}`))
			return
		}
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(server.Close)
	c := &config{url: server.URL, api: server.URL}

	repo := &model.Repo{Owner: "owner", Name: "repo", WebhookSecret: "repo-secret"}
	require.NoError(t, c.Activate(t.Context(), fakeUser, repo, "https://ci.example.com/api/hook?access_token=jwt"))
	require.NotNil(t, created)
	assert.Equal(t, "repo-secret", created["secret"])
}

func TestHookVerifiesSignature(t *testing.T) {
	// test vector of https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries,
	// Bitbucket signs the body the same way
	const (
		secret    = "It's a Secret to Everybody"
		body      = "Hello, World!"
		signature = "sha256=757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17"
	)
	c := &config{url: "https://bitbucket.org", api: "https://api.bitbucket.org"}

	for _, tc := range []struct {
		name      string
		secret    string
		body      string
		signature string
		valid     bool
	}{
		{name: "valid", secret: secret, body: body, signature: signature, valid: true},
		{name: "wrong secret", secret: "other secret", body: body, signature: signature},
		{name: "changed body", secret: secret, body: body + "!", signature: signature},
		{name: "missing signature", secret: secret, body: body},
		{name: "unsupported algorithm", secret: secret, body: body, signature: "sha1=757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17"},
		{name: "malformed signature", secret: secret, body: body, signature: "sha256=not hex"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewBufferString(tc.body))
			req.Header.Set(hookEvent, "repo:updated")
			if tc.signature != "" {
				req.Header.Set("X-Hub-Signature", tc.signature)
			}

			_, _, err := c.Hook(t.Context(), req, tc.secret)
			if tc.valid {
				// repo:updated is no pipeline event, so a valid hook ends here
				assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
			} else {
				assert.ErrorIs(t, err, types.ErrInvalidWebhookSignature)
			}
		})
	}
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

	fakeRepoHooksNoNext = &model.Repo{
		Owner:    "test_name",
		Name:     "repo_hooks_no_next",
		FullName: "test_name/repo_hooks_no_next",
	}

	fakeRepoFromHook = &model.Repo{
		Owner:    "martinherren1984",
		Name:     "publictestrepo",
		FullName: "martinherren1984/publictestrepo",
		UserID:   1,
	}

	fakePipeline = &model.Pipeline{
		Commit: "9ecad50",
	}

	fakeWorkflow = &model.Workflow{
		Name:  "test",
		State: model.StatusSuccess,
	}
)

// testHookSecret is the webhook secret signHookRequest signs with.
const testHookSecret = "test-secret"

// signHookRequest signs the body of req with testHookSecret like Bitbucket does.
func signHookRequest(t *testing.T, req *http.Request) *http.Request {
	t.Helper()
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	req.Body = io.NopCloser(bytes.NewReader(body))
	mac := hmac.New(sha256.New, []byte(testHookSecret))
	_, _ = mac.Write(body)
	req.Header.Set("X-Hub-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	return req
}
