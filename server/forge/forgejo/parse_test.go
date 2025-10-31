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

package forgejo

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/forgejo/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestForgejoParser(t *testing.T) {
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
					SHA:     "28c3613ae62640216bea5e7dc71aa65356e4298b",
					Message: "Delete '.woodpecker/.check.yml'\n",
					Author: model.CommitAuthor{
						Name:  "meisam",
						Email: "meisam@noreply.codeberg.org",
					},
					ForgeURL: "https://codeberg.org/meisam/woodpecktester/commit/28c3613ae62640216bea5e7dc71aa65356e4298b",
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
				Avatar:        "http://forgejo.golang.org/gordon/hello-world",
				ForgeURL:      "http://forgejo.golang.org/gordon/hello-world",
				Clone:         "http://forgejo.golang.org/gordon/hello-world.git",
				CloneSSH:      "git@forgejo.golang.org:gordon/hello-world.git",
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
					SHA: "ef98532add3b2feb7a137426bba1248724367df5",
					Author: model.CommitAuthor{
						Name:  "Gordon the Gopher",
						Email: "gordon@golang.org",
					},
					ForgeURL: "http://forgejo.golang.org/gordon/hello-world/commit/ef98532add3b2feb7a137426bba1248724367df5",
					Message:  "bump\n",
				},
				Branch:       "main",
				Ref:          "refs/heads/main",
				Author:       "gordon",
				AuthorAvatar: "http://1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
				ForgeURL:     "http://forgejo.golang.org/gordon/hello-world/commit/ef98532add3b2feb7a137426bba1248724367df5",
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
				ForgeURL:      "http://forgejo.golang.org/gordon/hello-world",
				Clone:         "http://forgejo.golang.org/gordon/hello-world.git",
				CloneSSH:      "git@forgejo.golang.org:gordon/hello-world.git",
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
				AuthorAvatar: "https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
				ForgeURL:     "http://forgejo.golang.org/gordon/hello-world/releases/tag/v1.0.0",
				TagTitle:     "v1.0.0",
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
				ForgeURL:      "http://forgejo.golang.org/gordon/hello-world",
				Clone:         "https://forgejo.golang.org/gordon/hello-world.git",
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
				ForgeURL:     "http://forgejo.golang.org/gordon/hello-world/pull/1",
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
				CloneSSH:      "ssh://git@codeberg.org/test_it/test_ci_thing.git",
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
			name:  "pull-request events should handle a PR edited hook when PR got edited",
			data:  fixtures.HookPullRequestEdited,
			event: "pull_request",
			repo: &model.Repo{
				ForgeRemoteID: "46534",
				Owner:         "anbraten",
				Name:          "test-repo",
				FullName:      "anbraten/test-repo",
				Avatar:        "https://seccdn.libravatar.org/avatar/fc9b6fe77c6b732a02925a62a81f05a0?d=identicon",
				ForgeURL:      "https://forgejo.com/anbraten/test-repo",
				Clone:         "https://forgejo.com/anbraten/test-repo.git",
				CloneSSH:      "git@forgejo.com:anbraten/test-repo.git",
				Branch:        "main",
				PREnabled:     true,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Author:      "anbraten",
				Event:       "pull_request_metadata",
				EventReason: []string{"edited"},
				Commit: &model.Commit{
					SHA: "d555a5dd07f4d0148a58d4686ec381502ae6a2d4",
				},
				Branch:  "main",
				Ref:     "refs/pull/1/head",
				Refspec: "anbraten-patch-1:main",
				PullRequest: &model.PullRequest{
					Title:  "Adjust file",
					Labels: []string{},
					Index:  "1",
				},
				AuthorAvatar: "https://seccdn.libravatar.org/avatar/fc9b6fe77c6b732a02925a62a81f05a0?d=identicon",
				ForgeURL:     "https://forgejo.com/anbraten/test-repo/pulls/1",
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
				ForgeURL:      "https://forgejo.com/anbraten/test-repo",
				Clone:         "https://forgejo.com/anbraten/test-repo.git",
				CloneSSH:      "git@forgejo.com:anbraten/test-repo.git",
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
				ForgeURL:     "https://forgejo.com/anbraten/test-repo/pulls/1",
				PullRequest: &model.PullRequest{
					Labels: []string{},
					Index:  "1",
					Title:  "Adjust file",
				},
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
				ForgeURL:      "https://forgejo.com/anbraten/test-repo",
				Clone:         "https://forgejo.com/anbraten/test-repo.git",
				CloneSSH:      "git@forgejo.com:anbraten/test-repo.git",
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
				ForgeURL:     "https://forgejo.com/anbraten/test-repo/pulls/1",
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
				Author:       "anbraten",
				AuthorAvatar: "https://git.xxx/user/avatar/anbraten/-1",
				ForgeURL:     "https://git.xxx/anbraten/demo/releases/tag/0.0.5",
				Release:      &model.Release{Title: "Version 0.0.5"},
				TagTitle:     "0.0.5",
			},
		},
		{
			name:  "pull-request events should handle a PR assignees added hook when assignees are added",
			data:  fixtures.HookPullRequestAssigneesAdded,
			event: "pull_request",
			repo: &model.Repo{
				ForgeRemoteID: "138564",
				Owner:         "test_it",
				Name:          "test_ci_thing",
				FullName:      "test_it/test_ci_thing",
				Avatar:        "https://codeberg.org/avatars/bb6f3159a98a869b43f20b350542f8fb",
				ForgeURL:      "https://codeberg.org/test_it/test_ci_thing",
				Clone:         "https://codeberg.org/test_it/test_ci_thing.git",
				CloneSSH:      "ssh://git@codeberg.org/test_it/test_ci_thing.git",
				Branch:        "main",
				PREnabled:     true,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Author:      "6543",
				Event:       "pull_request_metadata",
				EventReason: []string{"assigned"},
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
				ChangedFiles: nil,
			},
		},
		{
			name:  "pull-request events should handle a PR milestone added hook when milestone is added",
			data:  fixtures.HookPullRequestMilestoneAdded,
			event: "pull_request",
			repo: &model.Repo{
				ForgeRemoteID: "138564",
				Owner:         "test_it",
				Name:          "test_ci_thing",
				FullName:      "test_it/test_ci_thing",
				Avatar:        "https://codeberg.org/avatars/bb6f3159a98a869b43f20b350542f8fb",
				ForgeURL:      "https://codeberg.org/test_it/test_ci_thing",
				Clone:         "https://codeberg.org/test_it/test_ci_thing.git",
				CloneSSH:      "ssh://git@codeberg.org/test_it/test_ci_thing.git",
				Branch:        "main",
				PREnabled:     true,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Author:      "6543",
				Event:       "pull_request_metadata",
				EventReason: []string{"milestoned"},
				Commit: &model.Commit{
					SHA: "36b5813240a9d2daa29b05046d56a53e18f39a3e",
				},
				Branch:  "main",
				Ref:     "refs/pull/1/head",
				Refspec: "6543-patch-1:main",
				PullRequest: &model.PullRequest{
					Title: "Some ned more AAAA", Index: "1",
					Labels:    []string{},
					Milestone: "mile v2",
				},
				AuthorAvatar: "https://codeberg.org/avatars/09a234c768cb9bca78f6b2f82d6af173",
				ForgeURL:     "https://codeberg.org/test_it/test_ci_thing/pulls/1",
				ChangedFiles: nil,
			},
		},
		{
			name:  "pull-request events should handle a PR label updated hook when labels are updated",
			data:  fixtures.HookPullRequestLabelAdded,
			event: "pull_request",
			repo: &model.Repo{
				ForgeRemoteID: "138564",
				Owner:         "test_it",
				Name:          "test_ci_thing",
				FullName:      "test_it/test_ci_thing",
				Avatar:        "https://codeberg.org/avatars/bb6f3159a98a869b43f20b350542f8fb",
				ForgeURL:      "https://codeberg.org/test_it/test_ci_thing",
				Clone:         "https://codeberg.org/test_it/test_ci_thing.git",
				CloneSSH:      "ssh://git@codeberg.org/test_it/test_ci_thing.git",
				Branch:        "main",
				PREnabled:     true,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Author:      "6543",
				Event:       "pull_request_metadata",
				EventReason: []string{"label_updated"},
				Commit: &model.Commit{
					SHA: "36b5813240a9d2daa29b05046d56a53e18f39a3e",
				},
				Branch:  "main",
				Ref:     "refs/pull/1/head",
				Refspec: "6543-patch-1:main",
				PullRequest: &model.PullRequest{
					Title: "Some ned more AAAA", Index: "1",
					Labels:    []string{"Kind/Documentation", "Kind/Enhancement"},
					Milestone: "mile v2",
				},
				AuthorAvatar: "https://codeberg.org/avatars/09a234c768cb9bca78f6b2f82d6af173",
				ForgeURL:     "https://codeberg.org/test_it/test_ci_thing/pulls/1",
				ChangedFiles: nil,
			},
		},
		{
			name:  "pull-request events should handle a PR assignee cleared hook when assignee is removed",
			data:  fixtures.HookPullRequestAssigneeCleared,
			event: "pull_request",
			repo: &model.Repo{
				ForgeRemoteID: "138564",
				Owner:         "test_it",
				Name:          "test_ci_thing",
				FullName:      "test_it/test_ci_thing",
				Avatar:        "https://codeberg.org/avatars/bb6f3159a98a869b43f20b350542f8fb",
				ForgeURL:      "https://codeberg.org/test_it/test_ci_thing",
				Clone:         "https://codeberg.org/test_it/test_ci_thing.git",
				CloneSSH:      "ssh://git@codeberg.org/test_it/test_ci_thing.git",
				Branch:        "main",
				PREnabled:     true,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Author:      "6543",
				Event:       "pull_request_metadata",
				EventReason: []string{"unassigned"},
				Commit: &model.Commit{
					SHA: "36b5813240a9d2daa29b05046d56a53e18f39a3e",
				},
				Branch:  "main",
				Ref:     "refs/pull/1/head",
				Refspec: "6543-patch-1:main",
				PullRequest: &model.PullRequest{
					Title: "Some ned more AAAA", Index: "1",
					Labels:    []string{"Kind/Documentation", "Kind/Enhancement"},
					Milestone: "mile v2",
				},
				AuthorAvatar: "https://codeberg.org/avatars/09a234c768cb9bca78f6b2f82d6af173",
				ForgeURL:     "https://codeberg.org/test_it/test_ci_thing/pulls/1",
				ChangedFiles: nil,
			},
		},
		{
			name:  "pull-request events should handle a PR milestone changed hook when milestone is changed",
			data:  fixtures.HookPullRequestMilestoneChanged,
			event: "pull_request",
			repo: &model.Repo{
				ForgeRemoteID: "138564",
				Owner:         "test_it",
				Name:          "test_ci_thing",
				FullName:      "test_it/test_ci_thing",
				Avatar:        "https://codeberg.org/avatars/bb6f3159a98a869b43f20b350542f8fb",
				ForgeURL:      "https://codeberg.org/test_it/test_ci_thing",
				Clone:         "https://codeberg.org/test_it/test_ci_thing.git",
				CloneSSH:      "ssh://git@codeberg.org/test_it/test_ci_thing.git",
				Branch:        "main",
				PREnabled:     true,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Author:      "6543",
				Event:       "pull_request_metadata",
				EventReason: []string{"milestoned"},
				Commit: &model.Commit{
					SHA: "36b5813240a9d2daa29b05046d56a53e18f39a3e",
				},
				Branch:  "main",
				Ref:     "refs/pull/1/head",
				Refspec: "6543-patch-1:main",
				PullRequest: &model.PullRequest{
					Title: "Some ned more AAAA", Index: "1",
					Labels:    []string{"Kind/Documentation", "Kind/Enhancement"},
					Milestone: "mile v2",
				},
				AuthorAvatar: "https://codeberg.org/avatars/09a234c768cb9bca78f6b2f82d6af173",
				ForgeURL:     "https://codeberg.org/test_it/test_ci_thing/pulls/1",
				ChangedFiles: nil,
			},
		},
		{
			name:  "pull-request events should handle a PR labels updated hook when labels are updated",
			data:  fixtures.HookPullRequestLabelsUpdated,
			event: "pull_request",
			repo: &model.Repo{
				ForgeRemoteID: "138564",
				Owner:         "test_it",
				Name:          "test_ci_thing",
				FullName:      "test_it/test_ci_thing",
				Avatar:        "https://codeberg.org/avatars/bb6f3159a98a869b43f20b350542f8fb",
				ForgeURL:      "https://codeberg.org/test_it/test_ci_thing",
				Clone:         "https://codeberg.org/test_it/test_ci_thing.git",
				CloneSSH:      "ssh://git@codeberg.org/test_it/test_ci_thing.git",
				Branch:        "main",
				PREnabled:     true,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Author:      "6543",
				Event:       "pull_request_metadata",
				EventReason: []string{"label_updated"},
				Commit: &model.Commit{
					SHA: "36b5813240a9d2daa29b05046d56a53e18f39a3e",
				},
				Branch:  "main",
				Ref:     "refs/pull/1/head",
				Refspec: "6543-patch-1:main",
				PullRequest: &model.PullRequest{
					Title: "Some ned more AAAA", Index: "1",
					Labels:    []string{"Kind/Enhancement", "Kind/Testing"},
					Milestone: "mile v1",
				},
				AuthorAvatar: "https://codeberg.org/avatars/09a234c768cb9bca78f6b2f82d6af173",
				ForgeURL:     "https://codeberg.org/test_it/test_ci_thing/pulls/1",
				ChangedFiles: nil,
			},
		},
		{
			name:  "pull-request events should handle a PR labels cleared hook when labels are cleared",
			data:  fixtures.HookPullRequestLabelsCleared,
			event: "pull_request",
			repo: &model.Repo{
				ForgeRemoteID: "138564",
				Owner:         "test_it",
				Name:          "test_ci_thing",
				FullName:      "test_it/test_ci_thing",
				Avatar:        "https://codeberg.org/avatars/bb6f3159a98a869b43f20b350542f8fb",
				ForgeURL:      "https://codeberg.org/test_it/test_ci_thing",
				Clone:         "https://codeberg.org/test_it/test_ci_thing.git",
				CloneSSH:      "ssh://git@codeberg.org/test_it/test_ci_thing.git",
				Branch:        "main",
				PREnabled:     true,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Author:      "6543",
				Event:       "pull_request_metadata",
				EventReason: []string{"label_cleared"},
				Commit: &model.Commit{
					SHA: "36b5813240a9d2daa29b05046d56a53e18f39a3e",
				},
				Branch:  "main",
				Ref:     "refs/pull/1/head",
				Refspec: "6543-patch-1:main",
				PullRequest: &model.PullRequest{
					Title: "Some ned more AAAA", Index: "1",
					Labels:    []string{},
					Milestone: "mile v1",
				},
				AuthorAvatar: "https://codeberg.org/avatars/09a234c768cb9bca78f6b2f82d6af173",
				ForgeURL:     "https://codeberg.org/test_it/test_ci_thing/pulls/1",
				ChangedFiles: nil,
			},
		},
		{
			name:  "pull-request events should handle a PR milestone cleared hook when milestone is removed",
			data:  fixtures.HookPullRequestMilestoneCleared,
			event: "pull_request",
			repo: &model.Repo{
				ForgeRemoteID: "138564",
				Owner:         "test_it",
				Name:          "test_ci_thing",
				FullName:      "test_it/test_ci_thing",
				Avatar:        "https://codeberg.org/avatars/bb6f3159a98a869b43f20b350542f8fb",
				ForgeURL:      "https://codeberg.org/test_it/test_ci_thing",
				Clone:         "https://codeberg.org/test_it/test_ci_thing.git",
				CloneSSH:      "ssh://git@codeberg.org/test_it/test_ci_thing.git",
				Branch:        "main",
				PREnabled:     true,
				Perm: &model.Perm{
					Pull:  true,
					Push:  true,
					Admin: true,
				},
			},
			pipe: &model.Pipeline{
				Author:      "6543",
				Event:       "pull_request_metadata",
				EventReason: []string{"demilestoned"},
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
				ChangedFiles: nil,
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
