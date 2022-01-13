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

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote/bitbucket/internal"
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

			to := convertRepo(from)
			g.Assert(to.Avatar).Equal(from.Owner.Links.Avatar.Href)
			g.Assert(to.FullName).Equal(from.FullName)
			g.Assert(to.Owner).Equal("octocat")
			g.Assert(to.Name).Equal("hello-world")
			g.Assert(to.Branch).Equal("default")
			g.Assert(string(to.SCMKind)).Equal(from.Scm)
			g.Assert(to.IsSCMPrivate).Equal(from.IsPrivate)
			g.Assert(to.Clone).Equal(from.Links.HTML.Href)
			g.Assert(to.Link).Equal(from.Links.HTML.Href)
		})

		g.It("should convert team", func() {
			from := &internal.Account{Login: "octocat"}
			from.Links.Avatar.Href = "http://..."
			to := convertTeam(from)
			g.Assert(to.Avatar).Equal(from.Links.Avatar.Href)
			g.Assert(to.Login).Equal(from.Login)
		})

		g.It("should convert team list", func() {
			from := &internal.Account{Login: "octocat"}
			from.Links.Avatar.Href = "http://..."
			to := convertTeamList([]*internal.Account{from})
			g.Assert(to[0].Avatar).Equal(from.Links.Avatar.Href)
			g.Assert(to[0].Login).Equal(from.Login)
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
			g.Assert(result.Token).Equal(token.AccessToken)
			g.Assert(result.Secret).Equal(token.RefreshToken)
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

		g.It("should convert pull hook to build", func() {
			hook := &internal.PullRequestHook{}
			hook.Actor.Login = "octocat"
			hook.Actor.Links.Avatar.Href = "https://..."
			hook.PullRequest.Dest.Commit.Hash = "73f9c44d"
			hook.PullRequest.Dest.Branch.Name = "master"
			hook.PullRequest.Dest.Repo.Links.HTML.Href = "https://bitbucket.org/foo/bar"
			hook.PullRequest.Source.Branch.Name = "change"
			hook.PullRequest.Source.Repo.FullName = "baz/bar"
			hook.PullRequest.Links.HTML.Href = "https://bitbucket.org/foo/bar/pulls/5"
			hook.PullRequest.Desc = "updated README"
			hook.PullRequest.Updated = time.Now()

			build := convertPullHook(hook)
			g.Assert(build.Event).Equal(model.EventPull)
			g.Assert(build.Author).Equal(hook.Actor.Login)
			g.Assert(build.Avatar).Equal(hook.Actor.Links.Avatar.Href)
			g.Assert(build.Commit).Equal(hook.PullRequest.Dest.Commit.Hash)
			g.Assert(build.Branch).Equal(hook.PullRequest.Dest.Branch.Name)
			g.Assert(build.Link).Equal(hook.PullRequest.Links.HTML.Href)
			g.Assert(build.Ref).Equal("refs/heads/master")
			g.Assert(build.Refspec).Equal("change:master")
			g.Assert(build.Remote).Equal("https://bitbucket.org/baz/bar")
			g.Assert(build.Message).Equal(hook.PullRequest.Desc)
			g.Assert(build.Timestamp).Equal(hook.PullRequest.Updated.Unix())
		})

		g.It("should convert push hook to build", func() {
			change := internal.Change{}
			change.New.Target.Hash = "73f9c44d"
			change.New.Name = "master"
			change.New.Target.Links.HTML.Href = "https://bitbucket.org/foo/bar/commits/73f9c44d"
			change.New.Target.Message = "updated README"
			change.New.Target.Date = time.Now()
			change.New.Target.Author.Raw = "Test <test@domain.tld>"

			hook := internal.PushHook{}
			hook.Actor.Login = "octocat"
			hook.Actor.Links.Avatar.Href = "https://..."

			build := convertPushHook(&hook, &change)
			g.Assert(build.Event).Equal(model.EventPush)
			g.Assert(build.Email).Equal("test@domain.tld")
			g.Assert(build.Author).Equal(hook.Actor.Login)
			g.Assert(build.Avatar).Equal(hook.Actor.Links.Avatar.Href)
			g.Assert(build.Commit).Equal(change.New.Target.Hash)
			g.Assert(build.Branch).Equal(change.New.Name)
			g.Assert(build.Link).Equal(change.New.Target.Links.HTML.Href)
			g.Assert(build.Ref).Equal("refs/heads/master")
			g.Assert(build.Message).Equal(change.New.Target.Message)
			g.Assert(build.Timestamp).Equal(change.New.Target.Date.Unix())
		})

		g.It("should convert tag hook to build", func() {
			change := internal.Change{}
			change.New.Name = "v1.0.0"
			change.New.Type = "tag"

			hook := internal.PushHook{}

			build := convertPushHook(&hook, &change)
			g.Assert(build.Event).Equal(model.EventTag)
			g.Assert(build.Ref).Equal("refs/tags/v1.0.0")
		})
	})
}
