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

	"github.com/franela/goblin"
	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge/gitea/fixtures"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func TestGiteaParser(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("Gitea parser", func() {
		g.It("should ignore unsupported hook events", func() {
			buf := bytes.NewBufferString(fixtures.HookPullRequest)
			req, _ := http.NewRequest("POST", "/hook", buf)
			req.Header = http.Header{}
			req.Header.Set(hookEvent, "issues")
			r, b, err := parseHook(req)
			g.Assert(r).IsNil()
			g.Assert(b).IsNil()
			assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
		})

		g.Describe("push event", func() {
			g.It("should handle a push hook", func() {
				buf := bytes.NewBufferString(fixtures.HookPushBranch)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPush)
				r, p, err := parseHook(req)
				if assert.NoError(t, err) {
					assert.EqualValues(t, &model.Repo{
						ForgeRemoteID: "50820",
						Owner:         "meisam",
						Name:          "woodpecktester",
						FullName:      "meisam/woodpecktester",
						Avatar:        "https://codeberg.org/avatars/96512da76a14cf44e0bb32d1640e878e",
						ForgeURL:      "https://codeberg.org/meisam/woodpecktester",
						Clone:         "https://codeberg.org/meisam/woodpecktester.git",
						CloneSSH:      "git@codeberg.org:meisam/woodpecktester.git",
						Branch:        "main",
						SCMKind:       "git",
						PREnabled:     true,
						Perm: &model.Perm{
							Pull:  true,
							Push:  true,
							Admin: true,
						},
					}, r)
					p.Timestamp = 0
					assert.EqualValues(t, &model.Pipeline{
						Author:       "6543",
						Event:        "push",
						Commit:       "28c3613ae62640216bea5e7dc71aa65356e4298b",
						Branch:       "fdsafdsa",
						Ref:          "refs/heads/fdsafdsa",
						Message:      "Delete '.woodpecker/.check.yml'\n",
						Sender:       "6543",
						Avatar:       "https://codeberg.org/avatars/09a234c768cb9bca78f6b2f82d6af173",
						Email:        "6543@obermui.de",
						ForgeURL:     "https://codeberg.org/meisam/woodpecktester/commit/28c3613ae62640216bea5e7dc71aa65356e4298b",
						ChangedFiles: []string{".woodpecker/.check.yml"},
					}, p)
				}
			})

			g.It("should extract repository and pipeline details", func() {
				buf := bytes.NewBufferString(fixtures.HookPush)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPush)
				r, p, err := parseHook(req)
				if assert.NoError(t, err) {
					assert.EqualValues(t, &model.Repo{
						ForgeRemoteID: "1",
						Owner:         "gordon",
						Name:          "hello-world",
						FullName:      "gordon/hello-world",
						Avatar:        "http://gitea.golang.org/gordon/hello-world",
						ForgeURL:      "http://gitea.golang.org/gordon/hello-world",
						Clone:         "http://gitea.golang.org/gordon/hello-world.git",
						CloneSSH:      "git@gitea.golang.org:gordon/hello-world.git",
						SCMKind:       "git",
						IsSCMPrivate:  true,
						Perm: &model.Perm{
							Pull:  true,
							Push:  true,
							Admin: true,
						},
					}, r)
					p.Timestamp = 0
					assert.EqualValues(t, &model.Pipeline{
						Author:       "gordon",
						Event:        "push",
						Commit:       "ef98532add3b2feb7a137426bba1248724367df5",
						Branch:       "main",
						Ref:          "refs/heads/main",
						Message:      "bump\n",
						Sender:       "gordon",
						Avatar:       "http://1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
						Email:        "gordon@golang.org",
						ForgeURL:     "http://gitea.golang.org/gordon/hello-world/commit/ef98532add3b2feb7a137426bba1248724367df5",
						ChangedFiles: []string{"CHANGELOG.md", "app/controller/application.rb"},
					}, p)
				}
			})

			g.It("should handle multi commit push", func() {
				buf := bytes.NewBufferString(fixtures.HookPushMulti)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPush)
				r, p, err := parseHook(req)
				if assert.NoError(t, err) {
					assert.EqualValues(t, &model.Repo{
						ForgeRemoteID: "6",
						Owner:         "Test-CI",
						Name:          "multi-line-secrets",
						FullName:      "Test-CI/multi-line-secrets",
						Avatar:        "http://127.0.0.1:3000/avatars/5b0a83c2185b3cb1ebceb11062d6c2eb",
						ForgeURL:      "http://127.0.0.1:3000/Test-CI/multi-line-secrets",
						Clone:         "http://127.0.0.1:3000/Test-CI/multi-line-secrets.git",
						CloneSSH:      "ssh://git@127.0.0.1:2200/Test-CI/multi-line-secrets.git",
						Branch:        "main",
						SCMKind:       "git",
						Perm: &model.Perm{
							Pull:  true,
							Push:  true,
							Admin: true,
						},
					}, r)
					p.Timestamp = 0
					assert.EqualValues(t, &model.Pipeline{
						Author:       "test-user",
						Event:        "push",
						Commit:       "29be01c073851cf0db0c6a466e396b725a670453",
						Branch:       "main",
						Ref:          "refs/heads/main",
						Message:      "add some text\n",
						Sender:       "test-user",
						Avatar:       "http://127.0.0.1:3000/avatars/dd46a756faad4727fb679320751f6dea",
						Email:        "test@noreply.localhost",
						ForgeURL:     "http://127.0.0.1:3000/Test-CI/multi-line-secrets/compare/6efcf5b7c98f3e7a491675164b7a2e7acac27941...29be01c073851cf0db0c6a466e396b725a670453",
						ChangedFiles: []string{"aaa", "aa"},
					}, p)
				}
			})
		})

		g.Describe("tag event", func() {
			g.It("should handle a tag hook", func() {
				buf := bytes.NewBufferString(fixtures.HookTag)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookCreated)
				r, b, err := parseHook(req)
				g.Assert(r).IsNotNil()
				g.Assert(b).IsNotNil()
				g.Assert(err).IsNil()
				g.Assert(b.Event).Equal(model.EventTag)
				assert.EqualValues(t, "gordon@golang.org", b.Email)
			})
		})

		g.Describe("pull-request events", func() {
			// g.It("should handle a PR hook when PR got created")

			g.It("should handle a PR hook when PR got updated", func() {
				buf := bytes.NewBufferString(fixtures.HookPullRequest)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPullRequest)
				r, p, err := parseHook(req)
				if assert.NoError(t, err) {
					assert.EqualValues(t, &model.Repo{
						ForgeRemoteID: "35129377",
						Owner:         "gordon",
						Name:          "hello-world",
						FullName:      "gordon/hello-world",
						Avatar:        "https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
						ForgeURL:      "http://gitea.golang.org/gordon/hello-world",
						Clone:         "https://gitea.golang.org/gordon/hello-world.git",
						CloneSSH:      "",
						Branch:        "main",
						SCMKind:       "git",
						IsSCMPrivate:  true,
						Perm: &model.Perm{
							Pull:  true,
							Push:  true,
							Admin: true,
						},
					}, r)
					p.Timestamp = 0
					assert.EqualValues(t, &model.Pipeline{
						Author:            "gordon",
						Event:             "pull_request",
						Commit:            "0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c",
						Branch:            "main",
						Ref:               "refs/pull/1/head",
						Refspec:           "feature/changes:main",
						Title:             "Update the README with new information",
						Message:           "Update the README with new information",
						Sender:            "gordon",
						Avatar:            "http://1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
						Email:             "gordon@golang.org",
						ForgeURL:          "http://gitea.golang.org/gordon/hello-world/pull/1",
						PullRequestLabels: []string{},
					}, p)
				}
			})

			g.It("should handle a PR closed hook when PR got closed", func() {
				buf := bytes.NewBufferString(fixtures.HookPullRequestClosed)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPullRequest)
				r, p, err := parseHook(req)
				if assert.NoError(t, err) {
					assert.EqualValues(t, &model.Repo{
						ForgeRemoteID: "46534",
						Owner:         "anbraten",
						Name:          "test-repo",
						FullName:      "anbraten/test-repo",
						Avatar:        "https://seccdn.libravatar.org/avatar/fc9b6fe77c6b732a02925a62a81f05a0?d=identicon",
						ForgeURL:      "https://gitea.com/anbraten/test-repo",
						Clone:         "https://gitea.com/anbraten/test-repo.git",
						CloneSSH:      "git@gitea.com:anbraten/test-repo.git",
						Branch:        "main",
						SCMKind:       "git",
						PREnabled:     true,
						Perm: &model.Perm{
							Pull:  true,
							Push:  true,
							Admin: true,
						},
					}, r)
					p.Timestamp = 0
					assert.EqualValues(t, &model.Pipeline{
						Author:            "anbraten",
						Event:             "pull_request_closed",
						Commit:            "d555a5dd07f4d0148a58d4686ec381502ae6a2d4",
						Branch:            "main",
						Ref:               "refs/pull/1/head",
						Refspec:           "anbraten-patch-1:main",
						Title:             "Adjust file",
						Message:           "Adjust file",
						Sender:            "anbraten",
						Avatar:            "https://seccdn.libravatar.org/avatar/fc9b6fe77c6b732a02925a62a81f05a0?d=identicon",
						Email:             "anbraten@sender.gitea.com",
						ForgeURL:          "https://gitea.com/anbraten/test-repo/pulls/1",
						PullRequestLabels: []string{},
					}, p)
				}
			})

			g.It("should handle a PR closed hook when PR was merged", func() {
				buf := bytes.NewBufferString(fixtures.HookPullRequestMerged)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPullRequest)
				r, p, err := parseHook(req)
				if assert.NoError(t, err) {
					assert.EqualValues(t, &model.Repo{
						ForgeRemoteID: "46534",
						Owner:         "anbraten",
						Name:          "test-repo",
						FullName:      "anbraten/test-repo",
						Avatar:        "https://seccdn.libravatar.org/avatar/fc9b6fe77c6b732a02925a62a81f05a0?d=identicon",
						ForgeURL:      "https://gitea.com/anbraten/test-repo",
						Clone:         "https://gitea.com/anbraten/test-repo.git",
						CloneSSH:      "git@gitea.com:anbraten/test-repo.git",
						Branch:        "main",
						SCMKind:       "git",
						PREnabled:     true,
						Perm: &model.Perm{
							Pull:  true,
							Push:  true,
							Admin: true,
						},
					}, r)
					p.Timestamp = 0
					assert.EqualValues(t, &model.Pipeline{
						Author:            "anbraten",
						Event:             "pull_request_closed",
						Commit:            "d555a5dd07f4d0148a58d4686ec381502ae6a2d4",
						Branch:            "main",
						Ref:               "refs/pull/1/head",
						Refspec:           "anbraten-patch-1:main",
						Title:             "Adjust file",
						Message:           "Adjust file",
						Sender:            "anbraten",
						Avatar:            "https://seccdn.libravatar.org/avatar/fc9b6fe77c6b732a02925a62a81f05a0?d=identicon",
						Email:             "anbraten@noreply.gitea.com",
						ForgeURL:          "https://gitea.com/anbraten/test-repo/pulls/1",
						PullRequestLabels: []string{},
					}, p)
				}
			})

			g.It("should handle release hook", func() {
				buf := bytes.NewBufferString(fixtures.HookRelease)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookRelease)
				r, p, err := parseHook(req)
				if assert.NoError(t, err) {
					assert.EqualValues(t, &model.Repo{
						ForgeRemoteID: "77",
						Owner:         "anbraten",
						Name:          "demo",
						FullName:      "anbraten/demo",
						Avatar:        "https://git.xxx/user/avatar/anbraten/-1",
						ForgeURL:      "https://git.xxx/anbraten/demo",
						Clone:         "https://git.xxx/anbraten/demo.git",
						CloneSSH:      "ssh://git@git.xxx:22/anbraten/demo.git",
						Branch:        "main",
						SCMKind:       "git",
						PREnabled:     true,
						IsSCMPrivate:  true,
						Perm: &model.Perm{
							Pull:  true,
							Push:  true,
							Admin: true,
						},
					}, r)
					p.Timestamp = 0
					assert.EqualValues(t, &model.Pipeline{
						Author:   "anbraten",
						Event:    "release",
						Branch:   "main",
						Ref:      "refs/tags/0.0.5",
						Message:  "created release Version 0.0.5",
						Sender:   "anbraten",
						Avatar:   "https://git.xxx/user/avatar/anbraten/-1",
						Email:    "anbraten@noreply.xxx",
						ForgeURL: "https://git.xxx/anbraten/demo/releases/tag/0.0.5",
					}, p)
				}
			})
		})
	})
}
