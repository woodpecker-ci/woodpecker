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

package ccmenu

import (
	"testing"
	"time"

	"github.com/franela/goblin"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestCC(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("CC", func() {
		g.It("Should create a project", func() {
			now := time.Now().Unix()
			nowFmt := time.Unix(now, 0).Format(time.RFC3339)
			r := &model.Repo{
				FullName: "foo/bar",
			}
			b := &model.Pipeline{
				Status:  model.StatusSuccess,
				Number:  1,
				Started: now,
			}
			cc := New(r, b, "http://localhost/foo/bar/1")

			g.Assert(cc.Project.Name).Equal("foo/bar")
			g.Assert(cc.Project.Activity).Equal("Sleeping")
			g.Assert(cc.Project.LastBuildStatus).Equal("Success")
			g.Assert(cc.Project.LastBuildLabel).Equal("1")
			g.Assert(cc.Project.LastBuildTime).Equal(nowFmt)
			g.Assert(cc.Project.WebURL).Equal("http://localhost/foo/bar/1")
		})

		g.It("Should properly label exceptions", func() {
			r := &model.Repo{FullName: "foo/bar"}
			b := &model.Pipeline{
				Status:  model.StatusError,
				Number:  1,
				Started: 1257894000,
			}
			cc := New(r, b, "http://localhost/foo/bar/1")
			g.Assert(cc.Project.LastBuildStatus).Equal("Exception")
			g.Assert(cc.Project.Activity).Equal("Sleeping")
		})

		g.It("Should properly label success", func() {
			r := &model.Repo{FullName: "foo/bar"}
			b := &model.Pipeline{
				Status:  model.StatusSuccess,
				Number:  1,
				Started: 1257894000,
			}
			cc := New(r, b, "http://localhost/foo/bar/1")
			g.Assert(cc.Project.LastBuildStatus).Equal("Success")
			g.Assert(cc.Project.Activity).Equal("Sleeping")
		})

		g.It("Should properly label failure", func() {
			r := &model.Repo{FullName: "foo/bar"}
			b := &model.Pipeline{
				Status:  model.StatusFailure,
				Number:  1,
				Started: 1257894000,
			}
			cc := New(r, b, "http://localhost/foo/bar/1")
			g.Assert(cc.Project.LastBuildStatus).Equal("Failure")
			g.Assert(cc.Project.Activity).Equal("Sleeping")
		})

		g.It("Should properly label running", func() {
			r := &model.Repo{FullName: "foo/bar"}
			b := &model.Pipeline{
				Status:  model.StatusRunning,
				Number:  1,
				Started: 1257894000,
			}
			cc := New(r, b, "http://localhost/foo/bar/1")
			g.Assert(cc.Project.Activity).Equal("Building")
			g.Assert(cc.Project.LastBuildStatus).Equal("Unknown")
			g.Assert(cc.Project.LastBuildLabel).Equal("Unknown")
		})
	})
}
