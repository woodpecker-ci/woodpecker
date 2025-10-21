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
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	// Blank imports to register the sql drivers.
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

const (
	sqliteDB     = "./test-files/sqlite.db"
	postgresDump = "./test-files/postgres.sql"
)

func testDriver() string {
	driver := os.Getenv("WOODPECKER_DATABASE_DRIVER")
	if len(driver) == 0 {
		return "sqlite3"
	}
	return driver
}

func createSQLiteDB(t *testing.T) string {
	tmpF, err := os.CreateTemp("./test-files", "tmp_")
	require.NoError(t, err)
	dbF, err := os.ReadFile(sqliteDB)
	require.NoError(t, err)

	if !assert.NoError(t, os.WriteFile(tmpF.Name(), dbF, 0o644)) {
		t.FailNow()
	}
	return tmpF.Name()
}

func testDB(t *testing.T, initNewDB bool) (engine *xorm.Engine, closeDB func()) {
	driver := testDriver()
	var err error
	closeDB = func() {}
	switch driver {
	case "sqlite3":
		config := ":memory:"
		if !initNewDB {
			config = createSQLiteDB(t)
			closeDB = func() {
				_ = os.Remove(config)
			}
		}
		engine, err = xorm.NewEngine(driver, config)
		require.NoError(t, err)
		return engine, closeDB
	case "mysql":
		config := os.Getenv("WOODPECKER_DATABASE_DATASOURCE")
		if !initNewDB {
			t.Logf("do not have dump to test against")
			t.SkipNow()
		}
		engine, err = xorm.NewEngine(driver, config)
		require.NoError(t, err)
		return engine, closeDB
	case "postgres":
		config := os.Getenv("WOODPECKER_DATABASE_DATASOURCE")
		if !initNewDB {
			restorePostgresDump(t, config)
			closeDB = func() {
				cleanPostgresDB(t, config)
			}
		}
		engine, err = xorm.NewEngine(driver, config)
		require.NoError(t, err)
		return engine, closeDB
	default:
		t.Errorf("unsupported driver: %s", driver)
		t.FailNow()
	}
	return engine, closeDB
}

func restorePostgresDump(t *testing.T, config string) {
	dump, err := os.ReadFile(postgresDump)
	require.NoError(t, err)

	db, err := sql.Open("postgres", config)
	require.NoError(t, err)
	defer db.Close()

	// clean dump
	lines := strings.Split(string(dump), "\n")
	newLines := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		} else if strings.HasPrefix(line, "\\") {
			continue
		} else if strings.HasPrefix(line, "--") {
			continue
		} else if strings.HasPrefix(line, "\\restrict") {
			continue
		} else if strings.HasPrefix(line, "\\unrestrict") {
			continue
		}
		newLines = append(newLines, line)
	}

	for _, stmt := range strings.Split(strings.Join(newLines, "\n"), ";") {
		if stmt == "" {
			continue
		}

		_, err = db.Exec(stmt)
		if err != nil {
			t.Logf("Failed to execute statement: %s", stmt[:min(len(stmt), 100)])
			require.NoErrorf(t, err, "could not load postgres dump")
		}
	}
}

func cleanPostgresDB(t *testing.T, config string) {
	db, err := sql.Open("postgres", config)
	require.NoError(t, err)
	defer db.Close()

	// Drop and recreate the public schema
	// This removes all tables, indexes, constraints, sequences, etc.
	_, err = db.Exec(`
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
		GRANT ALL ON SCHEMA public TO postgres;
		GRANT ALL ON SCHEMA public TO public;
	`)
	require.NoError(t, err)
}

func TestMigrate(t *testing.T) {
	// init new db
	engine, closeDB := testDB(t, true)
	assert.NoError(t, Migrate(t.Context(), engine, true))
	closeDB()

	dbType := engine.Dialect().URI().DBType
	if dbType == schemas.MYSQL || dbType == schemas.POSTGRES {
		// wait for mysql/postgres to sync ...
		time.Sleep(100 * time.Millisecond)
	}

	// migrate old db
	engine, closeDB = testDB(t, false)
	assert.NoError(t, Migrate(t.Context(), engine, true))
	closeDB()
}
