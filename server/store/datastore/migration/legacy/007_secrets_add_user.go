// Copyright 2022 Woodpecker Authors
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

package legacy

import (
	"xorm.io/xorm"
)

type SecretV007 struct {
	Owner  string `json:"-"    xorm:"NOT NULL DEFAULT '' UNIQUE(s) INDEX 'secret_owner'"`
	RepoID int64  `json:"-"    xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'secret_repo_id'"`
	Name   string `json:"name" xorm:"NOT NULL UNIQUE(s) INDEX 'secret_name'"`
}

// TableName return database table name for xorm
func (SecretV007) TableName() string {
	return "secrets"
}

var alterTableSecretsAddUserCol = task{
	name: "alter-table-add-secrets-user-id",
	fn: func(sess *xorm.Session) error {
		if err := sess.Sync(new(SecretV007)); err != nil {
			return err
		}
		if err := alterColumnDefault(sess, "secrets", "secret_repo_id", "0"); err != nil {
			return err
		}
		if err := alterColumnNull(sess, "secrets", "secret_repo_id", false); err != nil {
			return err
		}
		return alterColumnNull(sess, "secrets", "secret_name", false)
	},
}
