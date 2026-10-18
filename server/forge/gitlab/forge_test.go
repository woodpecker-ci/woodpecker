// Copyright 2026 Woodpecker Authors
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

package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/gitlab/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func newTestServer(t *testing.T, routes map[string]string) *GitLab {
	t.Helper()
	mux := http.NewServeMux()
	for route, body := range routes {
		mux.HandleFunc(route, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(body))
		})
	}
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return load(server.URL + "?client_id=test&client_secret=test")
}

// storeContext returns a context with a store that resolves every hook repo
// to project 4 of a user with a valid token.
func storeContext(t *testing.T) context.Context {
	t.Helper()
	s := store_mocks.NewMockStore(t)
	s.On("GetRepoNameFallback", mock.Anything, mock.Anything, mock.Anything).Return(&model.Repo{
		ID:            1,
		UserID:        1,
		ForgeRemoteID: "4",
		Owner:         "diaspora",
		Name:          "diaspora-client",
	}, nil)
	s.On("GetUser", int64(1)).Return(&model.User{
		ID:          1,
		AccessToken: "token",
		Expiry:      time.Now().Add(time.Hour).Unix(),
	}, nil)
	return store.InjectToContext(t.Context(), s)
}

func hookRequest(t *testing.T, body []byte, header http.Header) *http.Request {
	t.Helper()
	req, err := http.NewRequest(fixtures.ServiceHookMethod, fixtures.ServiceHookURL.String(), bytes.NewReader(body))
	require.NoError(t, err)
	req.Header = header
	return req
}

func TestHookLoadsMetadataFromForge(t *testing.T) {
	t.Run("merge request gets changed files and milestone", func(t *testing.T) {
		client := newTestServer(t, map[string]string{
			"/api/v4/projects/4": `{"id":4}`,
			// a renamed file reports both paths, an edited one the same path twice
			"/api/v4/projects/4/merge_requests/3/diffs": `[` +
				`{"old_path":"old.go","new_path":"new.go"},` +
				`{"old_path":"main.go","new_path":"main.go"}]`,
			"/api/v4/projects/4/milestones/6088906": `{"id":6088906,"title":"v1.0"}`,
		})

		_, pipeline, err := client.Hook(storeContext(t), hookRequest(t, fixtures.HookPullRequestMilestoned, fixtures.ServiceHookHeaders))
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"new.go", "old.go", "main.go"}, pipeline.ChangedFiles)
		assert.Equal(t, "v1.0", pipeline.PullRequestMilestone)
	})

	t.Run("tag without message loads the tagged commit", func(t *testing.T) {
		var payload map[string]any
		require.NoError(t, json.Unmarshal(fixtures.HookTag, &payload))
		payload["message"] = ""
		body, err := json.Marshal(payload)
		require.NoError(t, err)
		sha, ok := payload["after"].(string)
		require.True(t, ok)

		client := newTestServer(t, map[string]string{
			"/api/v4/projects/4": `{"id":4}`,
			"/api/v4/projects/4/repository/commits/" + sha: `{"id":"` + sha + `","message":"Release v22",` +
				`"author_name":"Jane","author_email":"jane@example.com","committed_date":"2026-01-02T03:04:05Z"}`,
		})

		_, pipeline, err := client.Hook(storeContext(t), hookRequest(t, body, fixtures.ServiceHookHeaders))
		require.NoError(t, err)
		assert.Equal(t, model.EventTag, pipeline.Event)
		assert.Equal(t, "Release v22", pipeline.Message)
		assert.Equal(t, "Jane", pipeline.Author)
		assert.Equal(t, "jane@example.com", pipeline.Email)
		assert.Equal(t, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC).Unix(), pipeline.Timestamp)
		assert.NotEmpty(t, pipeline.Avatar)
	})

	t.Run("release gets the release author", func(t *testing.T) {
		client := newTestServer(t, map[string]string{
			"/api/v4/projects/4":                `{"id":4}`,
			"/api/v4/projects/4/releases/0.0.2": `{"tag_name":"0.0.2","author":{"username":"releaser","avatar_url":"https://avatar/releaser"}}`,
		})

		_, pipeline, err := client.Hook(storeContext(t), hookRequest(t, fixtures.WebhookReleaseBody, fixtures.ReleaseHookHeaders))
		require.NoError(t, err)
		assert.Equal(t, model.EventRelease, pipeline.Event)
		assert.Equal(t, "releaser", pipeline.Author)
		assert.Equal(t, "https://avatar/releaser", pipeline.Avatar)
	})

	t.Run("forge errors fail the hook", func(t *testing.T) {
		client := newTestServer(t, map[string]string{"/api/v4/projects/4": `{"id":4}`})

		_, _, err := client.Hook(storeContext(t), hookRequest(t, fixtures.HookPullRequestMilestoned, fixtures.ServiceHookHeaders))
		assert.Error(t, err)
		_, _, err = client.Hook(storeContext(t), hookRequest(t, fixtures.WebhookReleaseBody, fixtures.ReleaseHookHeaders))
		assert.Error(t, err)
	})
}

func TestOrgMembership(t *testing.T) {
	user := &model.User{Login: "test_user", AccessToken: "token"}

	// the first page is full, so the member on the second page must be found
	members := func(page int) string {
		if page == 1 {
			list := make([]string, defaultPerPage)
			for i := range list {
				list[i] = fmt.Sprintf(`{"id":%d,"username":"user%d","access_level":30}`, i+100, i)
			}
			return "[" + strings.Join(list, ",") + "]"
		}
		return `[{"id":3,"username":"test_user","access_level":50}]`
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/groups", func(w http.ResponseWriter, _ *http.Request) {
		// the search also matches groups whose path only contains the name
		_, _ = w.Write([]byte(`[{"id":1,"path":"group-other","full_path":"group-other"},{"id":2,"path":"group","full_path":"group"}]`))
	})
	mux.HandleFunc("/api/v4/groups/2/members/all", func(w http.ResponseWriter, r *http.Request) {
		page := 1
		if r.URL.Query().Get("page") == "2" {
			page = 2
		}
		_, _ = w.Write([]byte(members(page)))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	client := load(server.URL + "?client_id=test&client_secret=test")

	t.Run("owner on a later member page", func(t *testing.T) {
		perm, err := client.OrgMembership(t.Context(), user, "group")
		require.NoError(t, err)
		assert.Equal(t, &model.OrgPerm{Member: true, Admin: true}, perm)
	})

	t.Run("not a member", func(t *testing.T) {
		perm, err := client.OrgMembership(t.Context(), &model.User{Login: "stranger", AccessToken: "token"}, "group")
		require.NoError(t, err)
		assert.Equal(t, &model.OrgPerm{}, perm)
	})

	t.Run("no group with that path", func(t *testing.T) {
		perm, err := client.OrgMembership(t.Context(), user, "grou")
		require.NoError(t, err)
		assert.Equal(t, &model.OrgPerm{}, perm)
	})
}

func TestOrg(t *testing.T) {
	groups := map[string]string{
		// display names differ from the full path, only the full path counts
		"parent/sub": `[{"id":1,"path":"sub","name":"sub","full_path":"other/sub","visibility":"public"},` +
			`{"id":2,"path":"sub","name":"sub","full_path":"parent/sub","visibility":"internal"}]`,
		"public": `[{"id":3,"path":"public","full_path":"public","visibility":"public"}]`,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/users", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("username") == "jane" {
			_, _ = w.Write([]byte(`[{"id":5,"username":"jane","private_profile":true}]`))
			return
		}
		_, _ = w.Write([]byte(`[]`))
	})
	mux.HandleFunc("/api/v4/groups", func(w http.ResponseWriter, r *http.Request) {
		body, ok := groups[r.URL.Query().Get("search")]
		if !ok {
			body = `[]`
		}
		_, _ = w.Write([]byte(body))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	client := load(server.URL + "?client_id=test&client_secret=test")
	user := &model.User{AccessToken: "token"}

	org, err := client.Org(t.Context(), user, "jane")
	require.NoError(t, err)
	assert.Equal(t, &model.Org{Name: "jane", IsUser: true, Private: true}, org)

	org, err = client.Org(t.Context(), user, "parent/sub")
	require.NoError(t, err)
	assert.Equal(t, &model.Org{Name: "parent/sub", Private: true}, org)

	org, err = client.Org(t.Context(), user, "public")
	require.NoError(t, err)
	assert.Equal(t, &model.Org{Name: "public", Private: false}, org)

	_, err = client.Org(t.Context(), user, "missing")
	assert.ErrorContains(t, err, "could not find org missing")
}

func TestNetrc(t *testing.T) {
	client := load("https://gitlab.example.com?client_id=test&client_secret=test")
	repo := &model.Repo{Clone: "https://gitlab.example.com/group/repo.git"}

	netrc, err := client.Netrc(&model.User{AccessToken: "token"}, repo)
	require.NoError(t, err)
	assert.Equal(t, &model.Netrc{Machine: "gitlab.example.com", Login: "oauth2", Password: "token", Type: model.ForgeTypeGitlab}, netrc)

	// without a user only the machine is known
	netrc, err = client.Netrc(nil, repo)
	require.NoError(t, err)
	assert.Equal(t, &model.Netrc{Machine: "gitlab.example.com", Type: model.ForgeTypeGitlab}, netrc)
}

func TestFileAndDir(t *testing.T) {
	user := &model.User{AccessToken: "token"}
	repo := &model.Repo{ForgeRemoteID: "4"}
	pipeline := &model.Pipeline{Commit: "abc"}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects/4", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":4}`))
	})
	mux.HandleFunc("/api/v4/projects/4/repository/files/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "abc", r.URL.Query().Get("ref"))
		switch strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v4/projects/4/repository/files/"), "/raw") {
		case ".woodpecker/a.yaml":
			_, _ = w.Write([]byte("steps: a"))
		case ".woodpecker/b.yaml":
			_, _ = w.Write([]byte("steps: b"))
		default:
			http.NotFound(w, r)
		}
	})
	mux.HandleFunc("/api/v4/projects/4/repository/tree", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, ".woodpecker", r.URL.Query().Get("path"))
		_, _ = w.Write([]byte(`[` +
			`{"type":"blob","path":".woodpecker/a.yaml"},` +
			`{"type":"tree","path":".woodpecker/sub"},` +
			`{"type":"blob","path":".woodpecker/b.yaml"}]`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	client := load(server.URL + "?client_id=test&client_secret=test")

	data, err := client.File(t.Context(), user, repo, pipeline, ".woodpecker/a.yaml")
	require.NoError(t, err)
	assert.Equal(t, "steps: a", string(data))

	_, err = client.File(t.Context(), user, repo, pipeline, ".woodpecker.yaml")
	assert.ErrorIs(t, err, &types.ErrConfigNotFound{})

	// sub directories are skipped, files are loaded with their content
	metas, err := client.Dir(t.Context(), user, repo, pipeline, ".woodpecker")
	require.NoError(t, err)
	assert.Equal(t, []*types.FileMeta{
		{Name: ".woodpecker/a.yaml", Data: []byte("steps: a")},
		{Name: ".woodpecker/b.yaml", Data: []byte("steps: b")},
	}, metas)
}

func TestBranchHead(t *testing.T) {
	client := newTestServer(t, map[string]string{
		"/api/v4/projects/4":                          `{"id":4}`,
		"/api/v4/projects/4/repository/branches/main": `{"name":"main","commit":{"id":"abc","web_url":"https://gitlab.example.com/commit/abc"}}`,
	})

	commit, err := client.BranchHead(t.Context(), &model.User{AccessToken: "token"}, &model.Repo{ForgeRemoteID: "4"}, "main")
	require.NoError(t, err)
	assert.Equal(t, &model.Commit{SHA: "abc", ForgeURL: "https://gitlab.example.com/commit/abc"}, commit)
}
