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
	"testing"

	"github.com/franela/goblin"
	"github.com/google/go-github/v64/github"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func Test_helper(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("GitHub converter", func() {
		g.It("should convert passing status", func() {
			g.Assert(convertStatus(model.StatusSuccess)).Equal(statusSuccess)
		})

		g.It("should convert pending status", func() {
			g.Assert(convertStatus(model.StatusPending)).Equal(statusPending)
			g.Assert(convertStatus(model.StatusRunning)).Equal(statusPending)
		})

		g.It("should convert failing status", func() {
			g.Assert(convertStatus(model.StatusFailure)).Equal(statusFailure)
		})

		g.It("should convert error status", func() {
			g.Assert(convertStatus(model.StatusKilled)).Equal(statusError)
			g.Assert(convertStatus(model.StatusError)).Equal(statusError)
		})

		g.It("should convert passing desc", func() {
			g.Assert(convertDesc(model.StatusSuccess)).Equal(descSuccess)
		})

		g.It("should convert pending desc", func() {
			g.Assert(convertDesc(model.StatusPending)).Equal(descPending)
			g.Assert(convertDesc(model.StatusRunning)).Equal(descPending)
		})

		g.It("should convert failing desc", func() {
			g.Assert(convertDesc(model.StatusFailure)).Equal(descFailure)
		})

		g.It("should convert error desc", func() {
			g.Assert(convertDesc(model.StatusKilled)).Equal(descError)
			g.Assert(convertDesc(model.StatusError)).Equal(descError)
		})

		g.It("should convert repository list", func() {
			from := []*github.Repository{
				{
					Private:  github.Bool(false),
					FullName: github.String("octocat/hello-world"),
					Name:     github.String("hello-world"),
					Owner: &github.User{
						AvatarURL: github.String("http://..."),
						Login:     github.String("octocat"),
					},
					HTMLURL:  github.String("https://github.com/octocat/hello-world"),
					CloneURL: github.String("https://github.com/octocat/hello-world.git"),
					Permissions: map[string]bool{
						"push":  true,
						"pull":  true,
						"admin": true,
					},
				},
			}

			to := convertRepoList(from)
			g.Assert(to[0].Avatar).Equal("http://...")
			g.Assert(to[0].FullName).Equal("octocat/hello-world")
			g.Assert(to[0].Owner).Equal("octocat")
			g.Assert(to[0].Name).Equal("hello-world")
		})

		g.It("should convert repository", func() {
			from := github.Repository{
				FullName:      github.String("octocat/hello-world"),
				Name:          github.String("hello-world"),
				HTMLURL:       github.String("https://github.com/octocat/hello-world"),
				CloneURL:      github.String("https://github.com/octocat/hello-world.git"),
				DefaultBranch: github.String("develop"),
				Private:       github.Bool(true),
				Owner: &github.User{
					AvatarURL: github.String("http://..."),
					Login:     github.String("octocat"),
				},
				Permissions: map[string]bool{
					"push":  true,
					"pull":  true,
					"admin": true,
				},
			}

			to := convertRepo(&from)
			g.Assert(to.Avatar).Equal("http://...")
			g.Assert(to.FullName).Equal("octocat/hello-world")
			g.Assert(to.Owner).Equal("octocat")
			g.Assert(to.Name).Equal("hello-world")
			g.Assert(to.Branch).Equal("develop")
			g.Assert(string(to.SCMKind)).Equal("git")
			g.Assert(to.IsSCMPrivate).IsTrue()
			g.Assert(to.Clone).Equal("https://github.com/octocat/hello-world.git")
			g.Assert(to.ForgeURL).Equal("https://github.com/octocat/hello-world")
		})

		g.It("should convert repository permissions", func() {
			from := &github.Repository{
				Permissions: map[string]bool{
					"admin": true,
					"push":  true,
					"pull":  true,
				},
			}

			to := convertPerm(from.GetPermissions())
			g.Assert(to.Push).IsTrue()
			g.Assert(to.Pull).IsTrue()
			g.Assert(to.Admin).IsTrue()
		})

		g.It("should convert team", func() {
			from := &github.Organization{
				Login:     github.String("octocat"),
				AvatarURL: github.String("http://..."),
			}
			to := convertTeam(from)
			g.Assert(to.Login).Equal("octocat")
			g.Assert(to.Avatar).Equal("http://...")
		})

		g.It("should convert team list", func() {
			from := []*github.Organization{
				{
					Login:     github.String("octocat"),
					AvatarURL: github.String("http://..."),
				},
			}
			to := convertTeamList(from)
			g.Assert(to[0].Login).Equal("octocat")
			g.Assert(to[0].Avatar).Equal("http://...")
		})

		g.It("should convert a repository from webhook", func() {
			from := &github.PushEventRepository{Owner: &github.User{}}
			from.Owner.Login = github.String("octocat")
			from.Owner.Name = github.String("octocat")
			from.Name = github.String("hello-world")
			from.FullName = github.String("octocat/hello-world")
			from.Private = github.Bool(true)
			from.HTMLURL = github.String("https://github.com/octocat/hello-world")
			from.CloneURL = github.String("https://github.com/octocat/hello-world.git")
			from.DefaultBranch = github.String("develop")

			repo := convertRepoHook(from)
			g.Assert(repo.Owner).Equal(*from.Owner.Login)
			g.Assert(repo.Name).Equal(*from.Name)
			g.Assert(repo.FullName).Equal(*from.FullName)
			g.Assert(repo.IsSCMPrivate).Equal(*from.Private)
			g.Assert(repo.ForgeURL).Equal(*from.HTMLURL)
			g.Assert(repo.Clone).Equal(*from.CloneURL)
			g.Assert(repo.Branch).Equal(*from.DefaultBranch)
		})

		g.It("should convert a pull request from webhook", func() {
			from := &github.PullRequestEvent{
				Action: github.String(actionOpen),
				PullRequest: &github.PullRequest{
					State:   github.String(stateOpen),
					HTMLURL: github.String("https://github.com/octocat/hello-world/pulls/42"),
					Number:  github.Int(42),
					Title:   github.String("Updated README.md"),
					Base: &github.PullRequestBranch{
						Ref: github.String("main"),
					},
					Head: &github.PullRequestBranch{
						Ref: github.String("changes"),
						SHA: github.String("f72fc19"),
						Repo: &github.Repository{
							CloneURL: github.String("https://github.com/octocat/hello-world-fork"),
						},
					},
					User: &github.User{
						Login:     github.String("octocat"),
						AvatarURL: github.String("https://avatars1.githubusercontent.com/u/583231"),
					},
				}, Sender: &github.User{
					Login: github.String("octocat"),
				},
			}
			pull, _, pipeline, err := parsePullHook(from, true)
			g.Assert(err).IsNil()
			g.Assert(pull).IsNotNil()
			g.Assert(pipeline.Event).Equal(model.EventPull)
			g.Assert(pipeline.Branch).Equal(*from.PullRequest.Base.Ref)
			g.Assert(pipeline.Ref).Equal("refs/pull/42/merge")
			g.Assert(pipeline.Refspec).Equal("changes:main")
			g.Assert(pipeline.Commit).Equal(*from.PullRequest.Head.SHA)
			g.Assert(pipeline.Message).Equal(*from.PullRequest.Title)
			g.Assert(pipeline.Title).Equal(*from.PullRequest.Title)
			g.Assert(pipeline.Author).Equal(*from.PullRequest.User.Login)
			g.Assert(pipeline.Avatar).Equal(*from.PullRequest.User.AvatarURL)
			g.Assert(pipeline.Sender).Equal(*from.Sender.Login)
		})

		g.It("should convert a deployment from webhook", func() {
			from := &github.DeploymentEvent{Deployment: &github.Deployment{}, Sender: &github.User{}}
			from.Deployment.Description = github.String(":shipit:")
			from.Deployment.Environment = github.String("production")
			from.Deployment.Task = github.String("deploy")
			from.Deployment.ID = github.Int64(42)
			from.Deployment.Ref = github.String("main")
			from.Deployment.SHA = github.String("f72fc19")
			from.Deployment.URL = github.String("https://github.com/octocat/hello-world")
			from.Sender.Login = github.String("octocat")
			from.Sender.AvatarURL = github.String("https://avatars1.githubusercontent.com/u/583231")

			_, pipeline := parseDeployHook(from)
			g.Assert(pipeline.Event).Equal(model.EventDeploy)
			g.Assert(pipeline.Branch).Equal("main")
			g.Assert(pipeline.Ref).Equal("refs/heads/main")
			g.Assert(pipeline.Commit).Equal(*from.Deployment.SHA)
			g.Assert(pipeline.Message).Equal(*from.Deployment.Description)
			g.Assert(pipeline.ForgeURL).Equal(*from.Deployment.URL)
			g.Assert(pipeline.Author).Equal(*from.Sender.Login)
			g.Assert(pipeline.Avatar).Equal(*from.Sender.AvatarURL)
		})

		g.It("should convert a push from webhook", func() {
			from := &github.PushEvent{Sender: &github.User{}, Repo: &github.PushEventRepository{}, HeadCommit: &github.HeadCommit{Author: &github.CommitAuthor{}}}
			from.Sender.Login = github.String("octocat")
			from.Sender.AvatarURL = github.String("https://avatars1.githubusercontent.com/u/583231")
			from.Repo.CloneURL = github.String("https://github.com/octocat/hello-world.git")
			from.HeadCommit.Author.Email = github.String("github.String(octocat@github.com")
			from.HeadCommit.Message = github.String("updated README.md")
			from.HeadCommit.URL = github.String("https://github.com/octocat/hello-world")
			from.HeadCommit.ID = github.String("f72fc19")
			from.Ref = github.String("refs/heads/main")

			_, pipeline := parsePushHook(from)
			g.Assert(pipeline.Event).Equal(model.EventPush)
			g.Assert(pipeline.Branch).Equal("main")
			g.Assert(pipeline.Ref).Equal("refs/heads/main")
			g.Assert(pipeline.Commit).Equal(*from.HeadCommit.ID)
			g.Assert(pipeline.Message).Equal(*from.HeadCommit.Message)
			g.Assert(pipeline.ForgeURL).Equal(*from.HeadCommit.URL)
			g.Assert(pipeline.Author).Equal(*from.Sender.Login)
			g.Assert(pipeline.Avatar).Equal(*from.Sender.AvatarURL)
			g.Assert(pipeline.Email).Equal(*from.HeadCommit.Author.Email)
		})

		g.It("should convert a tag from webhook", func() {
			from := &github.PushEvent{}
			from.Ref = github.String("refs/tags/v1.0.0")

			_, pipeline := parsePushHook(from)
			g.Assert(pipeline.Event).Equal(model.EventTag)
			g.Assert(pipeline.Ref).Equal("refs/tags/v1.0.0")
		})

		g.It("should convert tag's base branch from webhook to pipeline's branch ", func() {
			from := &github.PushEvent{}
			from.Ref = github.String("refs/tags/v1.0.0")
			from.BaseRef = github.String("refs/heads/main")

			_, pipeline := parsePushHook(from)
			g.Assert(pipeline.Event).Equal(model.EventTag)
			g.Assert(pipeline.Branch).Equal("main")
		})

		g.It("should not convert tag's base_ref from webhook if not prefixed with 'ref/heads/'", func() {
			from := &github.PushEvent{}
			from.Ref = github.String("refs/tags/v1.0.0")
			from.BaseRef = github.String("refs/refs/main")

			_, pipeline := parsePushHook(from)
			g.Assert(pipeline.Event).Equal(model.EventTag)
			g.Assert(pipeline.Branch).Equal("refs/tags/v1.0.0")
		})
	})
}
