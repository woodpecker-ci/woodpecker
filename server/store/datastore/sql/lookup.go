// Copyright 2018 Drone.IO Inc.
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

package sql

import (
	"github.com/woodpecker-ci/woodpecker/server/store/datastore/sql/mysql"
	"github.com/woodpecker-ci/woodpecker/server/store/datastore/sql/postgres"
	"github.com/woodpecker-ci/woodpecker/server/store/datastore/sql/sqlite"
)

// Supported database drivers
const (
	DriverSqlite   = "sqlite3"
	DriverMysql    = "mysql"
	DriverPostgres = "postgres"
)

// Lookup returns the named sql statement compatible with
// the specified database driver.
func Lookup(driver string, name string) string {
	switch driver {
	case DriverPostgres:
		return postgres.Lookup(name)
	case DriverMysql:
		return mysql.Lookup(name)
	default:
		return sqlite.Lookup(name)
	}
}
