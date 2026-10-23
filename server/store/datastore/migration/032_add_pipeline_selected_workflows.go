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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// addPipelineSelectedWorkflows records which workflows a trigger selected, so a
// restart can re-apply the selection instead of trusting whatever the config
// service hands back. Existing pipelines get an empty selection, which means
// every workflow, matching what they ran.
var addPipelineSelectedWorkflows = xormigrate.Migration{
	ID: "add-pipeline-selected-workflows",
	MigrateSession: func(sess *xorm.Session) error {
		type pipelines struct {
			ID int64 `xorm:"pk autoincr 'id'"`

			// new selected_workflows field
			SelectedWorkflows []string `xorm:"json 'selected_workflows'"`
		}

		return sess.Sync(new(pipelines))
	},
}
