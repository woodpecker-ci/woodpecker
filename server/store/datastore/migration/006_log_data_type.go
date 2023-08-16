// Copyright 2023 Woodpecker Authors
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
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

var alterTableLogUpdateColumnLogDataType = task{
	name: "alter-table-logs-update-type-of-data",
	fn: func(sess *xorm.Session) (err error) {
		dialect := sess.Engine().Dialect().URI().DBType

		switch dialect {
		case schemas.POSTGRES:
			_, err = sess.Exec("ALTER TABLE logs ALTER COLUMN log_data TYPE BYTEA")
		case schemas.MYSQL:
			_, err = sess.Exec("ALTER TABLE logs MODIFY COLUMN log_data LONGBLOB")
		default:
			// sqlite does only know BLOB in all cases
			return nil
		}

		return err
	},
}
