// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// Registry represents a docker registry with credentials.
type RegistryV032 struct {
	ID       int64  `json:"id"       xorm:"pk autoincr 'id'"`
	OrgID    int64  `json:"org_id"   xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'org_id'"`
	RepoID   int64  `json:"repo_id"  xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'repo_id'"`
	Address  string `json:"address"  xorm:"NOT NULL UNIQUE(s) INDEX 'address'"`
	Username string `json:"username" xorm:"varchar(2000) 'username'"`
	Password string `json:"password" xorm:"TEXT 'password'"`
} //	@name Registry

func (r RegistryV032) TableName() string {
	return "registries"
}

var alterTableRegistriesAddOrgIDCol = xormigrate.Migration{
	ID: "alter-table-add-registries-org-id",
	MigrateSession: func(sess *xorm.Session) error {
		if err := sess.Sync2(new(RegistryV032)); err != nil {
			return err
		}
		if err := alterColumnDefault(sess, "registries", "repo_id", "0"); err != nil {
			return err
		}
		if err := alterColumnNull(sess, "registries", "repo_id", false); err != nil {
			return err
		}
		return alterColumnNull(sess, "registries", "address", false)
	},
}
