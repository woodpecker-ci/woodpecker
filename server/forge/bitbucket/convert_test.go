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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucket/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucket/internal"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

const (
	demoAvatarLinkRaw = "http://...avatar..."
	demoForgeURLRaw   = "https://bitbucket.org/foo/bar"
)

var (
	demoAvatarLinks = internal.WebhookLinks{
		linkKeyAvatar: struct {
			Href string `json:"href"`
		}{
			Href: demoAvatarLinkRaw,
		},
	}
	demoForgeURLLinks = internal.WebhookLinks{
		linkKeyHTML: struct {
			Href string `json:"href"`
		}{
			Href: demoForgeURLRaw,
		},
	}
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
	var from internal.Repo
	if !assert.NoError(t, json.Unmarshal([]byte(fixtures.APIRepo), &from)) {
		t.FailNow()
	}

	fromPerm := &internal.RepoPerm{
		Permission: "write",
	}

	to := convertRepo(&from, fromPerm)
	assert.Equal(t, "https://secure.gravatar.com/avatar/e3df5ba3ff85167eb228babbcd37481e?d=https%3A%2F%2Favatar-management--avatars.us-west-2.prod.public.atl-paas.net%2Fdefault-avatar-1.png", to.Avatar)
	assert.Equal(t, "6543/collect-webhooks", to.FullName)
	assert.Equal(t, "6543", to.Owner)
	assert.Equal(t, "collect-webhooks", to.Name)
	assert.Equal(t, "niam", to.Branch)
	assert.True(t, to.IsSCMPrivate)
	assert.Equal(t, "https://bitbucket.org/6543/collect-webhooks.git", to.Clone)
	assert.Equal(t, "https://bitbucket.org/6543/collect-webhooks", to.ForgeURL)
	assert.True(t, to.Perm.Push)
	assert.False(t, to.Perm.Admin)
}

func Test_convertWorkspace(t *testing.T) {
	from := &internal.Workspace{Slug: "octocat"}
	from.Links = demoAvatarLinks
	to := convertWorkspace(from)
	assert.Equal(t, demoAvatarLinkRaw, to.Avatar)
	assert.Equal(t, from.Slug, to.Login)
}

func Test_convertWorkspaceList(t *testing.T) {
	from := &internal.Workspace{Slug: "octocat"}
	from.Links = demoAvatarLinks
	to := convertWorkspaceList([]*internal.Workspace{from})
	assert.Equal(t, demoAvatarLinkRaw, to[0].Avatar)
	assert.Equal(t, from.Slug, to[0].Login)
}

func Test_convertUser(t *testing.T) {
	token := &oauth2.Token{
		AccessToken:  "foo",
		RefreshToken: "bar",
		Expiry:       time.Now(),
	}
	user := &internal.Account{Nickname: "octocat", DisplayName: "OctoCat"}
	user.Links = demoAvatarLinks

	result := convertUser(user, token)
	assert.Equal(t, demoAvatarLinkRaw, result.Avatar)
	assert.Equal(t, user.Nickname, result.Login)
	assert.Equal(t, token.AccessToken, result.AccessToken)
	assert.Equal(t, token.RefreshToken, result.RefreshToken)
	assert.Equal(t, token.Expiry.UTC().Unix(), result.Expiry)
}

func Test_cloneLink(t *testing.T) {
	var repo internal.Repo
	if !assert.NoError(t, json.Unmarshal([]byte(fixtures.APIRepo), &repo)) {
		t.FailNow()
	}
	assert.Equal(t, "https://bitbucket.org/6543/collect-webhooks.git", cloneLink(&repo))
	assert.Equal(t, "git@bitbucket.org:6543/collect-webhooks.git", sshCloneLink(&repo))

	repo = internal.Repo{}
	repo.Links.HTML.Href = "https://foo:bar@bitbucket.org/foo/bar.git"
	assert.Equal(t, "https://bitbucket.org/foo/bar.git", cloneLink(&repo))
}

func Test_convertPullHook(t *testing.T) {
	hook := &internal.PullRequestHook{}
	hook.Actor.Nickname = "octocat"
	hook.Actor.Links = demoAvatarLinks
	hook.PullRequest.Destination.Commit.Hash = "73f9c44d"
	hook.PullRequest.Destination.Branch.Name = "main"
	hook.PullRequest.Destination.Repo.Links = demoForgeURLLinks
	hook.PullRequest.Source.Branch.Name = "change"
	hook.PullRequest.Source.Repo.FullName = "baz/bar"
	hook.PullRequest.Source.Commit.Hash = "c8411d7"
	hook.PullRequest.Links = internal.WebhookLinks{
		linkKeyHTML: struct {
			Href string `json:"href"`
		}{
			Href: "https://bitbucket.org/foo/bar/pulls/5",
		},
	}

	hook.PullRequest.Title = "updated README"
	hook.PullRequest.Updated = time.Now()
	hook.PullRequest.ID = 1

	pipeline := convertPullHook(hook)
	assert.Equal(t, model.EventPull, pipeline.Event)
	assert.Equal(t, hook.Actor.Nickname, pipeline.Author)
	assert.Equal(t, demoAvatarLinkRaw, pipeline.Avatar)
	assert.Equal(t, hook.PullRequest.Source.Commit.Hash, pipeline.Commit)
	assert.Equal(t, hook.PullRequest.Source.Branch.Name, pipeline.Branch)
	assert.Equal(t, fmt.Sprintf("%s/pulls/5", demoForgeURLRaw), pipeline.ForgeURL)
	assert.Equal(t, "refs/pull-requests/1/from", pipeline.Ref)
	assert.Equal(t, "change:main", pipeline.Refspec)
	assert.Equal(t, hook.PullRequest.Title, pipeline.Message)
	assert.Equal(t, hook.PullRequest.Updated.Unix(), pipeline.Timestamp)
}

func Test_convertPushHook(t *testing.T) {
	change := internal.Change{}
	change.New.Target.Hash = "73f9c44d"
	change.New.Name = "main"
	change.New.Target.Links = internal.WebhookLinks{
		linkKeyHTML: struct {
			Href string `json:"href"`
		}{
			Href: "https://bitbucket.org/foo/bar/commits/73f9c44d",
		},
	}
	change.New.Target.Message = "updated README"
	change.New.Target.Date = time.Now()
	change.New.Target.Author.Raw = "Test <test@domain.tld>"

	hook := internal.PushHook{}
	hook.Actor.Nickname = "octocat"
	hook.Actor.Links = demoAvatarLinks

	pipeline := convertPushHook(&hook, &change)
	assert.Equal(t, model.EventPush, pipeline.Event)
	assert.Equal(t, "test@domain.tld", pipeline.Email)
	assert.Equal(t, hook.Actor.Nickname, pipeline.Author)
	assert.Equal(t, demoAvatarLinkRaw, pipeline.Avatar)
	assert.Equal(t, change.New.Target.Hash, pipeline.Commit)
	assert.Equal(t, change.New.Name, pipeline.Branch)
	assert.Equal(t, "https://bitbucket.org/foo/bar/commits/73f9c44d", pipeline.ForgeURL)
	assert.Equal(t, "refs/heads/main", pipeline.Ref)
	assert.Equal(t, change.New.Target.Message, pipeline.Message)
	assert.Equal(t, change.New.Target.Date.Unix(), pipeline.Timestamp)
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
