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

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

var renameTableProcsToSteps = xormigrate.Migration{
	ID: "rename-procs-to-steps",
	MigrateSession: func(sess *xorm.Session) error {
		err := renameTable(sess, "procs", "steps")
		if err != nil {
			return err
		}

		oldProcColumns := []*oldTable{
			{
				table: "steps",
				columns: []string{
					"proc_id",
					"proc_pipeline_id",
					"proc_pid",
					"proc_ppid",
					"proc_pgid",
					"proc_name",
					"proc_state",
					"proc_error",
					"proc_exit_code",
					"proc_started",
					"proc_stopped",
					"proc_machine",
					"proc_platform",
					"proc_environ",
				},
			},
			{
				table:   "files",
				columns: []string{"file_proc_id"},
			},
		}

		for _, table := range oldProcColumns {
			for _, column := range table.columns {
				err := renameColumn(sess, table.table, column, strings.Replace(column, "proc_", "step_", 1))
				if err != nil {
					return err
				}
			}
		}

		oldJobColumns := []*oldTable{
			{
				table: "logs",
				columns: []string{
					"log_job_id",
				},
			},
		}

		for _, table := range oldJobColumns {
			for _, column := range table.columns {
				err := renameColumn(sess, table.table, column, strings.Replace(column, "job_", "step_", 1))
				if err != nil {
					return err
				}
			}
		}

		return nil
	},
}
