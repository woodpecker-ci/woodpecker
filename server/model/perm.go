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

// Perm defines a repository permission for an individual user.
type Perm struct {
	UserID  int64 `json:"-"       xorm:"UNIQUE(s) INDEX NOT NULL 'user_id'"`
	RepoID  int64 `json:"-"       xorm:"UNIQUE(s) INDEX NOT NULL 'repo_id'"`
	Repo    *Repo `json:"-"       xorm:"-"`
	Pull    bool  `json:"pull"    xorm:"pull"`
	Push    bool  `json:"push"    xorm:"push"`
	Admin   bool  `json:"admin"   xorm:"admin"`
	Synced  int64 `json:"synced"  xorm:"synced"`
	Created int64 `json:"created" xorm:"created"`
	Updated int64 `json:"updated" xorm:"updated"`
} //	@name Perm

// TableName return database table name for xorm.
func (Perm) TableName() string {
	return "perms"
}

// OrgPerm defines an organization permission for an individual user.
type OrgPerm struct {
	Member bool `json:"member"`
	Admin  bool `json:"admin"`
} //	@name OrgPerm
