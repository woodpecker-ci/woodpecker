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
	"net/url"
	"testing"
	"time"

	"github.com/neticdk/go-bitbucket/bitbucket"
	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func Test_convertStatus(t *testing.T) {
	tests := []struct {
		from model.StatusValue
		to   bitbucket.BuildStatusState
	}{
		{
			from: model.StatusPending,
			to:   bitbucket.BuildStatusStateInProgress,
		},
		{
			from: model.StatusRunning,
			to:   bitbucket.BuildStatusStateInProgress,
		},
		{
			from: model.StatusSuccess,
			to:   bitbucket.BuildStatusStateSuccessful,
		},
		{
			from: model.StatusValue("other"),
			to:   bitbucket.BuildStatusStateFailed,
		},
	}
	for _, tt := range tests {
		to := convertStatus(tt.from)
		assert.Equal(t, tt.to, to)
	}
}

func Test_convertRepo(t *testing.T) {
	from := &bitbucket.Repository{
		ID:   uint64(1234),
		Slug: "REPO",
		Project: &bitbucket.Project{
			Key: "PRJ",
		},
		Links: map[string][]bitbucket.Link{
			"clone": {
				{
					Name: "http",
					Href: "https://user@git.domain/clone",
				},
			},
			"self": {
				{
					Href: "https://git.domain/self",
				},
			},
		},
	}
	perm := &model.Perm{}
	to := convertRepo(from, perm, "main")

	assert.Equal(t, &model.Repo{
		ForgeRemoteID: model.ForgeRemoteID("1234"),
		Name:          "REPO",
		Owner:         "PRJ",
		Branch:        "main",
		FullName:      "PRJ/REPO",
		Perm:          perm,
		Clone:         "https://git.domain/clone",
		ForgeURL:      "https://git.domain/self",
		PREnabled:     true,
		IsSCMPrivate:  true,
	}, to)
}

func Test_convertRepositoryPushEvent(t *testing.T) {
	now := time.Now()
	tests := []struct {
		from *bitbucket.RepositoryPushEvent
		to   *model.Pipeline
	}{
		{
			from: &bitbucket.RepositoryPushEvent{},
			to:   nil,
		},
		{
			from: &bitbucket.RepositoryPushEvent{
				Changes: []bitbucket.RepositoryPushEventChange{
					{
						FromHash: "1234567890abcdef",
						ToHash:   "0000000000000000000000000000000000000000",
					},
				},
			},
			to: nil,
		},
		{
			from: &bitbucket.RepositoryPushEvent{
				Changes: []bitbucket.RepositoryPushEventChange{
					{
						FromHash: "0000000000000000000000000000000000000000",
						ToHash:   "1234567890abcdef",
						Type:     bitbucket.RepositoryPushEventChangeTypeDelete,
					},
				},
			},
			to: nil,
		},
		{
			from: &bitbucket.RepositoryPushEvent{
				Event: bitbucket.Event{
					Date: bitbucket.ISOTime(now),
					Actor: bitbucket.User{
						Name:  "John Doe",
						Email: "john.doe@mail.com",
						Slug:  "john.doe_mail.com",
					},
				},
				Repository: bitbucket.Repository{
					Slug: "REPO",
					Project: &bitbucket.Project{
						Key: "PRJ",
					},
				},
				Changes: []bitbucket.RepositoryPushEventChange{
					{
						Ref: bitbucket.RepositoryPushEventRef{
							ID:        "refs/head/branch",
							DisplayID: "branch",
						},
						RefId:  "refs/head/branch",
						ToHash: "1234567890abcdef",
					},
				},
			},
			to: &model.Pipeline{
				Commit: &model.Commit{
					SHA:      "1234567890abcdef",
					ForgeURL: "https://base.url/projects/PRJ/repos/REPO/commits/1234567890abcdef",
				},
				Branch:       "branch",
				AuthorAvatar: "https://base.url/users/john.doe_mail.com/avatar.png",
				Author:       "John Doe",
				Ref:          "refs/head/branch",
				ForgeURL:     "https://base.url/projects/PRJ/repos/REPO/commits/1234567890abcdef",
				Event:        model.EventPush,
			},
		},
	}
	for _, tt := range tests {
		to := convertRepositoryPushEvent(tt.from, "https://base.url")
		assert.Equal(t, tt.to, to)
	}
}

func Test_convertPullRequestEvent(t *testing.T) {
	now := time.Now()
	from := &bitbucket.PullRequestEvent{
		Event: bitbucket.Event{
			Date:     bitbucket.ISOTime(now),
			EventKey: bitbucket.EventKeyPullRequestFrom,
			Actor: bitbucket.User{
				Name:  "John Doe",
				Email: "john.doe@mail.com",
				Slug:  "john.doe_mail.com",
			},
		},
		PullRequest: bitbucket.PullRequest{
			ID:    123,
			Title: "my title",
			Source: bitbucket.PullRequestRef{
				ID:        "refs/head/branch",
				DisplayID: "branch",
				Latest:    "1234567890abcdef",
				Repository: bitbucket.Repository{
					Slug: "REPO",
					Project: &bitbucket.Project{
						Key: "PRJ",
					},
				},
			},
			Target: bitbucket.PullRequestRef{
				ID:        "refs/head/main",
				DisplayID: "main",
				Latest:    "abcdef1234567890",
				Repository: bitbucket.Repository{
					Slug: "REPO",
					Project: &bitbucket.Project{
						Key: "PRJ",
					},
				},
			},
		},
	}
	to := convertPullRequestEvent(from, "https://base.url")
	assert.Equal(t, &model.Pipeline{
		Commit: &model.Commit{
			SHA:      "1234567890abcdef",
			ForgeURL: "https://base.url/projects/PRJ/repos/REPO/commits/1234567890abcdef",
		},
		Branch:       "branch",
		AuthorAvatar: "https://base.url/users/john.doe_mail.com/avatar.png",
		Author:       "John Doe",
		Ref:          "refs/pull-requests/123/from",
		ForgeURL:     "https://base.url/projects/PRJ/repos/REPO/pull-requests/123",
		Event:        model.EventPull,
		Refspec:      "branch:main",
		PullRequest: &model.PullRequest{
			Index: "123",
			Title: "my title",
		},
	}, to)
}

func Test_convertPullRequestCloseEvent(t *testing.T) {
	now := time.Now()
	from := &bitbucket.PullRequestEvent{
		Event: bitbucket.Event{
			Date:     bitbucket.ISOTime(now),
			EventKey: bitbucket.EventKeyPullRequestMerged,
			Actor: bitbucket.User{
				Name:  "John Doe",
				Email: "john.doe@mail.com",
				Slug:  "john.doe_mail.com",
			},
		},
		PullRequest: bitbucket.PullRequest{
			ID:    123,
			Title: "my title",
			Source: bitbucket.PullRequestRef{
				ID:        "refs/head/branch",
				DisplayID: "branch",
				Latest:    "1234567890abcdef",
				Repository: bitbucket.Repository{
					Slug: "REPO",
					Project: &bitbucket.Project{
						Key: "PRJ",
					},
				},
			},
			Target: bitbucket.PullRequestRef{
				ID:        "refs/head/main",
				DisplayID: "main",
				Latest:    "abcdef1234567890",
				Repository: bitbucket.Repository{
					Slug: "REPO",
					Project: &bitbucket.Project{
						Key: "PRJ",
					},
				},
			},
		},
	}
	to := convertPullRequestEvent(from, "https://base.url")
	assert.Equal(t, &model.Pipeline{
		Commit: &model.Commit{
			SHA:      "1234567890abcdef",
			ForgeURL: "https://base.url/projects/PRJ/repos/REPO/commits/1234567890abcdef",
		},
		Branch:       "branch",
		AuthorAvatar: "https://base.url/users/john.doe_mail.com/avatar.png",
		Author:       "John Doe",
		Ref:          "refs/pull-requests/123/from",
		ForgeURL:     "https://base.url/projects/PRJ/repos/REPO/pull-requests/123",
		Event:        model.EventPullClosed,
		Refspec:      "branch:main",
		PullRequest: &model.PullRequest{
			Title: "my title",
			Index: "123",
		},
	}, to)
}

func Test_convertUser(t *testing.T) {
	from := &bitbucket.User{
		Slug:  "slug",
		Email: "john.doe@mail.com",
		ID:    1,
	}
	to := convertUser(from, "https://base.url")
	assert.Equal(t, &model.User{
		Login:         "slug",
		Avatar:        "https://base.url/users/slug/avatar.png",
		Email:         "john.doe@mail.com",
		ForgeRemoteID: "1",
	}, to)
}

func Test_convertProjectsToTeams(t *testing.T) {
	tests := []struct {
		projects []*bitbucket.Project
		baseURL  string
		expected []*model.Team
	}{
		{
			projects: []*bitbucket.Project{
				{
					Key: "PRJ1",
				},
				{
					Key: "PRJ2",
				},
			},
			baseURL: "https://base.url",
			expected: []*model.Team{
				{
					Login:  "PRJ1",
					Avatar: "https://base.url/projects/PRJ1/avatar.png",
				},
				{
					Login:  "PRJ2",
					Avatar: "https://base.url/projects/PRJ2/avatar.png",
				},
			},
		},
		{
			projects: []*bitbucket.Project{},
			baseURL:  "https://base.url",
			expected: []*model.Team{},
		},
	}

	for _, tt := range tests {
		// Parse the baseURL string into a *url.URL
		parsedURL, err := url.Parse(tt.baseURL)
		assert.NoError(t, err)

		mockClient := &bitbucket.Client{BaseURL: parsedURL}
		actual := convertProjectsToTeams(tt.projects, mockClient)

		assert.Equal(t, tt.expected, actual)
	}
}
