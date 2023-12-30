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

package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type SecretV008 struct {
	Owner  string `json:"-"    xorm:"NOT NULL DEFAULT '' UNIQUE(s) INDEX 'secret_owner'"`
	RepoID int64  `json:"-"    xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'secret_repo_id'"`
	Name   string `json:"name" xorm:"NOT NULL UNIQUE(s) INDEX 'secret_name'"`
}

// TableName return database table name for xorm
func (SecretV008) TableName() string {
	return "secrets"
}

var alterTableSecretsAddUserCol = xormigrate.Migration{
	ID: "alter-table-add-secrets-user-id",
	MigrateSession: func(sess *xorm.Session) error {
		if err := sess.Sync(new(SecretV008)); err != nil {
			return err
		}
		if _, err := sess.SQL(`UPDATE secrets SET secret_repo_id=0 WHERE secret_repo_id=NULL;`).Exec(); err != nil {
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
