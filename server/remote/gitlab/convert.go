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

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/shared/utils"
)

const (
	mergeRefs = "refs/merge-requests/%d/head" // merge request merged with base
)

func (g *Gitlab) convertGitlabRepo(_repo *gitlab.Project) (*model.Repo, error) {
	parts := strings.Split(_repo.PathWithNamespace, "/")
	// TODO(648) save repo id (support nested repos)
	owner := strings.Join(parts[:len(parts)-1], "/")
	name := parts[len(parts)-1]
	repo := &model.Repo{
		RemoteID:     model.RemoteID(fmt.Sprint(_repo.ID)),
		Owner:        owner,
		Name:         name,
		FullName:     _repo.PathWithNamespace,
		Avatar:       _repo.AvatarURL,
		Link:         _repo.WebURL,
		Clone:        _repo.HTTPURLToRepo,
		Branch:       _repo.DefaultBranch,
		Visibility:   model.RepoVisibly(_repo.Visibility),
		IsSCMPrivate: !_repo.Public,
	}

	if len(repo.Branch) == 0 { // TODO: do we need that?
		repo.Branch = "master"
	}

	if len(repo.Avatar) != 0 && !strings.HasPrefix(repo.Avatar, "http") {
		repo.Avatar = fmt.Sprintf("%s/%s", g.URL, repo.Avatar)
	}

	return repo, nil
}

func convertMergeRequestHook(hook *gitlab.MergeEvent, req *http.Request) (int, *model.Repo, *model.Pipeline, error) {
	repo := &model.Repo{}
	pipeline := &model.Pipeline{}

	target := hook.ObjectAttributes.Target
	source := hook.ObjectAttributes.Source
	obj := hook.ObjectAttributes

	if target == nil && source == nil {
		return 0, nil, nil, fmt.Errorf("target and source keys expected in merge request hook")
	} else if target == nil {
		return 0, nil, nil, fmt.Errorf("target key expected in merge request hook")
	} else if source == nil {
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

	repo.RemoteID = model.RemoteID(fmt.Sprint(obj.TargetProjectID))
	repo.Link = target.WebURL

	if target.GitHTTPURL != "" {
		repo.Clone = target.GitHTTPURL
	} else {
		repo.Clone = target.HTTPURL
	}

	if target.DefaultBranch != "" {
		repo.Branch = target.DefaultBranch
	} else {
		repo.Branch = "master"
	}

	if target.AvatarURL != "" {
		repo.Avatar = target.AvatarURL
	}

	pipeline.Event = model.EventPull

	lastCommit := obj.LastCommit

	pipeline.Message = lastCommit.Message
	pipeline.Commit = lastCommit.ID
	pipeline.Remote = obj.Source.HTTPURL

	pipeline.Ref = fmt.Sprintf(mergeRefs, obj.IID)
	pipeline.Branch = obj.SourceBranch

	author := lastCommit.Author

	pipeline.Author = author.Name
	pipeline.Email = author.Email

	if len(pipeline.Email) != 0 {
		pipeline.Avatar = getUserAvatar(pipeline.Email)
	}

	pipeline.Title = obj.Title
	pipeline.Link = obj.URL
	pipeline.Labels = convertLabels(hook.Labels)

	return obj.IID, repo, pipeline, nil
}

func convertPushHook(hook *gitlab.PushEvent) (*model.Repo, *model.Pipeline, error) {
	repo := &model.Repo{}
	pipeline := &model.Pipeline{}

	var err error
	if repo.Owner, repo.Name, err = extractFromPath(hook.Project.PathWithNamespace); err != nil {
		return nil, nil, err
	}

	repo.RemoteID = model.RemoteID(fmt.Sprint(hook.ProjectID))
	repo.Avatar = hook.Project.AvatarURL
	repo.Link = hook.Project.WebURL
	repo.Clone = hook.Project.GitHTTPURL
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
	pipeline.ChangedFiles = utils.DedupStrings(files)

	return repo, pipeline, nil
}

func convertTagHook(hook *gitlab.TagEvent) (*model.Repo, *model.Pipeline, error) {
	repo := &model.Repo{}
	pipeline := &model.Pipeline{}

	var err error
	if repo.Owner, repo.Name, err = extractFromPath(hook.Project.PathWithNamespace); err != nil {
		return nil, nil, err
	}

	repo.RemoteID = model.RemoteID(fmt.Sprint(hook.ProjectID))
	repo.Avatar = hook.Project.AvatarURL
	repo.Link = hook.Project.WebURL
	repo.Clone = hook.Project.GitHTTPURL
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

func extractFromPath(str string) (string, string, error) {
	s := strings.Split(str, "/")
	if len(s) < 2 {
		return "", "", fmt.Errorf("Minimum match not found")
	}
	return s[0], s[1], nil
}

func convertLabels(from []*gitlab.Label) []string {
	labels := make([]string, len(from))
	for i, label := range from {
		labels[i] = label.Name
	}
	return labels
}
