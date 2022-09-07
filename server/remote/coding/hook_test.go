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

package coding

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/franela/goblin"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote/coding/fixtures"
)

func Test_hook(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("Coding hook", func() {
		g.It("Should parse hook", func() {
			reader := io.NopCloser(strings.NewReader(fixtures.PushHook))
			r := &http.Request{
				Header: map[string][]string{
					hookEvent: {hookPush},
				},
				Body: reader,
			}

			repo := &model.Repo{
				Owner:    "demo1",
				Name:     "test1",
				FullName: "demo1/test1",
				Link:     "https://coding.net/u/demo1/p/test1",
				Clone:    "https://git.coding.net/demo1/test1.git",
				SCMKind:  model.RepoGit,
			}

			//repo := &model.Repo{ID:0, UserID:0, RemoteID:"", Owner:"demo1", Name:"test1", FullName:"demo1/test1", Avatar:"", Link:"https://coding.net/u/demo1/p/test1", Clone:"https://git.coding.net/demo1/test1.git", Branch:"", SCMKind:"git", Timeout:0, Visibility:"", IsSCMPrivate:false, IsTrusted:false, IsStarred:false, IsGated:false, IsActive:false, AllowPull:false, Config:"", Hash:"", Perm:(*model.Perm)(nil), CancelPreviousPipelineEvents:[]model.WebhookEvent(nil)} does not equal &model.Repo{ID:0, UserID:0, RemoteID:"", Owner:"demo1", Name:"test1", FullName:"demo1/test1", Avatar:"", Link:"https://coding.net/u/demo1/p/test1", Clone:"", Branch:"", SCMKind:"git", Timeout:0, Visibility:"", IsSCMPrivate:false, IsTrusted:false, IsStarred:false, IsGated:false, IsActive:false, AllowPull:false, Config:"", Hash:"", Perm:(*model.Perm)(nil), CancelPreviousPipelineEvents:[]model.WebhookEvent(nil)}

			build := &model.Build{
				Event:   model.EventPush,
				Commit:  "5b9912a6ff272e9c93a4c44c278fe9b359ed1ab4",
				Ref:     "refs/heads/master",
				Link:    "https://coding.net/u/demo1/p/test1/git/commit/5b9912a6ff272e9c93a4c44c278fe9b359ed1ab4",
				Branch:  "master",
				Message: "new file .woodpecker.yml\n",
				Email:   "demo1@gmail.com",
				Avatar:  "/static/fruit_avatar/Fruit-20.png",
				Author:  "demo1",
				Remote:  "https://git.coding.net/demo1/test1.git",
			}

			actualRepo, actualBuild, err := parseHook(r)
			g.Assert(err).IsNil()
			g.Assert(actualRepo).Equal(repo)
			g.Assert(actualBuild).Equal(build)
		})

		g.It("Should find last commit", func() {
			commit1 := &Commit{SHA: "1234567890", Committer: &Committer{}}
			commit2 := &Commit{SHA: "abcdef1234", Committer: &Committer{}}
			commits := []*Commit{commit1, commit2}
			g.Assert(findLastCommit(commits, "abcdef1234")).Equal(commit2)
		})

		g.It("Should find last commit", func() {
			commit1 := &Commit{SHA: "1234567890", Committer: &Committer{}}
			commit2 := &Commit{SHA: "abcdef1234", Committer: &Committer{}}
			commits := []*Commit{commit1, commit2}
			emptyCommit := &Commit{Committer: &Committer{}}
			g.Assert(findLastCommit(commits, "00000000000")).Equal(emptyCommit)
		})

		g.It("Should convert repository", func() {
			repository := &Repository{
				Name:     "test_project",
				HTTPSURL: "https://git.coding.net/kelvin/test_project.git",
				SSHURL:   "git@git.coding.net:kelvin/test_project.git",
				WebURL:   "https://coding.net/u/kelvin/p/test_project",
				Owner: &User{
					GlobalKey: "kelvin",
					Avatar:    "https://dn-coding-net-production-static.qbox.me/9ed11de3-65e3-4cd8-b6aa-5abe7285ab43.jpeg?imageMogr2/auto-orient/format/jpeg/crop/!209x209a0a0",
				},
			}
			repo := &model.Repo{
				Owner:    "kelvin",
				Name:     "test_project",
				FullName: "kelvin/test_project",
				Link:     "https://coding.net/u/kelvin/p/test_project",
				Clone:    "https://git.coding.net/kelvin/test_project.git",
				SCMKind:  model.RepoGit,
			}
			actual, err := convertRepository(repository)
			g.Assert(err).IsNil()
			g.Assert(actual).Equal(repo)
		})

		g.It("Should parse push hook", func() {
			repo := &model.Repo{
				Owner:    "demo1",
				Name:     "test1",
				FullName: "demo1/test1",
				Link:     "https://coding.net/u/demo1/p/test1",
				Clone:    "https://git.coding.net/demo1/test1.git",
				SCMKind:  model.RepoGit,
			}

			build := &model.Build{
				Event:   model.EventPush,
				Commit:  "5b9912a6ff272e9c93a4c44c278fe9b359ed1ab4",
				Ref:     "refs/heads/master",
				Link:    "https://coding.net/u/demo1/p/test1/git/commit/5b9912a6ff272e9c93a4c44c278fe9b359ed1ab4",
				Branch:  "master",
				Message: "new file .woodpecker.yml\n",
				Email:   "demo1@gmail.com",
				Avatar:  "/static/fruit_avatar/Fruit-20.png",
				Author:  "demo1",
				Remote:  "https://git.coding.net/demo1/test1.git",
			}

			actualRepo, actualBuild, err := parsePushHook([]byte(fixtures.PushHook))
			g.Assert(err).IsNil()
			g.Assert(actualRepo).Equal(repo)
			g.Assert(actualBuild).Equal(build)
		})

		g.It("Should parse delete branch push hook", func() {
			actualRepo, actualBuild, err := parsePushHook([]byte(fixtures.DeleteBranchPushHook))
			g.Assert(err).IsNil()
			g.Assert(actualRepo).IsNil()
			g.Assert(actualBuild).IsNil()
		})

		g.It("Should parse pull request hook", func() {
			repo := &model.Repo{
				Owner:    "demo1",
				Name:     "test2",
				FullName: "demo1/test2",
				Link:     "https://coding.net/u/demo1/p/test2",
				Clone:    "https://git.coding.net/demo1/test2.git",
				SCMKind:  model.RepoGit,
			}

			build := &model.Build{
				Event:   model.EventPull,
				Commit:  "55e77b328b71d3ee4f9e70a5f67231b0acceeadc",
				Link:    "https://coding.net/u/demo1/p/test2/git/pull/1",
				Ref:     "refs/pull/1/MERGE",
				Branch:  "master",
				Message: "pr message",
				Author:  "demo2",
				Avatar:  "/static/fruit_avatar/Fruit-2.png",
				Title:   "pr1",
				Remote:  "https://git.coding.net/demo1/test2.git",
				Refspec: "master:master",
			}

			actualRepo, actualBuild, err := parsePullRequestHook([]byte(fixtures.PullRequestHook))
			g.Assert(err).IsNil()
			g.Assert(actualRepo).Equal(repo)
			g.Assert(actualBuild).Equal(build)
		})

		g.It("Should parse merge request hook", func() {
			repo := &model.Repo{
				Owner:    "demo1",
				Name:     "test1",
				FullName: "demo1/test1",
				Link:     "https://coding.net/u/demo1/p/test1",
				Clone:    "https://git.coding.net/demo1/test1.git",
				SCMKind:  model.RepoGit,
			}

			build := &model.Build{
				Event:   model.EventPull,
				Commit:  "74e6755580c34e9fd81dbcfcbd43ee5f30259436",
				Link:    "https://coding.net/u/demo1/p/test1/git/merge/1",
				Ref:     "refs/merge/1/MERGE",
				Branch:  "master",
				Message: "<p>mr message</p>",
				Author:  "demo1",
				Avatar:  "/static/fruit_avatar/Fruit-20.png",
				Title:   "mr1",
				Remote:  "https://git.coding.net/demo1/test1.git",
				Refspec: "branch1:master",
			}

			actualRepo, actualBuild, err := parseMergeReuqestHook([]byte(fixtures.MergeRequestHook))
			g.Assert(err).IsNil()
			g.Assert(actualRepo).Equal(repo)
			g.Assert(actualBuild).Equal(build)
		})
	})
}
