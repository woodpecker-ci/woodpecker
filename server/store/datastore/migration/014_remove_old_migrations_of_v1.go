// Copyright 2024 Woodpecker Authors
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

var removeOldMigrationsOfV1 = xormigrate.Migration{
	ID: "remove-old-migrations-of-v1",
	MigrateSession: func(sess *xorm.Session) (err error) {
		_, err = sess.Table(&xormigrate.Migration{}).In("id", []string{
			"xorm",
			"alter-table-drop-repo-fallback",
			"drop-allow-push-tags-deploys-columns",
			"fix-pr-secret-event-name",
			"alter-table-drop-counter",
			"drop-senders",
			"alter-table-logs-update-type-of-data",
			"alter-table-add-secrets-user-id",
			"lowercase-secret-names",
			"recreate-agents-table",
			"rename-builds-to-pipeline",
			"rename-columns-builds-to-pipeline",
			"rename-procs-to-steps",
			"rename-remote-to-forge",
			"rename-forge-id-to-forge-remote-id",
			"remove-active-from-users",
			"remove-inactive-repos",
			"drop-files",
			"remove-machine-col",
			"drop-old-col",
			"init-log_entries",
			"migrate-logs-to-log_entries",
			"parent-steps-to-workflows",
			"add-orgs",
		}).Delete()

		return err
	},
}
