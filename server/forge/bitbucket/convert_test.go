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

package bitbucket

import (
	"testing"
	"time"

	"github.com/franela/goblin"
	"golang.org/x/oauth2"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge/bitbucket/internal"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func Test_helper(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("Bitbucket converter", func() {
		g.It("should convert passing status", func() {
			g.Assert(convertStatus(model.StatusSuccess)).Equal(statusSuccess)
		})

		g.It("should convert pending status", func() {
			g.Assert(convertStatus(model.StatusPending)).Equal(statusPending)
			g.Assert(convertStatus(model.StatusRunning)).Equal(statusPending)
		})

		g.It("should convert failing status", func() {
			g.Assert(convertStatus(model.StatusFailure)).Equal(statusFailure)
			g.Assert(convertStatus(model.StatusKilled)).Equal(statusFailure)
			g.Assert(convertStatus(model.StatusError)).Equal(statusFailure)
		})

		g.It("should convert repository", func() {
			from := &internal.Repo{
				FullName:  "octocat/hello-world",
				IsPrivate: true,
				Scm:       "hg",
			}
			from.Owner.Links.Avatar.Href = "http://..."
			from.Links.HTML.Href = "https://bitbucket.org/foo/bar"
			fromPerm := &internal.RepoPerm{
				Permission: "write",
			}

			to := convertRepo(from, fromPerm)
			g.Assert(to.Avatar).Equal(from.Owner.Links.Avatar.Href)
			g.Assert(to.FullName).Equal(from.FullName)
			g.Assert(to.Owner).Equal("octocat")
			g.Assert(to.Name).Equal("hello-world")
			g.Assert(to.Branch).Equal("default")
			g.Assert(string(to.SCMKind)).Equal(from.Scm)
			g.Assert(to.IsSCMPrivate).Equal(from.IsPrivate)
			g.Assert(to.Clone).Equal(from.Links.HTML.Href)
			g.Assert(to.ForgeURL).Equal(from.Links.HTML.Href)
			g.Assert(to.Perm.Push).IsTrue()
			g.Assert(to.Perm.Admin).IsFalse()
		})

		g.It("should convert team", func() {
			from := &internal.Workspace{Slug: "octocat"}
			from.Links.Avatar.Href = "http://..."
			to := convertWorkspace(from)
			g.Assert(to.Avatar).Equal(from.Links.Avatar.Href)
			g.Assert(to.Login).Equal(from.Slug)
		})

		g.It("should convert team list", func() {
			from := &internal.Workspace{Slug: "octocat"}
			from.Links.Avatar.Href = "http://..."
			to := convertWorkspaceList([]*internal.Workspace{from})
			g.Assert(to[0].Avatar).Equal(from.Links.Avatar.Href)
			g.Assert(to[0].Login).Equal(from.Slug)
		})

		g.It("should convert user", func() {
			token := &oauth2.Token{
				AccessToken:  "foo",
				RefreshToken: "bar",
				Expiry:       time.Now(),
			}
			user := &internal.Account{Login: "octocat"}
			user.Links.Avatar.Href = "http://..."

			result := convertUser(user, token)
			g.Assert(result.Avatar).Equal(user.Links.Avatar.Href)
			g.Assert(result.Login).Equal(user.Login)
			g.Assert(result.AccessToken).Equal(token.AccessToken)
			g.Assert(result.RefreshToken).Equal(token.RefreshToken)
			g.Assert(result.Expiry).Equal(token.Expiry.UTC().Unix())
		})

		g.It("should use clone url", func() {
			repo := &internal.Repo{}
			repo.Links.Clone = append(repo.Links.Clone, internal.Link{
				Name: "https",
				Href: "https://bitbucket.org/foo/bar.git",
			})
			link := cloneLink(repo)
			g.Assert(link).Equal(repo.Links.Clone[0].Href)
		})

		g.It("should build clone url", func() {
			repo := &internal.Repo{}
			repo.Links.HTML.Href = "https://foo:bar@bitbucket.org/foo/bar.git"
			link := cloneLink(repo)
			g.Assert(link).Equal("https://bitbucket.org/foo/bar.git")
		})

		g.It("should convert pull hook to pipeline", func() {
			hook := &internal.PullRequestHook{}
			hook.Actor.Login = "octocat"
			hook.Actor.Links.Avatar.Href = "https://..."
			hook.PullRequest.Dest.Commit.Hash = "73f9c44d"
			hook.PullRequest.Dest.Branch.Name = "main"
			hook.PullRequest.Dest.Repo.Links.HTML.Href = "https://bitbucket.org/foo/bar"
			hook.PullRequest.Source.Branch.Name = "change"
			hook.PullRequest.Source.Repo.FullName = "baz/bar"
			hook.PullRequest.Source.Commit.Hash = "c8411d7"
			hook.PullRequest.Links.HTML.Href = "https://bitbucket.org/foo/bar/pulls/5"
			hook.PullRequest.Title = "updated README"
			hook.PullRequest.Updated = time.Now()
			hook.PullRequest.ID = 1

			pipeline := convertPullHook(hook)
			g.Assert(pipeline.Event).Equal(model.EventPull)
			g.Assert(pipeline.Author).Equal(hook.Actor.Login)
			g.Assert(pipeline.Avatar).Equal(hook.Actor.Links.Avatar.Href)
			g.Assert(pipeline.Commit).Equal(hook.PullRequest.Source.Commit.Hash)
			g.Assert(pipeline.Branch).Equal(hook.PullRequest.Source.Branch.Name)
			g.Assert(pipeline.ForgeURL).Equal(hook.PullRequest.Links.HTML.Href)
			g.Assert(pipeline.Ref).Equal("refs/pull-requests/1/from")
			g.Assert(pipeline.Refspec).Equal("change:main")
			g.Assert(pipeline.Message).Equal(hook.PullRequest.Title)
			g.Assert(pipeline.Timestamp).Equal(hook.PullRequest.Updated.Unix())
		})

		g.It("should convert push hook to pipeline", func() {
			change := internal.Change{}
			change.New.Target.Hash = "73f9c44d"
			change.New.Name = "main"
			change.New.Target.Links.HTML.Href = "https://bitbucket.org/foo/bar/commits/73f9c44d"
			change.New.Target.Message = "updated README"
			change.New.Target.Date = time.Now()
			change.New.Target.Author.Raw = "Test <test@domain.tld>"

			hook := internal.PushHook{}
			hook.Actor.Login = "octocat"
			hook.Actor.Links.Avatar.Href = "https://..."

			pipeline := convertPushHook(&hook, &change)
			g.Assert(pipeline.Event).Equal(model.EventPush)
			g.Assert(pipeline.Email).Equal("test@domain.tld")
			g.Assert(pipeline.Author).Equal(hook.Actor.Login)
			g.Assert(pipeline.Avatar).Equal(hook.Actor.Links.Avatar.Href)
			g.Assert(pipeline.Commit).Equal(change.New.Target.Hash)
			g.Assert(pipeline.Branch).Equal(change.New.Name)
			g.Assert(pipeline.ForgeURL).Equal(change.New.Target.Links.HTML.Href)
			g.Assert(pipeline.Ref).Equal("refs/heads/main")
			g.Assert(pipeline.Message).Equal(change.New.Target.Message)
			g.Assert(pipeline.Timestamp).Equal(change.New.Target.Date.Unix())
		})

		g.It("should convert tag hook to pipeline", func() {
			change := internal.Change{}
			change.New.Name = "v1.0.0"
			change.New.Type = "tag"

			hook := internal.PushHook{}

			pipeline := convertPushHook(&hook, &change)
			g.Assert(pipeline.Event).Equal(model.EventTag)
			g.Assert(pipeline.Ref).Equal("refs/tags/v1.0.0")
		})
	})
}
