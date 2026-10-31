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

package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v93/github"
	github_mock "github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/github/fixtures"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestNew(t *testing.T) {
	forge, _ := New(1, Opts{
		URL:               "http://localhost:8080/",
		OAuthClientID:     "0ZXh0IjoiI",
		OAuthClientSecret: "I1NiIsInR5",
		SkipVerify:        true,
	})
	f, _ := forge.(*client)
	assert.Equal(t, "http://localhost:8080", f.url)
	assert.Equal(t, "http://localhost:8080/api/v3/", f.API)
	assert.Equal(t, "0ZXh0IjoiI", f.Client)
	assert.Equal(t, "I1NiIsInR5", f.Secret)
	assert.True(t, f.SkipVerify)
}

func TestGithub(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := httptest.NewServer(fixtures.Handler())
	c, _ := New(1, Opts{
		URL:        s.URL,
		SkipVerify: true,
	})

	defer s.Close()

	ctx := t.Context()

	t.Run("netrc with user token", func(t *testing.T) {
		forge, _ := New(1, Opts{})
		netrc, _ := forge.Netrc(fakeUser, fakeRepo)
		assert.Equal(t, "github.com", netrc.Machine)
		assert.Equal(t, fakeUser.AccessToken, netrc.Login)
		assert.Equal(t, "x-oauth-basic", netrc.Password)
		assert.Equal(t, model.ForgeTypeGithub, netrc.Type)
	})
	t.Run("netrc with machine account", func(t *testing.T) {
		forge, _ := New(1, Opts{})
		netrc, _ := forge.Netrc(nil, fakeRepo)
		assert.Equal(t, "github.com", netrc.Machine)
		assert.Empty(t, netrc.Login)
		assert.Empty(t, netrc.Password)
	})

	t.Run("Should return the repository details", func(t *testing.T) {
		repo, err := c.Repo(ctx, fakeUser, fakeRepo.ForgeRemoteID, fakeRepo.Owner, fakeRepo.Name)
		assert.NoError(t, err)
		assert.Equal(t, fakeRepo.ForgeRemoteID, repo.ForgeRemoteID)
		assert.Equal(t, fakeRepo.Owner, repo.Owner)
		assert.Equal(t, fakeRepo.Name, repo.Name)
		assert.Equal(t, fakeRepo.FullName, repo.FullName)
		assert.True(t, repo.IsSCMPrivate)
		assert.Equal(t, fakeRepo.Clone, repo.Clone)
		assert.Equal(t, fakeRepo.ForgeURL, repo.ForgeURL)
	})
	t.Run("repo not found error", func(t *testing.T) {
		_, err := c.Repo(ctx, fakeUser, "0", fakeRepoNotFound.Owner, fakeRepoNotFound.Name)
		assert.Error(t, err)
	})

	t.Run("dir is fetched with a single graphql request", func(t *testing.T) {
		files, err := c.Dir(ctx, fakeUser, fakeRepo, &model.Pipeline{Commit: "abc123"}, "somedir")
		assert.NoError(t, err)
		require.Len(t, files, 1)
		assert.Equal(t, "somedir/a.yaml", files[0].Name)
		assert.Equal(t, "pipeline:", string(files[0].Data))
	})

	t.Run("dir returns config-not-found for missing directories", func(t *testing.T) {
		_, err := c.Dir(ctx, fakeUser, fakeRepo, &model.Pipeline{Commit: "abc123"}, "no-such-dir")
		assert.ErrorIs(t, err, &forge_types.ErrConfigNotFound{})
	})

	t.Run("dir skips symlinks", func(t *testing.T) {
		files, err := c.Dir(ctx, fakeUser, fakeRepo, &model.Pipeline{Commit: "abc123"}, "symlink-dir")
		require.NoError(t, err)
		assert.Equal(t, []*forge_types.FileMeta{
			{Name: "symlink-dir/a.yaml", Data: []byte("steps: a")},
		}, files)
	})

	t.Run("dir loads symlinks over rest when following them", func(t *testing.T) {
		followSymlinks = true
		t.Cleanup(func() { followSymlinks = false })

		// the graphql entry carries the link target as its text, which must
		// never be used as the file content
		files, err := c.Dir(ctx, fakeUser, fakeRepo, &model.Pipeline{Commit: "abc123"}, "symlink-dir")
		require.NoError(t, err)
		assert.Equal(t, []*forge_types.FileMeta{
			{Name: "symlink-dir/a.yaml", Data: []byte("steps: a")},
			{Name: "symlink-dir/link.yaml", Data: []byte("pipeline:")},
		}, files)
	})

	t.Run("dir loads large and binary files over rest", func(t *testing.T) {
		files, err := c.Dir(ctx, fakeUser, fakeRepo, &model.Pipeline{Commit: "abc123"}, "large-dir")
		require.NoError(t, err)
		assert.Equal(t, []*forge_types.FileMeta{
			{Name: "large-dir/a.yaml", Data: []byte("steps: a")},
			{Name: "large-dir/large.yaml", Data: []byte("pipeline:")},
			{Name: "large-dir/binary.yaml", Data: []byte("pipeline:")},
		}, files)
	})

	t.Run("dir falls back to per-file fetching when graphql fails", func(t *testing.T) {
		files, err := c.Dir(ctx, fakeUser, fakeRepo, &model.Pipeline{Commit: "abc123"}, "rest-fallback-dir")
		assert.NoError(t, err)
		require.Len(t, files, 1)
		assert.Equal(t, "rest-fallback-dir/b.yaml", files[0].Name)
		assert.Equal(t, "pipeline:", string(files[0].Data))
	})

	t.Run("dir falls back instead of reporting config-not-found for an unexpected graphql response", func(t *testing.T) {
		// a 200 without a repository object must not be mistaken for a missing
		// path: config-not-found makes the caller drop the pipeline silently
		files, err := c.Dir(ctx, fakeUser, fakeRepo, &model.Pipeline{Commit: "abc123"}, "null-repo-dir")
		require.NoError(t, err)
		assert.NotErrorIs(t, err, &forge_types.ErrConfigNotFound{})
		require.Len(t, files, 1)
		assert.Equal(t, "null-repo-dir/c.yaml", files[0].Name)
		assert.Equal(t, "pipeline:", string(files[0].Data))
	})
}

func TestStatusDeployment(t *testing.T) {
	var (
		method    string
		path      string
		decodeErr error
		body      struct {
			State       string `json:"state"`
			Description string `json:"description"`
			LogURL      string `json:"log_url"`
		}
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		decodeErr = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)

	gh, err := github.NewClient(
		github.WithURLs(new(server.URL+"/"), nil),
		github.WithHTTPClient(server.Client()),
	)
	require.NoError(t, err)

	ctx := context.WithValue(t.Context(), githubClientKey, gh)
	c := &client{}
	err = c.Status(ctx, fakeUser, &model.Repo{
		ID:    7,
		Owner: "octocat",
		Name:  "Hello-World",
	}, &model.Pipeline{
		Number:   9,
		Event:    model.EventDeploy,
		Status:   model.StatusSuccess,
		ForgeURL: "https://api.github.com/repos/octocat/Hello-World/deployments/42",
	}, nil)
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, method)
	assert.Equal(t, "/repos/octocat/Hello-World/deployments/42/statuses", path)
	require.NoError(t, decodeErr)
	assert.Equal(t, "success", body.State)
	assert.Equal(t, "Pipeline was successful", body.Description)
	assert.Contains(t, body.LogURL, "/repos/7/pipeline/9")
}

var (
	fakeUser = &model.User{
		Login:       "6543",
		AccessToken: "cfcd2084",
	}

	fakeRepo = &model.Repo{
		ForgeRemoteID: "5",
		Owner:         "octocat",
		Name:          "Hello-World",
		FullName:      "octocat/Hello-World",
		Avatar:        "https://github.com/images/error/octocat_happy.gif",
		ForgeURL:      "https://github.com/octocat/Hello-World",
		Clone:         "https://github.com/octocat/Hello-World.git",
		IsSCMPrivate:  true,
	}

	fakeRepoNotFound = &model.Repo{
		Owner:    "test_name",
		Name:     "repo_not_found",
		FullName: "test_name/repo_not_found",
	}
)

func TestHook(t *testing.T) {
	// Mock GitHub API for changed files
	mockedHTTPClient := github_mock.NewMockedHTTPClient(
		github_mock.WithRequestMatch(
			github_mock.GetReposCommitsByOwnerByRepoByRef,
			github.RepositoryCommit{
				Files: []*github.CommitFile{
					{Filename: new("README.md")},
					{Filename: new("main.go")},
				},
			},
		),
		github_mock.WithRequestMatch(
			github_mock.GetReposCompareByOwnerByRepoByBasehead,
			github.CommitsComparison{
				Files: []*github.CommitFile{
					{Filename: new("main.go")},
				},
			},
		),
		github_mock.WithRequestMatch(
			github_mock.GetReposPullsFilesByOwnerByRepoByPullNumber,
			[]*github.CommitFile{
				{Filename: new("README.md")},
				{Filename: new("main.go")},
			},
		),
	)

	// Create a GitHub client with the mocked HTTP client
	gh, err := github.NewClient(github.WithHTTPClient(mockedHTTPClient))
	require.NoError(t, err)

	// Use the custom type as the key
	ctx := context.WithValue(t.Context(), githubClientKey, gh)

	// Create a mock store using the proper mocking pattern
	mockStore := store_mocks.NewMockStore(t)
	mockStore.On("GetUser", mock.Anything).Return(&model.User{
		ID:          1,
		Login:       "6543",
		AccessToken: "token",
	}, nil)
	mockStore.On("GetRepoNameFallback", mock.Anything, mock.Anything, mock.Anything).Return(&model.Repo{
		ID:            1,
		ForgeRemoteID: "1",
		Owner:         "6543",
		Name:          "hello-world",
		UserID:        1,
	}, nil)

	// Set up context with mock store
	ctx = store.InjectToContext(ctx, mockStore)

	// Create a mock client
	c := &client{
		API: defaultAPI,
		url: defaultURL,
	}

	t.Run("convert push from webhook", func(t *testing.T) {
		// Create a mock HTTP request with a push event payload
		req := httptest.NewRequest("POST", "/hook", strings.NewReader(fixtures.HookPush))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GitHub-Event", "push")

		// Call the Hook function
		repo, pipeline, err := c.Hook(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, repo)
		assert.NotNil(t, pipeline)
		assert.Equal(t, model.EventPush, pipeline.Event)
		assert.Equal(t, "main", pipeline.Branch)
		assert.Equal(t, "refs/heads/main", pipeline.Ref)
		assert.Equal(t, "366701fde727cb7a9e7f21eb88264f59f6f9b89c", pipeline.Commit)
		assert.Equal(t, "Fix multiline secrets replacer (#700)\n\n* Fix multiline secrets replacer\r\n\r\n* Add tests", pipeline.Message)
		assert.Equal(t, "https://github.com/woodpecker-ci/woodpecker/commit/366701fde727cb7a9e7f21eb88264f59f6f9b89c", pipeline.ForgeURL)
		assert.Equal(t, "6543", pipeline.Author)
		assert.Equal(t, "https://avatars.githubusercontent.com/u/24977596?v=4", pipeline.Avatar)
		assert.Equal(t, "admin@philipp.info", pipeline.Email)
		assert.Equal(t, []string{"main.go"}, pipeline.ChangedFiles)
	})

	t.Run("convert pull request from webhook", func(t *testing.T) {
		// Create a mock HTTP request with a pull request event payload
		req := httptest.NewRequest("POST", "/hook", strings.NewReader(fixtures.HookPullRequest))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GitHub-Event", "pull_request")

		// Call the Hook function
		repo, pipeline, err := c.Hook(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, repo)
		assert.NotNil(t, pipeline)
		assert.Equal(t, model.EventPull, pipeline.Event)
		assert.Equal(t, "main", pipeline.Branch)
		assert.Equal(t, "refs/pull/1/head", pipeline.Ref)
		assert.Equal(t, "changes:main", pipeline.Refspec)
		assert.Equal(t, "0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c", pipeline.Commit)
		assert.Equal(t, "Update the README with new information", pipeline.Message)
		assert.Equal(t, "Update the README with new information", pipeline.Title)
		assert.Equal(t, "baxterthehacker", pipeline.Author)
		assert.Equal(t, "https://avatars.githubusercontent.com/u/6752317?v=3", pipeline.Avatar)
		assert.Equal(t, "octocat", pipeline.Sender)
		assert.Equal(t, []string{"README.md", "main.go"}, pipeline.ChangedFiles)
	})

	t.Run("convert deployment from webhook", func(t *testing.T) {
		// Create a mock HTTP request with a deployment event payload
		req := httptest.NewRequest("POST", "/hook", strings.NewReader(fixtures.HookDeploy))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GitHub-Event", "deployment")

		// Call the Hook function
		repo, pipeline, err := c.Hook(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, repo)
		assert.NotNil(t, pipeline)
		assert.Equal(t, model.EventDeploy, pipeline.Event)
		assert.Equal(t, "main", pipeline.Branch)
		assert.Equal(t, "refs/heads/main", pipeline.Ref)
		assert.Equal(t, "9049f1265b7d61be4a8904a9a27120d2064dab3b", pipeline.Commit)
		assert.Equal(t, "", pipeline.Message)
		assert.Equal(t, "https://api.github.com/repos/baxterthehacker/public-repo/deployments/710692", pipeline.ForgeURL)
		assert.Equal(t, "baxterthehacker", pipeline.Author)
		assert.Equal(t, "https://avatars.githubusercontent.com/u/6752317?v=3", pipeline.Avatar)
	})

	t.Run("convert tag from webhook", func(t *testing.T) {
		// Create a mock HTTP request with a tag event payload but push event header (tags create push events at github)
		req := httptest.NewRequest("POST", "/hook", strings.NewReader(fixtures.HookTag))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GitHub-Event", "push")

		// Call the Hook function
		repo, pipeline, err := c.Hook(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, repo)
		assert.NotNil(t, pipeline)
		assert.Equal(t, model.EventTag, pipeline.Event)
		assert.Equal(t, "main", pipeline.Branch)
		assert.Equal(t, "refs/tags/the-tag-v1", pipeline.Ref)
		assert.Equal(t, "67012991d6c69b1c58378346fca366b864d8d1a1", pipeline.Commit)
		assert.Equal(t, "the-tag-v1", pipeline.TagTitle)
		assert.Equal(t, "https://github.com/6543/test_ci_tmp/commit/67012991d6c69b1c58378346fca366b864d8d1a1", pipeline.ForgeURL)
		assert.Equal(t, "6543", pipeline.Author)
		assert.Equal(t, "https://avatars.githubusercontent.com/u/24977596?v=4", pipeline.Avatar)
		assert.Equal(t, "6543@obermui.de", pipeline.Email)
		assert.Empty(t, pipeline.ChangedFiles)
	})
}

func TestGetTagCommitSHA(t *testing.T) {
	// Tags API paginates 30 per page; put the target tag on the second page
	// to exercise pagination instead of a first-page match.
	mockedHTTPClient := github_mock.NewMockedHTTPClient(
		github_mock.WithRequestMatchPages(
			github_mock.GetReposTagsByOwnerByRepo,
			[]github.RepositoryTag{
				{Name: new("v1.0.0")},
				{Name: new("v1.0.1")},
			},
			[]github.RepositoryTag{
				{Name: new("v1.0.2")},
				{
					Name:   new("v1.0.3"),
					Commit: &github.Commit{SHA: new("deadbeefcafe")},
				},
			},
		),
	)

	gh, err := github.NewClient(github.WithHTTPClient(mockedHTTPClient))
	require.NoError(t, err)

	ctx := context.WithValue(t.Context(), githubClientKey, gh)

	mockStore := store_mocks.NewMockStore(t)
	mockStore.On("GetUser", mock.Anything).Return(&model.User{
		ID:          1,
		Login:       "6543",
		AccessToken: "token",
	}, nil)
	mockStore.On("GetRepoNameFallback", mock.Anything, mock.Anything, mock.Anything).Return(&model.Repo{
		ID:            1,
		ForgeRemoteID: "1",
		Owner:         "6543",
		Name:          "hello-world",
		UserID:        1,
	}, nil)
	ctx = store.InjectToContext(ctx, mockStore)

	c := &client{API: defaultAPI, url: defaultURL}

	t.Run("finds a tag beyond the first page", func(t *testing.T) {
		sha, err := c.getTagCommitSHA(ctx, &model.Repo{ForgeRemoteID: "1", FullName: "6543/hello-world"}, "v1.0.3")
		require.NoError(t, err)
		assert.Equal(t, "deadbeefcafe", sha)
	})

	t.Run("returns an error instead of looping forever when the tag does not exist", func(t *testing.T) {
		_, err := c.getTagCommitSHA(ctx, &model.Repo{ForgeRemoteID: "1", FullName: "6543/hello-world"}, "does-not-exist")
		require.Error(t, err)
	})
}

// mockedContext returns a context whose GitHub client talks to a mock API.
func mockedContext(t *testing.T, ctx context.Context, opts ...github_mock.MockBackendOption) context.Context {
	t.Helper()
	gh, err := github.NewClient(github.WithHTTPClient(github_mock.NewMockedHTTPClient(opts...)))
	require.NoError(t, err)
	return context.WithValue(ctx, githubClientKey, gh)
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	assert.NoError(t, json.NewEncoder(w).Encode(v))
}

func TestMatchingEmail(t *testing.T) {
	primaryUnverified := &github.UserEmail{Email: new("unverified@example.com"), Primary: new(true), Verified: new(false)}
	secondary := &github.UserEmail{Email: new("secondary@example.com"), Primary: new(false), Verified: new(true)}
	primary := &github.UserEmail{Email: new("primary@example.com"), Primary: new(true), Verified: new(true)}
	incomplete := &github.UserEmail{Email: new("incomplete@example.com")}

	assert.Equal(t, primary, matchingEmail([]*github.UserEmail{incomplete, primaryUnverified, secondary, primary}, defaultAPI))
	// github.com requires a verified primary address
	assert.Nil(t, matchingEmail([]*github.UserEmail{primaryUnverified, secondary}, defaultAPI))
	assert.Nil(t, matchingEmail(nil, defaultAPI))
	// GitHub Enterprise does not verify addresses, the first one is used
	assert.Equal(t, primaryUnverified, matchingEmail([]*github.UserEmail{primaryUnverified, secondary}, "https://github.example.com/api/v3/"))
	assert.Nil(t, matchingEmail(nil, "https://github.example.com/api/v3/"))
}

func TestMatchingHooks(t *testing.T) {
	hookWithoutID := &github.Hook{Config: &github.HookConfig{URL: new("https://ci.example.com/api/hook")}}
	otherHost := &github.Hook{ID: new(int64(1)), Config: &github.HookConfig{URL: new("https://other.example.com/api/hook")}}
	ours := &github.Hook{ID: new(int64(2)), Config: &github.HookConfig{URL: new("https://ci.example.com/api/hook?access_token=old")}}

	assert.Equal(t, ours, matchingHooks([]*github.Hook{hookWithoutID, otherHost, ours}, "https://ci.example.com/api/hook?access_token=new"))
	assert.Nil(t, matchingHooks([]*github.Hook{hookWithoutID, otherHost}, "https://ci.example.com/api/hook"))
	assert.Nil(t, matchingHooks([]*github.Hook{ours}, "://invalid"))
}

func TestPerPage(t *testing.T) {
	assert.Equal(t, defaultPageSize, perPage(0))
	assert.Equal(t, 10, perPage(10))
	assert.Equal(t, defaultPageSize, perPage(defaultPageSize+1))
}

func TestActivate(t *testing.T) {
	var deleted []string
	var created *github.Hook

	ctx := mockedContext(
		t, t.Context(),
		github_mock.WithRequestMatch(github_mock.GetReposByOwnerByRepo, github.Repository{
			ID: new(int64(1)), Name: new("hello-world"), FullName: new("octocat/hello-world"), Owner: &github.User{Login: new("octocat")},
		}),
		github_mock.WithRequestMatch(github_mock.GetReposHooksByOwnerByRepo, []*github.Hook{
			{ID: new(int64(1)), Config: &github.HookConfig{URL: new("https://other.example.com/hook")}},
			{ID: new(int64(2)), Config: &github.HookConfig{URL: new("https://ci.example.com/api/hook?access_token=old")}},
		}),
		github_mock.WithRequestMatchHandler(github_mock.DeleteReposHooksByOwnerByRepoByHookId, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			deleted = append(deleted, r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		})),
		github_mock.WithRequestMatchHandler(github_mock.PostReposHooksByOwnerByRepo, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			created = new(github.Hook)
			assert.NoError(t, json.NewDecoder(r.Body).Decode(created))
			w.WriteHeader(http.StatusCreated)
			writeJSON(t, w, created)
		})),
	)

	c := &client{API: defaultAPI, url: defaultURL}
	err := c.Activate(ctx, &model.User{AccessToken: "token"}, &model.Repo{Owner: "octocat", Name: "hello-world"}, "https://ci.example.com/api/hook?access_token=new")
	require.NoError(t, err)

	// the old hook of this server is replaced, hooks of others are kept
	assert.Equal(t, []string{"/repos/octocat/hello-world/hooks/2"}, deleted)
	require.NotNil(t, created)
	assert.Equal(t, "https://ci.example.com/api/hook?access_token=new", created.GetConfig().GetURL())
	assert.Equal(t, "form", created.GetConfig().GetContentType())
	assert.ElementsMatch(t, []string{"push", "pull_request", "pull_request_review", "deployment"}, created.Events)
}

func TestDeactivateWithoutMatchingHook(t *testing.T) {
	ctx := mockedContext(
		t, t.Context(),
		github_mock.WithRequestMatch(github_mock.GetReposByOwnerByRepo, github.Repository{
			ID: new(int64(1)), Name: new("hello-world"), FullName: new("octocat/hello-world"), Owner: &github.User{Login: new("octocat")},
		}),
		github_mock.WithRequestMatch(github_mock.GetReposHooksByOwnerByRepo, []*github.Hook{
			{ID: new(int64(1)), Config: &github.HookConfig{URL: new("https://other.example.com/hook")}},
		}),
		github_mock.WithRequestMatchHandler(github_mock.DeleteReposHooksByOwnerByRepoByHookId, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected hook deletion %s", r.URL.Path)
		})),
	)

	c := &client{API: defaultAPI, url: defaultURL}
	assert.NoError(t, c.Deactivate(ctx, &model.User{AccessToken: "token"}, &model.Repo{Owner: "octocat", Name: "hello-world"}, "https://ci.example.com/api/hook"))
}

func TestLoadChangedFilesFromCommits(t *testing.T) {
	const (
		prev = "1111111111111111111111111111111111111111"
		curr = "2222222222222222222222222222222222222222"
	)

	mockStore := store_mocks.NewMockStore(t)
	mockStore.On("GetRepoNameFallback", mock.Anything, mock.Anything, mock.Anything).Return(&model.Repo{
		ID: 1, ForgeRemoteID: "1", Owner: "octocat", Name: "hello-world", UserID: 1,
	}, nil).Maybe()
	mockStore.On("GetUser", mock.Anything).Return(&model.User{ID: 1, AccessToken: "token"}, nil).Maybe()

	c := &client{API: defaultAPI, url: defaultURL}
	repo := &model.Repo{ForgeRemoteID: "1", FullName: "octocat/hello-world"}

	t.Run("same commit before and after has no changes", func(t *testing.T) {
		// no API route is mocked, a forge call would fail the test
		ctx := store.InjectToContext(mockedContext(t, t.Context()), mockStore)
		p, err := c.loadChangedFilesFromCommits(ctx, repo, &model.Pipeline{}, curr, curr)
		require.NoError(t, err)
		assert.Empty(t, p.ChangedFiles)
	})

	t.Run("new branch lists the files of the current commit on all pages", func(t *testing.T) {
		ctx := store.InjectToContext(mockedContext(
			t, t.Context(),
			github_mock.WithRequestMatchPages(
				github_mock.GetReposCommitsByOwnerByRepoByRef,
				github.RepositoryCommit{Files: []*github.CommitFile{{Filename: new("a.go")}}},
				github.RepositoryCommit{Files: []*github.CommitFile{{Filename: new("b.go")}}},
			),
		), mockStore)
		p, err := c.loadChangedFilesFromCommits(ctx, repo, &model.Pipeline{}, curr, "0000000000000000000000000000000000000000")
		require.NoError(t, err)
		assert.Equal(t, []string{"a.go", "b.go"}, p.ChangedFiles)
	})

	t.Run("push compares both commits and keeps renamed paths once", func(t *testing.T) {
		ctx := store.InjectToContext(mockedContext(
			t, t.Context(),
			github_mock.WithRequestMatchHandler(github_mock.GetReposCompareByOwnerByRepoByBasehead, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/repos/octocat/hello-world/compare/"+prev+"..."+curr, r.URL.Path)
				writeJSON(t, w, github.CommitsComparison{Files: []*github.CommitFile{
					{Filename: new("new.go"), PreviousFilename: new("old.go")},
					{Filename: new("main.go")},
					{Filename: new("main.go")},
				}})
			})),
		), mockStore)
		p, err := c.loadChangedFilesFromCommits(ctx, repo, &model.Pipeline{}, curr, prev)
		require.NoError(t, err)
		assert.Equal(t, []string{"new.go", "old.go", "main.go"}, p.ChangedFiles)
	})

	t.Run("without store the pipeline is unchanged", func(t *testing.T) {
		p, err := c.loadChangedFilesFromCommits(t.Context(), repo, &model.Pipeline{Commit: curr}, curr, prev)
		require.NoError(t, err)
		assert.Equal(t, &model.Pipeline{Commit: curr}, p)
	})
}

func TestRepos(t *testing.T) {
	ctx := mockedContext(
		t, t.Context(),
		github_mock.WithRequestMatchPages(
			github_mock.GetUserRepos,
			[]github.Repository{
				{ID: new(int64(1)), FullName: new("octocat/active"), Name: new("active"), Owner: &github.User{Login: new("octocat")}},
				{ID: new(int64(2)), FullName: new("octocat/archived"), Name: new("archived"), Owner: &github.User{Login: new("octocat")}, Archived: new(true)},
			},
			[]github.Repository{
				{ID: new(int64(3)), FullName: new("octocat/second-page"), Name: new("second-page"), Owner: &github.User{Login: new("octocat")}},
			},
		),
	)
	c := &client{API: defaultAPI, url: defaultURL}
	user := &model.User{AccessToken: "token"}

	repos, err := c.Repos(ctx, user, &model.ListOptions{Page: 1})
	require.NoError(t, err)
	require.Len(t, repos, 2)
	assert.Equal(t, "octocat/active", repos[0].FullName)
	assert.Equal(t, "octocat/second-page", repos[1].FullName)

	// all repos come with the first page
	repos, err = c.Repos(ctx, user, &model.ListOptions{Page: 2})
	require.NoError(t, err)
	assert.Empty(t, repos)
}

func TestFileAndDir(t *testing.T) {
	encoded := func(s string) *string { return new(base64.StdEncoding.EncodeToString([]byte(s))) }

	ctx := mockedContext(
		t, t.Context(),
		github_mock.WithRequestMatchHandler(github_mock.GetReposContentsByOwnerByRepoByPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "abc", r.URL.Query().Get("ref"))
			switch r.URL.Path {
			case "/repos/octocat/hello-world/contents/.woodpecker/a.yaml":
				writeJSON(t, w, github.RepositoryContent{Type: new("file"), Encoding: new("base64"), Content: encoded("steps: a")})
			case "/repos/octocat/hello-world/contents/.woodpecker/b.yaml":
				writeJSON(t, w, github.RepositoryContent{Type: new("file"), Encoding: new("base64"), Content: encoded("steps: b")})
			case "/repos/octocat/hello-world/contents/.woodpecker":
				writeJSON(t, w, []github.RepositoryContent{
					{Type: new("file"), Name: new("a.yaml")},
					{Type: new("dir"), Name: new("sub")},
					{Type: new("symlink"), Name: new("link.yaml")},
					{Type: new("file"), Name: new("b.yaml")},
				})
			case "/repos/octocat/hello-world/contents/.broken":
				writeJSON(t, w, []github.RepositoryContent{
					{Type: new("file"), Name: new("a.yaml")},
					{Type: new("file"), Name: new("b.yaml")},
				})
			case "/repos/octocat/hello-world/contents/.broken/a.yaml", "/repos/octocat/hello-world/contents/.broken/b.yaml":
				w.WriteHeader(http.StatusInternalServerError)
				writeJSON(t, w, map[string]string{"message": "Server Error"})
			default:
				w.WriteHeader(http.StatusNotFound)
				writeJSON(t, w, map[string]string{"message": "Not Found"})
			}
		})),
	)

	c := &client{API: defaultAPI, url: defaultURL}
	user := &model.User{AccessToken: "token"}
	repo := &model.Repo{Owner: "octocat", Name: "hello-world"}
	pipeline := &model.Pipeline{Commit: "abc"}

	data, err := c.File(ctx, user, repo, pipeline, ".woodpecker/a.yaml")
	require.NoError(t, err)
	assert.Equal(t, "steps: a", string(data))

	_, err = c.File(ctx, user, repo, pipeline, ".woodpecker.yaml")
	assert.ErrorIs(t, err, &forge_types.ErrConfigNotFound{})

	_, err = c.File(ctx, user, repo, pipeline, ".woodpecker")
	assert.ErrorContains(t, err, "is a folder not a file")

	// dir fails the test instead of hanging when Dir does not return
	dir := func(path string) ([]*forge_types.FileMeta, error) {
		type result struct {
			metas []*forge_types.FileMeta
			err   error
		}
		done := make(chan result, 1)
		go func() {
			metas, err := c.Dir(ctx, user, repo, pipeline, path)
			done <- result{metas, err}
		}()
		select {
		case r := <-done:
			return r.metas, r.err
		case <-time.After(5 * time.Second):
			require.FailNow(t, "Dir did not return", path)
			return nil, nil
		}
	}

	// sub folders and symlinks are skipped, files are loaded with their content
	metas, err := dir(".woodpecker")
	require.NoError(t, err)
	assert.ElementsMatch(t, []*forge_types.FileMeta{
		{Name: ".woodpecker/a.yaml", Data: []byte("steps: a")},
		{Name: ".woodpecker/b.yaml", Data: []byte("steps: b")},
	}, metas)

	// a file that can not be loaded fails the folder
	_, err = dir(".broken")
	assert.ErrorContains(t, err, "500")

	_, err = dir(".missing")
	assert.ErrorIs(t, err, &forge_types.ErrConfigNotFound{})
}

func TestDirLimitsParallelFileFetches(t *testing.T) {
	// more files than workers, each fetch stays open for a moment so
	// unbounded fetching would have all of them in flight at once
	const files = 3 * maxParallelFileFetches
	var inFlight, maxInFlight atomic.Int32
	ctx := mockedContext(t, t.Context(),
		github_mock.WithRequestMatchHandler(github_mock.GetReposContentsByOwnerByRepoByPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/repos/octocat/hello-world/contents/.woodpecker" {
				list := make([]github.RepositoryContent, 0, files)
				for i := range files {
					list = append(list, github.RepositoryContent{Type: new("file"), Name: new(strconv.Itoa(i) + ".yaml")})
				}
				writeJSON(t, w, list)
				return
			}
			n := inFlight.Add(1)
			defer inFlight.Add(-1)
			for m := maxInFlight.Load(); n > m && !maxInFlight.CompareAndSwap(m, n); m = maxInFlight.Load() {
			}
			time.Sleep(20 * time.Millisecond)
			name := strings.TrimPrefix(r.URL.Path, "/repos/octocat/hello-world/contents/")
			writeJSON(t, w, github.RepositoryContent{Type: new("file"), Encoding: new("base64"), Content: new(base64.StdEncoding.EncodeToString([]byte("steps: " + name)))})
		})),
	)
	c := &client{API: defaultAPI, url: defaultURL}

	metas, err := c.Dir(ctx, &model.User{AccessToken: "token"}, &model.Repo{Owner: "octocat", Name: "hello-world"}, &model.Pipeline{Commit: "abc"}, ".woodpecker")
	require.NoError(t, err)
	assert.LessOrEqual(t, maxInFlight.Load(), int32(maxParallelFileFetches))
	// files keep the order of the folder listing
	require.Len(t, metas, files)
	for i, meta := range metas {
		name := ".woodpecker/" + strconv.Itoa(i) + ".yaml"
		assert.Equal(t, &forge_types.FileMeta{Name: name, Data: []byte("steps: " + name)}, meta)
	}
}

func TestOrgMembership(t *testing.T) {
	ctx := mockedContext(
		t, t.Context(),
		github_mock.WithRequestMatch(
			github_mock.GetOrgsMembershipsByOrgByUsername,
			github.Membership{State: new("active"), Role: new("admin")},
			github.Membership{State: new("pending"), Role: new("member")},
		),
	)
	c := &client{API: defaultAPI, url: defaultURL}
	user := &model.User{Login: "octocat", AccessToken: "token"}

	perm, err := c.OrgMembership(ctx, user, "org")
	require.NoError(t, err)
	assert.Equal(t, &model.OrgPerm{Member: true, Admin: true}, perm)

	// an invitation that is not accepted yet is no membership
	perm, err = c.OrgMembership(ctx, user, "org")
	require.NoError(t, err)
	assert.Equal(t, &model.OrgPerm{Member: false, Admin: false}, perm)
}

func TestOrg(t *testing.T) {
	ctx := mockedContext(
		t, t.Context(),
		github_mock.WithRequestMatchHandler(github_mock.GetOrgsByOrg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/orgs/org" {
				writeJSON(t, w, github.Organization{Login: new("org")})
				return
			}
			w.WriteHeader(http.StatusNotFound)
			writeJSON(t, w, map[string]string{"message": "Not Found"})
		})),
		github_mock.WithRequestMatchHandler(github_mock.GetUsersByUsername, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/users/octocat" {
				writeJSON(t, w, github.User{Login: new("octocat")})
				return
			}
			w.WriteHeader(http.StatusNotFound)
			writeJSON(t, w, map[string]string{"message": "Not Found"})
		})),
	)
	c := &client{API: defaultAPI, url: defaultURL}
	user := &model.User{AccessToken: "token"}

	org, err := c.Org(ctx, user, "org")
	require.NoError(t, err)
	assert.Equal(t, &model.Org{Name: "org"}, org)

	org, err = c.Org(ctx, user, "octocat")
	require.NoError(t, err)
	assert.Equal(t, &model.Org{Name: "octocat", IsUser: true}, org)

	_, err = c.Org(ctx, user, "missing")
	assert.Error(t, err)
}

func TestRefreshWithoutRefreshToken(t *testing.T) {
	// OAuth apps do not issue refresh tokens, so there is nothing to refresh
	c := &client{API: defaultAPI, url: defaultURL}
	user := &model.User{AccessToken: "token", Expiry: time.Now().Unix()}
	refreshed, err := c.Refresh(t.Context(), user)
	require.NoError(t, err)
	assert.False(t, refreshed)
	assert.Equal(t, "token", user.AccessToken)
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

	f, ok := forge.(*client)
	require.True(t, ok)

	updated, err := f.Refresh(t.Context(), user)
	assert.NoError(t, err)
	assert.True(t, updated)

	assert.Equal(t, int32(1), tokenRequests.Load(), "token endpoint must be called exactly once")
	assert.Equal(t, "new-access", user.AccessToken)
	assert.Equal(t, "new-refresh", user.RefreshToken)
	assert.WithinDuration(t, time.Now().Add(time.Hour), time.Unix(user.Expiry, 0), 2*time.Minute)
}
