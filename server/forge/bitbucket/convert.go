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
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/oauth2"

	"github.com/woodpecker-ci/woodpecker/server/forge/bitbucket/internal"
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
	case model.StatusPending, model.StatusRunning, model.StatusBlocked:
		return statusPending
	case model.StatusSuccess:
		return statusSuccess
	default:
		return statusFailure
	}
}

// convertRepo is a helper function used to convert a Bitbucket repository
// structure to the common Woodpecker repository structure.
func convertRepo(from *internal.Repo, perm *internal.RepoPerm) *model.Repo {
	repo := model.Repo{
		ForgeRemoteID: model.ForgeRemoteID(from.UUID),
		Clone:         cloneLink(from),
		Owner:         strings.Split(from.FullName, "/")[0],
		Name:          strings.Split(from.FullName, "/")[1],
		FullName:      from.FullName,
		Link:          from.Links.HTML.Href,
		IsSCMPrivate:  from.IsPrivate,
		Avatar:        from.Owner.Links.Avatar.Href,
		SCMKind:       model.SCMKind(from.Scm),
		Branch:        "master",
		Perm:          convertPerm(perm),
	}
	if repo.SCMKind == model.RepoHg {
		repo.Branch = "default"
	}
	return &repo
}

func convertPerm(from *internal.RepoPerm) *model.Perm {
	perms := new(model.Perm)
	switch from.Permission {
	case "admin":
		perms.Admin = true
		fallthrough
	case "write":
		perms.Push = true
		fallthrough
	default:
		perms.Pull = true
	}
	return perms
}

// cloneLink is a helper function that tries to extract the clone url from the
// repository object.
func cloneLink(repo *internal.Repo) string {
	var clone string

	// above we manually constructed the repository clone url. below we will
	// iterate through the list of clone links and attempt to instead use the
	// clone url provided by bitbucket.
	for _, link := range repo.Links.Clone {
		if link.Name == "https" {
			clone = link.Href
		}
	}

	// if no repository name is provided, we use the Html link. this excludes the
	// .git suffix, but will still clone the repo.
	if len(clone) == 0 {
		clone = repo.Links.HTML.Href
	}

	// if bitbucket tries to automatically populate the user in the url we must
	// strip it out.
	cloneurl, err := url.Parse(clone)
	if err == nil {
		cloneurl.User = nil
		clone = cloneurl.String()
	}

	return clone
}

// convertUser is a helper function used to convert a Bitbucket user account
// structure to the Woodpecker User structure.
func convertUser(from *internal.Account, token *oauth2.Token) *model.User {
	return &model.User{
		Login:         from.Login,
		Token:         token.AccessToken,
		Secret:        token.RefreshToken,
		Expiry:        token.Expiry.UTC().Unix(),
		Avatar:        from.Links.Avatar.Href,
		ForgeRemoteID: model.ForgeRemoteID(fmt.Sprint(from.ID)),
	}
}

// convertTeamList is a helper function used to convert a Bitbucket team list
// structure to the Woodpecker Team structure.
func convertWorkspaceList(from []*internal.Workspace) []*model.Team {
	var teams []*model.Team
	for _, workspace := range from {
		teams = append(teams, convertWorkspace(workspace))
	}
	return teams
}

// convertTeam is a helper function used to convert a Bitbucket team account
// structure to the Woodpecker Team structure.
func convertWorkspace(from *internal.Workspace) *model.Team {
	return &model.Team{
		Login:  from.Slug,
		Avatar: from.Links.Avatar.Href,
	}
}

// convertPullHook is a helper function used to convert a Bitbucket pull request
// hook to the Woodpecker pipeline struct holding commit information.
func convertPullHook(from *internal.PullRequestHook) *model.Pipeline {
	return &model.Pipeline{
		Event:  model.EventPull,
		Commit: from.PullRequest.Dest.Commit.Hash,
		Ref:    fmt.Sprintf("refs/heads/%s", from.PullRequest.Dest.Branch.Name),
		Refspec: fmt.Sprintf("%s:%s",
			from.PullRequest.Source.Branch.Name,
			from.PullRequest.Dest.Branch.Name,
		),
		CloneURL:  fmt.Sprintf("https://bitbucket.org/%s", from.PullRequest.Source.Repo.FullName),
		Link:      from.PullRequest.Links.HTML.Href,
		Branch:    from.PullRequest.Dest.Branch.Name,
		Message:   from.PullRequest.Desc,
		Avatar:    from.Actor.Links.Avatar.Href,
		Author:    from.Actor.Login,
		Sender:    from.Actor.Login,
		Timestamp: from.PullRequest.Updated.UTC().Unix(),
	}
}

// convertPushHook is a helper function used to convert a Bitbucket push
// hook to the Woodpecker pipeline struct holding commit information.
func convertPushHook(hook *internal.PushHook, change *internal.Change) *model.Pipeline {
	pipeline := &model.Pipeline{
		Commit:    change.New.Target.Hash,
		Link:      change.New.Target.Links.HTML.Href,
		Branch:    change.New.Name,
		Message:   change.New.Target.Message,
		Avatar:    hook.Actor.Links.Avatar.Href,
		Author:    hook.Actor.Login,
		Sender:    hook.Actor.Login,
		Timestamp: change.New.Target.Date.UTC().Unix(),
	}
	switch change.New.Type {
	case "tag", "annotated_tag", "bookmark":
		pipeline.Event = model.EventTag
		pipeline.Ref = fmt.Sprintf("refs/tags/%s", change.New.Name)
	default:
		pipeline.Event = model.EventPush
		pipeline.Ref = fmt.Sprintf("refs/heads/%s", change.New.Name)
	}
	if len(change.New.Target.Author.Raw) != 0 {
		pipeline.Email = extractEmail(change.New.Target.Author.Raw)
	}
	return pipeline
}

// regex for git author fields ("name <name@mail.tld>")
var reGitMail = regexp.MustCompile("<(.*)>")

// extracts the email from a git commit author string
func extractEmail(gitauthor string) (author string) {
	matches := reGitMail.FindAllStringSubmatch(gitauthor, -1)
	if len(matches) == 1 {
		author = matches[0][1]
	}
	return
}
