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
	"strings"

	"xorm.io/xorm"
)

var renameRemoteToForge = task{
	name:     "rename-remote-to-forge",
	required: true,
	fn: func(sess *xorm.Session) error {
		var cloneURLColumns []*oldTable

		cloneURLColumns = append(cloneURLColumns, &oldTable{
			table: "pipelines",
			columns: []string{
				"pipeline_remote",
			},
		},
		)

		for _, table := range cloneURLColumns {
			for _, column := range table.columns {
				err := renameColumn(sess, table.table, column, strings.Replace(column, "remote", "clone_url", 1))
				if err != nil {
					return err
				}
			}
		}

		var forgeColumns []*oldTable

		forgeColumns = append(forgeColumns, &oldTable{
			table: "repos",
			columns: []string{
				"remote_id",
			},
		},
		)

		for _, table := range forgeColumns {
			for _, column := range table.columns {
				err := renameColumn(sess, table.table, column, strings.Replace(column, "remote", "forge", 1))
				if err != nil {
					return err
				}
			}
		}

		return nil
	},
}
