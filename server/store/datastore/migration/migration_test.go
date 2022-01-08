package migration

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"

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
	tmpF, err := ioutil.TempFile("./testfiles", "tmp_")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	dbF, err := ioutil.ReadFile(sqliteDB)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.NoError(t, ioutil.WriteFile(tmpF.Name(), dbF, 0o644)) {
		t.FailNow()
	}
	return tmpF.Name()
}

func testDB(t *testing.T, new bool) (engine *xorm.Engine, close func(e *xorm.Engine)) {
	driver := testDriver()
	var err error
	close = func(*xorm.Engine) {}
	switch driver {
	case "sqlite3":
		config := ":memory:"
		if !new {
			config = createSQLiteDB(t)
			close = func(*xorm.Engine) {
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
			close = func(e *xorm.Engine) {
				cleanDB(t, e)
			}
		}
		engine, err = xorm.NewEngine(driver, config)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		restorePostgresDump(t, engine)
		return
	default:
		t.Errorf("unsupported driver: %s", driver)
		t.FailNow()
	}
	return
}

func restorePostgresDump(t *testing.T, e *xorm.Engine) {
	dump, err := ioutil.ReadFile(postgresDump)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	_, err = e.SQL(dump).Exec()
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
	// init new db
	engine, close := testDB(t, true)
	assert.NoError(t, Migrate(engine))
	close(engine)

	// migrate old db
	engine, close = testDB(t, false)
	assert.NoError(t, Migrate(engine))
	close(engine)
}
