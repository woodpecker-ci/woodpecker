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
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
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

func TestBitbucketDCRepoByIDStopsAtTheMatch(t *testing.T) {
	// the repository is on the first page of many, no further page is needed
	requests := 0
	s := mock.NewMockServer(
		mock.WithRequestMatchHandler(mock.SearchRepositories, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requests++
			writeJSON(t, w, bitbucket.RepositoryList{
				Repositories:  []*bitbucket.Repository{{ID: uint64(1234), Slug: "test-repo", Project: &bitbucket.Project{Key: "TEST"}}},
				LastPage:      requests > 3,
				NextPageStart: uint(requests * listLimit),
			})
		})),
		mock.WithRequestMatch(mock.GetDefaultBranch, bitbucket.Branch{DisplayID: "main", Default: true}),
	)
	defer s.Close()
	c := &client{urlAPI: s.URL}

	repo, err := c.Repo(t.Context(), fakeUser, model.ForgeRemoteID("1234"), "TEST", "test-repo")
	require.NoError(t, err)
	assert.Equal(t, "TEST/test-repo", repo.FullName)
	assert.Equal(t, 1, requests)
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

var (
	compareChanges = mock.EndpointPattern{Pattern: "/api/latest/projects/:projectKey/repos/:repositorySlug/compare/changes", Method: http.MethodGet}
	prChanges      = mock.EndpointPattern{Pattern: "/api/latest/projects/:projectKey/repos/:repositorySlug/pull-requests/:pullRequestId/changes", Method: http.MethodGet}
)

func changes(last bool, next uint, paths ...string) bitbucket.ChangeList {
	l := bitbucket.ChangeList{LastPage: last, NextPageStart: next}
	for _, p := range paths {
		l.Changes = append(l.Changes, &bitbucket.Change{Path: bitbucket.ChangePath{Title: p}})
	}
	return l
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	assert.NoError(t, json.NewEncoder(w).Encode(v))
}

func signedHookRequest(t *testing.T, eventKey, payload, secret string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/hook", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Event-Key", eventKey)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	req.Header.Set(bitbucket.EventSignatureHeader, "sha256="+hex.EncodeToString(mac.Sum(nil)))
	return req
}

func hookContext(t *testing.T, remoteID model.ForgeRemoteID, repo *model.Repo) context.Context {
	t.Helper()
	user := &model.User{ID: 1, AccessToken: "token", Expiry: time.Now().Add(time.Hour).Unix()}
	s := store_mocks.NewMockStore(t)
	s.On("GetRepoForgeID", int64(1), remoteID).Return(repo, nil)
	s.On("GetUser", int64(1)).Return(user, nil)
	return store.InjectToContext(t.Context(), s)
}

func TestHook(t *testing.T) {
	t.Run("push collects the commit message and changed files of all pages", func(t *testing.T) {
		s := mock.NewMockServer(
			mock.WithRequestMatchHandler(mock.GetCommit, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/latest/projects/DEV/repos/deployment-automation/commits/76797d54bca87db6d1e3e82ee40622c7908aa514", r.URL.Path)
				writeJSON(t, w, bitbucket.Commit{ID: "76797d54bca87db6d1e3e82ee40622c7908aa514", Message: "Update dependencies"})
			})),
			mock.WithRequestMatchHandler(compareChanges, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "76797d54bca87db6d1e3e82ee40622c7908aa514", r.URL.Query().Get("from"))
				assert.Equal(t, "e0e15221b987fd8296141c0faa6a79f7c86ca4ce", r.URL.Query().Get("to"))
				if r.URL.Query().Get("start") == "0" {
					writeJSON(t, w, changes(false, 2, "go.mod", "go.sum"))
					return
				}
				assert.Equal(t, "2", r.URL.Query().Get("start"))
				writeJSON(t, w, changes(true, 0, "renovate.json"))
			})),
		)
		defer s.Close()

		c := &client{forgeID: 1, urlAPI: s.URL, url: "https://bitbucket.example.com"}
		dbRepo := &model.Repo{ID: 2, UserID: 1, Owner: "DEV", Name: "deployment-automation", WebhookSecret: "secret"}
		ctx := hookContext(t, "1684", dbRepo)

		repo, pipeline, err := c.Hook(ctx, signedHookRequest(t, "repo:refs_changed", fixtures.HookPush, "secret"))
		require.NoError(t, err)
		assert.Equal(t, dbRepo, repo)
		assert.Equal(t, model.EventPush, pipeline.Event)
		assert.Equal(t, "Update dependencies", pipeline.Message)
		assert.Equal(t, []string{"go.mod", "go.sum", "renovate.json"}, pipeline.ChangedFiles)
	})

	t.Run("pull request collects the changed files of all pages", func(t *testing.T) {
		s := mock.NewMockServer(
			mock.WithRequestMatchHandler(prChanges, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/latest/projects/DEV/repos/network-monitor/pull-requests/125/changes", r.URL.Path)
				if r.URL.Query().Get("start") == "0" {
					writeJSON(t, w, changes(false, 1, "main.go"))
					return
				}
				writeJSON(t, w, changes(true, 0, "README.md"))
			})),
		)
		defer s.Close()

		c := &client{forgeID: 1, urlAPI: s.URL, url: "https://bitbucket.example.com"}
		dbRepo := &model.Repo{ID: 3, UserID: 1, Owner: "DEV", Name: "network-monitor", WebhookSecret: "secret"}
		ctx := hookContext(t, "1079", dbRepo)

		repo, pipeline, err := c.Hook(ctx, signedHookRequest(t, "pr:opened", fixtures.HookPull, "secret"))
		require.NoError(t, err)
		assert.Equal(t, dbRepo, repo)
		assert.Equal(t, model.EventPull, pipeline.Event)
		assert.Equal(t, []string{"main.go", "README.md"}, pipeline.ChangedFiles)
	})

	t.Run("rejects a payload signed with another secret", func(t *testing.T) {
		// no API route is registered, so any forge call after the failed check would fail differently
		s := mock.NewMockServer()
		defer s.Close()

		c := &client{forgeID: 1, urlAPI: s.URL, url: "https://bitbucket.example.com"}
		ctx := hookContext(t, "1684", &model.Repo{ID: 2, UserID: 1, Owner: "DEV", Name: "deployment-automation", WebhookSecret: "secret"})

		_, _, err := c.Hook(ctx, signedHookRequest(t, "repo:refs_changed", fixtures.HookPush, "other"))
		assert.ErrorContains(t, err, "signature does not match")
	})

	t.Run("fails without a store", func(t *testing.T) {
		c := &client{forgeID: 1, url: "https://bitbucket.example.com"}
		_, _, err := c.Hook(t.Context(), signedHookRequest(t, "repo:refs_changed", fixtures.HookPush, "secret"))
		assert.ErrorContains(t, err, "unable to get store from context")
	})
}

func TestUpdatePipelineFromCommits(t *testing.T) {
	const (
		tagObjectSHA = "1111111111111111111111111111111111111111"
		commitSHA    = "2222222222222222222222222222222222222222"
	)

	s := mock.NewMockServer(
		mock.WithRequestMatch(
			mock.GetCommit,
			bitbucket.Commit{ID: commitSHA, Message: "Release"},
			bitbucket.Commit{ID: commitSHA, Message: "Release"},
		),
		mock.WithRequestMatch(compareChanges, changes(true, 0), changes(true, 0)),
	)
	defer s.Close()

	c := &client{urlAPI: s.URL, url: "https://bitbucket.example.com"}
	repo := &model.Repo{Owner: "PRJ", Name: "repo"}

	t.Run("nil pipeline", func(t *testing.T) {
		p, err := c.updatePipelineFromCommits(t.Context(), fakeUser, repo, nil, "", "")
		assert.NoError(t, err)
		assert.Nil(t, p)
	})

	t.Run("annotated tag points the pipeline to the tagged commit", func(t *testing.T) {
		// for annotated tags Bitbucket sends the tag object SHA as ToHash
		p, err := c.updatePipelineFromCommits(t.Context(), fakeUser, repo, &model.Pipeline{Event: model.EventTag, Commit: tagObjectSHA}, tagObjectSHA, "")
		require.NoError(t, err)
		assert.Equal(t, commitSHA, p.Commit)
		assert.Equal(t, "https://bitbucket.example.com/projects/PRJ/repos/repo/commits/"+commitSHA, p.ForgeURL)
		assert.Equal(t, "Release", p.Message)
	})

	t.Run("push keeps its commit and url", func(t *testing.T) {
		p, err := c.updatePipelineFromCommits(t.Context(), fakeUser, repo, &model.Pipeline{Event: model.EventPush, Commit: tagObjectSHA, ForgeURL: "keep"}, tagObjectSHA, "")
		require.NoError(t, err)
		assert.Equal(t, tagObjectSHA, p.Commit)
		assert.Equal(t, "keep", p.ForgeURL)
	})
}

func TestOrgMembership(t *testing.T) {
	permissionDenied := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mock.WriteError(w, http.StatusUnauthorized, []bitbucket.ErrorMessage{{Message: "no project admin"}})
	})

	t.Run("unknown without project admin scope", func(t *testing.T) {
		c := &client{}
		perm, err := c.OrgMembership(t.Context(), fakeUser, "PRJ")
		assert.NoError(t, err)
		assert.Nil(t, perm)
	})

	t.Run("project admin", func(t *testing.T) {
		s := mock.NewMockServer(
			mock.WithRequestMatchHandler(mock.SearchProjectPermissions, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/latest/projects/PRJ/permissions/search", r.URL.Path)
				writeJSON(t, w, bitbucket.ProjectPermissionList{
					LastPage:    true,
					Permissions: []*bitbucket.ProjectPermission{{Permission: bitbucket.PermissionProjectAdmin}},
				})
			})),
		)
		defer s.Close()

		c := &client{urlAPI: s.URL, oauthEnableProjectAdminScope: true}
		perm, err := c.OrgMembership(t.Context(), fakeUser, "PRJ")
		require.NoError(t, err)
		assert.Equal(t, &model.OrgPerm{Member: true, Admin: true}, perm)
	})

	t.Run("member through write access on a later page", func(t *testing.T) {
		s := mock.NewMockServer(
			mock.WithRequestMatchHandler(mock.SearchProjectPermissions, permissionDenied),
			mock.WithRequestMatchHandler(mock.SearchRepositories, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				assert.Equal(t, "PRJ", q.Get("projectkey"))
				assert.Equal(t, string(bitbucket.PermissionRepoWrite), q.Get("permission"))
				assert.Equal(t, "ACTIVE", q.Get("archived"))
				if q.Get("start") == "0" {
					writeJSON(t, w, bitbucket.RepositoryList{NextPageStart: 25})
					return
				}
				assert.Equal(t, "25", q.Get("start"))
				writeJSON(t, w, bitbucket.RepositoryList{
					LastPage:     true,
					Repositories: []*bitbucket.Repository{{Slug: "repo"}},
				})
			})),
		)
		defer s.Close()

		c := &client{urlAPI: s.URL, oauthEnableProjectAdminScope: true}
		perm, err := c.OrgMembership(t.Context(), fakeUser, "PRJ")
		require.NoError(t, err)
		assert.Equal(t, &model.OrgPerm{Member: true, Admin: false}, perm)
	})

	t.Run("no member without project or repository access", func(t *testing.T) {
		s := mock.NewMockServer(
			mock.WithRequestMatch(mock.SearchProjectPermissions, bitbucket.ProjectPermissionList{LastPage: true}),
			mock.WithRequestMatch(mock.SearchRepositories, bitbucket.RepositoryList{LastPage: true}),
		)
		defer s.Close()

		c := &client{urlAPI: s.URL, oauthEnableProjectAdminScope: true}
		perm, err := c.OrgMembership(t.Context(), fakeUser, "PRJ")
		require.NoError(t, err)
		assert.Equal(t, &model.OrgPerm{Member: false, Admin: false}, perm)
	})

	t.Run("repository search error", func(t *testing.T) {
		s := mock.NewMockServer(
			mock.WithRequestMatchHandler(mock.SearchProjectPermissions, permissionDenied),
			mock.WithRequestMatchHandler(mock.SearchRepositories, permissionDenied),
		)
		defer s.Close()

		c := &client{urlAPI: s.URL, oauthEnableProjectAdminScope: true}
		_, err := c.OrgMembership(t.Context(), fakeUser, "PRJ")
		assert.ErrorContains(t, err, "failed to check repository access")
	})
}

func TestTeams(t *testing.T) {
	s := mock.NewMockServer(
		mock.WithRequestMatch(
			mock.ListProjects,
			bitbucket.ProjectList{NextPageStart: 1, Projects: []*bitbucket.Project{{Key: "ONE"}}},
			bitbucket.ProjectList{LastPage: true, Projects: []*bitbucket.Project{{Key: "TWO"}}},
		),
	)
	defer s.Close()

	c := &client{urlAPI: s.URL}

	t.Run("collects the projects of all pages on the first page", func(t *testing.T) {
		teams, err := c.Teams(t.Context(), fakeUser, &model.ListOptions{Page: 1})
		require.NoError(t, err)
		require.Len(t, teams, 2)
		assert.Equal(t, "ONE", teams[0].Login)
		assert.Equal(t, "TWO", teams[1].Login)
		assert.True(t, strings.HasSuffix(teams[1].Avatar, "/projects/TWO/avatar.png"), teams[1].Avatar)
	})

	t.Run("later pages are empty", func(t *testing.T) {
		teams, err := c.Teams(t.Context(), fakeUser, &model.ListOptions{Page: 2})
		require.NoError(t, err)
		assert.Empty(t, teams)
	})
}

func TestRepos(t *testing.T) {
	s := mock.NewMockServer(
		mock.WithRequestMatchHandler(mock.SearchRepositories, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Query().Get("permission") {
			case string(bitbucket.PermissionRepoWrite):
				writeJSON(t, w, bitbucket.RepositoryList{
					LastPage: true,
					Repositories: []*bitbucket.Repository{
						{ID: 1, Slug: "write", Project: &bitbucket.Project{Key: "PRJ"}},
						{ID: 2, Slug: "admin", Project: &bitbucket.Project{Key: "PRJ"}},
					},
				})
			case string(bitbucket.PermissionRepoAdmin):
				writeJSON(t, w, bitbucket.RepositoryList{
					LastPage:     true,
					Repositories: []*bitbucket.Repository{{ID: 2, Slug: "admin", Project: &bitbucket.Project{Key: "PRJ"}}},
				})
			default:
				t.Errorf("unexpected permission filter %q", r.URL.Query().Get("permission"))
			}
		})),
	)
	defer s.Close()

	c := &client{urlAPI: s.URL}

	t.Run("first page lists writable repos and marks admin ones", func(t *testing.T) {
		repos, err := c.Repos(t.Context(), fakeUser, &model.ListOptions{Page: 1})
		require.NoError(t, err)
		require.Len(t, repos, 2)
		assert.Equal(t, "PRJ/write", repos[0].FullName)
		assert.Equal(t, &model.Perm{Pull: true, Push: true}, repos[0].Perm)
		assert.Equal(t, "PRJ/admin", repos[1].FullName)
		assert.Equal(t, &model.Perm{Pull: true, Push: true, Admin: true}, repos[1].Perm)
	})

	t.Run("later pages are empty", func(t *testing.T) {
		repos, err := c.Repos(t.Context(), fakeUser, &model.ListOptions{Page: 2})
		require.NoError(t, err)
		assert.Empty(t, repos)
	})
}

func TestFileAndDir(t *testing.T) {
	browse := mock.EndpointPattern{Pattern: "/api/latest/projects/:projectKey/repos/:repositorySlug/browse/*path", Method: http.MethodGet}
	files := mock.EndpointPattern{Pattern: "/api/latest/projects/:projectKey/repos/:repositorySlug/files/*path", Method: http.MethodGet}

	s := mock.NewMockServer(
		mock.WithRequestMatchHandler(browse, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, fakePipeline.Commit, r.URL.Query().Get("at"))
			switch r.URL.Path {
			case "/api/latest/projects/PRJ/repos/repo/browse/.woodpecker/build.yaml":
				writeJSON(t, w, bitbucket.FileContent{
					LastPage: true,
					Lines:    []bitbucket.FileLine{{Text: "steps:"}, {Text: "  - name: build"}},
				})
			default:
				mock.WriteError(w, http.StatusNotFound, []bitbucket.ErrorMessage{{Message: "not found"}})
			}
		})),
		mock.WithRequestMatchHandler(files, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/latest/projects/PRJ/repos/repo/files/.woodpecker" {
				mock.WriteError(w, http.StatusNotFound, []bitbucket.ErrorMessage{{Message: "not found"}})
				return
			}
			writeJSON(t, w, bitbucket.FileList{LastPage: true, Files: []string{"build.yaml"}})
		})),
	)
	defer s.Close()

	c := &client{urlAPI: s.URL}
	repo := &model.Repo{Owner: "PRJ", Name: "repo"}

	t.Run("file content is joined by lines", func(t *testing.T) {
		data, err := c.File(t.Context(), fakeUser, repo, fakePipeline, ".woodpecker/build.yaml")
		require.NoError(t, err)
		assert.Equal(t, "steps:\n  - name: build", string(data))
	})

	t.Run("missing file is a config not found error", func(t *testing.T) {
		_, err := c.File(t.Context(), fakeUser, repo, fakePipeline, ".woodpecker.yaml")
		assert.ErrorIs(t, err, &forge_types.ErrConfigNotFound{})
	})

	t.Run("dir loads every file with its full path", func(t *testing.T) {
		metas, err := c.Dir(t.Context(), fakeUser, repo, fakePipeline, ".woodpecker")
		require.NoError(t, err)
		assert.Equal(t, []*forge_types.FileMeta{{Name: ".woodpecker/build.yaml", Data: []byte("steps:\n  - name: build")}}, metas)
	})

	t.Run("missing dir is a config not found error", func(t *testing.T) {
		_, err := c.Dir(t.Context(), fakeUser, repo, fakePipeline, ".missing")
		assert.ErrorIs(t, err, &forge_types.ErrConfigNotFound{})
	})
}

func TestNetrc(t *testing.T) {
	c := &client{username: "machine", password: "secret"}

	netrc, err := c.Netrc(fakeUser, &model.Repo{Clone: "https://bitbucket.example.com/scm/prj/repo.git"})
	require.NoError(t, err)
	assert.Equal(t, &model.Netrc{
		Machine:  "bitbucket.example.com",
		Login:    "machine",
		Password: "secret",
		Type:     model.ForgeTypeBitbucketDatacenter,
	}, netrc)

	_, err = c.Netrc(fakeUser, &model.Repo{Clone: "://invalid"})
	assert.Error(t, err)
}

func TestBranchHead(t *testing.T) {
	s := mock.NewMockServer(
		mock.WithRequestMatchHandler(mock.SearchBranches, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Bitbucket filters by substring, so "main" also returns "main-old"
			var branches []*bitbucket.Branch
			if r.URL.Query().Get("filterText") == "main" {
				branches = []*bitbucket.Branch{
					{DisplayID: "main-old", LatestCommit: "old"},
					{DisplayID: "main", LatestCommit: "abc"},
				}
			}
			if r.URL.Query().Get("filterText") == "dev" {
				branches = []*bitbucket.Branch{{DisplayID: "dev-old", LatestCommit: "old"}}
			}
			writeJSON(t, w, bitbucket.BranchList{LastPage: true, Branches: branches})
		})),
	)
	defer s.Close()

	c := &client{urlAPI: s.URL}
	repo := &model.Repo{Owner: "PRJ", Name: "repo", ForgeURL: "https://bitbucket.example.com/projects/PRJ/repos/repo/browse"}

	commit, err := c.BranchHead(t.Context(), fakeUser, repo, "main")
	require.NoError(t, err)
	assert.Equal(t, &model.Commit{SHA: "abc", ForgeURL: "https://bitbucket.example.com/projects/PRJ/repos/repo/commits/abc"}, commit)

	_, err = c.BranchHead(t.Context(), fakeUser, repo, "dev")
	assert.ErrorContains(t, err, "no matching branches found")

	_, err = c.BranchHead(t.Context(), fakeUser, repo, "none")
	assert.ErrorContains(t, err, "no matching branches returned")
}

func TestActivate(t *testing.T) {
	var deleted []string
	var created *bitbucket.Webhook

	s := mock.NewMockServer(
		mock.WithRequestMatch(mock.SearchRepositories, bitbucket.RepositoryList{
			LastPage:     true,
			Repositories: []*bitbucket.Repository{{ID: 1, Slug: "repo", Project: &bitbucket.Project{Key: "PRJ"}}},
		}),
		mock.WithRequestMatch(mock.GetDefaultBranch, bitbucket.Branch{DisplayID: "main"}),
		mock.WithRequestMatchHandler(mock.ListWebhooks, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("start") != "1" {
				writeJSON(t, w, bitbucket.WebhookList{
					NextPageStart: 1,
					Webhooks: []*bitbucket.Webhook{
						{ID: 10, URL: "https://ci.example.com/api/hook?access_token=old"},
						{ID: 11, URL: "https://other.example.com/hook"},
					},
				})
				return
			}
			writeJSON(t, w, bitbucket.WebhookList{
				LastPage: true,
				Webhooks: []*bitbucket.Webhook{{ID: 12, URL: "https://ci.example.com/api/hook?access_token=older"}},
			})
		})),
		mock.WithRequestMatchHandler(mock.DeleteWebhook, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			deleted = append(deleted, r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		})),
		mock.WithRequestMatchHandler(mock.CreateWebhook, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			created = new(bitbucket.Webhook)
			assert.NoError(t, json.NewDecoder(r.Body).Decode(created))
			writeJSON(t, w, created)
		})),
	)
	defer s.Close()

	c := &client{urlAPI: s.URL}
	repo := &model.Repo{ForgeRemoteID: "1", Owner: "PRJ", Name: "repo", WebhookSecret: "repo-secret"}

	require.NoError(t, c.Activate(t.Context(), fakeUser, repo, "https://ci.example.com/api/hook?access_token=new"))

	// only hooks pointing to this server are replaced, across all pages
	assert.Equal(t, []string{
		"/api/latest/projects/PRJ/repos/repo/webhooks/10",
		"/api/latest/projects/PRJ/repos/repo/webhooks/12",
	}, deleted)
	require.NotNil(t, created)
	assert.Equal(t, "https://ci.example.com/api/hook?access_token=new", created.URL)
	assert.True(t, created.Active)
	require.NotNil(t, created.Config)
	assert.Equal(t, "repo-secret", created.Config.Secret)
	assert.Contains(t, created.Events, bitbucket.EventKeyRepoRefsChanged)
	assert.Contains(t, created.Events, bitbucket.EventKeyPullRequestOpened)
}
