package migration

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

	if dbType == schemas.SQLITE {
		// skip migration of old db as this is covered by TestCopy for sqlite
		return
	}

	// migrate old db
	engine, closeDB = testDB(t, false)
	assert.NoError(t, Migrate(engine))
	closeDB()
}
