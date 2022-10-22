package migration

import (
	"database/sql"
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
	sqliteDB     = "./testfiles/sqlite.db"
	postgresDump = "./testfiles/postgres.sql"
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

func testDB(t *testing.T, new bool) (engine *xorm.Engine, close func()) {
	driver := testDriver()
	var err error
	close = func() {}
	switch driver {
	case "sqlite3":
		config := ":memory:"
		if !new {
			config = createSQLiteDB(t)
			close = func() {
				_ = os.Remove(config)
			}
		}
		engine, err = xorm.NewEngine(driver, config)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		return
	case "mysql":
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
	case "postgres":
		config := os.Getenv("WOODPECKER_DATABASE_DATASOURCE")
		if !new {
			restorePostgresDump(t, config)
			close = func() {
				cleanDB(t, engine)
			}
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

func restorePostgresDump(t *testing.T, config string) {
	dump, err := os.ReadFile(postgresDump)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	db, err := sql.Open("postgres", config)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer db.Close()

	_, err = db.Exec("DROP DATABASE woodpeckerci")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	_, err = db.Exec("CREATE DATABASE woodpeckerci")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	_, err = db.Exec(string(dump))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
}

func cleanDB(t *testing.T, e *xorm.Engine) {
	for _, bean := range allBeans {
		if !assert.NoError(t, e.DropTables(bean)) {
			t.FailNow()
		}
	}
}

func TestMigrate(t *testing.T) {
	// make all tasks required for tests
	for _, task := range migrationTasks {
		task.required = true
	}

	// init new db
	engine, close := testDB(t, true)
	assert.NoError(t, Migrate(engine))
	close()

	dbType := engine.Dialect().URI().DBType
	if dbType == schemas.MYSQL || dbType == schemas.POSTGRES {
		// wait for mysql/postgres to sync ...
		time.Sleep(10 * time.Millisecond)
	}
	assert.NoError(t, engine.ClearCache(allBeans...))

	// migrate old db
	engine, close = testDB(t, false)
	assert.NoError(t, Migrate(engine))
	close()
}
