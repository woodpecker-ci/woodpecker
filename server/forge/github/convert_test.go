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
	"testing"

	"github.com/google/go-github/v75/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"

	gh_mock "github.com/migueleliasweb/go-github-mock/src/mock"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

const (
	stateOpen = "open"
)

func Test_convertStatus(t *testing.T) {
	assert.Equal(t, statusSuccess, convertStatus(model.StatusSuccess))
	assert.Equal(t, statusPending, convertStatus(model.StatusPending))
	assert.Equal(t, statusPending, convertStatus(model.StatusRunning))
	assert.Equal(t, statusFailure, convertStatus(model.StatusFailure))
	assert.Equal(t, statusError, convertStatus(model.StatusKilled))
	assert.Equal(t, statusError, convertStatus(model.StatusError))
}

func Test_convertDesc(t *testing.T) {
	assert.Equal(t, descSuccess, convertDesc(model.StatusSuccess))
	assert.Equal(t, descPending, convertDesc(model.StatusPending))
	assert.Equal(t, descPending, convertDesc(model.StatusRunning))
	assert.Equal(t, descFailure, convertDesc(model.StatusFailure))
	assert.Equal(t, descError, convertDesc(model.StatusKilled))
	assert.Equal(t, descError, convertDesc(model.StatusError))
}

func Test_convertRepoList(t *testing.T) {
	from := []*github.Repository{
		{
			Private:  github.Ptr(false),
			FullName: github.Ptr("octocat/hello-world"),
			Name:     github.Ptr("hello-world"),
			Owner: &github.User{
				AvatarURL: github.Ptr("http://..."),
				Login:     github.Ptr("octocat"),
			},
			HTMLURL:  github.Ptr("https://github.com/octocat/hello-world"),
			CloneURL: github.Ptr("https://github.com/octocat/hello-world.git"),
			Permissions: map[string]bool{
				"push":  true,
				"pull":  true,
				"admin": true,
			},
		},
	}

	to := convertRepoList(from)
	assert.Equal(t, "http://...", to[0].Avatar)
	assert.Equal(t, "octocat/hello-world", to[0].FullName)
	assert.Equal(t, "octocat", to[0].Owner)
	assert.Equal(t, "hello-world", to[0].Name)
}

func Test_convertRepo(t *testing.T) {
	from := github.Repository{
		FullName:      github.Ptr("octocat/hello-world"),
		Name:          github.Ptr("hello-world"),
		HTMLURL:       github.Ptr("https://github.com/octocat/hello-world"),
		CloneURL:      github.Ptr("https://github.com/octocat/hello-world.git"),
		DefaultBranch: github.Ptr("develop"),
		Private:       github.Ptr(true),
		Owner: &github.User{
			AvatarURL: github.Ptr("http://..."),
			Login:     github.Ptr("octocat"),
		},
		Permissions: map[string]bool{
			"push":  true,
			"pull":  true,
			"admin": true,
		},
	}

	to := convertRepo(&from)
	assert.Equal(t, "http://...", to.Avatar)
	assert.Equal(t, "octocat/hello-world", to.FullName)
	assert.Equal(t, "octocat", to.Owner)
	assert.Equal(t, "hello-world", to.Name)
	assert.Equal(t, "develop", to.Branch)
	assert.True(t, to.IsSCMPrivate)
	assert.Equal(t, "https://github.com/octocat/hello-world.git", to.Clone)
	assert.Equal(t, "https://github.com/octocat/hello-world", to.ForgeURL)
}

func Test_convertPerm(t *testing.T) {
	from := &github.Repository{
		Permissions: map[string]bool{
			"admin": true,
			"push":  true,
			"pull":  true,
		},
	}

	to := convertPerm(from.GetPermissions())
	assert.True(t, to.Push)
	assert.True(t, to.Pull)
	assert.True(t, to.Admin)
}

func Test_convertTeam(t *testing.T) {
	from := &github.Organization{
		Login:     github.Ptr("octocat"),
		AvatarURL: github.Ptr("http://..."),
	}
	to := convertTeam(from)
	assert.Equal(t, "octocat", to.Login)
	assert.Equal(t, "http://...", to.Avatar)
}

func Test_convertTeamList(t *testing.T) {
	from := []*github.Organization{
		{
			Login:     github.Ptr("octocat"),
			AvatarURL: github.Ptr("http://..."),
		},
	}
	to := convertTeamList(from)
	assert.Equal(t, "octocat", to[0].Login)
	assert.Equal(t, "http://...", to[0].Avatar)
}

func Test_convertRepoHook(t *testing.T) {
	t.Run("should convert a repository from webhook", func(t *testing.T) {
		from := &github.PushEventRepository{Owner: &github.User{}}
		from.Owner.Login = github.Ptr("octocat")
		from.Owner.Name = github.Ptr("octocat")
		from.Name = github.Ptr("hello-world")
		from.FullName = github.Ptr("octocat/hello-world")
		from.Private = github.Ptr(true)
		from.HTMLURL = github.Ptr("https://github.com/octocat/hello-world")
		from.CloneURL = github.Ptr("https://github.com/octocat/hello-world.git")
		from.DefaultBranch = github.Ptr("develop")

		repo := convertRepoHook(from)
		assert.Equal(t, *from.Owner.Login, repo.Owner)
		assert.Equal(t, *from.Name, repo.Name)
		assert.Equal(t, *from.FullName, repo.FullName)
		assert.Equal(t, *from.Private, repo.IsSCMPrivate)
		assert.Equal(t, *from.HTMLURL, repo.ForgeURL)
		assert.Equal(t, *from.CloneURL, repo.Clone)
		assert.Equal(t, *from.DefaultBranch, repo.Branch)
	})
}

func Test_parsePullHook(t *testing.T) {
	from := &github.PullRequestEvent{
		Action: github.Ptr(actionOpen),
		PullRequest: &github.PullRequest{
			State:   github.Ptr(stateOpen),
			HTMLURL: github.Ptr("https://github.com/octocat/hello-world/pulls/42"),
			Number:  github.Ptr(42),
			Title:   github.Ptr("Updated README.md"),
			Base: &github.PullRequestBranch{
				Ref: github.Ptr("main"),
			},
			Head: &github.PullRequestBranch{
				Ref: github.Ptr("changes"),
				SHA: github.Ptr("f72fc19"),
				Repo: &github.Repository{
					CloneURL: github.Ptr("https://github.com/octocat/hello-world-fork"),
				},
			},
			User: &github.User{
				Login:     github.Ptr("octocat"),
				AvatarURL: github.Ptr("https://avatars1.githubusercontent.com/u/583231"),
			},
		}, Sender: &github.User{
			Login: github.Ptr("octocat"),
		},
	}
	pull, _, pipeline, err := parsePullHook(from, true)
	assert.NoError(t, err)
	assert.NotNil(t, pull)
	assert.Equal(t, model.EventPull, pipeline.Event)
	assert.Equal(t, *from.PullRequest.Base.Ref, pipeline.Branch)
	assert.Equal(t, "refs/pull/42/merge", pipeline.Ref)
	assert.Equal(t, "changes:main", pipeline.Refspec)
	assert.Equal(t, *from.PullRequest.Head.SHA, pipeline.Commit)
	assert.Equal(t, *from.PullRequest.Title, pipeline.Message)
	assert.Equal(t, *from.PullRequest.Title, pipeline.Title)
	assert.Equal(t, *from.PullRequest.User.Login, pipeline.Author)
	assert.Equal(t, *from.PullRequest.User.AvatarURL, pipeline.Avatar)
	assert.Equal(t, *from.Sender.Login, pipeline.Sender)
}

func Test_parseDeployHook(t *testing.T) {
	from := &github.DeploymentEvent{Deployment: &github.Deployment{}, Sender: &github.User{}}
	from.Deployment.Description = github.Ptr(":shipit:")
	from.Deployment.Environment = github.Ptr("production")
	from.Deployment.Task = github.Ptr("deploy")
	from.Deployment.ID = github.Ptr(int64(42))
	from.Deployment.Ref = github.Ptr("main")
	from.Deployment.SHA = github.Ptr("f72fc19")
	from.Deployment.URL = github.Ptr("https://github.com/octocat/hello-world")
	from.Sender.Login = github.Ptr("octocat")
	from.Sender.AvatarURL = github.Ptr("https://avatars1.githubusercontent.com/u/583231")

	_, pipeline := parseDeployHook(from)
	assert.Equal(t, model.EventDeploy, pipeline.Event)
	assert.Equal(t, "main", pipeline.Branch)
	assert.Equal(t, "refs/heads/main", pipeline.Ref)
	assert.Equal(t, *from.Deployment.SHA, pipeline.Commit)
	assert.Equal(t, *from.Deployment.Description, pipeline.Message)
	assert.Equal(t, *from.Deployment.URL, pipeline.ForgeURL)
	assert.Equal(t, *from.Sender.Login, pipeline.Author)
	assert.Equal(t, *from.Sender.AvatarURL, pipeline.Avatar)
}

func Test_parsePushHook(t *testing.T) {
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
	)

	// Create a GitHub client with the mocked HTTP client
	gh_client := github.NewClient(mockedHTTPClient)

	// Set up context with mocked client
	ctx := context.WithValue(t.Context(), "github_client", gh_client)

	// Create a mock store using the proper mocking pattern
	mockStore := store_mocks.NewMockStore(t)
	mockStore.On("GetUser", mock.Anything).Return(&model.User{
		ID:          1,
		Login:       "octocat",
		AccessToken: "token",
	}, nil)
	mockStore.On("GetRepoNameFallback", mock.Anything, mock.Anything).Return(&model.Repo{
		ID:            1,
		ForgeRemoteID: "1",
		Owner:         "octocat",
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
		from := &github.PushEvent{Sender: &github.User{}, Repo: &github.PushEventRepository{}, HeadCommit: &github.HeadCommit{Author: &github.CommitAuthor{}}}
		from.Sender.Login = github.Ptr("octocat")
		from.Sender.AvatarURL = github.Ptr("https://avatars1.githubusercontent.com/u/583231")
		from.Repo.CloneURL = github.Ptr("https://github.com/octocat/hello-world.git")
		from.HeadCommit.Author.Email = github.Ptr("octocat@github.com")
		from.HeadCommit.Message = github.Ptr("updated README.md")
		from.HeadCommit.URL = github.Ptr("https://github.com/octocat/hello-world")
		from.HeadCommit.ID = github.Ptr("f72fc19")
		from.Ref = github.Ptr("refs/heads/main")

		// Call the function that loads changed files
		_, pipeline, _, _ := parsePushHook(from)

		assert.Equal(t, model.EventPush, pipeline.Event)
		assert.Equal(t, "main", pipeline.Branch)
		assert.Equal(t, "refs/heads/main", pipeline.Ref)
		assert.Equal(t, *from.HeadCommit.ID, pipeline.Commit)
		assert.Equal(t, *from.HeadCommit.Message, pipeline.Message)
		assert.Equal(t, *from.HeadCommit.URL, pipeline.ForgeURL)
		assert.Equal(t, *from.Sender.Login, pipeline.Author)
		assert.Equal(t, *from.Sender.AvatarURL, pipeline.Avatar)
		assert.Equal(t, *from.HeadCommit.Author.Email, pipeline.Email)
	})

	t.Run("load changed files from single commit", func(t *testing.T) {
		from := &github.PushEvent{Sender: &github.User{}, Repo: &github.PushEventRepository{}, HeadCommit: &github.HeadCommit{Author: &github.CommitAuthor{}}}
		from.HeadCommit.ID = github.Ptr("f72fc19")
		from.Ref = github.Ptr("refs/heads/main")

		// Call the function that loads changed files
		repo, pipeline, curr, prev := parsePushHook(from)

		var err error
		pipeline, err = c.loadChangedFilesFromCommits(ctx, repo, pipeline, curr, prev)
		assert.NoError(t, err)

		assert.Equal(t, model.EventPush, pipeline.Event)
		assert.Equal(t, "main", pipeline.Branch)
		assert.Equal(t, "refs/heads/main", pipeline.Ref)
		assert.Equal(t, *from.HeadCommit.ID, curr)
		assert.NotNil(t, pipeline.ChangedFiles)
		assert.Contains(t, pipeline.ChangedFiles, "README.md")
		assert.Contains(t, pipeline.ChangedFiles, "main.go")
	})

	t.Run("load changed files from compare commits", func(t *testing.T) {
		from := &github.PushEvent{Sender: &github.User{}, Repo: &github.PushEventRepository{}, HeadCommit: &github.HeadCommit{Author: &github.CommitAuthor{}}}
		from.HeadCommit.ID = github.Ptr("f72fc19")
		from.Before = github.Ptr("91cf27f")
		from.Ref = github.Ptr("refs/heads/main")

		// Call the function that loads changed files
		repo, pipeline, curr, prev := parsePushHook(from)

		var err error
		pipeline, err = c.loadChangedFilesFromCommits(ctx, repo, pipeline, curr, prev)
		assert.NoError(t, err)

		assert.Equal(t, model.EventPush, pipeline.Event)
		assert.Equal(t, "main", pipeline.Branch)
		assert.Equal(t, "refs/heads/main", pipeline.Ref)
		assert.Equal(t, *from.HeadCommit.ID, curr)
		assert.Equal(t, *from.Before, prev)
		assert.NotNil(t, pipeline.ChangedFiles)
		assert.NotContains(t, pipeline.ChangedFiles, "README.md")
		assert.Contains(t, pipeline.ChangedFiles, "main.go")
	})

	t.Run("convert tag from webhook", func(t *testing.T) {
		from := &github.PushEvent{}
		from.Ref = github.Ptr("refs/tags/v1.0.0")

		_, pipeline, _, _ := parsePushHook(from)
		assert.Equal(t, model.EventTag, pipeline.Event)
		assert.Equal(t, "refs/tags/v1.0.0", pipeline.Ref)
	})

	t.Run("convert tag's base branch to pipeline's branch ", func(t *testing.T) {
		from := &github.PushEvent{}
		from.Ref = github.Ptr("refs/tags/v1.0.0")
		from.BaseRef = github.Ptr("refs/heads/main")

		_, pipeline, _, _ := parsePushHook(from)
		assert.Equal(t, model.EventTag, pipeline.Event)
		assert.Equal(t, "main", pipeline.Branch)
	})

	t.Run("not convert tag's base_ref from webhook if not prefixed with 'ref/heads/'", func(t *testing.T) {
		from := &github.PushEvent{}
		from.Ref = github.Ptr("refs/tags/v1.0.0")
		from.BaseRef = github.Ptr("refs/refs/main")

		_, pipeline, _, _ := parsePushHook(from)
		assert.Equal(t, model.EventTag, pipeline.Event)
		assert.Equal(t, "refs/tags/v1.0.0", pipeline.Branch)
	})
}
