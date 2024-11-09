// Copyright 2021 Woodpecker Authors
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

package datastore

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func testDriverConfig() (driver, config string) {
	driver = "sqlite3"
	config = ":memory:"

	if os.Getenv("WOODPECKER_DATABASE_DRIVER") != "" {
		driver = os.Getenv("WOODPECKER_DATABASE_DRIVER")
		config = os.Getenv("WOODPECKER_DATABASE_DATASOURCE")
	}
	return
}

// newTestStore creates a new database connection for testing purposes.
// The database driver and connection string are provided by
// environment variables, with fallback to in-memory sqlite.
func newTestStore(t *testing.T, tables ...any) (*storage, func()) {
	engine, err := xorm.NewEngine(testDriverConfig())
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	for _, table := range tables {
		if err := engine.Sync(table); err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	return &storage{
			engine: engine,
		}, func() {
			for _, bean := range tables {
				if err := engine.DropIndexes(bean); err != nil {
					t.Error(err)
					t.FailNow()
				}
			}
			if err := engine.DropTables(tables...); err != nil {
				t.Error(err)
				t.FailNow()
			}
			if err := engine.Close(); err != nil {
				t.Error(err)
				t.FailNow()
			}

			dbType := engine.Dialect().URI().DBType
			if dbType == schemas.MYSQL || dbType == schemas.POSTGRES {
				// wait for mysql/postgres to sync ...
				time.Sleep(10 * time.Millisecond)
			}
		}
}
