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

package bitbucketdatacenter

import (
	"testing"
	"time"

	"github.com/franela/goblin"
	bb "github.com/neticdk/go-bitbucket/bitbucket"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

//nolint:misspell
func TestHelper(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("Bitbucket Server converter", func() {
		g.It("should convert status", func() {
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
				g.Assert(to).Equal(tt.to)
			}
		})

		g.It("should convert repository", func() {
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
							Href: "https://git.domain/clone",
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
			g.Assert(to.ForgeRemoteID).Equal(model.ForgeRemoteID("1234"))
			g.Assert(to.Name).Equal("REPO")
			g.Assert(to.Owner).Equal("PRJ")
			g.Assert(to.Branch).Equal("main")
			g.Assert(to.SCMKind).Equal(model.RepoGit)
			g.Assert(to.FullName).Equal("PRJ/REPO")
			g.Assert(to.Perm).Equal(perm)
		})

		g.It("should convert repository push event", func() {
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
						Link:      "https://base.url/projects/PRJ/repos/REPO/commits/1234567890abcdef",
						Event:     model.EventPush,
					},
				},
			}
			for _, tt := range tests {
				to := convertRepositoryPushEvent(tt.from, "https://base.url")
				g.Assert(to).Equal(tt.to)
			}
		})

		g.It("should convert pull request event", func() {
			now := time.Now()
			from := &bb.PullRequestEvent{
				Event: bb.Event{
					Date: bb.ISOTime(now),
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
			g.Assert(to.Commit).Equal("1234567890abcdef")
			g.Assert(to.Branch).Equal("branch")
			g.Assert(to.Avatar).Equal("https://base.url/users/john.doe_mail.com/avatar.png")
			g.Assert(to.Author).Equal("John Doe")
			g.Assert(to.Email).Equal("john.doe@mail.com")
			g.Assert(to.Timestamp).Equal(now.UTC().Unix())
			g.Assert(to.Ref).Equal("refs/pull-requests/123/from")
			g.Assert(to.Link).Equal("https://base.url/projects/PRJ/repos/REPO/commits/1234567890abcdef")
			g.Assert(to.Event).Equal(model.EventPull)
			g.Assert(to.Refspec).Equal("branch:main")
		})

		g.It("should truncate author", func() {
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
					to:   "Some Very Long Author That May Includ...",
				},
			}
			for _, tt := range tests {
				g.Assert(authorLabel(tt.from)).Equal(tt.to)
			}
		})

		g.It("should convert user", func() {
			from := &bb.User{
				Slug:  "slug",
				Email: "john.doe@mail.com",
			}
			to := convertUser(from, "token", "https://base.url")
			g.Assert(to.Login).Equal("slug")
			g.Assert(to.Avatar).Equal("https://base.url/users/slug/avatar.png")
			g.Assert(to.Email).Equal("john.doe@mail.com")
			g.Assert(to.Token).Equal("token")
		})
	})
}
