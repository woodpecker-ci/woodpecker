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

package bitbucket

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/franela/goblin"

	"github.com/woodpecker-ci/woodpecker/server/forge/bitbucket/fixtures"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func Test_parser(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("Bitbucket parser", func() {
		g.It("Should ignore unsupported hook", func() {
			buf := bytes.NewBufferString(fixtures.HookPush)
			req, _ := http.NewRequest("POST", "/hook", buf)
			req.Header = http.Header{}
			req.Header.Set(hookEvent, "issue:created")

			r, b, err := parseHook(req)
			g.Assert(r).IsNil()
			g.Assert(b).IsNil()
			g.Assert(err).IsNil()
		})

		g.Describe("Given a pull request hook payload", func() {
			g.It("Should return err when malformed", func() {
				buf := bytes.NewBufferString("[]")
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPullCreated)

				_, _, err := parseHook(req)
				g.Assert(err).IsNotNil()
			})

			g.It("Should return nil if not open", func() {
				buf := bytes.NewBufferString(fixtures.HookMerged)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPullCreated)

				r, b, err := parseHook(req)
				g.Assert(r).IsNil()
				g.Assert(b).IsNil()
				g.Assert(err).IsNil()
			})

			g.It("Should return pull request details", func() {
				buf := bytes.NewBufferString(fixtures.HookPull)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPullCreated)

				r, b, err := parseHook(req)
				g.Assert(err).IsNil()
				g.Assert(r.FullName).Equal("user_name/repo_name")
				g.Assert(b.Commit).Equal("ce5965ddd289")
			})
		})

		g.Describe("Given a push hook payload", func() {
			g.It("Should return err when malformed", func() {
				buf := bytes.NewBufferString("[]")
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPush)

				_, _, err := parseHook(req)
				g.Assert(err).IsNotNil()
			})

			g.It("Should return nil if missing commit sha", func() {
				buf := bytes.NewBufferString(fixtures.HookPushEmptyHash)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPush)

				r, b, err := parseHook(req)
				g.Assert(r).IsNil()
				g.Assert(b).IsNil()
				g.Assert(err).IsNil()
			})

			g.It("Should return push details", func() {
				buf := bytes.NewBufferString(fixtures.HookPush)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPush)

				r, b, err := parseHook(req)
				g.Assert(err).IsNil()
				g.Assert(r.FullName).Equal("martinherren1984/publictestrepo")
				g.Assert(r.SCMKind).Equal(model.RepoGit)
				g.Assert(r.Clone).Equal("https://bitbucket.org/martinherren1984/publictestrepo")
				g.Assert(b.Commit).Equal("c14c1bb05dfb1fdcdf06b31485fff61b0ea44277")
				g.Assert(b.Message).Equal("a\n")
			})
		})
	})
}
