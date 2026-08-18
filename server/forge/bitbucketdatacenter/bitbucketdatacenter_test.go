// Copyright 2024 Woodpecker Authors
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

package bitbucketdatacenter

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucketdatacenter/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestNew(t *testing.T) {
	forge, err := New(1, Opts{
		URL:               "http://localhost:8080",
		Username:          "0ZXh0IjoiI",
		Password:          "I1NiIsInR5",
		OAuthClientID:     "client-id",
		OAuthClientSecret: "client-secret",
	})
	assert.NoError(t, err)
	assert.NotNil(t, forge)
	cl, ok := forge.(*client)
	assert.True(t, ok)
	assert.Equal(t, &client{
		forgeID:      1,
		url:          "http://localhost:8080",
		urlAPI:       "http://localhost:8080/rest",
		username:     "0ZXh0IjoiI",
		password:     "I1NiIsInR5",
		clientID:     "client-id",
		clientSecret: "client-secret",
	}, cl)
}

func TestBitbucketDC(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := fixtures.Server()
	defer s.Close()
	c := &client{
		urlAPI: s.URL,
	}

	server.Config.Server.StatusContext = "ci/woodpecker"
	server.Config.Server.StatusContextFormat = "{{ .context }}/{{ .event }}/{{ .workflow }}"

	ctx := t.Context()

	repo, err := c.Repo(ctx, fakeUser, model.ForgeRemoteID("1234"), "PRJ", "repo-slug")
	assert.NoError(t, err)
	assert.Equal(t, &model.Repo{
		Name:          "repo-slug-2",
		Owner:         "PRJ",
		Perm:          &model.Perm{Pull: true, Push: true},
		Branch:        "main",
		IsSCMPrivate:  true,
		PREnabled:     true,
		ForgeRemoteID: model.ForgeRemoteID("1234"),
		FullName:      "PRJ/repo-slug-2",
	}, repo)

	// org
	org, err := c.Org(ctx, fakeUser, "ORG")
	assert.NoError(t, err)
	assert.Equal(t, &model.Org{
		Name:   "ORG",
		IsUser: false,
	}, org)

	// user
	org, err = c.Org(ctx, fakeUser, "~ORG")
	assert.NoError(t, err)
	assert.Equal(t, &model.Org{
		Name:   "~ORG",
		IsUser: true,
	}, org)

	// Execute the Status method
	err = c.Status(ctx, fakeUser, fakeRepo, fakePipeline, fakeWorkflow)
	assert.NoError(t, err)
}

func TestTagHeadResolvesTagBeforeFetchingCommit(t *testing.T) {
	var tagCalls, commitCalls int
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/latest/projects/PRJ/repos/repo/tags":
			tagCalls++
			assert.Equal(t, "v1.0.0", r.URL.Query().Get("filterText"))
			_, _ = w.Write([]byte(`{"isLastPage":true,"values":[{"displayId":"v1.0.0","latestCommit":"resolved-sha"}]}`))
		case "/api/latest/projects/PRJ/repos/repo/commits/resolved-sha":
			commitCalls++
			_, _ = w.Write([]byte(`{"id":"resolved-sha"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(s.Close)

	c := &client{urlAPI: s.URL}
	commit, err := c.TagHead(t.Context(), &model.User{AccessToken: "token"}, &model.Repo{
		Owner:    "PRJ",
		Name:     "repo",
		ForgeURL: "https://bitbucket.example/projects/PRJ/repos/repo/browse",
	}, "v1.0.0")

	require.NoError(t, err)
	assert.Equal(t, &model.Commit{
		SHA:      "resolved-sha",
		ForgeURL: "https://bitbucket.example/projects/PRJ/repos/repo/commits/resolved-sha",
	}, commit)
	assert.Equal(t, 1, tagCalls)
	assert.Equal(t, 1, commitCalls)
}

func TestTagsFetchesSinglePage(t *testing.T) {
	var calls int
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			_, _ = w.Write([]byte(`{"isLastPage":false,"nextPageStart":25,"values":[{"displayId":"v1.0.0"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"isLastPage":true,"values":[{"displayId":"v2.0.0"}]}`))
	}))
	t.Cleanup(s.Close)

	c := &client{urlAPI: s.URL}
	tags, err := c.Tags(t.Context(), &model.User{AccessToken: "token"}, &model.Repo{
		Owner: "PRJ",
		Name:  "repo",
	}, &model.ListOptions{All: true, Page: 1, PerPage: 25})

	require.NoError(t, err)
	assert.Equal(t, []string{"v1.0.0"}, tags)
	assert.Equal(t, 1, calls)
}

var (
	fakeUser = &model.User{
		AccessToken: "fake",
		Expiry:      time.Now().Add(1 * time.Hour).Unix(),
	}

	fakeRepo = &model.Repo{
		ID:     1,
		Owner:  "test-owner",
		Name:   "test-repo",
		Branch: "main",
	}

	fakePipeline = &model.Pipeline{
		ID:       1,
		Number:   42,
		Commit:   "3ce383490b3d90d79460c60f67ba2580acc6cc59",
		Started:  1759825800,
		Finished: 1759825883,
		Branch:   "feature-branch",
		Ref:      "refs/pull-requests/123/from",
		Event:    model.EventPush,
	}

	fakeWorkflow = &model.Workflow{
		ID:    1,
		PID:   1,
		Name:  "build",
		State: model.StatusSuccess,
	}
)
