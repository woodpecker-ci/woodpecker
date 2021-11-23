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
	ID           int64       `json:"id"                      xorm:"pk autoincr 'build_id'"`
	RepoID       int64       `json:"-"                       xorm:"UNIQUE(s) INDEX 'build_repo_id'"`
	Number       int64       `json:"number"                  xorm:"UNIQUE(s) 'build_number'"`
	Author       string      `json:"author"                  xorm:"INDEX 'build_author'"`
	ConfigID     int64       `json:"-"                       xorm:"build_config_id"`
	Parent       int64       `json:"parent"                  xorm:"build_parent"`
	Event        string      `json:"event"                   xorm:"build_event"`
	Status       StatusValue `json:"status"                  xorm:"INDEX 'build_status'"`
	Error        string      `json:"error"                   xorm:"build_error"`
	Enqueued     int64       `json:"enqueued_at"             xorm:"build_enqueued"`
	Created      int64       `json:"created_at"              xorm:"build_created"`
	Started      int64       `json:"started_at"              xorm:"build_started"`
	Finished     int64       `json:"finished_at"             xorm:"build_finished"`
	Deploy       string      `json:"deploy_to"               xorm:"build_deploy"`
	Commit       string      `json:"commit"                  xorm:"build_commit"`
	Branch       string      `json:"branch"                  xorm:"build_branch"`
	Ref          string      `json:"ref"                     xorm:"build_ref"`
	Refspec      string      `json:"refspec"                 xorm:"build_refspec"`
	Remote       string      `json:"remote"                  xorm:"build_remote"`
	Title        string      `json:"title"                   xorm:"build_title"`
	Message      string      `json:"message"                 xorm:"build_message"`
	Timestamp    int64       `json:"timestamp"               xorm:"build_timestamp"`
	Sender       string      `json:"sender"                  xorm:"build_sender"`
	Avatar       string      `json:"author_avatar"           xorm:"build_avatar"`
	Email        string      `json:"author_email"            xorm:"build_email"`
	Link         string      `json:"link_url"                xorm:"build_link"`
	Signed       bool        `json:"signed"                  xorm:"build_signed"`   // deprecate
	Verified     bool        `json:"verified"                xorm:"build_verified"` // deprecate
	Reviewer     string      `json:"reviewed_by"             xorm:"build_reviewer"`
	Reviewed     int64       `json:"reviewed_at"             xorm:"build_reviewed"`
	Procs        []*Proc     `json:"procs,omitempty"         xorm:"-"`
	Files        []*File     `json:"files,omitempty"         xorm:"-"`
	ChangedFiles []string    `json:"changed_files,omitempty" xorm:"json 'changed_files'"`
}

// TableName return database table name for xorm
func (Build) TableName() string {
	return "builds"
}
