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

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote/gitea/fixtures"
)

func Test_parser(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("Gitea parser", func() {
		g.It("should ignore unsupported hook events", func() {
			buf := bytes.NewBufferString(fixtures.HookPullRequest)
			req, _ := http.NewRequest("POST", "/hook", buf)
			req.Header = http.Header{}
			req.Header.Set(hookEvent, "issues")
			r, b, err := parseHook(req)
			g.Assert(r == nil).IsTrue()
			g.Assert(b == nil).IsTrue()
			g.Assert(err == nil).IsTrue()
		})
		g.Describe("given a push hook", func() {
			g.It("should extract repository and build details", func() {
				buf := bytes.NewBufferString(fixtures.HookPush)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPush)
				r, b, err := parseHook(req)
				g.Assert(err == nil).IsTrue()
				g.Assert(r != nil).IsTrue()
				g.Assert(b != nil).IsTrue()
				g.Assert(b.Event).Equal(model.EventPush)
				g.Assert(b.ChangedFiles).Equal([]string{"CHANGELOG.md", "app/controller/application.rb"})
			})
		})
	})
}
