// Copyright 2025 Woodpecker Authors
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

var addTaskPipelineRepo = xormigrate.Migration{
	ID: "add-task-pipeline-repo",
	MigrateSession: func(sess *xorm.Session) (err error) {
		// Add the pipeline_id and repo_id columns to the tasks table
		_, err = sess.Exec("ALTER TABLE tasks ADD COLUMN pipeline_id BIGINT NOT NULL DEFAULT 0;")
		if err != nil {
			return err
		}
		_, err = sess.Exec("ALTER TABLE tasks ADD COLUMN repo_id BIGINT NOT NULL DEFAULT 0;")
		return err
	},
}
