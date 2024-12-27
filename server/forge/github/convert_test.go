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

	"github.com/google/go-github/v67/github"
	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func Test_convertStatus(t *testing.T) {
	assert.Equal(t, statusSuccess, convertStatus(model.StatusSuccess))
	assert.Equal(t, statusPending, convertStatus(model.StatusPending))
	assert.Equal(t, statusPending, convertStatus(model.StatusRunning))
	assert.Equal(t, statusFailure, convertStatus(model.StatusFailure))
	assert.Equal(t, statusError, convertStatus(model.StatusKilled))
	assert.Equal(t, statusError, convertStatus(model.StatusError))
}

func Test_convertDesc(t *testing.T) {
	assert.Equal(t, descSuccess, convertDesc(model.StatusSuccess))
	assert.Equal(t, descPending, convertDesc(model.StatusPending))
	assert.Equal(t, descPending, convertDesc(model.StatusRunning))
	assert.Equal(t, descFailure, convertDesc(model.StatusFailure))
	assert.Equal(t, descError, convertDesc(model.StatusKilled))
	assert.Equal(t, descError, convertDesc(model.StatusError))
}

func Test_convertRepoList(t *testing.T) {
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
	assert.Equal(t, "http://...", to[0].Avatar)
	assert.Equal(t, "octocat/hello-world", to[0].FullName)
	assert.Equal(t, "octocat", to[0].Owner)
	assert.Equal(t, "hello-world", to[0].Name)
}

func Test_convertRepo(t *testing.T) {
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
	assert.Equal(t, "http://...", to.Avatar)
	assert.Equal(t, "octocat/hello-world", to.FullName)
	assert.Equal(t, "octocat", to.Owner)
	assert.Equal(t, "hello-world", to.Name)
	assert.Equal(t, "develop", to.Branch)
	assert.Equal(t, "git", string(to.SCMKind))
	assert.True(t, to.IsSCMPrivate)
	assert.Equal(t, "https://github.com/octocat/hello-world.git", to.Clone)
	assert.Equal(t, "https://github.com/octocat/hello-world", to.ForgeURL)
}

func Test_convertPerm(t *testing.T) {
	from := &github.Repository{
		Permissions: map[string]bool{
			"admin": true,
			"push":  true,
			"pull":  true,
		},
	}

	to := convertPerm(from.GetPermissions())
	assert.True(t, to.Push)
	assert.True(t, to.Pull)
	assert.True(t, to.Admin)
}

func Test_convertTeam(t *testing.T) {
	from := &github.Organization{
		Login:     github.String("octocat"),
		AvatarURL: github.String("http://..."),
	}
	to := convertTeam(from)
	assert.Equal(t, "octocat", to.Login)
	assert.Equal(t, "http://...", to.Avatar)
}

func Test_convertTeamList(t *testing.T) {
	from := []*github.Organization{
		{
			Login:     github.String("octocat"),
			AvatarURL: github.String("http://..."),
		},
	}
	to := convertTeamList(from)
	assert.Equal(t, "octocat", to[0].Login)
	assert.Equal(t, "http://...", to[0].Avatar)
}

func Test_convertRepoHook(t *testing.T) {
	t.Run("should convert a repository from webhook", func(t *testing.T) {
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
		assert.Equal(t, *from.Owner.Login, repo.Owner)
		assert.Equal(t, *from.Name, repo.Name)
		assert.Equal(t, *from.FullName, repo.FullName)
		assert.Equal(t, *from.Private, repo.IsSCMPrivate)
		assert.Equal(t, *from.HTMLURL, repo.ForgeURL)
		assert.Equal(t, *from.CloneURL, repo.Clone)
		assert.Equal(t, *from.DefaultBranch, repo.Branch)
	})
}

func Test_parsePullHook(t *testing.T) {
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
	assert.NoError(t, err)
	assert.NotNil(t, pull)
	assert.Equal(t, model.EventPull, pipeline.Event)
	assert.Equal(t, *from.PullRequest.Base.Ref, pipeline.Branch)
	assert.Equal(t, "refs/pull/42/merge", pipeline.Ref)
	assert.Equal(t, "changes:main", pipeline.Refspec)
	assert.Equal(t, *from.PullRequest.Head.SHA, pipeline.Commit)
	assert.Equal(t, *from.PullRequest.Title, pipeline.Message)
	assert.Equal(t, *from.PullRequest.Title, pipeline.Title)
	assert.Equal(t, *from.PullRequest.User.Login, pipeline.Author)
	assert.Equal(t, *from.PullRequest.User.AvatarURL, pipeline.Avatar)
	assert.Equal(t, *from.Sender.Login, pipeline.Sender)
}

func Test_parseDeployHook(t *testing.T) {
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
	assert.Equal(t, model.EventDeploy, pipeline.Event)
	assert.Equal(t, "main", pipeline.Branch)
	assert.Equal(t, "refs/heads/main", pipeline.Ref)
	assert.Equal(t, *from.Deployment.SHA, pipeline.Commit)
	assert.Equal(t, *from.Deployment.Description, pipeline.Message)
	assert.Equal(t, *from.Deployment.URL, pipeline.ForgeURL)
	assert.Equal(t, *from.Sender.Login, pipeline.Author)
	assert.Equal(t, *from.Sender.AvatarURL, pipeline.Avatar)
}

func Test_parsePushHook(t *testing.T) {
	t.Run("convert push from webhook", func(t *testing.T) {
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
		assert.Equal(t, model.EventPush, pipeline.Event)
		assert.Equal(t, "main", pipeline.Branch)
		assert.Equal(t, "refs/heads/main", pipeline.Ref)
		assert.Equal(t, *from.HeadCommit.ID, pipeline.Commit)
		assert.Equal(t, *from.HeadCommit.Message, pipeline.Message)
		assert.Equal(t, *from.HeadCommit.URL, pipeline.ForgeURL)
		assert.Equal(t, *from.Sender.Login, pipeline.Author)
		assert.Equal(t, *from.Sender.AvatarURL, pipeline.Avatar)
		assert.Equal(t, *from.HeadCommit.Author.Email, pipeline.Email)
	})

	t.Run("convert tag from webhook", func(t *testing.T) {
		from := &github.PushEvent{}
		from.Ref = github.String("refs/tags/v1.0.0")

		_, pipeline := parsePushHook(from)
		assert.Equal(t, model.EventTag, pipeline.Event)
		assert.Equal(t, "refs/tags/v1.0.0", pipeline.Ref)
	})

	t.Run("convert tag's base branch to pipeline's branch ", func(t *testing.T) {
		from := &github.PushEvent{}
		from.Ref = github.String("refs/tags/v1.0.0")
		from.BaseRef = github.String("refs/heads/main")

		_, pipeline := parsePushHook(from)
		assert.Equal(t, model.EventTag, pipeline.Event)
		assert.Equal(t, "main", pipeline.Branch)
	})

	t.Run("not convert tag's base_ref from webhook if not prefixed with 'ref/heads/'", func(t *testing.T) {
		from := &github.PushEvent{}
		from.Ref = github.String("refs/tags/v1.0.0")
		from.BaseRef = github.String("refs/refs/main")

		_, pipeline := parsePushHook(from)
		assert.Equal(t, model.EventTag, pipeline.Event)
		assert.Equal(t, "refs/tags/v1.0.0", pipeline.Branch)
	})
}
