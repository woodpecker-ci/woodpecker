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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

var alterTableConfigUpdateColumnConfigDataType = xormigrate.Migration{
	ID: "alter-table-config-update-type-of-config-data",
	MigrateSession: func(sess *xorm.Session) (err error) {
		dialect := sess.Engine().Dialect().URI().DBType

		switch dialect {
		case schemas.MYSQL:
			_, err = sess.Exec("ALTER TABLE config MODIFY COLUMN config_data LONGBLOB")
		default:
			// xorm uses the same type for all blob sizes in sqlite and postgres
			return nil
		}

		return err
	},
}
