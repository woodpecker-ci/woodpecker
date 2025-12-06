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

package sourcehut

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/sourcehut/git"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// toRepo converts a SourceHut repository to a Woodpecker repository.
func (c *SourceHut) toRepo(from *git.Repository, ver *git.Version) *model.Repo {
	u, _ := url.Parse(c.gitURL)
	fullName := fmt.Sprintf("%s/%s", from.Owner.CanonicalName, from.Name)
	cloneURL := fmt.Sprintf("%s/%s", c.gitURL, fullName)
	sshURL := fmt.Sprintf("%s@%s:%s", ver.Settings.SshUser, u.Hostname(), fullName)
	return &model.Repo{
		ForgeRemoteID: model.ForgeRemoteID(fullName),
		Name:          from.Name,
		Owner:         from.Owner.CanonicalName,
		FullName:      fullName,
		ForgeURL:      fmt.Sprintf("%s/%s", c.gitURL, fullName),
		IsSCMPrivate:  from.Visibility == git.VisibilityPrivate,
		Clone:         cloneURL,
		CloneSSH:      sshURL,
		Branch:        from.HEAD.Name,
		Perm: &model.Perm{
			Pull:  true,
			Push:  from.Access == git.AccessModeRw,
			Admin: from.Access == git.AccessModeRw,
		},
		PREnabled: false, // TODO
	}
}

func (c *SourceHut) toPushPipeline(from *git.GitEvent) *model.Pipeline {
	// TODO: Figure out what to do with several updates pushed at once
	update := from.Updates[0]
	if update.New == nil {
		return nil
	}

	commit, ok := update.New.Value.(*git.Commit)
	if !ok {
		return nil
	}

	forgeUrl := fmt.Sprintf("%s/%s/%s/commit/%s",
		c.gitURL, from.Repository.Owner.CanonicalName, from.Repository.Name,
		update.New.Id)

	return &model.Pipeline{
		Event:     model.EventPush,
		Commit:    update.New.Id,
		Ref:       strings.TrimPrefix(update.Ref.Name, "refs/heads/"),
		ForgeURL:  forgeUrl,
		Branch:    strings.TrimPrefix(update.Ref.Name, "refs/heads/"),
		Message:   commit.Message,
		Author:    commit.Author.Name,
		Email:     commit.Author.Email,
		Timestamp: time.Now().UTC().Unix(),
		Sender:    from.Pusher.CanonicalName[1:],
	}
}
