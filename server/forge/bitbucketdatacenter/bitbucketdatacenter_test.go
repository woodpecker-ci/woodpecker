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
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neticdk/go-bitbucket/bitbucket"
	"github.com/neticdk/go-bitbucket/mock"
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

func TestBitbucketDCRepoByIDWithSpacesInName(t *testing.T) {
	// Bitbucket matches the "name" search filter against the repository
	// display name, while Woodpecker stores the repository slug. For a
	// repository with a space in its display name the slug differs from the
	// name, so looking the repository up by its forge remote ID must still
	// find it. See https://github.com/woodpecker-ci/woodpecker/issues/5821
	repos := []*bitbucket.Repository{
		{
			ID:   uint64(1234),
			Slug: "test-repo",
			Name: "test repo",
			Project: &bitbucket.Project{
				ID:  uint64(456),
				Key: "TEST",
			},
		},
	}

	s := mock.NewMockServer(
		mock.WithRequestMatchHandler(mock.SearchRepositories, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// emulate Bitbucket: the "name" filter matches the display name
			nameFilter := strings.ToLower(r.URL.Query().Get("name"))
			filtered := make([]*bitbucket.Repository, 0, len(repos))
			for _, repo := range repos {
				if nameFilter == "" || strings.Contains(strings.ToLower(repo.Name), nameFilter) {
					filtered = append(filtered, repo)
				}
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(bitbucket.RepositoryList{
				Repositories: filtered,
				Size:         uint(len(filtered)),
				LastPage:     true,
			})
			assert.NoError(t, err)
		})),
		mock.WithRequestMatch(mock.GetDefaultBranch, bitbucket.Branch{
			ID:        "refs/head/main",
			DisplayID: "main",
			Default:   true,
		}),
	)
	defer s.Close()

	c := &client{urlAPI: s.URL}

	repo, err := c.Repo(t.Context(), fakeUser, model.ForgeRemoteID("1234"), "TEST", "test-repo")
	require.NoError(t, err)
	assert.Equal(t, &model.Repo{
		Name:          "test-repo",
		Owner:         "TEST",
		Perm:          &model.Perm{Pull: true, Push: true},
		Branch:        "main",
		IsSCMPrivate:  true,
		PREnabled:     true,
		ForgeRemoteID: model.ForgeRemoteID("1234"),
		FullName:      "TEST/test-repo",
	}, repo)
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
