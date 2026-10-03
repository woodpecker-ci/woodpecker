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

// addCronWorkflows adds the workflow selection to crons. Existing crons get an
// empty selection, which keeps their current behavior of running every
// workflow config found in the repo.
var addCronWorkflows = xormigrate.Migration{
	ID: "add-cron-workflows",
	MigrateSession: func(sess *xorm.Session) error {
		type crons struct {
			ID int64 `xorm:"pk autoincr 'id'"`

			// new workflows field
			Workflows []string `xorm:"json 'workflows'"`
		}

		return sess.Sync(new(crons))
	},
}
