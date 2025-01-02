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

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucket/internal"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func Test_convertStatus(t *testing.T) {
	assert.Equal(t, statusSuccess, convertStatus(model.StatusSuccess))
	assert.Equal(t, statusPending, convertStatus(model.StatusPending))
	assert.Equal(t, statusPending, convertStatus(model.StatusRunning))
	assert.Equal(t, statusFailure, convertStatus(model.StatusFailure))
	assert.Equal(t, statusFailure, convertStatus(model.StatusKilled))
	assert.Equal(t, statusFailure, convertStatus(model.StatusError))
}

func Test_convertRepo(t *testing.T) {
	from := &internal.Repo{
		FullName:  "octocat/hello-world",
		IsPrivate: true,
		Scm:       "git",
	}
	from.Owner.Links.Avatar.Href = "http://..."
	from.Links.HTML.Href = "https://bitbucket.org/foo/bar"
	from.MainBranch.Name = "default"
	fromPerm := &internal.RepoPerm{
		Permission: "write",
	}

	to := convertRepo(from, fromPerm)
	assert.Equal(t, from.Owner.Links.Avatar.Href, to.Avatar)
	assert.Equal(t, from.FullName, to.FullName)
	assert.Equal(t, "octocat", to.Owner)
	assert.Equal(t, "hello-world", to.Name)
	assert.Equal(t, "default", to.Branch)
	assert.Equal(t, from.IsPrivate, to.IsSCMPrivate)
	assert.Equal(t, from.Links.HTML.Href, to.Clone)
	assert.Equal(t, from.Links.HTML.Href, to.ForgeURL)
	assert.True(t, to.Perm.Push)
	assert.False(t, to.Perm.Admin)
}

func Test_convertWorkspace(t *testing.T) {
	from := &internal.Workspace{Slug: "octocat"}
	from.Links.Avatar.Href = "http://..."
	to := convertWorkspace(from)
	assert.Equal(t, from.Links.Avatar.Href, to.Avatar)
	assert.Equal(t, from.Slug, to.Login)
}

func Test_convertWorkspaceList(t *testing.T) {
	from := &internal.Workspace{Slug: "octocat"}
	from.Links.Avatar.Href = "http://..."
	to := convertWorkspaceList([]*internal.Workspace{from})
	assert.Equal(t, from.Links.Avatar.Href, to[0].Avatar)
	assert.Equal(t, from.Slug, to[0].Login)
}

func Test_convertUser(t *testing.T) {
	token := &oauth2.Token{
		AccessToken:  "foo",
		RefreshToken: "bar",
		Expiry:       time.Now(),
	}
	user := &internal.Account{Login: "octocat"}
	user.Links.Avatar.Href = "http://..."

	result := convertUser(user, token)
	assert.Equal(t, user.Links.Avatar.Href, result.Avatar)
	assert.Equal(t, user.Login, result.Login)
	assert.Equal(t, token.AccessToken, result.AccessToken)
	assert.Equal(t, token.RefreshToken, result.RefreshToken)
	assert.Equal(t, token.Expiry.UTC().Unix(), result.Expiry)
}

func Test_cloneLink(t *testing.T) {
	repo := &internal.Repo{}
	repo.Links.Clone = append(repo.Links.Clone, internal.Link{
		Name: "https",
		Href: "https://bitbucket.org/foo/bar.git",
	})
	link := cloneLink(repo)
	assert.Equal(t, repo.Links.Clone[0].Href, link)

	repo = &internal.Repo{}
	repo.Links.HTML.Href = "https://foo:bar@bitbucket.org/foo/bar.git"
	link = cloneLink(repo)
	assert.Equal(t, "https://bitbucket.org/foo/bar.git", link)
}

func Test_convertPullHook(t *testing.T) {
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
	hook.PullRequest.ID = 1

	pipeline := convertPullHook(hook)
	assert.Equal(t, model.EventPull, pipeline.Event)
	assert.Equal(t, hook.Actor.Login, pipeline.Author)
	assert.Equal(t, hook.Actor.Links.Avatar.Href, pipeline.Author.Author)
	assert.Equal(t, hook.PullRequest.Source.Commit.Hash, pipeline.Commit)
	assert.Equal(t, hook.PullRequest.Source.Branch.Name, pipeline.Branch)
	assert.Equal(t, hook.PullRequest.Links.HTML.Href, pipeline.ForgeURL)
	assert.Equal(t, "refs/pull-requests/1/from", pipeline.Ref)
	assert.Equal(t, "change:main", pipeline.Refspec)
	assert.Equal(t, hook.PullRequest.Title, pipeline.PullRequest.Title)
}

func Test_convertPushHook(t *testing.T) {
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
	assert.Equal(t, model.EventPush, pipeline.Event)
	assert.Equal(t, "test@domain.tld", pipeline.Commit.Author.Email)
	assert.Equal(t, hook.Actor.Login, pipeline.Author)
	assert.Equal(t, hook.Actor.Links.Avatar.Href, pipeline.Author.Avatar)
	assert.Equal(t, change.New.Target.Hash, pipeline.Commit)
	assert.Equal(t, change.New.Name, pipeline.Branch)
	assert.Equal(t, change.New.Target.Links.HTML.Href, pipeline.ForgeURL)
	assert.Equal(t, "refs/heads/main", pipeline.Ref)
	assert.Equal(t, change.New.Target.Message, pipeline.Commit.Message)
}

func Test_convertPushHookTag(t *testing.T) {
	change := internal.Change{}
	change.New.Name = "v1.0.0"
	change.New.Type = "tag"

	hook := internal.PushHook{}

	pipeline := convertPushHook(&hook, &change)
	assert.Equal(t, model.EventTag, pipeline.Event)
	assert.Equal(t, "refs/tags/v1.0.0", pipeline.Ref)
}
