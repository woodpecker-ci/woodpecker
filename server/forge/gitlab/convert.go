// Copyright 2021 Woodpecker Authors
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

package gitlab

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/xanzy/go-gitlab"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/shared/utils"
)

const (
	mergeRefs               = "refs/merge-requests/%d/head" // merge request merged with base
	VisibilityLevelInternal = 10
)

func (g *GitLab) convertGitLabRepo(_repo *gitlab.Project, projectMember *gitlab.ProjectMember) (*model.Repo, error) {
	parts := strings.Split(_repo.PathWithNamespace, "/")
	owner := strings.Join(parts[:len(parts)-1], "/")
	name := parts[len(parts)-1]
	repo := &model.Repo{
		ForgeRemoteID: model.ForgeRemoteID(fmt.Sprint(_repo.ID)),
		Owner:         owner,
		Name:          name,
		FullName:      _repo.PathWithNamespace,
		Avatar:        _repo.AvatarURL,
		ForgeURL:      _repo.WebURL,
		Clone:         _repo.HTTPURLToRepo,
		CloneSSH:      _repo.SSHURLToRepo,
		Branch:        _repo.DefaultBranch,
		Visibility:    model.RepoVisibility(_repo.Visibility),
		IsSCMPrivate:  _repo.Visibility == gitlab.InternalVisibility || _repo.Visibility == gitlab.PrivateVisibility,
		Perm: &model.Perm{
			Pull:  isRead(_repo, projectMember),
			Push:  isWrite(projectMember),
			Admin: isAdmin(projectMember),
		},
		PREnabled: _repo.MergeRequestsEnabled,
	}

	if len(repo.Avatar) != 0 && !strings.HasPrefix(repo.Avatar, "http") {
		repo.Avatar = fmt.Sprintf("%s/%s", g.url, repo.Avatar)
	}

	return repo, nil
}

func convertMergeRequestHook(hook *gitlab.MergeEvent, req *http.Request) (int, *model.Repo, *model.Pipeline, error) {
	repo := &model.Repo{}
	pipeline := &model.Pipeline{}

	target := hook.ObjectAttributes.Target
	source := hook.ObjectAttributes.Source
	obj := hook.ObjectAttributes

	switch {
	case target == nil && source == nil:
		return 0, nil, nil, fmt.Errorf("target and source keys expected in merge request hook")
	case target == nil:
		return 0, nil, nil, fmt.Errorf("target key expected in merge request hook")
	case source == nil:
		return 0, nil, nil, fmt.Errorf("source key expected in merge request hook")
	}

	if target.PathWithNamespace != "" {
		var err error
		if repo.Owner, repo.Name, err = extractFromPath(target.PathWithNamespace); err != nil {
			return 0, nil, nil, err
		}
		repo.FullName = target.PathWithNamespace
	} else {
		repo.Owner = req.FormValue("owner")
		repo.Name = req.FormValue("name")
		repo.FullName = fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
	}

	repo.ForgeRemoteID = model.ForgeRemoteID(fmt.Sprint(obj.TargetProjectID))
	repo.ForgeURL = target.WebURL

	if target.GitHTTPURL != "" {
		repo.Clone = target.GitHTTPURL
	} else {
		repo.Clone = target.HTTPURL
	}
	if target.GitSSHURL != "" {
		repo.CloneSSH = target.GitSSHURL
	} else {
		repo.CloneSSH = target.SSHURL
	}

	repo.Branch = target.DefaultBranch

	if target.AvatarURL != "" {
		repo.Avatar = target.AvatarURL
	}

	pipeline.Event = model.EventPull
	if obj.State == "closed" || obj.State == "merged" {
		pipeline.Event = model.EventPullClosed
	}

	lastCommit := obj.LastCommit

	pipeline.Message = lastCommit.Message
	pipeline.Commit = lastCommit.ID

	pipeline.Ref = fmt.Sprintf(mergeRefs, obj.IID)
	pipeline.Branch = obj.SourceBranch
	pipeline.Refspec = fmt.Sprintf("%s:%s", obj.SourceBranch, obj.TargetBranch)

	author := lastCommit.Author

	pipeline.Author = author.Name
	pipeline.Email = author.Email

	if len(pipeline.Email) != 0 {
		pipeline.Avatar = getUserAvatar(pipeline.Email)
	}

	pipeline.Title = obj.Title
	pipeline.ForgeURL = obj.URL
	pipeline.PullRequestLabels = convertLabels(hook.Labels)

	return obj.IID, repo, pipeline, nil
}

func convertPushHook(hook *gitlab.PushEvent) (*model.Repo, *model.Pipeline, error) {
	repo := &model.Repo{}
	pipeline := &model.Pipeline{}

	var err error
	if repo.Owner, repo.Name, err = extractFromPath(hook.Project.PathWithNamespace); err != nil {
		return nil, nil, err
	}

	repo.ForgeRemoteID = model.ForgeRemoteID(fmt.Sprint(hook.ProjectID))
	repo.Avatar = hook.Project.AvatarURL
	repo.ForgeURL = hook.Project.WebURL
	repo.Clone = hook.Project.GitHTTPURL
	repo.CloneSSH = hook.Project.GitSSHURL
	repo.FullName = hook.Project.PathWithNamespace
	repo.Branch = hook.Project.DefaultBranch

	switch hook.Project.Visibility {
	case gitlab.PrivateVisibility:
		repo.IsSCMPrivate = true
	case gitlab.InternalVisibility:
		repo.IsSCMPrivate = true
	case gitlab.PublicVisibility:
		repo.IsSCMPrivate = false
	}

	pipeline.Event = model.EventPush
	pipeline.Commit = hook.After
	pipeline.Branch = strings.TrimPrefix(hook.Ref, "refs/heads/")
	pipeline.Ref = hook.Ref

	// assume a capacity of 4 changed files per commit
	files := make([]string, 0, len(hook.Commits)*4)
	for _, cm := range hook.Commits {
		if hook.After == cm.ID {
			pipeline.Author = cm.Author.Name
			pipeline.Email = cm.Author.Email
			pipeline.Message = cm.Message
			pipeline.Timestamp = cm.Timestamp.Unix()
			if len(pipeline.Email) != 0 {
				pipeline.Avatar = getUserAvatar(pipeline.Email)
			}
		}

		files = append(files, cm.Added...)
		files = append(files, cm.Removed...)
		files = append(files, cm.Modified...)
	}
	pipeline.ChangedFiles = utils.DeduplicateStrings(files)

	return repo, pipeline, nil
}

func convertTagHook(hook *gitlab.TagEvent) (*model.Repo, *model.Pipeline, error) {
	repo := &model.Repo{}
	pipeline := &model.Pipeline{}

	var err error
	if repo.Owner, repo.Name, err = extractFromPath(hook.Project.PathWithNamespace); err != nil {
		return nil, nil, err
	}

	repo.ForgeRemoteID = model.ForgeRemoteID(fmt.Sprint(hook.ProjectID))
	repo.Avatar = hook.Project.AvatarURL
	repo.ForgeURL = hook.Project.WebURL
	repo.Clone = hook.Project.GitHTTPURL
	repo.CloneSSH = hook.Project.GitSSHURL
	repo.FullName = hook.Project.PathWithNamespace
	repo.Branch = hook.Project.DefaultBranch

	switch hook.Project.Visibility {
	case gitlab.PrivateVisibility:
		repo.IsSCMPrivate = true
	case gitlab.InternalVisibility:
		repo.IsSCMPrivate = true
	case gitlab.PublicVisibility:
		repo.IsSCMPrivate = false
	}

	pipeline.Event = model.EventTag
	pipeline.Commit = hook.After
	pipeline.Branch = strings.TrimPrefix(hook.Ref, "refs/heads/")
	pipeline.Ref = hook.Ref

	for _, cm := range hook.Commits {
		if hook.After == cm.ID {
			pipeline.Author = cm.Author.Name
			pipeline.Email = cm.Author.Email
			pipeline.Message = cm.Message
			pipeline.Timestamp = cm.Timestamp.Unix()
			if len(pipeline.Email) != 0 {
				pipeline.Avatar = getUserAvatar(pipeline.Email)
			}
			break
		}
	}

	return repo, pipeline, nil
}

func convertReleaseHook(hook *gitlab.ReleaseEvent) (*model.Repo, *model.Pipeline, error) {
	repo := &model.Repo{}

	var err error
	if repo.Owner, repo.Name, err = extractFromPath(hook.Project.PathWithNamespace); err != nil {
		return nil, nil, err
	}

	repo.ForgeRemoteID = model.ForgeRemoteID(fmt.Sprint(hook.Project.ID))
	repo.Avatar = ""
	if hook.Project.AvatarURL != nil {
		repo.Avatar = *hook.Project.AvatarURL
	}
	repo.ForgeURL = hook.Project.WebURL
	repo.Clone = hook.Project.GitHTTPURL
	repo.CloneSSH = hook.Project.GitSSHURL
	repo.FullName = hook.Project.PathWithNamespace
	repo.Branch = hook.Project.DefaultBranch
	repo.IsSCMPrivate = hook.Project.VisibilityLevel > VisibilityLevelInternal

	pipeline := &model.Pipeline{
		Event:    model.EventRelease,
		Commit:   hook.Commit.ID,
		ForgeURL: hook.URL,
		Message:  fmt.Sprintf("created release %s", hook.Name),
		Sender:   hook.Commit.Author.Name,
		Author:   hook.Commit.Author.Name,
		Email:    hook.Commit.Author.Email,

		// Tag name here is the ref. We should add the refs/tags, so
		// it is known it's a tag (git-plugin looks for it)
		Ref: "refs/tags/" + hook.Tag,
	}
	if len(pipeline.Email) != 0 {
		pipeline.Avatar = getUserAvatar(pipeline.Email)
	}

	return repo, pipeline, nil
}

func getUserAvatar(email string) string {
	hasher := md5.New()
	hasher.Write([]byte(email))

	return fmt.Sprintf(
		"%s/%v.jpg?s=%s",
		gravatarBase,
		hex.EncodeToString(hasher.Sum(nil)),
		"128",
	)
}

// extractFromPath splits a repository path string into owner and name components.
// It requires at least two path components, otherwise an error is returned.
func extractFromPath(str string) (string, string, error) {
	const minPathComponents = 2

	s := strings.Split(str, "/")
	if len(s) < minPathComponents {
		return "", "", fmt.Errorf("minimum match not found")
	}
	return s[0], s[1], nil
}

func convertLabels(from []*gitlab.EventLabel) []string {
	labels := make([]string, len(from))
	for i, label := range from {
		labels[i] = label.Title
	}
	return labels
}
