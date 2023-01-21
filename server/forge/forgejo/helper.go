// Copyright The Forgejo Authors.
// SPDX-License-Identifier: Apache-2.0
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
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/woodpecker-ci/woodpecker/server/forge/forgejo/client"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/shared/utils"
)

// helper function that converts a Forgejo repository to a Woodpecker repository.
func toRepo(from *client.Repository) *model.Repo {
	name := strings.Split(from.FullName, "/")[1]
	avatar := expandAvatar(
		from.HTMLURL,
		from.Owner.AvatarURL,
	)
	return &model.Repo{
		ForgeRemoteID: model.ForgeRemoteID(fmt.Sprint(from.ID)),
		SCMKind:       model.RepoGit,
		Name:          name,
		Owner:         from.Owner.UserName,
		FullName:      from.FullName,
		Avatar:        avatar,
		Link:          from.HTMLURL,
		IsSCMPrivate:  from.Private || from.Owner.Visibility != client.VisibleTypePublic,
		Clone:         from.CloneURL,
		Branch:        from.DefaultBranch,
	}
}

// helper function that converts a Forgejo permission to a Woodpecker permission.
func toPerm(from *client.Permission) *model.Perm {
	return &model.Perm{
		Pull:  from.Pull,
		Push:  from.Push,
		Admin: from.Admin,
	}
}

// helper function that converts a Forgejo team to a Woodpecker team.
func toTeam(from *client.Organization, link string) *model.Team {
	return &model.Team{
		Login:  from.UserName,
		Avatar: expandAvatar(link, from.AvatarURL),
	}
}

// helper function that extracts the Pipeline data from a Forgejo push hook
func pipelineFromPush(hook *pushHook) *model.Pipeline {
	avatar := expandAvatar(
		hook.Repo.HTMLURL,
		fixMalformedAvatar(hook.Sender.AvatarURL),
	)

	message := ""
	link := hook.Compare
	if len(hook.Commits) > 0 {
		message = hook.Commits[0].Message
	} else {
		message = hook.HeadCommit.Message
	}

	if len(hook.Commits) == 1 {
		link = hook.Commits[0].URL
	}

	return &model.Pipeline{
		Event:        model.EventPush,
		Commit:       hook.After,
		Ref:          hook.Ref,
		Link:         link,
		Branch:       strings.TrimPrefix(hook.Ref, "refs/heads/"),
		Message:      message,
		Avatar:       avatar,
		Author:       hook.Sender.UserName,
		Email:        hook.Sender.Email,
		Timestamp:    time.Now().UTC().Unix(),
		Sender:       hook.Sender.UserName,
		ChangedFiles: getChangedFilesFromPushHook(hook),
	}
}

func getChangedFilesFromPushHook(hook *pushHook) []string {
	// assume a capacity of 4 changed files per commit
	files := make([]string, 0, len(hook.Commits)*4)
	for _, c := range hook.Commits {
		files = append(files, c.Added...)
		files = append(files, c.Removed...)
		files = append(files, c.Modified...)
	}

	files = append(files, hook.HeadCommit.Added...)
	files = append(files, hook.HeadCommit.Removed...)
	files = append(files, hook.HeadCommit.Modified...)

	return utils.DedupStrings(files)
}

// helper function that extracts the Pipeline data from a Forgejo tag hook
func pipelineFromTag(hook *pushHook) *model.Pipeline {
	avatar := expandAvatar(
		hook.Repo.HTMLURL,
		fixMalformedAvatar(hook.Sender.AvatarURL),
	)

	return &model.Pipeline{
		Event:     model.EventTag,
		Commit:    hook.Sha,
		Ref:       fmt.Sprintf("refs/tags/%s", hook.Ref),
		Link:      fmt.Sprintf("%s/src/tag/%s", hook.Repo.HTMLURL, hook.Ref),
		Branch:    fmt.Sprintf("refs/tags/%s", hook.Ref),
		Message:   fmt.Sprintf("created tag %s", hook.Ref),
		Avatar:    avatar,
		Author:    hook.Sender.UserName,
		Sender:    hook.Sender.UserName,
		Timestamp: time.Now().UTC().Unix(),
	}
}

// helper function that extracts the Pipeline data from a Forgejo pull_request hook
func pipelineFromPullRequest(hook *pullRequestHook) *model.Pipeline {
	avatar := expandAvatar(
		hook.Repo.HTMLURL,
		fixMalformedAvatar(hook.PullRequest.Poster.AvatarURL),
	)
	pipeline := &model.Pipeline{
		Event:   model.EventPull,
		Commit:  hook.PullRequest.Head.Sha,
		Link:    hook.PullRequest.URL,
		Ref:     fmt.Sprintf("refs/pull/%d/head", hook.Number),
		Branch:  hook.PullRequest.Base.Ref,
		Message: hook.PullRequest.Title,
		Author:  hook.PullRequest.Poster.UserName,
		Avatar:  avatar,
		Sender:  hook.Sender.UserName,
		Title:   hook.PullRequest.Title,
		Refspec: fmt.Sprintf("%s:%s",
			hook.PullRequest.Head.Ref,
			hook.PullRequest.Base.Ref,
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
// (currently a known bug with forgejo)
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

func matchingHooks(hooks []*client.Hook, rawurl string) *client.Hook {
	link, err := url.Parse(rawurl)
	if err != nil {
		return nil
	}
	for _, hook := range hooks {
		if val, ok := hook.Config["url"]; ok {
			hookurl, err := url.Parse(val)
			if err == nil && hookurl.Host == link.Host {
				return hook
			}
		}
	}
	return nil
}
