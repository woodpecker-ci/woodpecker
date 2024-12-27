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

package gitea

import (
	"bytes"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/gitea/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func Test_parsePush(t *testing.T) {
	t.Run("Should parse push hook payload", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPush)
		hook, err := parsePush(buf)
		assert.NoError(t, err)
		assert.Equal(t, "refs/heads/main", hook.Ref)
		assert.Equal(t, "ef98532add3b2feb7a137426bba1248724367df5", hook.After)
		assert.Equal(t, "4b2626259b5a97b6b4eab5e6cca66adb986b672b", hook.Before)
		assert.Equal(t, "http://gitea.golang.org/gordon/hello-world/compare/4b2626259b5a97b6b4eab5e6cca66adb986b672b...ef98532add3b2feb7a137426bba1248724367df5", hook.Compare)
		assert.Equal(t, "hello-world", hook.Repo.Name)
		assert.Equal(t, "http://gitea.golang.org/gordon/hello-world", hook.Repo.HTMLURL)
		assert.Equal(t, "gordon", hook.Repo.Owner.UserName)
		assert.Equal(t, "gordon/hello-world", hook.Repo.FullName)
		assert.Equal(t, "gordon@golang.org", hook.Repo.Owner.Email)
		assert.True(t, hook.Repo.Private)
		assert.Equal(t, "gordon@golang.org", hook.Pusher.Email)
		assert.Equal(t, "gordon", hook.Pusher.UserName)
		assert.Equal(t, "gordon", hook.Sender.UserName)
		assert.Equal(t, "http://gitea.golang.org///1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87", hook.Sender.AvatarURL)
	})
	t.Run("Should parse tag hook payload", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookTag)
		hook, err := parsePush(buf)
		assert.NoError(t, err)
		assert.Equal(t, "v1.0.0", hook.Ref)
		assert.Equal(t, "ef98532add3b2feb7a137426bba1248724367df5", hook.Sha)
		assert.Equal(t, "hello-world", hook.Repo.Name)
		assert.Equal(t, "http://gitea.golang.org/gordon/hello-world", hook.Repo.HTMLURL)
		assert.Equal(t, "gordon/hello-world", hook.Repo.FullName)
		assert.Equal(t, "gordon@golang.org", hook.Repo.Owner.Email)
		assert.Equal(t, "gordon", hook.Repo.Owner.UserName)
		assert.True(t, hook.Repo.Private)
		assert.Equal(t, "gordon", hook.Sender.UserName)
		assert.Equal(t, "https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87", hook.Sender.AvatarURL)
	})

	t.Run("Should return a Pipeline struct from a push hook", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPush)
		hook, _ := parsePush(buf)
		pipeline := pipelineFromPush(hook)
		assert.Equal(t, model.EventPush, pipeline.Event)
		assert.Equal(t, hook.After, pipeline.Commit)
		assert.Equal(t, hook.Ref, pipeline.Ref)
		assert.Equal(t, hook.Commits[0].URL, pipeline.ForgeURL)
		assert.Equal(t, "main", pipeline.Branch)
		assert.Equal(t, hook.Commits[0].Message, pipeline.Message)
		assert.Equal(t, "http://1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87", pipeline.Avatar)
		assert.Equal(t, hook.Sender.UserName, pipeline.Author)
		assert.Equal(t, []string{"CHANGELOG.md", "app/controller/application.rb"}, pipeline.ChangedFiles)
	})

	t.Run("Should return a Repo struct from a push hook", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPush)
		hook, _ := parsePush(buf)
		repo := toRepo(hook.Repo)
		assert.Equal(t, hook.Repo.Name, repo.Name)
		assert.Equal(t, hook.Repo.Owner.UserName, repo.Owner)
		assert.Equal(t, "gordon/hello-world", repo.FullName)
		assert.Equal(t, hook.Repo.HTMLURL, repo.ForgeURL)
	})

	t.Run("Should return a Pipeline struct from a tag hook", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookTag)
		hook, _ := parsePush(buf)
		pipeline := pipelineFromTag(hook)
		assert.Equal(t, model.EventTag, pipeline.Event)
		assert.Equal(t, hook.Sha, pipeline.Commit)
		assert.Equal(t, "refs/tags/v1.0.0", pipeline.Ref)
		assert.Empty(t, pipeline.Branch)
		assert.Equal(t, "http://gitea.golang.org/gordon/hello-world/src/tag/v1.0.0", pipeline.ForgeURL)
		assert.Equal(t, "created tag v1.0.0", pipeline.Message)
	})
}

func Test_parsePullRequest(t *testing.T) {
	t.Run("Should parse pull_request hook payload", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPullRequest)
		hook, err := parsePullRequest(buf)
		assert.NoError(t, err)
		assert.Equal(t, "opened", hook.Action)
		assert.Equal(t, int64(1), hook.Number)

		assert.Equal(t, "hello-world", hook.Repo.Name)
		assert.Equal(t, "http://gitea.golang.org/gordon/hello-world", hook.Repo.HTMLURL)
		assert.Equal(t, "gordon/hello-world", hook.Repo.FullName)
		assert.Equal(t, "gordon@golang.org", hook.Repo.Owner.Email)
		assert.Equal(t, "gordon", hook.Repo.Owner.UserName)
		assert.True(t, hook.Repo.Private)
		assert.Equal(t, "gordon", hook.Sender.UserName)
		assert.Equal(t, "https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87", hook.Sender.AvatarURL)

		assert.Equal(t, "Update the README with new information", hook.PullRequest.Title)
		assert.Equal(t, "please merge", hook.PullRequest.Body)
		assert.Equal(t, gitea.StateOpen, hook.PullRequest.State)
		assert.Equal(t, "gordon", hook.PullRequest.Poster.UserName)
		assert.Equal(t, "main", hook.PullRequest.Base.Name)
		assert.Equal(t, "main", hook.PullRequest.Base.Ref)
		assert.Equal(t, "feature/changes", hook.PullRequest.Head.Name)
		assert.Equal(t, "feature/changes", hook.PullRequest.Head.Ref)
	})

	t.Run("Should return a Pipeline struct from a pull_request hook", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPullRequest)
		hook, _ := parsePullRequest(buf)
		pipeline := pipelineFromPullRequest(hook)
		assert.Equal(t, model.EventPull, pipeline.Event)
		assert.Equal(t, hook.PullRequest.Head.Sha, pipeline.Commit)
		assert.Equal(t, "refs/pull/1/head", pipeline.Ref)
		assert.Equal(t, "http://gitea.golang.org/gordon/hello-world/pull/1", pipeline.ForgeURL)
		assert.Equal(t, "main", pipeline.Branch)
		assert.Equal(t, "feature/changes:main", pipeline.Refspec)
		assert.Equal(t, hook.PullRequest.Title, pipeline.Message)
		assert.Equal(t, "http://1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87", pipeline.Avatar)
		assert.Equal(t, hook.PullRequest.Poster.UserName, pipeline.Author)
	})

	t.Run("Should return a Repo struct from a pull_request hook", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPullRequest)
		hook, _ := parsePullRequest(buf)
		repo := toRepo(hook.Repo)
		assert.Equal(t, hook.Repo.Name, repo.Name)
		assert.Equal(t, hook.Repo.Owner.UserName, repo.Owner)
		assert.Equal(t, "gordon/hello-world", repo.FullName)
		assert.Equal(t, hook.Repo.HTMLURL, repo.ForgeURL)
	})
}

func Test_toPerm(t *testing.T) {
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
		assert.Equal(t, from.Pull, perm.Pull)
		assert.Equal(t, from.Push, perm.Push)
		assert.Equal(t, from.Admin, perm.Admin)
	}
}

func Test_toTeam(t *testing.T) {
	from := &gitea.Organization{
		UserName:  "woodpecker",
		AvatarURL: "/avatars/1",
	}

	to := toTeam(from, "http://localhost:80")
	assert.Equal(t, from.UserName, to.Login)
	assert.Equal(t, "http://localhost:80/avatars/1", to.Avatar)
}

func Test_toRepo(t *testing.T) {
	from := gitea.Repository{
		FullName: "gophers/hello-world",
		Owner: &gitea.User{
			UserName:  "gordon",
			AvatarURL: "http://1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
		},
		CloneURL:      "http://gitea.golang.org/gophers/hello-world.git",
		HTMLURL:       "http://gitea.golang.org/gophers/hello-world",
		Private:       true,
		DefaultBranch: "main",
		Permissions:   &gitea.Permission{Admin: true},
	}
	repo := toRepo(&from)
	assert.Equal(t, from.FullName, repo.FullName)
	assert.Equal(t, from.Owner.UserName, repo.Owner)
	assert.Equal(t, "hello-world", repo.Name)
	assert.Equal(t, "main", repo.Branch)
	assert.Equal(t, from.HTMLURL, repo.ForgeURL)
	assert.Equal(t, from.CloneURL, repo.Clone)
	assert.Equal(t, from.Owner.AvatarURL, repo.Avatar)
	assert.Equal(t, from.Private, repo.IsSCMPrivate)
	assert.True(t, repo.Perm.Admin)
}

func Test_fixMalformedAvatar(t *testing.T) {
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
		assert.Equal(t, url.After, got)
	}
}

func Text_expandAvatar(t *testing.T) {
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
		assert.Equal(t, url.After, got)
	}
}
