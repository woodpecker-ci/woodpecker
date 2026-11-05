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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/gitea/fixtures"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestNew(t *testing.T) {
	forge, _ := New(1, Opts{
		URL:        "http://localhost:8080",
		SkipVerify: true,
	})

	f, _ := forge.(*Gitea)
	assert.Equal(t, "http://localhost:8080", f.url)
	assert.True(t, f.skipVerify)
}

func TestGitea(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := httptest.NewServer(fixtures.Handler())
	defer s.Close()
	c, _ := New(1, Opts{
		URL:        s.URL,
		SkipVerify: true,
	})

	mockStore := store_mocks.NewMockStore(t)
	ctx := store.InjectToContext(t.Context(), mockStore)

	t.Run("netrc with user token", func(t *testing.T) {
		forge, _ := New(1, Opts{})
		netrc, _ := forge.Netrc(fakeUser, fakeRepo)
		assert.Equal(t, "gitea.com", netrc.Machine)
		assert.Equal(t, fakeUser.Login, netrc.Login)
		assert.Equal(t, fakeUser.AccessToken, netrc.Password)
		assert.Equal(t, model.ForgeTypeGitea, netrc.Type)
	})
	t.Run("netrc with machine account", func(t *testing.T) {
		forge, _ := New(1, Opts{})
		netrc, _ := forge.Netrc(nil, fakeRepo)
		assert.Equal(t, "gitea.com", netrc.Machine)
		assert.Empty(t, netrc.Login)
		assert.Empty(t, netrc.Password)
	})

	t.Run("repository details", func(t *testing.T) {
		repo, err := c.Repo(ctx, fakeUser, fakeRepo.ForgeRemoteID, fakeRepo.Owner, fakeRepo.Name)
		assert.NoError(t, err)
		assert.Equal(t, fakeRepo.Owner, repo.Owner)
		assert.Equal(t, fakeRepo.Name, repo.Name)
		assert.Equal(t, fakeRepo.Owner+"/"+fakeRepo.Name, repo.FullName)
		assert.True(t, repo.IsSCMPrivate)
		assert.Equal(t, "http://localhost/test_name/repo_name.git", repo.Clone)
		assert.Equal(t, "http://localhost/test_name/repo_name", repo.ForgeURL)
	})
	t.Run("repo not found", func(t *testing.T) {
		_, err := c.Repo(ctx, fakeUser, "0", fakeRepoNotFound.Owner, fakeRepoNotFound.Name)
		assert.Error(t, err)
	})

	t.Run("repository list", func(t *testing.T) {
		repos, err := c.Repos(ctx, fakeUser, &model.ListOptions{Page: 1, PerPage: 10})
		assert.NoError(t, err)
		assert.Equal(t, fakeRepo.ForgeRemoteID, repos[0].ForgeRemoteID)
		assert.Equal(t, fakeRepo.Owner, repos[0].Owner)
		assert.Equal(t, fakeRepo.Name, repos[0].Name)
		assert.Equal(t, fakeRepo.Owner+"/"+fakeRepo.Name, repos[0].FullName)
	})
	t.Run("not found error", func(t *testing.T) {
		_, err := c.Repos(ctx, fakeUserNoRepos, &model.ListOptions{Page: 1, PerPage: 10})
		assert.Error(t, err)
	})

	t.Run("register repository", func(t *testing.T) {
		err := c.Activate(ctx, fakeUser, fakeRepo, "http://localhost")
		assert.NoError(t, err)
	})

	t.Run("remove hooks", func(t *testing.T) {
		err := c.Deactivate(ctx, fakeUser, fakeRepo, "http://localhost")
		assert.NoError(t, err)
	})

	t.Run("repository file", func(t *testing.T) {
		raw, err := c.File(ctx, fakeUser, fakeRepo, fakePipeline, ".woodpecker.yml")
		assert.NoError(t, err)
		assert.Equal(t, "{ platform: linux/amd64 }", string(raw))
	})

	t.Run("pipeline status", func(t *testing.T) {
		err := c.Status(ctx, fakeUser, fakeRepo, fakePipeline, fakeWorkflow)
		assert.NoError(t, err)
	})

	t.Run("PR hook", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPullRequest)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullRequest)
		mockStore.On("GetRepoNameFallback", mock.Anything, mock.Anything, mock.Anything).Return(fakeRepo, nil)
		mockStore.On("GetUser", mock.Anything).Return(fakeUser, nil)
		r, b, err := c.Hook(ctx, req)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.NoError(t, err)
		assert.Equal(t, model.EventPull, b.Event)
		assert.Equal(t, []string{"README.md"}, b.ChangedFiles)
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

func TestGetStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status model.StatusValue
		want   gitea.StatusState
	}{
		{model.StatusPending, gitea.StatusPending},
		{model.StatusBlocked, gitea.StatusPending},
		{model.StatusCreated, gitea.StatusPending},
		{model.StatusRunning, gitea.StatusPending},
		{model.StatusSuccess, gitea.StatusSuccess},
		{model.StatusFailure, gitea.StatusFailure},
		{model.StatusKilled, gitea.StatusFailure},
		{model.StatusSkipped, gitea.StatusFailure},
		{model.StatusCanceled, gitea.StatusFailure},
		{model.StatusDeclined, gitea.StatusWarning},
		{model.StatusError, gitea.StatusError},
		{model.StatusValue("bogus"), gitea.StatusFailure},
	}

	for _, tt := range tests {
		assert.Equalf(t, tt.want, getStatus(tt.status), "status %q", tt.status)
	}
}

// newTestForge returns a forge talking to a fake Gitea API with the given routes.
func newTestForge(t *testing.T, routes map[string]http.HandlerFunc) *Gitea {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"version":"1.22.0"}`))
	})
	for route, handler := range routes {
		mux.HandleFunc(route, handler)
	}
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	f, err := New(1, Opts{URL: server.URL})
	require.NoError(t, err)
	g, ok := f.(*Gitea)
	require.True(t, ok)
	return g
}

func respond(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func TestHookLoadsTagCommit(t *testing.T) {
	newRequest := func() *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewBufferString(fixtures.HookRelease))
		req.Header.Set(hookEvent, hookRelease)
		return req
	}
	newTagRequest := func() *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewBufferString(fixtures.HookTag))
		req.Header.Set(hookEvent, hookCreated)
		return req
	}

	s := store_mocks.NewMockStore(t)
	s.On("GetRepoNameFallback", mock.Anything, mock.Anything, mock.Anything).Return(&model.Repo{
		ID: 1, UserID: 1, ForgeRemoteID: "77", Owner: "anbraten", Name: "demo",
	}, nil)
	s.On("GetUser", int64(1)).Return(&model.User{ID: 1, AccessToken: "token", Expiry: time.Now().Add(time.Hour).Unix()}, nil)
	ctx := store.InjectToContext(t.Context(), s)

	t.Run("release events get the commit of their tag", func(t *testing.T) {
		g := newTestForge(t, map[string]http.HandlerFunc{
			"/api/v1/repos/anbraten/demo/tags/0.0.5": respond(http.StatusOK, `{"name":"0.0.5","commit":{"sha":"abc123"}}`),
		})

		_, pipeline, err := g.Hook(ctx, newRequest())
		require.NoError(t, err)
		assert.Equal(t, model.EventRelease, pipeline.Event)
		assert.Equal(t, "0.0.5", pipeline.TagTitle)
		assert.Equal(t, "abc123", pipeline.Commit)
	})

	t.Run("tag events get the commit of their tag, not the annotated tag object", func(t *testing.T) {
		// the create hook of a pushed annotated tag holds the SHA of the tag object
		g := newTestForge(t, map[string]http.HandlerFunc{
			"/api/v1/repos/anbraten/demo/tags/v1.0.0": respond(http.StatusOK, `{"name":"v1.0.0","id":"ef98532add3b2feb7a137426bba1248724367df5","commit":{"sha":"abc123"}}`),
		})

		_, pipeline, err := g.Hook(ctx, newTagRequest())
		require.NoError(t, err)
		assert.Equal(t, model.EventTag, pipeline.Event)
		assert.Equal(t, "v1.0.0", pipeline.TagTitle)
		assert.Equal(t, "abc123", pipeline.Commit)
	})

	t.Run("missing tag fails the hook", func(t *testing.T) {
		g := newTestForge(t, nil)

		_, _, err := g.Hook(ctx, newRequest())
		assert.Error(t, err)
	})
}

func TestOrgMembership(t *testing.T) {
	g := newTestForge(t, map[string]http.HandlerFunc{
		"/api/v1/orgs/org/members/owner":            respond(http.StatusNoContent, ""),
		"/api/v1/users/owner/orgs/org/permissions":  respond(http.StatusOK, `{"is_owner":true}`),
		"/api/v1/orgs/org/members/admin":            respond(http.StatusNoContent, ""),
		"/api/v1/users/admin/orgs/org/permissions":  respond(http.StatusOK, `{"is_admin":true}`),
		"/api/v1/orgs/org/members/member":           respond(http.StatusNoContent, ""),
		"/api/v1/users/member/orgs/org/permissions": respond(http.StatusOK, `{"can_read":true}`),
		"/api/v1/orgs/org/members/stranger":         respond(http.StatusNotFound, ""),
		"/api/v1/orgs/org/members/broken":           respond(http.StatusNoContent, ""),
		"/api/v1/users/broken/orgs/org/permissions": respond(http.StatusInternalServerError, `{"message":"boom"}`),
		"/api/v1/orgs/unreachable/members/owner":    respond(http.StatusInternalServerError, ""),
	})

	for _, tc := range []struct {
		login string
		want  *model.OrgPerm
	}{
		{login: "owner", want: &model.OrgPerm{Member: true, Admin: true}},
		{login: "admin", want: &model.OrgPerm{Member: true, Admin: true}},
		{login: "member", want: &model.OrgPerm{Member: true}},
		{login: "stranger", want: &model.OrgPerm{}},
	} {
		t.Run(tc.login, func(t *testing.T) {
			perm, err := g.OrgMembership(t.Context(), &model.User{Login: tc.login, AccessToken: "token"}, "org")
			require.NoError(t, err)
			assert.Equal(t, tc.want, perm)
		})
	}

	t.Run("membership is kept when permissions fail", func(t *testing.T) {
		perm, err := g.OrgMembership(t.Context(), &model.User{Login: "broken", AccessToken: "token"}, "org")
		assert.Error(t, err)
		assert.Equal(t, &model.OrgPerm{Member: true}, perm)
	})

	t.Run("membership check error", func(t *testing.T) {
		_, err := g.OrgMembership(t.Context(), &model.User{Login: "owner", AccessToken: "token"}, "unreachable")
		assert.Error(t, err)
	})
}

func TestOrg(t *testing.T) {
	g := newTestForge(t, map[string]http.HandlerFunc{
		"/api/v1/orgs/public-org":    respond(http.StatusOK, `{"name":"public-org","visibility":"public"}`),
		"/api/v1/orgs/private-org":   respond(http.StatusOK, `{"name":"private-org","visibility":"private"}`),
		"/api/v1/orgs/":              respond(http.StatusNotFound, `{"message":"not found"}`),
		"/api/v1/users/public-user":  respond(http.StatusOK, `{"login":"public-user","visibility":"public"}`),
		"/api/v1/users/limited-user": respond(http.StatusOK, `{"login":"limited-user","visibility":"limited"}`),
		"/api/v1/users/":             respond(http.StatusNotFound, `{"message":"not found"}`),
	})
	user := &model.User{AccessToken: "token"}

	for _, tc := range []struct {
		name string
		want *model.Org
	}{
		{name: "public-org", want: &model.Org{Name: "public-org"}},
		{name: "private-org", want: &model.Org{Name: "private-org", Private: true}},
		{name: "public-user", want: &model.Org{Name: "public-user", IsUser: true}},
		{name: "limited-user", want: &model.Org{Name: "limited-user", IsUser: true, Private: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			org, err := g.Org(t.Context(), user, tc.name)
			require.NoError(t, err)
			assert.Equal(t, tc.want, org)
		})
	}

	t.Run("neither org nor user", func(t *testing.T) {
		_, err := g.Org(t.Context(), user, "missing")
		assert.Error(t, err)
	})
}

func TestTeams(t *testing.T) {
	// a page size of two makes the second, shorter page the last one
	g := newTestForge(t, map[string]http.HandlerFunc{
		"/api/v1/settings/api": respond(http.StatusOK, `{"max_response_items":2}`),
		"/api/v1/user/orgs": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "2", r.URL.Query().Get("limit"))
			switch r.URL.Query().Get("page") {
			case "1":
				_, _ = w.Write([]byte(`[{"name":"one","avatar_url":"/avatars/one"},{"name":"two","avatar_url":"https://avatars.example.com/two"}]`))
			default:
				_, _ = w.Write([]byte(`[{"name":"three"}]`))
			}
		},
	})
	user := &model.User{AccessToken: "token"}

	teams, err := g.Teams(t.Context(), user, &model.ListOptions{Page: 1})
	require.NoError(t, err)
	require.Len(t, teams, 3)
	assert.Equal(t, "one", teams[0].Login)
	assert.Equal(t, g.url+"/avatars/one", teams[0].Avatar)
	assert.Equal(t, "https://avatars.example.com/two", teams[1].Avatar)
	assert.Equal(t, "three", teams[2].Login)

	// all teams come with the first page
	teams, err = g.Teams(t.Context(), user, &model.ListOptions{Page: 2})
	require.NoError(t, err)
	assert.Empty(t, teams)
}

func TestFileAndDir(t *testing.T) {
	g := newTestForge(t, map[string]http.HandlerFunc{
		"/api/v1/repos/owner/repo/contents/.woodpecker": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "abc", r.URL.Query().Get("ref"))
			_, _ = w.Write([]byte(`[` +
				`{"type":"file","path":".woodpecker/a.yaml"},` +
				`{"type":"dir","path":".woodpecker/sub"},` +
				`{"type":"file","path":".woodpecker/b.yaml"}]`))
		},
		"/api/v1/repos/owner/repo/contents/": respond(http.StatusNotFound, `{"message":"not found"}`),
		"/api/v1/repos/owner/repo/raw/": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "abc", r.URL.Query().Get("ref"))
			switch strings.TrimPrefix(r.URL.Path, "/api/v1/repos/owner/repo/raw/") {
			case ".woodpecker/a.yaml":
				_, _ = w.Write([]byte("steps: a"))
			case ".woodpecker/b.yaml":
				_, _ = w.Write([]byte("steps: b"))
			default:
				http.NotFound(w, r)
			}
		},
	})
	user := &model.User{AccessToken: "token"}
	repo := &model.Repo{Owner: "owner", Name: "repo"}
	pipeline := &model.Pipeline{Commit: "abc"}

	_, err := g.File(t.Context(), user, repo, pipeline, ".woodpecker.yaml")
	assert.ErrorIs(t, err, &forge_types.ErrConfigNotFound{})

	// sub directories are skipped, files are loaded with their content
	metas, err := g.Dir(t.Context(), user, repo, pipeline, ".woodpecker")
	require.NoError(t, err)
	assert.Equal(t, []*forge_types.FileMeta{
		{Name: ".woodpecker/a.yaml", Data: []byte("steps: a")},
		{Name: ".woodpecker/b.yaml", Data: []byte("steps: b")},
	}, metas)

	_, err = g.Dir(t.Context(), user, repo, pipeline, ".missing")
	assert.ErrorIs(t, err, &forge_types.ErrConfigNotFound{})
}

func TestBranchHead(t *testing.T) {
	g := newTestForge(t, map[string]http.HandlerFunc{
		"/api/v1/repos/owner/repo/branches/main": respond(http.StatusOK, `{"name":"main","commit":{"id":"abc","url":"https://gitea.example.com/owner/repo/commit/abc"}}`),
	})

	commit, err := g.BranchHead(t.Context(), &model.User{AccessToken: "token"}, &model.Repo{Owner: "owner", Name: "repo"}, "main")
	require.NoError(t, err)
	assert.Equal(t, &model.Commit{SHA: "abc", ForgeURL: "https://gitea.example.com/owner/repo/commit/abc"}, commit)
}

func TestToPermWithoutPermissions(t *testing.T) {
	// repos of other users come without permissions
	assert.Equal(t, &model.Perm{}, toPerm(nil))
}

func TestMatchingHooks(t *testing.T) {
	hooks := []*gitea.Hook{
		{ID: 1, Config: make(map[string]string)},
		{ID: 2, Config: map[string]string{"url": "https://other.example.com/hook"}},
		{ID: 3, Config: map[string]string{"url": "https://ci.example.com/api/hook?access_token=old"}},
	}

	assert.Equal(t, hooks[2], matchingHooks(hooks, "https://ci.example.com/api/hook?access_token=new"))
	assert.Nil(t, matchingHooks(hooks[:2], "https://ci.example.com/api/hook"))
	assert.Nil(t, matchingHooks(hooks, "://invalid"))
}

func TestPerPage(t *testing.T) {
	t.Run("uses the server limit", func(t *testing.T) {
		g := newTestForge(t, map[string]http.HandlerFunc{
			"/api/v1/settings/api": respond(http.StatusOK, `{"max_response_items":42}`),
		})
		assert.Equal(t, 42, g.perPage(t.Context()))
	})

	t.Run("falls back to the default", func(t *testing.T) {
		g := newTestForge(t, map[string]http.HandlerFunc{
			"/api/v1/settings/api": respond(http.StatusInternalServerError, ""),
		})
		assert.Equal(t, defaultPageSize, g.perPage(t.Context()))
	})

	t.Run("keeps a configured size", func(t *testing.T) {
		g := &Gitea{pageSize: 7}
		assert.Equal(t, 7, g.perPage(t.Context()))
	})
}

func TestConfig(t *testing.T) {
	f, err := New(1, Opts{URL: "https://example.com/gitea"})
	require.NoError(t, err)

	serverForge, ok := f.(metadata.ServerForge)
	require.True(t, ok, "gitea must implement the metadata forge interface")

	assert.Equal(t, metadata.Forge{Type: "gitea", URL: "https://example.com/gitea", Host: "example.com"}, serverForge.Config())
}

func TestRefreshExchangesStillValidToken(t *testing.T) {
	var tokenRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login/oauth/access_token" {
			http.NotFound(w, r)
			return
		}
		tokenRequests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"new-access","token_type":"bearer","expires_in":3600,"refresh_token":"new-refresh"}`)
	}))
	defer server.Close()

	forge, err := New(1, Opts{URL: server.URL, OAuthClientID: "id", OAuthClientSecret: "secret"})
	assert.NoError(t, err)

	user := &model.User{
		AccessToken:  "old-access",
		RefreshToken: "old-refresh",
		Expiry:       time.Now().Add(5 * time.Minute).UTC().Unix(),
	}

	f, ok := forge.(*Gitea)
	require.True(t, ok)

	updated, err := f.Refresh(t.Context(), user)
	assert.NoError(t, err)
	assert.True(t, updated)

	assert.Equal(t, int32(1), tokenRequests.Load(), "token endpoint must be called exactly once")
	assert.Equal(t, "new-access", user.AccessToken)
	assert.Equal(t, "new-refresh", user.RefreshToken)
	assert.WithinDuration(t, time.Now().Add(time.Hour), time.Unix(user.Expiry, 0), 2*time.Minute)
}
