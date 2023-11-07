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
	"bytes"
	"net/http"
	"sort"
	"testing"

	"github.com/franela/goblin"
	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/server/forge/github/fixtures"
	"go.woodpecker-ci.org/woodpecker/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/server/model"
)

const (
	hookEvent  = "X-GitHub-Event"
	hookDeploy = "deployment"
	hookPush   = "push"
	hookPull   = "pull_request"
)

func testHookRequest(payload []byte, event string) *http.Request {
	buf := bytes.NewBuffer(payload)
	req, _ := http.NewRequest("POST", "/hook", buf)
	req.Header = http.Header{}
	req.Header.Set(hookEvent, event)
	return req
}

func Test_parser(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("GitHub parser", func() {
		g.It("should ignore unsupported hook events", func() {
			req := testHookRequest([]byte(fixtures.HookPullRequest), "issues")
			p, r, b, err := parseHook(req, false)
			g.Assert(r).IsNil()
			g.Assert(b).IsNil()
			g.Assert(p).IsNil()
			assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
		})

		g.Describe("given a push hook", func() {
			g.It("should skip when action is deleted", func() {
				req := testHookRequest([]byte(fixtures.HookPushDeleted), hookPush)
				p, r, b, err := parseHook(req, false)
				g.Assert(r).IsNil()
				g.Assert(b).IsNil()
				g.Assert(err).IsNil()
				g.Assert(p).IsNil()
			})
			g.It("should extract repository and pipeline details", func() {
				req := testHookRequest([]byte(fixtures.HookPush), hookPush)
				p, r, b, err := parseHook(req, false)
				g.Assert(err).IsNil()
				g.Assert(p).IsNil()
				g.Assert(r).IsNotNil()
				g.Assert(b).IsNotNil()
				g.Assert(b.Event).Equal(model.EventPush)
				sort.Strings(b.ChangedFiles)
				g.Assert(b.ChangedFiles).Equal([]string{"pipeline/shared/replace_secrets.go", "pipeline/shared/replace_secrets_test.go"})
			})
		})

		g.Describe("given a pull request hook", func() {
			g.It("should skip when action is not open or sync", func() {
				req := testHookRequest([]byte(fixtures.HookPullRequestInvalidAction), hookPull)
				p, r, b, err := parseHook(req, false)
				g.Assert(r).IsNil()
				g.Assert(b).IsNil()
				g.Assert(err).IsNil()
				g.Assert(p).IsNil()
			})
			g.It("should skip when state is not open", func() {
				req := testHookRequest([]byte(fixtures.HookPullRequestInvalidState), hookPull)
				p, r, b, err := parseHook(req, false)
				g.Assert(r).IsNil()
				g.Assert(b).IsNil()
				g.Assert(err).IsNil()
				g.Assert(p).IsNil()
			})
			g.It("should extract repository and pipeline details", func() {
				req := testHookRequest([]byte(fixtures.HookPullRequest), hookPull)
				p, r, b, err := parseHook(req, false)
				g.Assert(err).IsNil()
				g.Assert(r).IsNotNil()
				g.Assert(b).IsNotNil()
				g.Assert(p).IsNotNil()
				g.Assert(b.Event).Equal(model.EventPull)
			})
		})

		g.Describe("given a deployment hook", func() {
			g.It("should extract repository and pipeline details", func() {
				req := testHookRequest([]byte(fixtures.HookDeploy), hookDeploy)
				p, r, b, err := parseHook(req, false)
				g.Assert(err).IsNil()
				g.Assert(r).IsNotNil()
				g.Assert(b).IsNotNil()
				g.Assert(p).IsNil()
				g.Assert(b.Event).Equal(model.EventDeploy)
			})
		})
	})
}
