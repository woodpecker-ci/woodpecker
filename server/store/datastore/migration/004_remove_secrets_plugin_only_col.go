// Copyright 2023 Woodpecker Authors
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

type oldSecret004 struct {
	ID          int64    `json:"id"              xorm:"pk autoincr 'secret_id'"`
	PluginsOnly bool     `json:"plugins_only"    xorm:"secret_plugins_only"`
	SkipVerify  bool     `json:"-"               xorm:"secret_skip_verify"`
	Conceal     bool     `json:"-"               xorm:"secret_conceal"`
	Images      []string `json:"images"          xorm:"json 'secret_images'"`
}

func (oldSecret004) TableName() string {
	return "secrets"
}

var removePluginOnlyOptionFromSecretsTable = xormigrate.Migration{
	ID: "remove-plugin-only-option-from-secrets-table",
	MigrateSession: func(sess *xorm.Session) (err error) {
		// make sure plugin_only column exists
		if err := sess.Sync(new(oldSecret004)); err != nil {
			return err
		}

		return dropTableColumns(sess, "secrets", "secret_plugins_only", "secret_skip_verify", "secret_conceal")
	},
}
