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

package bitbucketserver

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/mrjones/oauth"

	"github.com/woodpecker-ci/woodpecker/server/forge/bitbucketserver/internal"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

const (
	statusPending = "INPROGRESS"
	statusSuccess = "SUCCESSFUL"
	statusFailure = "FAILED"
)

// convertStatus is a helper function used to convert a Woodpecker status to a
// Bitbucket commit status.
func convertStatus(status model.StatusValue) string {
	switch status {
	case model.StatusPending, model.StatusRunning:
		return statusPending
	case model.StatusSuccess:
		return statusSuccess
	default:
		return statusFailure
	}
}

// convertRepo is a helper function used to convert a Bitbucket server repository
// structure to the common Woodpecker repository structure.
func convertRepo(from *internal.Repo, perm *model.Perm) *model.Repo {
	repo := model.Repo{
		ForgeRemoteID: model.ForgeRemoteID(fmt.Sprint(from.ID)),
		Name:          from.Slug,
		Owner:         from.Project.Key,
		Branch:        "master",
		SCMKind:       model.RepoGit,
		IsSCMPrivate:  true, // Since we have to use Netrc it has to always be private :/
		FullName:      fmt.Sprintf("%s/%s", from.Project.Key, from.Slug),
		Perm:          perm,
	}

	for _, item := range from.Links.Clone {
		if item.Name == "http" {
			uri, err := url.Parse(item.Href)
			if err != nil {
				return nil
			}
			uri.User = nil
			repo.Clone = uri.String()
		}
	}
	for _, item := range from.Links.Self {
		if item.Href != "" {
			repo.Link = item.Href
		}
	}
	return &repo
}

// convertPushHook is a helper function used to convert a Bitbucket push
// hook to the Woodpecker pipeline struct holding commit information.
func convertPushHook(hook *internal.PostHook, baseURL string) *model.Pipeline {
	branch := strings.TrimPrefix(
		strings.TrimPrefix(
			hook.RefChanges[0].RefID,
			"refs/heads/",
		),
		"refs/tags/",
	)

	// Ensuring the author label is not longer then 40 for the label of the commit author (default size in the db)
	authorLabel := hook.Changesets.Values[0].ToCommit.Author.Name
	if len(authorLabel) > 40 {
		authorLabel = authorLabel[0:37] + "..."
	}

	pipeline := &model.Pipeline{
		Commit:    hook.RefChanges[0].ToHash, // TODO check for index value
		Branch:    branch,
		Message:   hook.Changesets.Values[0].ToCommit.Message, // TODO check for index Values
		Avatar:    avatarLink(hook.Changesets.Values[0].ToCommit.Author.EmailAddress),
		Author:    authorLabel,
		Email:     hook.Changesets.Values[0].ToCommit.Author.EmailAddress,
		Timestamp: time.Now().UTC().Unix(),
		Ref:       hook.RefChanges[0].RefID, // TODO check for index Values
		Link:      fmt.Sprintf("%s/projects/%s/repos/%s/commits/%s", baseURL, hook.Repository.Project.Key, hook.Repository.Slug, hook.RefChanges[0].ToHash),
	}
	if strings.HasPrefix(hook.RefChanges[0].RefID, "refs/tags/") {
		pipeline.Event = model.EventTag
	} else {
		pipeline.Event = model.EventPush
	}

	return pipeline
}

// convertUser is a helper function used to convert a Bitbucket user account
// structure to the Woodpecker User structure.
func convertUser(from *internal.User, token *oauth.AccessToken) *model.User {
	return &model.User{
		Login:         from.Slug,
		Token:         token.Token,
		Email:         from.EmailAddress,
		Avatar:        avatarLink(from.EmailAddress),
		ForgeRemoteID: model.ForgeRemoteID(fmt.Sprint(from.ID)),
	}
}

func avatarLink(email string) string {
	hasher := md5.New()
	hasher.Write([]byte(strings.ToLower(email)))
	emailHash := fmt.Sprintf("%v", hex.EncodeToString(hasher.Sum(nil)))
	avatarURL := fmt.Sprintf("https://www.gravatar.com/avatar/%s.jpg", emailHash)
	return avatarURL
}
