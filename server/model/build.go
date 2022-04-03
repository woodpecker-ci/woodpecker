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

// swagger:model build
type Build struct {
	ID           int64        `json:"id"                      xorm:"pk autoincr 'build_id'"`
	RepoID       int64        `json:"-"                       xorm:"UNIQUE(s) INDEX 'build_repo_id'"`
	Number       int64        `json:"number"                  xorm:"UNIQUE(s) 'build_number'"`
	ConfigID     int64        `json:"-"                       xorm:"build_config_id"`
	Parent       int64        `json:"parent"                  xorm:"build_parent"` // used if a pipeline is restarted
	Author       string       `json:"author"                  xorm:"INDEX 'build_author'"`
	Avatar       string       `json:"author_avatar"           xorm:"build_avatar"`
	Event        WebhookEvent `json:"event"                   xorm:"build_event"`
	Email        string       `json:"author_email"            xorm:"build_email"`
	Commit       string       `json:"commit"                  xorm:"build_commit"`
	Branch       string       `json:"branch"                  xorm:"build_branch"`
	Ref          string       `json:"ref"                     xorm:"build_ref"`
	Refspec      string       `json:"refspec"                 xorm:"build_refspec"`
	Remote       string       `json:"remote"                  xorm:"build_remote"`
	Message      string       `json:"message"                 xorm:"build_message"` // (last) commit message
	Sender       string       `json:"sender"                  xorm:"build_sender"`
	Link         string       `json:"link_url"                xorm:"build_link"`      // link to forge (exp. to PR)
	Deploy       string       `json:"deploy_to"               xorm:"build_deploy"`    // deploy destination
	Timestamp    int64        `json:"timestamp"               xorm:"build_timestamp"` // timestamp of initial event like git commit
	ChangedFiles []string     `json:"changed_files,omitempty" xorm:"json 'changed_files'"`

	// execution data
	Status   StatusValue `json:"status"                  xorm:"INDEX 'build_status'"`
	Error    string      `json:"error"                   xorm:"build_error"`
	Enqueued int64       `json:"enqueued_at"             xorm:"build_enqueued"`
	Created  int64       `json:"created_at"              xorm:"build_created"`
	Updated  int64       `json:"updated_at"              xorm:"updated NOT NULL DEFAULT 0 'updated'"`
	Started  int64       `json:"started_at"              xorm:"build_started"`
	Finished int64       `json:"finished_at"             xorm:"build_finished"`
	Reviewer string      `json:"reviewed_by"             xorm:"build_reviewer"` // user who approved a pipeline
	Reviewed int64       `json:"reviewed_at"             xorm:"build_reviewed"` // timestamp the pipeline got approved at

	// TODO: deprecate / remove properties
	// Signed   bool   `json:"signed"                  xorm:"build_signed"`
	// Verified bool   `json:"verified"                xorm:"build_verified"`
	Title string  `json:"title"                   xorm:"build_title"` // title of pull-request
	Procs []*Proc `json:"procs,omitempty"         xorm:"-"`
}

// TableName return database table name for xorm
func (Build) TableName() string {
	return "builds"
}
