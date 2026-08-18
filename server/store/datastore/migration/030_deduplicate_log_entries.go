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

// Clears the way for the UNIQUE(step_id, line) index on log_entries by keeping
// the lowest id of every line number and dropping the rest.
//
// The inner query is restricted to line numbers that actually occur more than
// once, so the work stays proportional to the duplicates rather than to the
// size of the table. The extra nesting around it lets MySQL read from the
// table it deletes from.
//
// This migration must not be marked Long: the UNIQUE index is created from the
// model on every start, and skipping the cleanup would leave that failing.
var deduplicateLogEntries = xormigrate.Migration{
	ID: "deduplicate-log-entries",
	MigrateSession: func(sess *xorm.Session) error {
		_, err := sess.Exec(`DELETE FROM log_entries WHERE id IN (
			SELECT id FROM (
				SELECT entries.id AS id
				FROM log_entries entries
				JOIN (
					SELECT step_id, line, MIN(id) AS keep_id
					FROM log_entries
					GROUP BY step_id, line
					HAVING COUNT(*) > 1
				) duplicates
				ON duplicates.step_id = entries.step_id AND duplicates.line = entries.line
				WHERE entries.id > duplicates.keep_id
			) removable
		);`)

		return err
	},
}
