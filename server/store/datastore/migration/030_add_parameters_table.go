// Copyright 2026 Woodpecker Authors
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

var addParametersTable = xormigrate.Migration{
	ID: "add-parameters-table",
	MigrateSession: func(sess *xorm.Session) error {
		type parameters struct {
			ID          int64    `xorm:"pk autoincr 'id'"`
			RepoID      int64    `xorm:"UNIQUE(s) INDEX 'repo_id'"`
			Name        string   `xorm:"UNIQUE(s) INDEX 'name'"`
			Type        string   `xorm:"'param_type'"`
			Description string   `xorm:"TEXT 'description'"`
			Default     string   `xorm:"TEXT 'default_value'"`
			Options     []string `xorm:"json 'options'"`
			Required    bool     `xorm:"required"`
			Order       int      `xorm:"display_order"`
			Source      string   `xorm:"source"`
		}

		if err := sess.Sync(new(parameters)); err != nil {
			return fmt.Errorf("sync models failed: %w", err)
		}
		return nil
	},
}
