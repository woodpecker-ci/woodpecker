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
)

func (g *Gitlab) convertGitlabRepo(_repo *gitlab.Project) (*model.Repo, error) {
	parts := strings.Split(_repo.PathWithNamespace, "/")
	// TODO(648) save repo id (support nested repos)
	owner := strings.Join(parts[:len(parts)-1], "/")
	name := parts[len(parts)-1]
	repo := &model.Repo{
		Owner:      owner,
		Name:       name,
		FullName:   _repo.PathWithNamespace,
		Avatar:     _repo.AvatarURL,
		Link:       _repo.WebURL,
		Clone:      _repo.HTTPURLToRepo,
		Branch:     _repo.DefaultBranch,
		Visibility: model.RepoVisibly(_repo.Visibility),
	}

	if len(repo.Branch) == 0 { // TODO: do we need that?
		repo.Branch = "master"
	}

	if len(repo.Avatar) != 0 && !strings.HasPrefix(repo.Avatar, "http") {
		repo.Avatar = fmt.Sprintf("%s/%s", g.URL, repo.Avatar)
	}

	if g.PrivateMode {
		repo.IsSCMPrivate = true
	} else {
		repo.IsSCMPrivate = !_repo.Public
	}

	return repo, nil
}

func convertMergeRequestHock(hook *gitlab.MergeEvent, req *http.Request) (*model.Repo, *model.Build, error) {
	repo := &model.Repo{}
	build := &model.Build{}

	target := hook.ObjectAttributes.Target
	source := hook.ObjectAttributes.Source
	obj := hook.ObjectAttributes

	if target == nil && source == nil {
		return nil, nil, fmt.Errorf("target and source keys expected in merge request hook")
	} else if target == nil {
		return nil, nil, fmt.Errorf("target key expected in merge request hook")
	} else if source == nil {
		return nil, nil, fmt.Errorf("source key expected in merge request hook")
	}

	if target.PathWithNamespace != "" {
		var err error
		if repo.Owner, repo.Name, err = extractFromPath(target.PathWithNamespace); err != nil {
			return nil, nil, err
		}
		repo.FullName = target.PathWithNamespace
	} else {
		repo.Owner = req.FormValue("owner")
		repo.Name = req.FormValue("name")
		repo.FullName = fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
	}

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

	build.Event = model.EventPull

	lastCommit := obj.LastCommit

	build.Message = lastCommit.Message
	build.Commit = lastCommit.ID
	build.Remote = obj.Source.HTTPURL

	build.Ref = fmt.Sprintf("refs/merge-requests/%d/head", obj.IID)

	build.Branch = obj.SourceBranch

	author := lastCommit.Author

	build.Author = author.Name
	build.Email = author.Email

	if len(build.Email) != 0 {
		build.Avatar = getUserAvatar(build.Email)
	}

	build.Title = obj.Title
	build.Link = obj.URL

	return repo, build, nil
}

func convertPushHock(hook *gitlab.PushEvent) (*model.Repo, *model.Build, error) {
	repo := &model.Repo{}
	build := &model.Build{}

	var err error
	if repo.Owner, repo.Name, err = extractFromPath(hook.Project.PathWithNamespace); err != nil {
		return nil, nil, err
	}

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

	build.Event = model.EventPush
	build.Commit = hook.After
	build.Branch = strings.TrimPrefix(hook.Ref, "refs/heads/")
	build.Ref = hook.Ref

	for _, cm := range hook.Commits {
		if hook.After == cm.ID {
			build.Author = cm.Author.Name
			build.Email = cm.Author.Email
			build.Message = cm.Message
			build.Timestamp = cm.Timestamp.Unix()
			if len(build.Email) != 0 {
				build.Avatar = getUserAvatar(build.Email)
			}
			break
		}
	}

	return repo, build, nil
}

func convertTagHock(hook *gitlab.TagEvent) (*model.Repo, *model.Build, error) {
	repo := &model.Repo{}
	build := &model.Build{}

	var err error
	if repo.Owner, repo.Name, err = extractFromPath(hook.Project.PathWithNamespace); err != nil {
		return nil, nil, err
	}

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

	build.Event = model.EventTag
	build.Commit = hook.After
	build.Branch = strings.TrimPrefix(hook.Ref, "refs/heads/")
	build.Ref = hook.Ref

	for _, cm := range hook.Commits {
		if hook.After == cm.ID {
			build.Author = cm.Author.Name
			build.Email = cm.Author.Email
			build.Message = cm.Message
			build.Timestamp = cm.Timestamp.Unix()
			if len(build.Email) != 0 {
				build.Avatar = getUserAvatar(build.Email)
			}
			break
		}
	}

	return repo, build, nil
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
