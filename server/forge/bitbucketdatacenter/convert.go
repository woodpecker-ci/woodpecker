// Copyright 2024 Woodpecker Authors
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

package bitbucketdatacenter

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	bb "github.com/neticdk/go-bitbucket/bitbucket"
	"golang.org/x/oauth2"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func convertStatus(status model.StatusValue) bb.BuildStatusState {
	switch status {
	case model.StatusPending, model.StatusRunning:
		return bb.BuildStatusStateInProgress
	case model.StatusSuccess:
		return bb.BuildStatusStateSuccessful
	default:
		return bb.BuildStatusStateFailed
	}
}

func convertID(id uint64) model.ForgeRemoteID {
	return model.ForgeRemoteID(fmt.Sprintf("%d", id))
}

func anonymizeLink(link string) (href string) {
	parsed, err := url.Parse(link)
	if err != nil {
		return link
	}
	parsed.User = nil
	return parsed.String()
}

func convertRepo(from *bb.Repository, perm *model.Perm, branch string) *model.Repo {
	r := &model.Repo{
		ForgeRemoteID: convertID(from.ID),
		Name:          from.Slug,
		Owner:         from.Project.Key,
		Branch:        branch,
		SCMKind:       model.RepoGit,
		IsSCMPrivate:  true, // Since we have to use Netrc it has to always be private :/ TODO: Is this really true?
		FullName:      fmt.Sprintf("%s/%s", from.Project.Key, from.Slug),
		Perm:          perm,
		PREnabled:     true,
	}

	for _, l := range from.Links["clone"] {
		if l.Name == "http" {
			r.Clone = anonymizeLink(l.Href)
		}
	}

	if l, ok := from.Links["self"]; ok && len(l) > 0 {
		r.ForgeURL = l[0].Href
	}

	return r
}

func convertRepositoryPushEvent(ev *bb.RepositoryPushEvent, baseURL string) *model.Pipeline {
	if len(ev.Changes) == 0 {
		return nil
	}
	change := ev.Changes[0]
	if change.ToHash == "0000000000000000000000000000000000000000" {
		// No ToHash present - could be "DELETE"
		return nil
	}
	if change.Type == bb.RepositoryPushEventChangeTypeDelete {
		return nil
	}

	pipeline := &model.Pipeline{
		Commit:    change.ToHash,
		Branch:    change.Ref.DisplayID,
		Message:   "",
		Avatar:    bitbucketAvatarURL(baseURL, ev.Actor.Slug),
		Author:    authorLabel(ev.Actor.Name),
		Email:     ev.Actor.Email,
		Timestamp: time.Time(ev.Date).UTC().Unix(),
		Ref:       ev.Changes[0].RefId,
		ForgeURL:  fmt.Sprintf("%s/projects/%s/repos/%s/commits/%s", baseURL, ev.Repository.Project.Key, ev.Repository.Slug, change.ToHash),
	}

	if strings.HasPrefix(ev.Changes[0].RefId, "refs/tags/") {
		pipeline.Event = model.EventTag
	} else {
		pipeline.Event = model.EventPush
	}

	return pipeline
}

func convertPullRequestEvent(ev *bb.PullRequestEvent, baseURL string) *model.Pipeline {
	pipeline := &model.Pipeline{
		Commit:    ev.PullRequest.Source.Latest,
		Branch:    ev.PullRequest.Source.DisplayID,
		Title:     ev.PullRequest.Title,
		Message:   "",
		Avatar:    bitbucketAvatarURL(baseURL, ev.Actor.Slug),
		Author:    authorLabel(ev.Actor.Name),
		Email:     ev.Actor.Email,
		Timestamp: time.Time(ev.Date).UTC().Unix(),
		Ref:       fmt.Sprintf("refs/pull-requests/%d/from", ev.PullRequest.ID),
		ForgeURL:  fmt.Sprintf("%s/projects/%s/repos/%s/commits/%s", baseURL, ev.PullRequest.Source.Repository.Project.Key, ev.PullRequest.Source.Repository.Slug, ev.PullRequest.Source.Latest),
		Refspec:   fmt.Sprintf("%s:%s", ev.PullRequest.Source.DisplayID, ev.PullRequest.Target.DisplayID),
		FromFork:  ev.PullRequest.Source.Repository.ID != ev.PullRequest.Target.Repository.ID,
	}

	if ev.EventKey == bb.EventKeyPullRequestMerged || ev.EventKey == bb.EventKeyPullRequestDeclined || ev.EventKey == bb.EventKeyPullRequestDeleted {
		pipeline.Event = model.EventPullClosed
	} else {
		pipeline.Event = model.EventPull
	}

	return pipeline
}

func authorLabel(name string) string {
	var result string

	const maxNameLength = 40

	if len(name) > maxNameLength {
		result = name[0:37] + "..."
	} else {
		result = name
	}
	return result
}

func convertUser(user *bb.User, baseURL string) *model.User {
	return &model.User{
		ForgeRemoteID: model.ForgeRemoteID(fmt.Sprintf("%d", user.ID)),
		Login:         user.Slug,
		Email:         user.Email,
		Avatar:        bitbucketAvatarURL(baseURL, user.Slug),
	}
}

func bitbucketAvatarURL(baseURL, slug string) string {
	return fmt.Sprintf("%s/users/%s/avatar.png", baseURL, slug)
}

func convertListOptions(p *model.ListOptions) bb.ListOptions {
	if p.All {
		return bb.ListOptions{}
	}
	return bb.ListOptions{Limit: uint(p.PerPage), Start: uint((p.Page - 1) * p.PerPage)}
}

func updateUserCredentials(u *model.User, t *oauth2.Token) {
	u.AccessToken = t.AccessToken
	u.RefreshToken = t.RefreshToken
	u.Expiry = t.Expiry.UTC().Unix()
}
