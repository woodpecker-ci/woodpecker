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
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v78/github"
	gh_mock "github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/github/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestNew(t *testing.T) {
	forge, _ := New(Opts{
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
	c, _ := New(Opts{
		URL:        s.URL,
		SkipVerify: true,
	})

	defer s.Close()

	ctx := t.Context()

	t.Run("netrc with user token", func(t *testing.T) {
		forge, _ := New(Opts{})
		netrc, _ := forge.Netrc(fakeUser, fakeRepo)
		assert.Equal(t, "github.com", netrc.Machine)
		assert.Equal(t, fakeUser.AccessToken, netrc.Login)
		assert.Equal(t, "x-oauth-basic", netrc.Password)
		assert.Equal(t, model.ForgeTypeGithub, netrc.Type)
	})
	t.Run("netrc with machine account", func(t *testing.T) {
		forge, _ := New(Opts{})
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
	mockedHTTPClient := gh_mock.NewMockedHTTPClient(
		gh_mock.WithRequestMatch(
			gh_mock.GetReposCommitsByOwnerByRepoByRef,
			github.RepositoryCommit{
				Files: []*github.CommitFile{
					{Filename: github.Ptr("README.md")},
					{Filename: github.Ptr("main.go")},
				},
			},
		),
		gh_mock.WithRequestMatch(
			gh_mock.GetReposCompareByOwnerByRepoByBasehead,
			github.CommitsComparison{
				Files: []*github.CommitFile{
					{Filename: github.Ptr("main.go")},
				},
			},
		),
		gh_mock.WithRequestMatch(
			gh_mock.GetReposPullsFilesByOwnerByRepoByPullNumber,
			[]*github.CommitFile{
				{Filename: github.Ptr("README.md")},
				{Filename: github.Ptr("main.go")},
			},
		),
	)

	// Create a GitHub client with the mocked HTTP client
	gh := github.NewClient(mockedHTTPClient)

	// Use the custom type as the key
	ctx := context.WithValue(context.Background(), githubClientKey, gh)

	// Create a mock store using the proper mocking pattern
	mockStore := store_mocks.NewMockStore(t)
	mockStore.On("GetUser", mock.Anything).Return(&model.User{
		ID:          1,
		Login:       "6543",
		AccessToken: "token",
	}, nil)
	mockStore.On("GetRepoNameFallback", mock.Anything, mock.Anything).Return(&model.Repo{
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
		assert.Equal(t, "Update .woodpecker.yml", pipeline.Message)
		assert.Equal(t, "https://github.com/6543/test_ci_tmp/commit/67012991d6c69b1c58378346fca366b864d8d1a1", pipeline.ForgeURL)
		assert.Equal(t, "6543", pipeline.Author)
		assert.Equal(t, "https://avatars.githubusercontent.com/u/24977596?v=4", pipeline.Avatar)
		assert.Equal(t, "6543@obermui.de", pipeline.Email)
		assert.Empty(t, pipeline.ChangedFiles)
	})
}
