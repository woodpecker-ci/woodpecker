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
type Feed struct {
	RepoID   int64  `json:"repo_id"                 xorm:"repo_id"`
	ID       int64  `json:"id,omitempty"            xorm:"pipeline_id"`
	Number   int64  `json:"number,omitempty"        xorm:"pipeline_number"`
	Event    string `json:"event,omitempty"         xorm:"pipeline_event"`
	Status   string `json:"status,omitempty"        xorm:"pipeline_status"`
	Created  int64  `json:"created,omitempty"       xorm:"pipeline_created"`
	Started  int64  `json:"started,omitempty"       xorm:"pipeline_started"`
	Finished int64  `json:"finished,omitempty"      xorm:"pipeline_finished"`
	Commit   string `json:"commit,omitempty"        xorm:"pipeline_commit"`
	Branch   string `json:"branch,omitempty"        xorm:"pipeline_branch"`
	Ref      string `json:"ref,omitempty"           xorm:"pipeline_ref"`
	Refspec  string `json:"refspec,omitempty"       xorm:"pipeline_refspec"`
	Title    string `json:"title,omitempty"         xorm:"pipeline_title"`
	Message  string `json:"message,omitempty"       xorm:"pipeline_message"`
	Author   string `json:"author,omitempty"        xorm:"pipeline_author"`
	Avatar   string `json:"author_avatar,omitempty" xorm:"pipeline_avatar"`
	Email    string `json:"author_email,omitempty"  xorm:"pipeline_email"`
} //	@name Feed
