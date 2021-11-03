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

// Feed represents an item in the user's feed or timeline.
//
// swagger:model feed
type Feed struct {
	Owner    string `json:"owner"                   meddler:"repo_owner"                 xorm:"repo_owner"`
	Name     string `json:"name"                    meddler:"repo_name"                  xorm:"repo_name"`
	FullName string `json:"full_name"               meddler:"repo_full_name"             xorm:"repo_full_name"`

	Number   int    `json:"number,omitempty"        meddler:"build_number,zeroisnull"    xorm:"build_number"`
	Event    string `json:"event,omitempty"         meddler:"build_event,zeroisnull"     xorm:"build_event"`
	Status   string `json:"status,omitempty"        meddler:"build_status,zeroisnull"    xorm:"build_status"`
	Created  int64  `json:"created_at,omitempty"    meddler:"build_created,zeroisnull"   xorm:"build_created"`
	Started  int64  `json:"started_at,omitempty"    meddler:"build_started,zeroisnull"   xorm:"build_started"`
	Finished int64  `json:"finished_at,omitempty"   meddler:"build_finished,zeroisnull"  xorm:"build_finished"`
	Commit   string `json:"commit,omitempty"        meddler:"build_commit,zeroisnull"    xorm:"build_commit"`
	Branch   string `json:"branch,omitempty"        meddler:"build_branch,zeroisnull"    xorm:"build_branch"`
	Ref      string `json:"ref,omitempty"           meddler:"build_ref,zeroisnull"       xorm:"build_ref"`
	Refspec  string `json:"refspec,omitempty"       meddler:"build_refspec,zeroisnull"   xorm:"build_refspec"`
	Remote   string `json:"remote,omitempty"        meddler:"build_remote,zeroisnull"    xorm:"build_remote"`
	Title    string `json:"title,omitempty"         meddler:"build_title,zeroisnull"     xorm:"build_title"`
	Message  string `json:"message,omitempty"       meddler:"build_message,zeroisnull"   xorm:"build_message"`
	Author   string `json:"author,omitempty"        meddler:"build_author,zeroisnull"    xorm:"build_author"`
	Avatar   string `json:"author_avatar,omitempty" meddler:"build_avatar,zeroisnull"    xorm:"build_avatar"`
	Email    string `json:"author_email,omitempty"  meddler:"build_email,zeroisnull"     xorm:"build_email"`
}
