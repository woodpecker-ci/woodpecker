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

package forgejo

import (
	"bytes"
	"testing"

	"github.com/franela/goblin"

	"github.com/woodpecker-ci/woodpecker/server/forge/forgejo/client"
	"github.com/woodpecker-ci/woodpecker/server/forge/forgejo/fixtures"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/shared/utils"
)

func Test_parse(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("Forgejo", func() {
		g.It("Should parse push hook payload", func() {
			buf := bytes.NewBufferString(fixtures.HookPush)
			hook, err := parsePush(buf)
			g.Assert(err).IsNil()
			g.Assert(hook.Ref).Equal("refs/heads/master")
			g.Assert(hook.After).Equal("ef98532add3b2feb7a137426bba1248724367df5")
			g.Assert(hook.Before).Equal("4b2626259b5a97b6b4eab5e6cca66adb986b672b")
			g.Assert(hook.Compare).Equal("http://forgejo.golang.org/gordon/hello-world/compare/4b2626259b5a97b6b4eab5e6cca66adb986b672b...ef98532add3b2feb7a137426bba1248724367df5")
			g.Assert(hook.Repo.Name).Equal("hello-world")
			g.Assert(hook.Repo.HTMLURL).Equal("http://forgejo.golang.org/gordon/hello-world")
			g.Assert(hook.Repo.Owner.UserName).Equal("gordon")
			g.Assert(hook.Repo.FullName).Equal("gordon/hello-world")
			g.Assert(hook.Repo.Owner.Email).Equal("gordon@golang.org")
			g.Assert(hook.Repo.Private).Equal(true)
			g.Assert(hook.Pusher.Email).Equal("gordon@golang.org")
			g.Assert(hook.Pusher.UserName).Equal("gordon")
			g.Assert(hook.Sender.UserName).Equal("gordon")
			g.Assert(hook.Sender.AvatarURL).Equal("http://forgejo.golang.org///1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87")
		})

		g.It("Should parse tag hook payload", func() {
			buf := bytes.NewBufferString(fixtures.HookPushTag)
			hook, err := parsePush(buf)
			g.Assert(err).IsNil()
			g.Assert(hook.Ref).Equal("v1.0.0")
			g.Assert(hook.Sha).Equal("ef98532add3b2feb7a137426bba1248724367df5")
			g.Assert(hook.Repo.Name).Equal("hello-world")
			g.Assert(hook.Repo.HTMLURL).Equal("http://forgejo.golang.org/gordon/hello-world")
			g.Assert(hook.Repo.FullName).Equal("gordon/hello-world")
			g.Assert(hook.Repo.Owner.Email).Equal("gordon@golang.org")
			g.Assert(hook.Repo.Owner.UserName).Equal("gordon")
			g.Assert(hook.Repo.Private).Equal(true)
			g.Assert(hook.Sender.UserName).Equal("gordon")
			g.Assert(hook.Sender.AvatarURL).Equal("https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87")
		})

		g.It("Should parse pull_request hook payload", func() {
			buf := bytes.NewBufferString(fixtures.HookPullRequest)
			hook, err := parsePullRequest(buf)
			g.Assert(err).IsNil()
			g.Assert(hook.Action).Equal("opened")
			g.Assert(hook.Number).Equal(int64(1))

			g.Assert(hook.Repo.Name).Equal("hello-world")
			g.Assert(hook.Repo.HTMLURL).Equal("http://forgejo.golang.org/gordon/hello-world")
			g.Assert(hook.Repo.FullName).Equal("gordon/hello-world")
			g.Assert(hook.Repo.Owner.Email).Equal("gordon@golang.org")
			g.Assert(hook.Repo.Owner.UserName).Equal("gordon")
			g.Assert(hook.Repo.Private).Equal(true)
			g.Assert(hook.Sender.UserName).Equal("gordon")
			g.Assert(hook.Sender.AvatarURL).Equal("https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87")

			g.Assert(hook.PullRequest.Title).Equal("Update the README with new information")
			//			g.Assert(hook.PullRequest.Body).Equal("please merge")
			//			g.Assert(hook.PullRequest.State).Equal(client.StateOpen)
			g.Assert(hook.PullRequest.Poster.UserName).Equal("gordon")
			//			g.Assert(hook.PullRequest.Base.Name).Equal("master")
			//			g.Assert(hook.PullRequest.Base.Ref).Equal("master")
			//			g.Assert(hook.PullRequest.Head.Name).Equal("feature/changes")
			g.Assert(hook.PullRequest.Head.Ref).Equal("feature/changes")
		})

		g.It("Should return a Pipeline struct from a push hook", func() {
			buf := bytes.NewBufferString(fixtures.HookPush)
			hook, _ := parsePush(buf)
			pipeline := pipelineFromPush(hook)
			g.Assert(pipeline.Event).Equal(model.EventPush)
			g.Assert(pipeline.Commit).Equal(hook.After)
			g.Assert(pipeline.Ref).Equal(hook.Ref)
			g.Assert(pipeline.Link).Equal(hook.Commits[0].URL)
			g.Assert(pipeline.Branch).Equal("master")
			g.Assert(pipeline.Message).Equal(hook.Commits[0].Message)
			g.Assert(pipeline.Avatar).Equal("http://1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87")
			g.Assert(pipeline.Author).Equal(hook.Sender.UserName)
			g.Assert(utils.EqualStringSlice(pipeline.ChangedFiles, []string{"CHANGELOG.md", "app/controller/application.rb"})).IsTrue()
		})

		g.It("Should return a Repo struct from a push hook", func() {
			buf := bytes.NewBufferString(fixtures.HookPush)
			hook, _ := parsePush(buf)
			repo := toRepo(hook.Repo)
			g.Assert(repo.Name).Equal(hook.Repo.Name)
			g.Assert(repo.Owner).Equal(hook.Repo.Owner.UserName)
			g.Assert(repo.FullName).Equal("gordon/hello-world")
			g.Assert(repo.Link).Equal(hook.Repo.HTMLURL)
		})

		g.It("Should return a Pipeline struct from a tag hook", func() {
			buf := bytes.NewBufferString(fixtures.HookPushTag)
			hook, _ := parsePush(buf)
			pipeline := pipelineFromTag(hook)
			g.Assert(pipeline.Event).Equal(model.EventTag)
			g.Assert(pipeline.Commit).Equal(hook.Sha)
			g.Assert(pipeline.Ref).Equal("refs/tags/v1.0.0")
			g.Assert(pipeline.Branch).Equal("refs/tags/v1.0.0")
			g.Assert(pipeline.Link).Equal("http://forgejo.golang.org/gordon/hello-world/src/tag/v1.0.0")
			g.Assert(pipeline.Message).Equal("created tag v1.0.0")
		})

		g.It("Should return a Pipeline struct from a pull_request hook", func() {
			buf := bytes.NewBufferString(fixtures.HookPullRequest)
			hook, _ := parsePullRequest(buf)
			pipeline := pipelineFromPullRequest(hook)
			g.Assert(pipeline.Event).Equal(model.EventPull)
			g.Assert(pipeline.Commit).Equal(hook.PullRequest.Head.Sha)
			g.Assert(pipeline.Ref).Equal("refs/pull/1/head")
			g.Assert(pipeline.Link).Equal(hook.PullRequest.URL)
			g.Assert(pipeline.Branch).Equal("master")
			g.Assert(pipeline.Refspec).Equal("feature/changes:master")
			g.Assert(pipeline.Message).Equal(hook.PullRequest.Title)
			g.Assert(pipeline.Avatar).Equal("http://1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87")
			g.Assert(pipeline.Author).Equal(hook.PullRequest.Poster.UserName)
		})

		g.It("Should return a Repo struct from a pull_request hook", func() {
			buf := bytes.NewBufferString(fixtures.HookPullRequest)
			hook, _ := parsePullRequest(buf)
			repo := toRepo(hook.Repo)
			g.Assert(repo.Name).Equal(hook.Repo.Name)
			g.Assert(repo.Owner).Equal(hook.Repo.Owner.UserName)
			g.Assert(repo.FullName).Equal("gordon/hello-world")
			g.Assert(repo.Link).Equal(hook.Repo.HTMLURL)
		})

		g.It("Should return a Perm struct from a Forgejo Perm", func() {
			perms := []client.Permission{
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

		g.It("Should return a Team struct from a Forgejo Org", func() {
			from := &client.Organization{
				UserName:  "woodpecker",
				AvatarURL: "/avatars/1",
			}

			to := toTeam(from, "http://localhost:80")
			g.Assert(to.Login).Equal(from.UserName)
			g.Assert(to.Avatar).Equal("http://localhost:80/avatars/1")
		})

		g.It("Should return a Repo struct from a Forgejo Repo", func() {
			from := client.Repository{
				FullName: "gophers/hello-world",
				Owner: &client.User{
					UserName:  "gordon",
					AvatarURL: "http://1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
				},
				CloneURL:      "http://forgejo.golang.org/gophers/hello-world.git",
				HTMLURL:       "http://forgejo.golang.org/gophers/hello-world",
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
					"http://forgejo.golang.org///1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
					"//1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
				},
				{
					"//1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
					"//1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
				},
				{
					"http://forgejo.golang.org/avatars/1",
					"http://forgejo.golang.org/avatars/1",
				},
				{
					"http://forgejo.golang.org//avatars/1",
					"http://forgejo.golang.org/avatars/1",
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
					"http://forgejo.io/avatars/1",
				},
				{
					"//1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
					"http://1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
				},
				{
					"/forgejo/avatars/2",
					"http://forgejo.io/forgejo/avatars/2",
				},
			}

			repo := "http://forgejo.io/foo/bar"
			for _, url := range urls {
				got := expandAvatar(repo, url.Before)
				g.Assert(got).Equal(url.After)
			}
		})
	})
}
