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
		name string
		from *bitbucket.RepositoryPushEvent
		to   *model.Pipeline
	}{
		{
			name: "empty event",
			from: &bitbucket.RepositoryPushEvent{},
			to:   nil,
		},
		{
			name: "delete event with zero ToHash",
			from: &bb.RepositoryPushEvent{
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
			name: "delete event",
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
			name: "branch push event",
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
				Commit:    "1234567890abcdef",
				Branch:    "branch",
				Message:   "",
				Avatar:    "https://base.url/users/john.doe_mail.com/avatar.png",
				Author:    "John Doe",
				Email:     "john.doe@mail.com",
				Timestamp: now.UTC().Unix(),
				Ref:       "refs/head/branch",
				ForgeURL:  "https://base.url/projects/PRJ/repos/REPO/commits/1234567890abcdef",
				Event:     model.EventPush,
			},
		},
		{
			name: "tag push event",
			from: &bb.RepositoryPushEvent{
				Event: bb.Event{
					Date: bb.ISOTime(now),
					Actor: bb.User{
						Name:  "John Doe",
						Email: "john.doe@mail.com",
						Slug:  "john.doe_mail.com",
					},
				},
				Repository: bb.Repository{
					Slug: "REPO",
					Project: &bb.Project{
						Key: "PRJ",
					},
				},
				Changes: []bb.RepositoryPushEventChange{
					{
						Ref: bb.RepositoryPushEventRef{
							ID:        "refs/tags/v1.0.0",
							DisplayID: "v1.0.0",
						},
						RefId:  "refs/tags/v1.0.0",
						ToHash: "abcdef1234567890",
					},
				},
			},
			to: &model.Pipeline{
				Commit:    "abcdef1234567890",
				Branch:    "v1.0.0",
				Message:   "",
				Avatar:    "https://base.url/users/john.doe_mail.com/avatar.png",
				Author:    "John Doe",
				Email:     "john.doe@mail.com",
				Timestamp: now.UTC().Unix(),
				Ref:       "refs/tags/v1.0.0",
				ForgeURL:  "https://base.url/projects/PRJ/repos/REPO/commits/abcdef1234567890",
				Event:     model.EventTag,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			to := convertRepositoryPushEvent(tt.from, "https://base.url")
			assert.Equal(t, tt.to, to)
		})
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
		Commit:    "1234567890abcdef",
		Branch:    "branch",
		Avatar:    "https://base.url/users/john.doe_mail.com/avatar.png",
		Author:    "John Doe",
		Email:     "john.doe@mail.com",
		Timestamp: now.UTC().Unix(),
		Ref:       "refs/pull-requests/123/from",
		ForgeURL:  "https://base.url/projects/PRJ/repos/REPO/commits/1234567890abcdef",
		Event:     model.EventPull,
		Refspec:   "branch:main",
		Title:     "my title",
		Message:   "my title",
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
		Commit:    "1234567890abcdef",
		Branch:    "branch",
		Avatar:    "https://base.url/users/john.doe_mail.com/avatar.png",
		Author:    "John Doe",
		Email:     "john.doe@mail.com",
		Timestamp: now.UTC().Unix(),
		Ref:       "refs/pull-requests/123/from",
		ForgeURL:  "https://base.url/projects/PRJ/repos/REPO/commits/1234567890abcdef",
		Event:     model.EventPullClosed,
		Refspec:   "branch:main",
		Title:     "my title",
		Message:   "my title",
	}, to)
}

func Test_authorLabel(t *testing.T) {
	tests := []struct {
		from string
		to   string
	}{
		{
			from: "Some Short Author",
			to:   "Some Short Author",
		},
		{
			from: "Some Very Long Author That May Include Multiple Names Here",
			//nolint:misspell
			to: "Some Very Long Author That May Includ...",
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.to, authorLabel(tt.from))
	}
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
