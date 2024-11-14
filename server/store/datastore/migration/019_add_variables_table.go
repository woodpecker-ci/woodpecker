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
	"fmt"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type variable035 struct {
	ID     int64  `json:"id"              xorm:"pk autoincr 'id'"`
	OrgID  int64  `json:"org_id"          xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'org_id'"`
	RepoID int64  `json:"repo_id"         xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'repo_id'"`
	Name   string `json:"name"            xorm:"NOT NULL UNIQUE(s) INDEX 'name'"`
	Value  string `json:"value,omitempty" xorm:"TEXT 'value'"`
} //	@name Variable

// TableName return database table name for xorm.
func (variable035) TableName() string {
	return "variables"
}

var addVariablesTable = xormigrate.Migration{
	ID: "add-variables-table",
	MigrateSession: func(sess *xorm.Session) error {
		if err := sess.Sync(new(variable035)); err != nil {
			return fmt.Errorf("sync models failed: %w", err)
		}

		return nil
	},
}
