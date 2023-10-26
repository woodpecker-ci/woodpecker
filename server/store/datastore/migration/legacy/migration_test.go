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

package legacy

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"

	// blank imports to register the sql drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const (
	sqliteDB = "./testfiles/sqlite.db"
)

func testDriver() string {
	driver := os.Getenv("WOODPECKER_DATABASE_DRIVER")
	if len(driver) == 0 {
		return "sqlite3"
	}
	return driver
}

func createSQLiteDB(t *testing.T) string {
	tmpF, err := os.CreateTemp("./testfiles", "tmp_")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	dbF, err := os.ReadFile(sqliteDB)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.NoError(t, os.WriteFile(tmpF.Name(), dbF, 0o644)) {
		t.FailNow()
	}
	return tmpF.Name()
}

func testDB(t *testing.T, new bool) (engine *xorm.Engine, closeDB func()) {
	driver := testDriver()
	var err error
	closeDB = func() {}
	switch driver {
	case "sqlite3":
		config := ":memory:"
		if !new {
			config = createSQLiteDB(t)
			closeDB = func() {
				_ = os.Remove(config)
			}
		}
		engine, err = xorm.NewEngine(driver, config)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		return
	case "mysql", "postgres":
		config := os.Getenv("WOODPECKER_DATABASE_DATASOURCE")
		if !new {
			t.Logf("do not have dump to test against")
			t.SkipNow()
		}
		engine, err = xorm.NewEngine(driver, config)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		return
	default:
		t.Errorf("unsupported driver: %s", driver)
		t.FailNow()
	}
	return
}

func TestMigrate(t *testing.T) {
	// make all tasks required for tests
	for _, task := range migrationTasks {
		task.required = true
	}

	// init new db
	engine, closeDB := testDB(t, true)
	assert.NoError(t, Migrate(engine))
	closeDB()

	dbType := engine.Dialect().URI().DBType
	if dbType == schemas.MYSQL || dbType == schemas.POSTGRES {
		// wait for mysql/postgres to sync ...
		time.Sleep(100 * time.Millisecond)
	}

	// migrate old db
	engine, closeDB = testDB(t, false)
	assert.NoError(t, Migrate(engine))
	closeDB()
}
