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

package gitea

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/gitea/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestGiteaParser(t *testing.T) {
	pullMetaWebhookRepo := &model.Repo{
		ForgeRemoteID: "1234",
		Owner:         "a_nice_user",
		Name:          "hello_world_ci",
		FullName:      "a_nice_user/hello_world_ci",
		Avatar:        "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
		ForgeURL:      "https://gitea.com/a_nice_user/hello_world_ci",
		Clone:         "https://gitea.com/a_nice_user/hello_world_ci.git",
		CloneSSH:      "ssh://git@gitea.com:3344/a_nice_user/hello_world_ci.git",
		Branch:        "main",
		PREnabled:     true,
		Perm: &model.Perm{
			Pull:  true,
			Push:  true,
			Admin: true,
		},
	}

	tests := []struct {
		name  string
		data  string
		event string
		err   error
		repo  *model.Repo
		pipe  *model.Pipeline
	}{
		{
			name:  "should ignore unsupported hook events",
			data:  fixtures.HookPullRequest,
			event: "issues",
			err:   &types.ErrIgnoreEvent{},
		},
		{
			name:  "push event should handle a push hook",
			data:  fixtures.HookPushBranch,
			event: "push",
			repo: &model.Repo{
				ForgeRemoteID: "50820",
				Owner:         "meisam",
				Name:          "woodpecktester",
				FullName:      "meisam/woodpecktester",
				Avatar:        "https://codeberg.org/avatars/96512da76a14cf44e0bb32d1640e878e",
				ForgeURL:      "https://codeberg.org/meisam/woodpecktester",
				Clone:         "https://codeberg.org/meisam/woodpecktester.git",
				CloneSSH:      "git@codeberg.org:meisam/woodpecktester.git",
				Branch:        "main",
				PREnabled:     true,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Event: model.EventPush,
				Commit: &model.Commit{
					SHA:      "28c3613ae62640216bea5e7dc71aa65356e4298b",
					Message:  "Delete '.woodpecker/.check.yml'\n",
					ForgeURL: "https://codeberg.org/meisam/woodpecktester/commit/28c3613ae62640216bea5e7dc71aa65356e4298b",
					Author: model.CommitAuthor{
						Name:  "meisam",
						Email: "meisam@noreply.codeberg.org",
					},
				},
				Branch:       "fdsafdsa",
				Ref:          "refs/heads/fdsafdsa",
				Author:       "6543",
				AuthorAvatar: "https://codeberg.org/avatars/09a234c768cb9bca78f6b2f82d6af173",

				ForgeURL:     "https://codeberg.org/meisam/woodpecktester/commit/28c3613ae62640216bea5e7dc71aa65356e4298b",
				ChangedFiles: []string{".woodpecker/.check.yml"},
			},
		},
		{
			name:  "push event should extract repository and pipeline details",
			data:  fixtures.HookPush,
			event: "push",
			repo: &model.Repo{
				ForgeRemoteID: "1",
				Owner:         "gordon",
				Name:          "hello-world",
				FullName:      "gordon/hello-world",
				Avatar:        "http://gitea.golang.org/gordon/hello-world",
				ForgeURL:      "http://gitea.golang.org/gordon/hello-world",
				Clone:         "http://gitea.golang.org/gordon/hello-world.git",
				CloneSSH:      "git@gitea.golang.org:gordon/hello-world.git",
				IsSCMPrivate:  true,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Event: model.EventPush,
				Commit: &model.Commit{
					SHA:      "ef98532add3b2feb7a137426bba1248724367df5",
					Message:  "bump\n",
					ForgeURL: "http://gitea.golang.org/gordon/hello-world/commit/ef98532add3b2feb7a137426bba1248724367df5",
					Author: model.CommitAuthor{
						Name:  "Gordon the Gopher",
						Email: "gordon@golang.org",
					},
				},
				Branch:       "main",
				Ref:          "refs/heads/main",
				Author:       "gordon",
				AuthorAvatar: "http://1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
				ForgeURL:     "http://gitea.golang.org/gordon/hello-world/commit/ef98532add3b2feb7a137426bba1248724367df5",
				ChangedFiles: []string{"CHANGELOG.md", "app/controller/application.rb"},
			},
		},
		{
			name:  "push event should handle multi commit push",
			data:  fixtures.HookPushMulti,
			event: "push",
			repo: &model.Repo{
				ForgeRemoteID: "6",
				Owner:         "Test-CI",
				Name:          "multi-line-secrets",
				FullName:      "Test-CI/multi-line-secrets",
				Avatar:        "http://127.0.0.1:3000/avatars/5b0a83c2185b3cb1ebceb11062d6c2eb",
				ForgeURL:      "http://127.0.0.1:3000/Test-CI/multi-line-secrets",
				Clone:         "http://127.0.0.1:3000/Test-CI/multi-line-secrets.git",
				CloneSSH:      "ssh://git@127.0.0.1:2200/Test-CI/multi-line-secrets.git",
				Branch:        "main",
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Event: model.EventPush,
				Commit: &model.Commit{
					SHA:      "29be01c073851cf0db0c6a466e396b725a670453",
					Message:  "add some text\n",
					ForgeURL: "http://127.0.0.1:3000/Test-CI/multi-line-secrets/commit/29be01c073851cf0db0c6a466e396b725a670453",
					Author: model.CommitAuthor{
						Name:  "6543",
						Email: "6543@obermui.de",
					},
				},
				Branch:       "main",
				Ref:          "refs/heads/main",
				Author:       "test-user",
				AuthorAvatar: "http://127.0.0.1:3000/avatars/dd46a756faad4727fb679320751f6dea",
				ForgeURL:     "http://127.0.0.1:3000/Test-CI/multi-line-secrets/compare/6efcf5b7c98f3e7a491675164b7a2e7acac27941...29be01c073851cf0db0c6a466e396b725a670453",
				ChangedFiles: []string{"aaa", "aa"},
			},
		},
		{
			name:  "tag event should handle a tag hook",
			data:  fixtures.HookTag,
			event: "create",
			repo: &model.Repo{
				ForgeRemoteID: "12",
				Owner:         "gordon",
				Name:          "hello-world",
				FullName:      "gordon/hello-world",
				Avatar:        "https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
				ForgeURL:      "http://gitea.golang.org/gordon/hello-world",
				Clone:         "http://gitea.golang.org/gordon/hello-world.git",
				CloneSSH:      "git@gitea.golang.org:gordon/hello-world.git",
				Branch:        "main",
				IsSCMPrivate:  true,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Event:        model.EventTag,
				Commit:       &model.Commit{SHA: "ef98532add3b2feb7a137426bba1248724367df5"},
				Ref:          "refs/tags/v1.0.0",
				Author:       "gordon",
				TagTitle:     "v1.0.0",
				AuthorAvatar: "https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
				ForgeURL:     "http://gitea.golang.org/gordon/hello-world/releases/tag/v1.0.0",
			},
		},
		{
			name:  "pull-request events should handle a PR hook when PR got created",
			data:  fixtures.HookPullRequest,
			event: "pull_request",
			repo: &model.Repo{
				ForgeRemoteID: "35129377",
				Owner:         "gordon",
				Name:          "hello-world",
				FullName:      "gordon/hello-world",
				Avatar:        "https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
				ForgeURL:      "http://gitea.golang.org/gordon/hello-world",
				Clone:         "https://gitea.golang.org/gordon/hello-world.git",
				CloneSSH:      "",
				Branch:        "main",
				IsSCMPrivate:  true,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Event:        model.EventPull,
				Commit:       &model.Commit{SHA: "0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c"},
				Branch:       "main",
				Ref:          "refs/pull/1/head",
				Refspec:      "feature/changes:main",
				Author:       "gordon",
				AuthorAvatar: "https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
				ForgeURL:     "http://gitea.golang.org/gordon/hello-world/pull/1",
				PullRequest: &model.PullRequest{
					Labels: []string{},
					Index:  "1",
					Title:  "Update the README with new information",
				},
			},
		},
		{
			name:  "pull-request reopen events should handle a PR as it was first created",
			data:  fixtures.HookPullRequestReopened,
			event: "pull_request",
			repo: &model.Repo{
				ForgeRemoteID: "138564",
				Owner:         "test_it",
				Name:          "test_ci_thing",
				FullName:      "test_it/test_ci_thing",
				Avatar:        "https://codeberg.org/avatars/bb6f3159a98a869b43f20b350542f8fb",
				ForgeURL:      "https://codeberg.org/test_it/test_ci_thing",
				Clone:         "https://codeberg.org/test_it/test_ci_thing.git",
				CloneSSH:      "git@codeberg.org/test_it/test_ci_thing.git",
				Branch:        "main",
				PREnabled:     true,
				IsSCMPrivate:  false,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Author: "6543",
				Event:  "pull_request",
				Commit: &model.Commit{
					SHA: "36b5813240a9d2daa29b05046d56a53e18f39a3e",
				},
				Branch:  "main",
				Ref:     "refs/pull/1/head",
				Refspec: "6543-patch-1:main",
				PullRequest: &model.PullRequest{
					Title: "Some ned more AAAA", Index: "1",
					Labels: []string{},
				},
				AuthorAvatar: "https://codeberg.org/avatars/09a234c768cb9bca78f6b2f82d6af173",
				ForgeURL:     "https://codeberg.org/test_it/test_ci_thing/pulls/1",
			},
		},
		{
			name:  "pull-request events should handle a PR hook when PR got updated",
			data:  fixtures.HookPullRequestUpdated,
			event: "pull_request",
			repo: &model.Repo{
				ForgeRemoteID: "6",
				Owner:         "Test-CI",
				Name:          "multi-line-secrets",
				FullName:      "Test-CI/multi-line-secrets",
				Avatar:        "http://127.0.0.1:3000/avatars/5b0a83c2185b3cb1ebceb11062d6c2eb",
				ForgeURL:      "http://127.0.0.1:3000/Test-CI/multi-line-secrets",
				Clone:         "http://127.0.0.1:3000/Test-CI/multi-line-secrets.git",
				CloneSSH:      "ssh://git@127.0.0.1:2200/Test-CI/multi-line-secrets.git",
				Branch:        "main",
				PREnabled:     true,
				IsSCMPrivate:  false,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Event:        model.EventPull,
				Commit:       &model.Commit{SHA: "788ed8d02d3b7fcfcf6386dbcbca696aa1d4dc25"},
				Branch:       "main",
				Ref:          "refs/pull/2/head",
				Refspec:      "test-patch-1:main",
				Author:       "test",
				AuthorAvatar: "http://127.0.0.1:3000/avatars/dd46a756faad4727fb679320751f6dea",
				ForgeURL:     "http://127.0.0.1:3000/Test-CI/multi-line-secrets/pulls/2",
				PullRequest: &model.PullRequest{
					Labels: []string{
						"Kind/Bug",
						"Kind/Security",
					},
					Index: "2",
					Title: "New Pull",
				},
			},
		},
		{
			name:  "pull-request events should handle a PR closed hook when PR got closed",
			data:  fixtures.HookPullRequestClosed,
			event: "pull_request",
			repo: &model.Repo{
				ForgeRemoteID: "46534",
				Owner:         "anbraten",
				Name:          "test-repo",
				FullName:      "anbraten/test-repo",
				Avatar:        "https://seccdn.libravatar.org/avatar/fc9b6fe77c6b732a02925a62a81f05a0?d=identicon",
				ForgeURL:      "https://gitea.com/anbraten/test-repo",
				Clone:         "https://gitea.com/anbraten/test-repo.git",
				CloneSSH:      "git@gitea.com:anbraten/test-repo.git",
				Branch:        "main",
				PREnabled:     true,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Event:        model.EventPullClosed,
				Commit:       &model.Commit{SHA: "d555a5dd07f4d0148a58d4686ec381502ae6a2d4"},
				Branch:       "main",
				Ref:          "refs/pull/1/head",
				Refspec:      "anbraten-patch-1:main",
				Author:       "anbraten",
				AuthorAvatar: "https://seccdn.libravatar.org/avatar/fc9b6fe77c6b732a02925a62a81f05a0?d=identicon",
				ForgeURL:     "https://gitea.com/anbraten/test-repo/pulls/1",
				PullRequest: &model.PullRequest{
					Labels: []string{},
					Index:  "1",
					Title:  "Adjust file",
				},
			},
		},
		{
			name:  "pull-request events should handle a PR title change hook",
			data:  fixtures.HookPullRequestChangeTitle,
			event: "pull_request",
			repo:  pullMetaWebhookRepo,
			pipe: &model.Pipeline{
				Author:      "a_nice_user",
				Event:       model.EventPullMetadata,
				EventReason: []string{"edited"},
				Commit:      &model.Commit{SHA: "07977177c2cd7d46bad37b8472a9d50e7acb9d1f"},
				Branch:      "main",
				Ref:         "refs/pull/7/head",
				Refspec:     "jony-patch-1:main",
				PullRequest: &model.PullRequest{
					Title:  "Edit pull title :D",
					Index:  "7",
					Labels: []string{},
				},
				AuthorAvatar: "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
				ForgeURL:     "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
			},
		},
		{
			name:  "pull-request events should handle a PR body change hook",
			data:  fixtures.HookPullRequestChangeBody,
			event: "pull_request",
			repo:  pullMetaWebhookRepo,
			pipe: &model.Pipeline{
				Author:      "a_nice_user",
				Event:       model.EventPullMetadata,
				EventReason: []string{"edited"},
				Commit:      &model.Commit{SHA: "07977177c2cd7d46bad37b8472a9d50e7acb9d1f"},
				Branch:      "main",
				Ref:         "refs/pull/7/head",
				Refspec:     "jony-patch-1:main",
				PullRequest: &model.PullRequest{
					Title:  "somepull",
					Index:  "7",
					Labels: []string{},
				},
				AuthorAvatar: "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
				ForgeURL:     "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
			},
		},
		{
			name: "pull-request events should ignore a PR add review request hook",
			data: fixtures.HookPullRequestAddReviewRequest,
			err:  &types.ErrIgnoreEvent{},
		},
		{
			name: "pull-request events should ignore a PR add approval review request hook",
			data: fixtures.HookPullRequestReviewAck,
			err:  &types.ErrIgnoreEvent{},
		},
		{
			name: "pull-request events should ignore a PR add reject review request hook",
			data: fixtures.HookPullRequestReviewDeny,
			err:  &types.ErrIgnoreEvent{},
		},
		{
			name: "pull-request events should ignore a PR add comment review request hook",
			data: fixtures.HookPullRequestReviewComment,
			err:  &types.ErrIgnoreEvent{},
		},
		{
			name:  "pull-request events should handle a PR add label hook",
			data:  fixtures.HookPullRequestAddLabel,
			event: "pull_request", // type: pull_request_label
			repo:  pullMetaWebhookRepo,
			pipe: &model.Pipeline{
				Author:      "a_nice_user",
				Event:       model.EventPullMetadata,
				EventReason: []string{"label_updated"},
				Commit:      &model.Commit{SHA: "07977177c2cd7d46bad37b8472a9d50e7acb9d1f"},
				Branch:      "main",
				Ref:         "refs/pull/7/head",
				Refspec:     "jony-patch-1:main",
				PullRequest: &model.PullRequest{
					Title:  "somepull",
					Index:  "7",
					Labels: []string{"bug", "help wanted"},
				},
				AuthorAvatar: "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
				ForgeURL:     "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
			},
		},
		{
			name:  "pull-request events should handle a PR change label hook",
			data:  fixtures.HookPullRequestChangeLabel,
			event: "pull_request", // type: pull_request_label
			repo:  pullMetaWebhookRepo,
			pipe: &model.Pipeline{
				Author:      "a_nice_user",
				Event:       model.EventPullMetadata,
				EventReason: []string{"label_updated"},
				Commit:      &model.Commit{SHA: "07977177c2cd7d46bad37b8472a9d50e7acb9d1f"},
				Branch:      "main",
				Ref:         "refs/pull/7/head",
				Refspec:     "jony-patch-1:main",
				PullRequest: &model.PullRequest{
					Title:  "somepull",
					Index:  "7",
					Labels: []string{"bug"},
				},
				AuthorAvatar: "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
				ForgeURL:     "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
			},
		},
		{
			name:  "pull-request events should handle a PR remove label hook",
			data:  fixtures.HookPullRequestRemoveLabel,
			event: "pull_request", // type: pull_request_label
			repo:  pullMetaWebhookRepo,
			pipe: &model.Pipeline{
				Author:      "a_nice_user",
				Event:       model.EventPullMetadata,
				EventReason: []string{"label_cleared"},
				Commit:      &model.Commit{SHA: "07977177c2cd7d46bad37b8472a9d50e7acb9d1f"},
				Branch:      "main",
				Ref:         "refs/pull/7/head",
				Refspec:     "jony-patch-1:main",
				PullRequest: &model.PullRequest{
					Title:  "somepull",
					Index:  "7",
					Labels: []string{},
				},
				AuthorAvatar: "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
				ForgeURL:     "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
			},
		},
		{
			name:  "pull-request events should handle a PR add milestone hook",
			data:  fixtures.HookPullRequestAddMile,
			event: "pull_request", // type: pull_request_milestone
			repo:  pullMetaWebhookRepo,
			pipe: &model.Pipeline{
				Author:      "a_nice_user",
				Event:       model.EventPullMetadata,
				EventReason: []string{"milestoned"},
				Commit:      &model.Commit{SHA: "07977177c2cd7d46bad37b8472a9d50e7acb9d1f"},
				Branch:      "main",
				Ref:         "refs/pull/7/head",
				Refspec:     "jony-patch-1:main",
				PullRequest: &model.PullRequest{
					Title:     "somepull",
					Index:     "7",
					Labels:    []string{"bug", "help wanted"},
					Milestone: "new mile",
				},
				AuthorAvatar: "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
				ForgeURL:     "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
			},
		},
		{
			name:  "pull-request events should handle a PR change milestone hook",
			data:  fixtures.HookPullRequestChangeMile,
			event: "pull_request", // type: pull_request_milestone
			repo:  pullMetaWebhookRepo,
			pipe: &model.Pipeline{
				Author:      "a_nice_user",
				Event:       model.EventPullMetadata,
				EventReason: []string{"milestoned"},
				Commit:      &model.Commit{SHA: "07977177c2cd7d46bad37b8472a9d50e7acb9d1f"},
				Branch:      "main",
				Ref:         "refs/pull/7/head",
				Refspec:     "jony-patch-1:main",
				PullRequest: &model.PullRequest{
					Title:     "somepull",
					Index:     "7",
					Labels:    []string{"bug", "help wanted"},
					Milestone: "closed mile",
				},
				AuthorAvatar: "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
				ForgeURL:     "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
			},
		},
		{
			name:  "pull-request events should handle a PR remove milestone hook",
			data:  fixtures.HookPullRequestRemoveMile,
			event: "pull_request", // type: pull_request_milestone
			repo:  pullMetaWebhookRepo,
			pipe: &model.Pipeline{
				Author:      "a_nice_user",
				Event:       model.EventPullMetadata,
				EventReason: []string{"demilestoned"},
				Commit:      &model.Commit{SHA: "07977177c2cd7d46bad37b8472a9d50e7acb9d1f"},
				Branch:      "main",
				Ref:         "refs/pull/7/head",
				Refspec:     "jony-patch-1:main",
				PullRequest: &model.PullRequest{
					Title:  "somepull",
					Index:  "7",
					Labels: []string{"bug", "help wanted"},
				},
				AuthorAvatar: "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
				ForgeURL:     "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
			},
		},
		{
			name:  "pull-request events should handle a PR add assignee hook",
			data:  fixtures.HookPullRequestAssigneesAdded,
			event: "pull_request", // type: pull_request_assign
			repo:  pullMetaWebhookRepo,
			pipe: &model.Pipeline{
				Author:      "a_nice_user",
				Event:       model.EventPullMetadata,
				EventReason: []string{"assigned"},
				Commit:      &model.Commit{SHA: "07977177c2cd7d46bad37b8472a9d50e7acb9d1f"},
				Branch:      "main",
				Ref:         "refs/pull/7/head",
				Refspec:     "jony-patch-1:main",
				PullRequest: &model.PullRequest{
					Title:  "somepull",
					Index:  "7",
					Labels: []string{"bug"},
				},
				AuthorAvatar: "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
				ForgeURL:     "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
			},
		},
		{
			name:  "pull-request events should handle a PR remove assignee hook",
			data:  fixtures.HookPullRequestAssigneesRemoved,
			event: "pull_request", // type: pull_request_assign
			repo:  pullMetaWebhookRepo,
			pipe: &model.Pipeline{
				Author:      "a_nice_user",
				Event:       model.EventPullMetadata,
				EventReason: []string{"unassigned"},
				Commit:      &model.Commit{SHA: "07977177c2cd7d46bad37b8472a9d50e7acb9d1f"},
				Branch:      "main",
				Ref:         "refs/pull/7/head",
				Refspec:     "jony-patch-1:main",
				PullRequest: &model.PullRequest{
					Title:  "somepull",
					Index:  "7",
					Labels: []string{"bug"},
				},
				AuthorAvatar: "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
				ForgeURL:     "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
			},
		},
		{
			name:  "pull-request events should handle a PR closed hook when PR was merged",
			data:  fixtures.HookPullRequestMerged,
			event: "pull_request",
			repo: &model.Repo{
				ForgeRemoteID: "46534",
				Owner:         "anbraten",
				Name:          "test-repo",
				FullName:      "anbraten/test-repo",
				Avatar:        "https://seccdn.libravatar.org/avatar/fc9b6fe77c6b732a02925a62a81f05a0?d=identicon",
				ForgeURL:      "https://gitea.com/anbraten/test-repo",
				Clone:         "https://gitea.com/anbraten/test-repo.git",
				CloneSSH:      "git@gitea.com:anbraten/test-repo.git",
				Branch:        "main",
				PREnabled:     true,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Event:        model.EventPullClosed,
				Commit:       &model.Commit{SHA: "d555a5dd07f4d0148a58d4686ec381502ae6a2d4"},
				Branch:       "main",
				Ref:          "refs/pull/1/head",
				Refspec:      "anbraten-patch-1:main",
				Author:       "anbraten",
				AuthorAvatar: "https://seccdn.libravatar.org/avatar/fc9b6fe77c6b732a02925a62a81f05a0?d=identicon",
				ForgeURL:     "https://gitea.com/anbraten/test-repo/pulls/1",
				PullRequest: &model.PullRequest{
					Labels: []string{},
					Index:  "1",
					Title:  "Adjust file",
				},
			},
		},
		{
			name:  "release events should handle release hook",
			data:  fixtures.HookRelease,
			event: "release",
			repo: &model.Repo{
				ForgeRemoteID: "77",
				Owner:         "anbraten",
				Name:          "demo",
				FullName:      "anbraten/demo",
				Avatar:        "https://git.xxx/user/avatar/anbraten/-1",
				ForgeURL:      "https://git.xxx/anbraten/demo",
				Clone:         "https://git.xxx/anbraten/demo.git",
				CloneSSH:      "ssh://git@git.xxx:22/anbraten/demo.git",
				Branch:        "main",
				PREnabled:     true,
				IsSCMPrivate:  true,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Event:        model.EventRelease,
				Branch:       "main",
				Ref:          "refs/tags/0.0.5",
				Release:      &model.Release{Title: "Version 0.0.5"},
				Author:       "anbraten",
				AuthorAvatar: "https://git.xxx/user/avatar/anbraten/-1",
				ForgeURL:     "https://git.xxx/anbraten/demo/releases/tag/0.0.5",
				TagTitle:     "0.0.5",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, "/api/hook", bytes.NewBufferString(tc.data))
			req.Header = http.Header{}
			req.Header.Set(hookEvent, tc.event)
			r, p, err := parseHook(req)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else if assert.NoError(t, err) {
				assert.EqualValues(t, tc.repo, r)
				assert.EqualValues(t, tc.pipe, p)
			}
		})
	}
}
