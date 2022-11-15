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

package gogs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/gogits/go-gogs-client"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

// helper function that converts a Gogs repository to a Woodpecker repository.
func toRepo(from *gogs.Repository, privateMode bool) *model.Repo {
	name := strings.Split(from.FullName, "/")[1]
	avatar := expandAvatar(
		from.HTMLURL,
		from.Owner.AvatarUrl,
	)
	return &model.Repo{
		ForgeRemoteID: model.ForgeRemoteID(fmt.Sprint(from.ID)),
		SCMKind:       model.RepoGit,
		Name:          name,
		Owner:         from.Owner.UserName,
		FullName:      from.FullName,
		Avatar:        avatar,
		Link:          from.HTMLURL,
		IsSCMPrivate:  from.Private || privateMode,
		Clone:         from.CloneURL,
		Branch:        from.DefaultBranch,
	}
}

// helper function that converts a Gogs permission to a Woodpecker permission.
func toPerm(from *gogs.Permission) *model.Perm {
	return &model.Perm{
		Pull:  from.Pull,
		Push:  from.Push,
		Admin: from.Admin,
	}
}

// helper function that converts a Gogs team to a Woodpecker team.
func toTeam(from *gogs.Organization, link string) *model.Team {
	return &model.Team{
		Login:  from.UserName,
		Avatar: expandAvatar(link, from.AvatarUrl),
	}
}

// helper function that extracts the Pipeline data from a Gogs push hook
func pipelineFromPush(hook *pushHook) *model.Pipeline {
	avatar := expandAvatar(
		hook.Repo.HTMLURL,
		fixMalformedAvatar(hook.Sender.AvatarUrl),
	)
	author := hook.Sender.Login
	if author == "" {
		author = hook.Sender.UserName
	}
	sender := hook.Sender.UserName
	if sender == "" {
		sender = hook.Sender.Login
	}

	return &model.Pipeline{
		Event:     model.EventPush,
		Commit:    hook.After,
		Ref:       hook.Ref,
		Link:      hook.Compare,
		Branch:    strings.TrimPrefix(hook.Ref, "refs/heads/"),
		Message:   hook.Commits[0].Message,
		Avatar:    avatar,
		Author:    author,
		Email:     hook.Sender.Email,
		Timestamp: time.Now().UTC().Unix(),
		Sender:    sender,
	}
}

// helper function that extracts the pipeline data from a Gogs tag hook
func pipelineFromTag(hook *pushHook) *model.Pipeline {
	avatar := expandAvatar(
		hook.Repo.HTMLURL,
		fixMalformedAvatar(hook.Sender.AvatarUrl),
	)
	author := hook.Sender.Login
	if author == "" {
		author = hook.Sender.UserName
	}
	sender := hook.Sender.UserName
	if sender == "" {
		sender = hook.Sender.Login
	}

	return &model.Pipeline{
		Event:     model.EventTag,
		Commit:    hook.After,
		Ref:       fmt.Sprintf("refs/tags/%s", hook.Ref),
		Link:      fmt.Sprintf("%s/src/%s", hook.Repo.HTMLURL, hook.Ref),
		Branch:    fmt.Sprintf("refs/tags/%s", hook.Ref),
		Message:   fmt.Sprintf("created tag %s", hook.Ref),
		Avatar:    avatar,
		Author:    author,
		Sender:    sender,
		Timestamp: time.Now().UTC().Unix(),
	}
}

// helper function that extracts the Pipeline data from a Gogs pull_request hook
func pipelineFromPullRequest(hook *pullRequestHook) *model.Pipeline {
	avatar := expandAvatar(
		hook.Repo.HTMLURL,
		fixMalformedAvatar(hook.PullRequest.User.AvatarUrl),
	)
	sender := hook.Sender.UserName
	if sender == "" {
		sender = hook.Sender.Login
	}
	pipeline := &model.Pipeline{
		Event:   model.EventPull,
		Commit:  hook.PullRequest.Head.Sha,
		Link:    hook.PullRequest.URL,
		Ref:     fmt.Sprintf("refs/pull/%d/head", hook.Number),
		Branch:  hook.PullRequest.BaseBranch,
		Message: hook.PullRequest.Title,
		Author:  hook.PullRequest.User.UserName,
		Avatar:  avatar,
		Sender:  sender,
		Title:   hook.PullRequest.Title,
		Refspec: fmt.Sprintf("%s:%s",
			hook.PullRequest.HeadBranch,
			hook.PullRequest.BaseBranch,
		),
	}
	return pipeline
}

// helper function that parses a push hook from a read closer.
func parsePush(r io.Reader) (*pushHook, error) {
	push := new(pushHook)
	err := json.NewDecoder(r).Decode(push)
	return push, err
}

func parsePullRequest(r io.Reader) (*pullRequestHook, error) {
	pr := new(pullRequestHook)
	err := json.NewDecoder(r).Decode(pr)
	return pr, err
}

// fixMalformedAvatar is a helper function that fixes an avatar url if malformed
// (currently a known bug with gogs)
func fixMalformedAvatar(url string) string {
	index := strings.Index(url, "///")
	if index != -1 {
		return url[index+1:]
	}
	index = strings.Index(url, "//avatars/")
	if index != -1 {
		return strings.Replace(url, "//avatars/", "/avatars/", -1)
	}
	return url
}

// expandAvatar is a helper function that converts a relative avatar URL to the
// absolute url.
func expandAvatar(repo, rawurl string) string {
	aurl, err := url.Parse(rawurl)
	if err != nil {
		return rawurl
	}
	if aurl.IsAbs() {
		// Url is already absolute
		return aurl.String()
	}

	// Resolve to base
	burl, err := url.Parse(repo)
	if err != nil {
		return rawurl
	}
	aurl = burl.ResolveReference(aurl)

	return aurl.String()
}
