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
	"testing"
	"time"

	bb "github.com/neticdk/go-bitbucket/bitbucket"
	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func Test_convertStatus(t *testing.T) {
	tests := []struct {
		from model.StatusValue
		to   bb.BuildStatusState
	}{
		{
			from: model.StatusPending,
			to:   bb.BuildStatusStateInProgress,
		},
		{
			from: model.StatusRunning,
			to:   bb.BuildStatusStateInProgress,
		},
		{
			from: model.StatusSuccess,
			to:   bb.BuildStatusStateSuccessful,
		},
		{
			from: model.StatusValue("other"),
			to:   bb.BuildStatusStateFailed,
		},
	}
	for _, tt := range tests {
		to := convertStatus(tt.from)
		assert.Equal(t, tt.to, to)
	}
}

func Test_convertRepo(t *testing.T) {
	from := &bb.Repository{
		ID:   uint64(1234),
		Slug: "REPO",
		Project: &bb.Project{
			Key: "PRJ",
		},
		Links: map[string][]bb.Link{
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
		from *bb.RepositoryPushEvent
		to   *model.Pipeline
	}{
		{
			from: &bb.RepositoryPushEvent{},
			to:   nil,
		},
		{
			from: &bb.RepositoryPushEvent{
				Changes: []bb.RepositoryPushEventChange{
					{
						FromHash: "1234567890abcdef",
						ToHash:   "0000000000000000000000000000000000000000",
					},
				},
			},
			to: nil,
		},
		{
			from: &bb.RepositoryPushEvent{
				Changes: []bb.RepositoryPushEventChange{
					{
						FromHash: "0000000000000000000000000000000000000000",
						ToHash:   "1234567890abcdef",
						Type:     bb.RepositoryPushEventChangeTypeDelete,
					},
				},
			},
			to: nil,
		},
		{
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
	}
	for _, tt := range tests {
		to := convertRepositoryPushEvent(tt.from, "https://base.url")
		assert.Equal(t, tt.to, to)
	}
}

func Test_convertPullRequestEvent(t *testing.T) {
	now := time.Now()
	from := &bb.PullRequestEvent{
		Event: bb.Event{
			Date:     bb.ISOTime(now),
			EventKey: bb.EventKeyPullRequestFrom,
			Actor: bb.User{
				Name:  "John Doe",
				Email: "john.doe@mail.com",
				Slug:  "john.doe_mail.com",
			},
		},
		PullRequest: bb.PullRequest{
			ID:    123,
			Title: "my title",
			Source: bb.PullRequestRef{
				ID:        "refs/head/branch",
				DisplayID: "branch",
				Latest:    "1234567890abcdef",
				Repository: bb.Repository{
					Slug: "REPO",
					Project: &bb.Project{
						Key: "PRJ",
					},
				},
			},
			Target: bb.PullRequestRef{
				ID:        "refs/head/main",
				DisplayID: "main",
				Latest:    "abcdef1234567890",
				Repository: bb.Repository{
					Slug: "REPO",
					Project: &bb.Project{
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
	}, to)
}

func Test_convertPullRequestCloseEvent(t *testing.T) {
	now := time.Now()
	from := &bb.PullRequestEvent{
		Event: bb.Event{
			Date:     bb.ISOTime(now),
			EventKey: bb.EventKeyPullRequestMerged,
			Actor: bb.User{
				Name:  "John Doe",
				Email: "john.doe@mail.com",
				Slug:  "john.doe_mail.com",
			},
		},
		PullRequest: bb.PullRequest{
			ID:    123,
			Title: "my title",
			Source: bb.PullRequestRef{
				ID:        "refs/head/branch",
				DisplayID: "branch",
				Latest:    "1234567890abcdef",
				Repository: bb.Repository{
					Slug: "REPO",
					Project: &bb.Project{
						Key: "PRJ",
					},
				},
			},
			Target: bb.PullRequestRef{
				ID:        "refs/head/main",
				DisplayID: "main",
				Latest:    "abcdef1234567890",
				Repository: bb.Repository{
					Slug: "REPO",
					Project: &bb.Project{
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
	from := &bb.User{
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
