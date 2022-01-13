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
	// init new db
	engine, close := testDB(t, true)
	assert.NoError(t, Migrate(engine))
	close()

	// migrate old db
	engine, close = testDB(t, false)
	assert.NoError(t, Migrate(engine))
	close()
}
