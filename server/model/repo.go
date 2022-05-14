// Copyright 2021 Woodpecker Authors
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

// Repo represents a repository.
//
// swagger:model repo
type Repo struct {
	ID                           int64          `json:"id,omitempty"                    xorm:"pk autoincr 'repo_id'"`
	UserID                       int64          `json:"-"                               xorm:"repo_user_id"`
	Owner                        string         `json:"owner"                           xorm:"UNIQUE(name) 'repo_owner'"`
	Name                         string         `json:"name"                            xorm:"UNIQUE(name) 'repo_name'"`
	FullName                     string         `json:"full_name"                       xorm:"UNIQUE 'repo_full_name'"`
	Avatar                       string         `json:"avatar_url,omitempty"            xorm:"varchar(500) 'repo_avatar'"`
	Link                         string         `json:"link_url,omitempty"              xorm:"varchar(1000) 'repo_link'"`
	Clone                        string         `json:"clone_url,omitempty"             xorm:"varchar(1000) 'repo_clone'"`
	Branch                       string         `json:"default_branch,omitempty"        xorm:"varchar(500) 'repo_branch'"`
	SCMKind                      SCMKind        `json:"scm,omitempty"                   xorm:"varchar(50) 'repo_scm'"`
	Timeout                      int64          `json:"timeout,omitempty"               xorm:"repo_timeout"`
	Visibility                   RepoVisibly    `json:"visibility"                      xorm:"varchar(10) 'repo_visibility'"`
	IsSCMPrivate                 bool           `json:"private"                         xorm:"repo_private"`
	IsTrusted                    bool           `json:"trusted"                         xorm:"repo_trusted"`
	IsStarred                    bool           `json:"starred,omitempty"               xorm:"-"`
	IsGated                      bool           `json:"gated"                           xorm:"repo_gated"`
	IsActive                     bool           `json:"active"                          xorm:"repo_active"`
	AllowPull                    bool           `json:"allow_pr"                        xorm:"repo_allow_pr"`
	Config                       string         `json:"config_file"                     xorm:"varchar(500) 'repo_config_path'"`
	Hash                         string         `json:"-"                               xorm:"varchar(500) 'repo_hash'"`
	Perm                         *Perm          `json:"-"                               xorm:"-"`
	CancelPreviousPipelineEvents []WebhookEvent `json:"cancel_previous_pipeline_events" xorm:"json 'cancel_previous_pipeline_events'"`
	SecretEndpoint               string         `json:"secrect_endpoint"                xorm:"varchar(500) 'secrect_endpoint'"`
	RegistryEndpoint             string         `json:"registry_endpoint"               xorm:"varchar(500) 'registry_endpoint'"`
	ConfigEndpoint               string         `json:"config_endpoint"                 xorm:"varchar(500) 'config_endpoint'"`
}

// TableName return database table name for xorm
func (Repo) TableName() string {
	return "repos"
}

func (r *Repo) ResetVisibility() {
	r.Visibility = VisibilityPublic
	if r.IsSCMPrivate {
		r.Visibility = VisibilityPrivate
	}
}

// ParseRepo parses the repository owner and name from a string.
func ParseRepo(str string) (user, repo string, err error) {
	parts := strings.Split(str, "/")
	if len(parts) != 2 {
		err = fmt.Errorf("Error: Invalid or missing repository. eg octocat/hello-world")
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
	r.SCMKind = from.SCMKind
	r.Clone = from.Clone
	r.Branch = from.Branch
	if from.IsSCMPrivate != r.IsSCMPrivate {
		if from.IsSCMPrivate {
			r.Visibility = VisibilityPrivate
		} else {
			r.Visibility = VisibilityPublic
		}
	}
	r.IsSCMPrivate = from.IsSCMPrivate
}

// RepoPatch represents a repository patch object.
type RepoPatch struct {
	Config                       *string         `json:"config_file,omitempty"`
	IsTrusted                    *bool           `json:"trusted,omitempty"`
	IsGated                      *bool           `json:"gated,omitempty"`
	Timeout                      *int64          `json:"timeout,omitempty"`
	Visibility                   *string         `json:"visibility,omitempty"`
	AllowPull                    *bool           `json:"allow_pr,omitempty"`
	CancelPreviousPipelineEvents *[]WebhookEvent `json:"cancel_previous_pipeline_events"`
}
