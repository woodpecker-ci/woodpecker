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
	ID           int64    `json:"id"            meddler:"build_id,pk"     xorm:"pk autoincr 'build_id'"`
	RepoID       int64    `json:"-"             meddler:"build_repo_id"   xorm:"build_repo_id"`
	ConfigID     int64    `json:"-"             meddler:"build_config_id" xorm:"build_config_id"`
	Number       int      `json:"number"        meddler:"build_number"    xorm:"build_number"`
	Parent       int      `json:"parent"        meddler:"build_parent"    xorm:"build_parent"`
	Event        string   `json:"event"         meddler:"build_event"     xorm:"build_event"`
	Status       string   `json:"status"        meddler:"build_status"    xorm:"build_status"`
	Error        string   `json:"error"         meddler:"build_error"     xorm:"build_error"`
	Enqueued     int64    `json:"enqueued_at"   meddler:"build_enqueued"  xorm:"build_enqueued"`
	Created      int64    `json:"created_at"    meddler:"build_created"   xorm:"build_created"`
	Started      int64    `json:"started_at"    meddler:"build_started"   xorm:"build_started"`
	Finished     int64    `json:"finished_at"   meddler:"build_finished"  xorm:"build_finished"`
	Deploy       string   `json:"deploy_to"     meddler:"build_deploy"    xorm:"build_deploy"`
	Commit       string   `json:"commit"        meddler:"build_commit"    xorm:"build_commit"`
	Branch       string   `json:"branch"        meddler:"build_branch"    xorm:"build_branch"`
	Ref          string   `json:"ref"           meddler:"build_ref"       xorm:"build_ref"`
	Refspec      string   `json:"refspec"       meddler:"build_refspec"   xorm:"build_refspec"`
	Remote       string   `json:"remote"        meddler:"build_remote"    xorm:"build_remote"`
	Title        string   `json:"title"         meddler:"build_title"     xorm:"build_title"`
	Message      string   `json:"message"       meddler:"build_message"   xorm:"build_message"`
	Timestamp    int64    `json:"timestamp"     meddler:"build_timestamp" xorm:"build_timestamp"`
	Sender       string   `json:"sender"        meddler:"build_sender"    xorm:"build_sender"`
	Author       string   `json:"author"        meddler:"build_author"    xorm:"build_author"`
	Avatar       string   `json:"author_avatar" meddler:"build_avatar"    xorm:"build_avatar"`
	Email        string   `json:"author_email"  meddler:"build_email"     xorm:"build_email"`
	Link         string   `json:"link_url"      meddler:"build_link"      xorm:"build_link"`
	Signed       bool     `json:"signed"        meddler:"build_signed"    xrom:"build_signed"`   // deprecate
	Verified     bool     `json:"verified"      meddler:"build_verified"  xorm:"build_verified"` // deprecate
	Reviewer     string   `json:"reviewed_by"   meddler:"build_reviewer"  xorm:"build_reviewer"`
	Reviewed     int64    `json:"reviewed_at"   meddler:"build_reviewed"  xorm:"build_reviewed"`
	Procs        []*Proc  `json:"procs,omitempty" meddler:"-"             xorm:"-"`
	Files        []*File  `json:"files,omitempty" meddler:"-"             xorm:"-"`
	ChangedFiles []string `json:"changed_files,omitempty" meddler:"changed_files,json" xorm:"-"` // TODO: Xorm and json
}

// TableName return database table name for xorm
func (Build) TableName() string {
	return "builds"
}
