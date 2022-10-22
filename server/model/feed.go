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
	Owner    string `json:"owner"                   xorm:"repo_owner"`
	Name     string `json:"name"                    xorm:"repo_name"`
	FullName string `json:"full_name"               xorm:"repo_full_name"`

	Number   int64  `json:"number,omitempty"        xorm:"pipeline_number"`
	Event    string `json:"event,omitempty"         xorm:"pipeline_event"`
	Status   string `json:"status,omitempty"        xorm:"pipeline_status"`
	Created  int64  `json:"created_at,omitempty"    xorm:"pipeline_created"`
	Started  int64  `json:"started_at,omitempty"    xorm:"pipeline_started"`
	Finished int64  `json:"finished_at,omitempty"   xorm:"pipeline_finished"`
	Commit   string `json:"commit,omitempty"        xorm:"pipeline_commit"`
	Branch   string `json:"branch,omitempty"        xorm:"pipeline_branch"`
	Ref      string `json:"ref,omitempty"           xorm:"pipeline_ref"`
	Refspec  string `json:"refspec,omitempty"       xorm:"pipeline_refspec"`
	Remote   string `json:"remote,omitempty"        xorm:"pipeline_remote"`
	Title    string `json:"title,omitempty"         xorm:"pipeline_title"`
	Message  string `json:"message,omitempty"       xorm:"pipeline_message"`
	Author   string `json:"author,omitempty"        xorm:"pipeline_author"`
	Avatar   string `json:"author_avatar,omitempty" xorm:"pipeline_avatar"`
	Email    string `json:"author_email,omitempty"  xorm:"pipeline_email"`
}
