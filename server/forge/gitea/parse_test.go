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
	"go.woodpecker-ci.org/woodpecker/v2/shared/utils"
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
						ForgeURL:     "https://codeberg.org/meisam/woodpecktester/compare/main...28c3613ae62640216bea5e7dc71aa65356e4298b",
						ChangedFiles: []string{".woodpecker/.check.yml"},
						Title:        "", // TODO!
					}, p)
				}
			})

			g.It("should extract repository and pipeline details", func() {
				buf := bytes.NewBufferString(fixtures.HookPush)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPush)
				r, b, err := parseHook(req)
				g.Assert(err).IsNil()
				g.Assert(r).IsNotNil()
				g.Assert(b).IsNotNil()
				g.Assert(b.Event).Equal(model.EventPush)
				assert.EqualValues(t, "gordon@golang.org", b.Email)
				g.Assert(utils.EqualSliceValues(b.ChangedFiles, []string{"CHANGELOG.md", "app/controller/application.rb"})).IsTrue()
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
				r, b, err := parseHook(req)
				g.Assert(r).IsNotNil()
				g.Assert(b).IsNotNil()
				assert.EqualValues(t, "gordon@golang.org", b.Email)
				g.Assert(err).IsNil()
				g.Assert(b.Event).Equal(model.EventPull)
			})

			g.It("should handle a PR closed hook when PR got closed", func() {
				buf := bytes.NewBufferString(fixtures.HookPullRequestClosed)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPullRequest)
				r, b, err := parseHook(req)
				g.Assert(r).IsNotNil()
				g.Assert(b).IsNotNil()
				assert.EqualValues(t, "anbraten@sender.gitea.com", b.Email)
				g.Assert(err).IsNil()
				g.Assert(b.Event).Equal(model.EventPullClosed)
			})

			g.It("should handle a PR closed hook when PR was merged", func() {
				buf := bytes.NewBufferString(fixtures.HookPullRequestMerged)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPullRequest)
				r, b, err := parseHook(req)
				g.Assert(r).IsNotNil()
				g.Assert(b).IsNotNil()
				g.Assert(err).IsNil()
				g.Assert(b.Event).Equal(model.EventPullClosed)
			})
			g.It("should handle release hook", func() {
				buf := bytes.NewBufferString(fixtures.HookRelease)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookRelease)
				r, b, err := parseHook(req)
				g.Assert(err).IsNil()
				g.Assert(r).IsNotNil()
				g.Assert(b).IsNotNil()
				g.Assert(b.Event).Equal(model.EventRelease)
			})
		})
	})
}
