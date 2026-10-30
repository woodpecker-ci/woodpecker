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
	"xorm.io/xorm/schemas"
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
	MigrateSession: func(sess *xorm.Session) (err error) {
		// To speedup delete query we first create an index.
		// The error can be ignored as it does not matter if creation failed.
		_, _ = sess.Exec(`CREATE INDEX idx_log_entries_tmp ON log_entries (step_id, line, id);`)

		// Find duplicat log entries and delete second ones.
		dialect := sess.Engine().Dialect().URI().DBType
		switch dialect {
		case schemas.MYSQL:
			_, err = sess.Exec(`
DELETE a FROM log_entries a
JOIN log_entries b
ON a.step_id = b.step_id
AND a.line = b.line
AND a.id > b.id;`)

			_, _ = sess.Exec(`DROP INDEX IF EXISTS idx_log_entries_tmp ON log_entries;`)
		case schemas.POSTGRES:
			_, err = sess.Exec(`
DELETE FROM log_entries a
USING log_entries b
WHERE a.step_id = b.step_id
AND a.line = b.line
AND a.id > b.id;`)

			_, _ = sess.Exec(`DROP INDEX IF EXISTS idx_log_entries_tmp;`)
		case schemas.SQLITE:
			_, err = sess.Exec(`
DELETE FROM log_entries AS a
WHERE EXISTS (
SELECT 1 FROM log_entries b
WHERE b.step_id = a.step_id AND b.line = a.line AND b.id < a.id
);`)

			_, _ = sess.Exec(`DROP INDEX IF EXISTS idx_log_entries_tmp;`)
		default:
			err = fmt.Errorf("dialect '%s' not supported", dialect)
		}

		return err
	},
}
