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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v89/github"
	github_mock "github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/github/fixtures"
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

func Test_github(t *testing.T) {
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
}

func TestGithubApp(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := httptest.NewServer(fixtures.Handler())
	defer s.Close()

	_, appKey := generateAppKey(t)

	c, err := New(1, Opts{
		URL:           s.URL,
		SkipVerify:    true,
		AppID:         "12345",
		AppPrivateKey: appKey,
	})
	require.NoError(t, err)

	ctx := t.Context()

	notInstalledRepo := &model.Repo{
		UserID:   1,
		Owner:    "not-installed",
		Name:     "some-repo",
		FullName: "not-installed/some-repo",
		Clone:    "https://github.com/not-installed/some-repo.git",
	}

	t.Run("app id and private key must be set together", func(t *testing.T) {
		_, err := New(1, Opts{AppID: "12345"})
		assert.Error(t, err)
		_, err = New(1, Opts{AppPrivateKey: appKey})
		assert.Error(t, err)
	})

	t.Run("invalid private key", func(t *testing.T) {
		_, err := New(1, Opts{AppID: "12345", AppPrivateKey: "not-a-key"})
		assert.Error(t, err)
	})

	t.Run("netrc uses a repo-scoped installation token", func(t *testing.T) {
		netrc, err := c.Netrc(fakeUser, fakeRepo)
		assert.NoError(t, err)
		assert.Equal(t, "github.com", netrc.Machine)
		assert.Equal(t, "x-access-token", netrc.Login)
		// the fixture only hands this token out for requests restricted to
		// specific repositories with read-only contents access
		assert.Equal(t, fixtures.ScopedInstallationToken, netrc.Password)
		assert.Equal(t, model.ForgeTypeGithub, netrc.Type)
	})

	t.Run("netrc works without user", func(t *testing.T) {
		netrc, err := c.Netrc(nil, fakeRepo)
		assert.NoError(t, err)
		assert.Equal(t, "x-access-token", netrc.Login)
		assert.Equal(t, fixtures.ScopedInstallationToken, netrc.Password)
	})

	t.Run("netrc uses an installation-wide token when configured", func(t *testing.T) {
		wide, err := New(1, Opts{
			URL:                s.URL,
			SkipVerify:         true,
			AppID:              "12345",
			AppPrivateKey:      appKey,
			AppCloneTokenScope: AppCloneTokenScopeInstallation,
		})
		require.NoError(t, err)
		netrc, err := wide.Netrc(fakeUser, fakeRepo)
		assert.NoError(t, err)
		assert.Equal(t, fixtures.InstallationToken, netrc.Password)
	})

	t.Run("invalid clone token scope", func(t *testing.T) {
		_, err := New(1, Opts{AppID: "12345", AppPrivateKey: appKey, AppCloneTokenScope: "bogus"})
		assert.ErrorContains(t, err, "clone token scope")
	})

	t.Run("netrc falls back to user token when app is not installed", func(t *testing.T) {
		netrc, err := c.Netrc(fakeUser, notInstalledRepo)
		assert.NoError(t, err)
		assert.Equal(t, fakeUser.AccessToken, netrc.Login)
		assert.Equal(t, "x-oauth-basic", netrc.Password)
	})

	t.Run("repo token prefers installation token", func(t *testing.T) {
		cl, _ := c.(*client)
		assert.Equal(t, fixtures.InstallationToken, cl.repoToken(ctx, fakeUser, fakeRepo))
		assert.Equal(t, fakeUser.AccessToken, cl.repoToken(ctx, fakeUser, notInstalledRepo))
	})

	t.Run("app health reports app name and installations", func(t *testing.T) {
		cl, _ := c.(*client)
		name, installations, err := cl.AppHealth(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "Woodpecker Test App", name)
		assert.Equal(t, 1, installations)
	})

	t.Run("app health errors without app", func(t *testing.T) {
		forge, _ := New(1, Opts{})
		cl, _ := forge.(*client)
		_, _, err := cl.AppHealth(ctx)
		assert.Error(t, err)
	})

	t.Run("status is sent with installation token", func(t *testing.T) {
		// the fixture handler rejects any other token
		err := c.Status(ctx, fakeUser, fakeRepo, &model.Pipeline{
			Commit: "366701fde727cb7a9e7f21eb88264f59f6f9b89c",
			Event:  model.EventPush,
		}, &model.Workflow{
			Name:  "test",
			State: model.StatusSuccess,
		})
		assert.NoError(t, err)
	})

	t.Run("file and dir are fetched with installation token", func(t *testing.T) {
		// the fixture handler rejects any other token
		data, err := c.File(ctx, fakeUser, fakeRepo, &model.Pipeline{Commit: "abc123"}, ".woodpecker.yml")
		assert.NoError(t, err)
		assert.Equal(t, "pipeline:", string(data))

		files, err := c.Dir(ctx, fakeUser, fakeRepo, &model.Pipeline{Commit: "abc123"}, "somedir")
		assert.NoError(t, err)
		require.Len(t, files, 1)
		assert.Equal(t, "somedir/a.yaml", files[0].Name)
	})

	t.Run("branches are listed with installation token", func(t *testing.T) {
		branches, err := c.Branches(ctx, fakeUser, fakeRepo, &model.ListOptions{Page: 1, PerPage: 10})
		assert.NoError(t, err)
		assert.Equal(t, []string{"main", "dev"}, branches)

		commit, err := c.BranchHead(ctx, fakeUser, fakeRepo, "main")
		assert.NoError(t, err)
		assert.Equal(t, "abc123", commit.SHA)
	})

	t.Run("pull requests are listed with installation token", func(t *testing.T) {
		prs, err := c.PullRequests(ctx, fakeUser, fakeRepo, &model.ListOptions{Page: 1, PerPage: 10})
		assert.NoError(t, err)
		require.Len(t, prs, 1)
		assert.Equal(t, model.ForgeRemoteID("7"), prs[0].Index)
	})

	t.Run("netrc falls back to user token on installation lookup errors", func(t *testing.T) {
		netrc, err := c.Netrc(fakeUser, &model.Repo{
			Owner:    "lookup-error",
			Name:     "some-repo",
			FullName: "lookup-error/some-repo",
			Clone:    "https://github.com/lookup-error/some-repo.git",
		})
		assert.NoError(t, err)
		assert.Equal(t, fakeUser.AccessToken, netrc.Login)
		assert.Equal(t, "x-oauth-basic", netrc.Password)
	})

	t.Run("netrc errors on invalid clone url", func(t *testing.T) {
		_, err := c.Netrc(fakeUser, &model.Repo{Clone: "://not-a-url"})
		assert.Error(t, err)
	})

	t.Run("repo token without app uses user token", func(t *testing.T) {
		forge, _ := New(1, Opts{})
		cl, _ := forge.(*client)
		assert.Equal(t, fakeUser.AccessToken, cl.repoToken(ctx, fakeUser, fakeRepo))
	})

	t.Run("hook client falls back to the repo owner", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("GetUser", int64(1)).Return(fakeUser, nil)
		mockStore.On("UpdateUser", mock.Anything).Return(nil).Maybe()
		storeCtx := store.InjectToContext(ctx, mockStore)

		cl, _ := c.(*client)
		gh, err := cl.newRepoHookClient(storeCtx, mockStore, notInstalledRepo)
		assert.NoError(t, err)
		assert.NotNil(t, gh)
	})

	t.Run("hook client errors when the repo owner is gone", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("GetUser", int64(1)).Return(nil, errors.New("user not found"))
		storeCtx := store.InjectToContext(ctx, mockStore)

		cl, _ := c.(*client)
		_, err := cl.newRepoHookClient(storeCtx, mockStore, notInstalledRepo)
		assert.Error(t, err)
	})

	t.Run("app health surfaces api errors", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v3/app", func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{"id": 12345, "name": "Woodpecker Test App"}`)
		})
		mux.HandleFunc("/api/v3/app/installations", func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "boom", http.StatusInternalServerError)
		})
		errServer := httptest.NewServer(mux)
		defer errServer.Close()

		forge, err := New(1, Opts{URL: errServer.URL, AppID: "12345", AppPrivateKey: appKey})
		require.NoError(t, err)
		cl, _ := forge.(*client)
		_, _, err = cl.AppHealth(ctx)
		assert.ErrorContains(t, err, "installations")

		deadForge, err := New(1, Opts{URL: "http://127.0.0.1:1", AppID: "12345", AppPrivateKey: appKey})
		require.NoError(t, err)
		cl, _ = deadForge.(*client)
		_, _, err = cl.AppHealth(ctx)
		assert.Error(t, err)
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
		github.WithURLs(github.Ptr(server.URL+"/"), nil),
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
					{Filename: github.Ptr("README.md")},
					{Filename: github.Ptr("main.go")},
				},
			},
		),
		github_mock.WithRequestMatch(
			github_mock.GetReposCompareByOwnerByRepoByBasehead,
			github.CommitsComparison{
				Files: []*github.CommitFile{
					{Filename: github.Ptr("main.go")},
				},
			},
		),
		github_mock.WithRequestMatch(
			github_mock.GetReposPullsFilesByOwnerByRepoByPullNumber,
			[]*github.CommitFile{
				{Filename: github.Ptr("README.md")},
				{Filename: github.Ptr("main.go")},
			},
		),
	)

	// Create a GitHub client with the mocked HTTP client
	gh, err := github.NewClient(github.WithHTTPClient(mockedHTTPClient))
	require.NoError(t, err)

	// Use the custom type as the key
	ctx := context.WithValue(context.Background(), githubClientKey, gh)

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
				{Name: github.Ptr("v1.0.0")},
				{Name: github.Ptr("v1.0.1")},
			},
			[]github.RepositoryTag{
				{Name: github.Ptr("v1.0.2")},
				{
					Name:   github.Ptr("v1.0.3"),
					Commit: &github.Commit{SHA: github.Ptr("deadbeefcafe")},
				},
			},
		),
	)

	gh, err := github.NewClient(github.WithHTTPClient(mockedHTTPClient))
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), githubClientKey, gh)

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
