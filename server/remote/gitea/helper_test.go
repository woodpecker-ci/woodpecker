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
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/franela/goblin"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote/gitea/fixtures"
	"github.com/woodpecker-ci/woodpecker/shared/utils"
)

func Test_parse(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("Gitea", func() {
		g.It("Should parse push hook payload", func() {
			buf := bytes.NewBufferString(fixtures.HookPush)
			hook, err := parsePush(buf)
			g.Assert(err).IsNil()
			g.Assert(hook.Ref).Equal("refs/heads/master")
			g.Assert(hook.After).Equal("ef98532add3b2feb7a137426bba1248724367df5")
			g.Assert(hook.Before).Equal("4b2626259b5a97b6b4eab5e6cca66adb986b672b")
			g.Assert(hook.Compare).Equal("http://gitea.golang.org/gordon/hello-world/compare/4b2626259b5a97b6b4eab5e6cca66adb986b672b...ef98532add3b2feb7a137426bba1248724367df5")
			g.Assert(hook.Repo.Name).Equal("hello-world")
			g.Assert(hook.Repo.URL).Equal("http://gitea.golang.org/gordon/hello-world")
			g.Assert(hook.Repo.Owner.Name).Equal("gordon")
			g.Assert(hook.Repo.FullName).Equal("gordon/hello-world")
			g.Assert(hook.Repo.Owner.Email).Equal("gordon@golang.org")
			g.Assert(hook.Repo.Owner.Username).Equal("gordon")
			g.Assert(hook.Repo.Private).Equal(true)
			g.Assert(hook.Pusher.Name).Equal("gordon")
			g.Assert(hook.Pusher.Email).Equal("gordon@golang.org")
			g.Assert(hook.Pusher.Username).Equal("gordon")
			g.Assert(hook.Pusher.Login).Equal("gordon")
			g.Assert(hook.Sender.Login).Equal("gordon")
			g.Assert(hook.Sender.Username).Equal("gordon")
			g.Assert(hook.Sender.Avatar).Equal("http://gitea.golang.org///1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87")
		})

		g.It("Should parse tag hook payload", func() {
			buf := bytes.NewBufferString(fixtures.HookPushTag)
			hook, err := parsePush(buf)
			g.Assert(err).IsNil()
			g.Assert(hook.Ref).Equal("v1.0.0")
			g.Assert(hook.Sha).Equal("ef98532add3b2feb7a137426bba1248724367df5")
			g.Assert(hook.Repo.Name).Equal("hello-world")
			g.Assert(hook.Repo.URL).Equal("http://gitea.golang.org/gordon/hello-world")
			g.Assert(hook.Repo.FullName).Equal("gordon/hello-world")
			g.Assert(hook.Repo.Owner.Email).Equal("gordon@golang.org")
			g.Assert(hook.Repo.Owner.Username).Equal("gordon")
			g.Assert(hook.Repo.Private).Equal(true)
			g.Assert(hook.Sender.Username).Equal("gordon")
			g.Assert(hook.Sender.Avatar).Equal("https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87")
		})

		g.It("Should parse pull_request hook payload", func() {
			buf := bytes.NewBufferString(fixtures.HookPullRequest)
			hook, err := parsePullRequest(buf)
			g.Assert(err).IsNil()
			g.Assert(hook.Action).Equal("opened")
			g.Assert(hook.Number).Equal(int64(1))

			g.Assert(hook.Repo.Name).Equal("hello-world")
			g.Assert(hook.Repo.URL).Equal("http://gitea.golang.org/gordon/hello-world")
			g.Assert(hook.Repo.FullName).Equal("gordon/hello-world")
			g.Assert(hook.Repo.Owner.Email).Equal("gordon@golang.org")
			g.Assert(hook.Repo.Owner.Username).Equal("gordon")
			g.Assert(hook.Repo.Private).Equal(true)
			g.Assert(hook.Sender.Username).Equal("gordon")
			g.Assert(hook.Sender.Avatar).Equal("https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87")

			g.Assert(hook.PullRequest.Title).Equal("Update the README with new information")
			g.Assert(hook.PullRequest.Body).Equal("please merge")
			g.Assert(hook.PullRequest.State).Equal("open")
			g.Assert(hook.PullRequest.User.Username).Equal("gordon")
			g.Assert(hook.PullRequest.Base.Label).Equal("master")
			g.Assert(hook.PullRequest.Base.Ref).Equal("master")
			g.Assert(hook.PullRequest.Head.Label).Equal("feature/changes")
			g.Assert(hook.PullRequest.Head.Ref).Equal("feature/changes")
		})

		g.It("Should return a Build struct from a push hook", func() {
			buf := bytes.NewBufferString(fixtures.HookPush)
			hook, _ := parsePush(buf)
			build := buildFromPush(hook)
			g.Assert(build.Event).Equal(model.EventPush)
			g.Assert(build.Commit).Equal(hook.After)
			g.Assert(build.Ref).Equal(hook.Ref)
			g.Assert(build.Link).Equal(hook.Commits[0].URL)
			g.Assert(build.Branch).Equal("master")
			g.Assert(build.Message).Equal(hook.Commits[0].Message)
			g.Assert(build.Avatar).Equal("http://1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87")
			g.Assert(build.Author).Equal(hook.Sender.Login)
			g.Assert(utils.EqualStringSlice(build.ChangedFiles, []string{"CHANGELOG.md", "app/controller/application.rb"})).IsTrue()
		})

		g.It("Should return a Repo struct from a push hook", func() {
			buf := bytes.NewBufferString(fixtures.HookPush)
			hook, _ := parsePush(buf)
			repo := repoFromPush(hook)
			g.Assert(repo.Name).Equal(hook.Repo.Name)
			g.Assert(repo.Owner).Equal(hook.Repo.Owner.Username)
			g.Assert(repo.FullName).Equal("gordon/hello-world")
			g.Assert(repo.Link).Equal(hook.Repo.URL)
		})

		g.It("Should return a Build struct from a tag hook", func() {
			buf := bytes.NewBufferString(fixtures.HookPushTag)
			hook, _ := parsePush(buf)
			build := buildFromTag(hook)
			g.Assert(build.Event).Equal(model.EventTag)
			g.Assert(build.Commit).Equal(hook.Sha)
			g.Assert(build.Ref).Equal("refs/tags/v1.0.0")
			g.Assert(build.Branch).Equal("refs/tags/v1.0.0")
			g.Assert(build.Link).Equal("http://gitea.golang.org/gordon/hello-world/src/tag/v1.0.0")
			g.Assert(build.Message).Equal("created tag v1.0.0")
		})

		g.It("Should return a Build struct from a pull_request hook", func() {
			buf := bytes.NewBufferString(fixtures.HookPullRequest)
			hook, _ := parsePullRequest(buf)
			build := buildFromPullRequest(hook)
			g.Assert(build.Event).Equal(model.EventPull)
			g.Assert(build.Commit).Equal(hook.PullRequest.Head.Sha)
			g.Assert(build.Ref).Equal("refs/pull/1/head")
			g.Assert(build.Link).Equal(hook.PullRequest.URL)
			g.Assert(build.Branch).Equal("master")
			g.Assert(build.Refspec).Equal("feature/changes:master")
			g.Assert(build.Message).Equal(hook.PullRequest.Title)
			g.Assert(build.Avatar).Equal("http://1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87")
			g.Assert(build.Author).Equal(hook.PullRequest.User.Username)
		})

		g.It("Should return a Repo struct from a pull_request hook", func() {
			buf := bytes.NewBufferString(fixtures.HookPullRequest)
			hook, _ := parsePullRequest(buf)
			repo := repoFromPullRequest(hook)
			g.Assert(repo.Name).Equal(hook.Repo.Name)
			g.Assert(repo.Owner).Equal(hook.Repo.Owner.Username)
			g.Assert(repo.FullName).Equal("gordon/hello-world")
			g.Assert(repo.Link).Equal(hook.Repo.URL)
		})

		g.It("Should return a Perm struct from a Gitea Perm", func() {
			perms := []gitea.Permission{
				{
					Admin: true,
					Push:  true,
					Pull:  true,
				},
				{
					Admin: true,
					Push:  true,
					Pull:  false,
				},
				{
					Admin: true,
					Push:  false,
					Pull:  false,
				},
			}
			for _, from := range perms {
				perm := toPerm(&from)
				g.Assert(perm.Pull).Equal(from.Pull)
				g.Assert(perm.Push).Equal(from.Push)
				g.Assert(perm.Admin).Equal(from.Admin)
			}
		})

		g.It("Should return a Team struct from a Gitea Org", func() {
			from := &gitea.Organization{
				UserName:  "woodpecker",
				AvatarURL: "/avatars/1",
			}

			to := toTeam(from, "http://localhost:80")
			g.Assert(to.Login).Equal(from.UserName)
			g.Assert(to.Avatar).Equal("http://localhost:80/avatars/1")
		})

		g.It("Should return a Repo struct from a Gitea Repo", func() {
			from := gitea.Repository{
				FullName: "gophers/hello-world",
				Owner: &gitea.User{
					UserName:  "gordon",
					AvatarURL: "http://1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
				},
				CloneURL:      "http://gitea.golang.org/gophers/hello-world.git",
				HTMLURL:       "http://gitea.golang.org/gophers/hello-world",
				Private:       true,
				DefaultBranch: "master",
			}
			repo := toRepo(&from)
			g.Assert(repo.FullName).Equal(from.FullName)
			g.Assert(repo.Owner).Equal(from.Owner.UserName)
			g.Assert(repo.Name).Equal("hello-world")
			g.Assert(repo.Branch).Equal("master")
			g.Assert(repo.Link).Equal(from.HTMLURL)
			g.Assert(repo.Clone).Equal(from.CloneURL)
			g.Assert(repo.Avatar).Equal(from.Owner.AvatarURL)
			g.Assert(repo.IsSCMPrivate).Equal(from.Private)
		})

		g.It("Should correct a malformed avatar url", func() {
			urls := []struct {
				Before string
				After  string
			}{
				{
					"http://gitea.golang.org///1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
					"//1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
				},
				{
					"//1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
					"//1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
				},
				{
					"http://gitea.golang.org/avatars/1",
					"http://gitea.golang.org/avatars/1",
				},
				{
					"http://gitea.golang.org//avatars/1",
					"http://gitea.golang.org/avatars/1",
				},
			}

			for _, url := range urls {
				got := fixMalformedAvatar(url.Before)
				g.Assert(got).Equal(url.After)
			}
		})

		g.It("Should expand the avatar url", func() {
			urls := []struct {
				Before string
				After  string
			}{
				{
					"/avatars/1",
					"http://gitea.io/avatars/1",
				},
				{
					"//1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
					"http://1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
				},
				{
					"/gitea/avatars/2",
					"http://gitea.io/gitea/avatars/2",
				},
			}

			repo := "http://gitea.io/foo/bar"
			for _, url := range urls {
				got := expandAvatar(repo, url.Before)
				g.Assert(got).Equal(url.After)
			}
		})
	})
}
