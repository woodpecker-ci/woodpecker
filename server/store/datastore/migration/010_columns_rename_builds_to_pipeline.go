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

type oldBuildColumn struct {
	table   string
	columns []string
}

var renameColumnsBuildsToPipeline = task{
	name:     "rename-columns-builds-to-pipeline",
	required: true,
	fn: func(sess *xorm.Session) error {
		var oldColumns []*oldBuildColumn

		oldColumns = append(oldColumns, &oldBuildColumn{
			table: "pipelines",
			columns: []string{
				"build_id",
				"build_repo_id",
				"build_number",
				"build_author",
				"build_config_id",
				"build_parent",
				"build_event",
				"build_status",
				"build_error",
				"build_enqueued",
				"build_created",
				"build_started",
				"build_finished",
				"build_deploy",
				"build_commit",
				"build_branch",
				"build_ref",
				"build_refspec",
				"build_remote",
				"build_title",
				"build_message",
				"build_timestamp",
				"build_sender",
				"build_avatar",
				"build_email",
				"build_link",
				"build_signed",
				"build_verified",
				"build_reviewer",
				"build_reviewed",
			},
		},
		)

		oldColumns = append(oldColumns, &oldBuildColumn{
			table:   "pipeline_config",
			columns: []string{"build_id"},
		})

		oldColumns = append(oldColumns, &oldBuildColumn{
			table:   "files",
			columns: []string{"file_build_id"},
		})

		oldColumns = append(oldColumns, &oldBuildColumn{
			table:   "steps",
			columns: []string{"step_build_id"},
		})

		for _, table := range oldColumns {
			for _, column := range table.columns {
				err := renameColumn(sess, table.table, column, strings.Replace(column, "build_", "pipeline_", 1))
				if err != nil {
					return err
				}
			}
		}

		return nil
	},
}
