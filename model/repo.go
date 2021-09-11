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

package model

import (
	"fmt"
	"strings"
)

type RepoLite struct {
	Owner    string `json:"owner"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Avatar   string `json:"avatar_url"`
}

// Repo represents a repository.
//
// swagger:model repo
type Repo struct {
	ID          int64  `json:"id,omitempty"             meddler:"repo_id,pk"`
	UserID      int64  `json:"-"                        meddler:"repo_user_id"`
	Owner       string `json:"owner"                    meddler:"repo_owner"`
	Name        string `json:"name"                     meddler:"repo_name"`
	FullName    string `json:"full_name"                meddler:"repo_full_name"`
	Avatar      string `json:"avatar_url,omitempty"     meddler:"repo_avatar"`
	Link        string `json:"link_url,omitempty"       meddler:"repo_link"`
	Kind        string `json:"scm,omitempty"            meddler:"repo_scm"`
	Clone       string `json:"clone_url,omitempty"      meddler:"repo_clone"`
	Branch      string `json:"default_branch,omitempty" meddler:"repo_branch"`
	Timeout     int64  `json:"timeout,omitempty"        meddler:"repo_timeout"`
	Visibility  string `json:"visibility"               meddler:"repo_visibility"`
	IsPrivate   bool   `json:"private"                  meddler:"repo_private"`
	IsTrusted   bool   `json:"trusted"                  meddler:"repo_trusted"`
	IsStarred   bool   `json:"starred,omitempty"        meddler:"-"`
	IsGated     bool   `json:"gated"                    meddler:"repo_gated"`
	IsActive    bool   `json:"active"                   meddler:"repo_active"`
	AllowPull   bool   `json:"allow_pr"                 meddler:"repo_allow_pr"`
	AllowPush   bool   `json:"allow_push"               meddler:"repo_allow_push"`
	AllowDeploy bool   `json:"allow_deploys"            meddler:"repo_allow_deploys"`
	AllowTag    bool   `json:"allow_tags"               meddler:"repo_allow_tags"`
	Counter     int    `json:"last_build"               meddler:"repo_counter"`
	Config      string `json:"config_file"              meddler:"repo_config_path"`
	Hash        string `json:"-"                        meddler:"repo_hash"`
	Perm        *Perm  `json:"-"                        meddler:"-"`
}

func (r *Repo) ResetVisibility() {
	r.Visibility = VisibilityPublic
	if r.IsPrivate {
		r.Visibility = VisibilityPrivate
	}
}

// ParseRepo parses the repository owner and name from a string.
func ParseRepo(str string) (user, repo string, err error) {
	var parts = strings.Split(str, "/")
	if len(parts) != 2 {
		err = fmt.Errorf("Error: Invalid or missing repository. eg octocat/hello-world.")
		return
	}
	user = parts[0]
	repo = parts[1]
	return
}

// Update updates the repository with values from the given Repo.
func (r *Repo) Update(from *Repo) {
	r.Avatar = from.Avatar
	r.Link = from.Link
	r.Kind = from.Kind
	r.Clone = from.Clone
	r.Branch = from.Branch
	if from.IsPrivate != r.IsPrivate {
		if from.IsPrivate {
			r.Visibility = VisibilityPrivate
		} else {
			r.Visibility = VisibilityPublic
		}
	}
	r.IsPrivate = from.IsPrivate
}

// RepoPatch represents a repository patch object.
type RepoPatch struct {
	Config       *string `json:"config_file,omitempty"`
	IsTrusted    *bool   `json:"trusted,omitempty"`
	IsGated      *bool   `json:"gated,omitempty"`
	Timeout      *int64  `json:"timeout,omitempty"`
	Visibility   *string `json:"visibility,omitempty"`
	AllowPull    *bool   `json:"allow_pr,omitempty"`
	AllowPush    *bool   `json:"allow_push,omitempty"`
	AllowDeploy  *bool   `json:"allow_deploy,omitempty"`
	AllowTag     *bool   `json:"allow_tag,omitempty"`
	BuildCounter *int    `json:"build_counter,omitempty"`
}
