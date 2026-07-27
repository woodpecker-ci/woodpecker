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

package atomgit

import (
	"fmt"
	"net/url"
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

// toUser converts a AtomGit user payload to a Woodpecker user.
func toUser(from *user) *model.User {
	avatar := expandAvatar(from.HTMLURL, from.AvatarURL)
	return &model.User{
		ForgeRemoteID: model.ForgeRemoteID(from.ID),
		Login:         from.Username,
		Email:         from.Email,
		Avatar:        avatar,
	}
}

// repoOwnerName splits a full name ("owner/repo") into owner and name.
func repoOwnerName(fullName string) (owner, name string) {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return fullName, fullName
}

// toRepo converts a AtomGit repository payload to a Woodpecker repository.
func toRepo(from *repository) *model.Repo {
	owner, name := repoOwnerName(from.FullName)
	if owner == "" && from.Namespace != nil {
		owner = from.Namespace.Path
		name = from.Name
	}
	if from.FullName == "" {
		from.FullName = fmt.Sprintf("%s/%s", owner, name)
	}

	avatar := expandAvatar(from.WebURL, from.AvatarURL)
	clone := from.GitURL
	if clone == "" && from.HTTPURL() != "" {
		clone = from.HTTPURL()
	}

	// AtomGit's API does not return `html_url` on repository objects; the web
	// page URL is carried by `web_url` (or derivable from `http_url_to_repo`
	// by stripping the `.git` suffix). Default to that so the UI "Open in
	// forge" link renders as an <a> rather than an empty <button>.
	forgeURL := from.WebURL
	if forgeURL == "" && clone != "" {
		forgeURL = strings.TrimSuffix(clone, ".git")
	}

	// AtomGit supports merge requests (Pull requests) but the API has no
	// `has_pull_requests` field, so HasPullRequests is always false. Treat PRs
	// as enabled so the Pull requests tab is shown.
	//
	// Derive the SSH clone URL from the HTTP clone URL host so clone and
	// cloneSSH always point at the same forge host.
	cloneSSH := sshURLFromClone(clone, owner, name)

	repo := &model.Repo{
		ForgeRemoteID: model.ForgeRemoteID(from.ID),
		Owner:         owner,
		Name:          name,
		FullName:      from.FullName,
		Avatar:        avatar,
		ForgeURL:      forgeURL,
		Clone:         clone,
		CloneSSH:      cloneSSH,
		Branch:        from.DefaultBranch,
		IsSCMPrivate:  !from.Public.Bool(),
		PREnabled:     true,
	}
	if repo.Branch == "" {
		repo.Branch = "master"
	}
	if perm := toPerm(from); perm != nil {
		repo.Perm = perm
	}
	return repo
}

// toPerm derives Woodpecker permissions from a AtomGit repository's permissions.
func toPerm(from *repository) *model.Perm {
	// AtomGit access levels: 10 Guest, 20 Reporter, 30 Developer, 40 Maintainer, 50 Owner.
	if from.Permissions == nil {
		// The repository listing endpoint (/user/repos) does not include a
		// permissions object. These repos are returned for the authenticated
		// user, so treat them as fully accessible (pull/push/admin) so the
		// add-repository page and permission sync can use them without a nil
		// pointer dereference upstream.
		return &model.Perm{
			Pull:  true,
			Push:  true,
			Admin: true,
		}
	}
	level := 0
	if from.Permissions.ProjectAccess != nil {
		level = from.Permissions.ProjectAccess.AccessLevel
	}
	if from.Permissions.GroupAccess != nil && from.Permissions.GroupAccess.AccessLevel > level {
		level = from.Permissions.GroupAccess.AccessLevel
	}
	return &model.Perm{
		Pull:  level >= 20,
		Push:  level >= 30,
		Admin: level >= 40,
	}
}

// toTeam converts a AtomGit namespace into a Woodpecker team.
func toTeam(from *namespace, link string) *model.Team {
	return &model.Team{
		Login:  from.Path,
		Avatar: expandAvatar(link, from.AvatarURL),
	}
}

// toOrg converts a AtomGit namespace into a Woodpecker org.
func toOrg(from *namespace) *model.Org {
	return &model.Org{
		Name:    from.Path,
		Private: from.VisibilityLevel != 0 && from.VisibilityLevel != 20,
	}
}

// pipelineFromPush converts a AtomGit push webhook into a Woodpecker pipeline.
func pipelineFromPush(hook *pushHook) *model.Pipeline {
	avatar := expandAvatar(hook.Repository.HTMLURL, hook.UserAvatar)
	author := hook.UserName
	if author == "" {
		author = hook.UserEmail
	}

	var message, link string
	if len(hook.Commits) > 0 {
		message = hook.Commits[0].Message
		link = hook.Commits[0].URL
	}
	if message == "" {
		message = fmt.Sprintf("push %s", hook.After)
	}

	return &model.Pipeline{
		Event:        model.EventPush,
		Commit:       hook.After,
		Ref:          hook.Ref,
		ForgeURL:     link,
		Branch:       strings.TrimPrefix(hook.Ref, "refs/heads/"),
		Message:      message,
		Avatar:       avatar,
		Author:       author,
		Email:        hook.UserEmail,
		Timestamp:    0,
		Sender:       author,
		ChangedFiles: getChangedFilesFromPushHook(hook),
	}
}

// pipelineFromTag converts a AtomGit tag push webhook into a Woodpecker pipeline.
func pipelineFromTag(hook *tagPushHook) *model.Pipeline {
	avatar := expandAvatar(hook.Repository.HTMLURL, hook.UserAvatar)
	author := hook.UserName
	if author == "" {
		author = hook.UserEmail
	}
	ref := strings.TrimPrefix(hook.Ref, "refs/tags/")

	return &model.Pipeline{
		Event:     model.EventTag,
		Commit:    hook.After,
		Ref:       fmt.Sprintf("refs/tags/%s", ref),
		ForgeURL:  fmt.Sprintf("%s/-/tags/%s", hook.Repository.HTMLURL, ref),
		Message:   fmt.Sprintf("created tag %s", ref),
		Avatar:    avatar,
		Author:    author,
		Email:     hook.UserEmail,
		Timestamp: 0,
		Sender:    author,
	}
}

// pipelineFromMergeRequest converts a AtomGit merge request webhook into a Woodpecker pipeline.
func pipelineFromMergeRequest(hook *mergeRequestHook) *model.Pipeline {
	pr := hook.ObjectAttributes
	avatar := expandAvatar(hook.Project.HTMLURL, "")
	if pr.Author != nil {
		avatar = expandAvatar(hook.Project.HTMLURL, pr.Author.AvatarURL)
	}

	event := model.EventPull
	switch hook.EventType {
	case actionClose, actionMerge:
		event = model.EventPullClosed
	case actionUpdate:
		event = model.EventPull
	}

	base := pr.TargetBranch
	head := pr.SourceBranch

	commitSHA := ""
	if pr.LastCommit != nil {
		commitSHA = pr.LastCommit.ID
	}

	pipeline := &model.Pipeline{
		Event:    event,
		Commit:   commitSHA,
		ForgeURL: pr.HTTPURL,
		Ref:      fmt.Sprintf("refs/pull/%s/head", pr.IID),
		Branch:   base,
		Message:  pr.Title,
		Avatar:   avatar,
		Author:   authorLogin(pr.Author),
		Sender:   authorLogin(hook.User),
		Email:    authorEmail(pr.Author),
		Title:    pr.Title,
		Refspec:  fmt.Sprintf("%s:%s", head, base),
		FromFork: pr.SourceRepo != nil && pr.TargetRepo != nil && pr.SourceRepo.ID != pr.TargetRepo.ID,
	}
	if labels := convertLabels(hook.Labels); len(labels) > 0 {
		pipeline.PullRequestLabels = labels
	} else if labels := convertLabels(pr.Labels); len(labels) > 0 {
		pipeline.PullRequestLabels = labels
	}
	if pr.Milestone != nil {
		pipeline.PullRequestMilestone = pr.Milestone.Title
	}
	pipeline.PullRequestDraft = pr.Draft
	return pipeline
}

func authorLogin(u *user) string {
	if u == nil {
		return ""
	}
	return u.Username
}

func authorEmail(u *user) string {
	if u == nil {
		return ""
	}
	return u.Email
}

func convertLabels(from []label) []string {
	labels := make([]string, 0, len(from))
	for _, label := range from {
		labels = append(labels, label.Name)
	}
	return labels
}

// getChangedFilesFromPushHook collects changed files from a push webhook payload.
func getChangedFilesFromPushHook(hook *pushHook) []string {
	files := make([]string, 0)
	for _, c := range hook.Commits {
		files = append(files, c.Added...)
		files = append(files, c.Modified...)
		files = append(files, c.Removed...)
	}
	return utils.DeduplicateStrings(files)
}

// HTTPURL returns the HTTP clone URL of a repository.
func (r *repository) HTTPURL() string {
	if r.GitURL != "" {
		return r.GitURL
	}
	return r.WebURL
}

// sshURLFromClone derives the SSH clone URL from the HTTP clone URL's host,
// so both clone methods point at the same forge host.
func sshURLFromClone(clone, owner, name string) string {
	if clone == "" || owner == "" || name == "" {
		return ""
	}
	host := clone
	if i := strings.Index(clone, "://"); i >= 0 {
		host = clone[i+len("://"):]
	}
	if i := strings.IndexByte(host, '/'); i >= 0 {
		host = host[:i]
	}
	return fmt.Sprintf("git@%s:%s/%s.git", host, owner, name)
}

// expandAvatar resolves a possibly-relative avatar URL against the base URL.
func expandAvatar(repo, rawURL string) string {
	if rawURL == "" {
		return rawURL
	}
	aURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if aURL.IsAbs() {
		return aURL.String()
	}
	bURL, err := url.Parse(repo)
	if err != nil {
		return rawURL
	}
	return bURL.ResolveReference(aURL).String()
}
