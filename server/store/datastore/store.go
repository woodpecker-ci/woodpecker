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

package datastore

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/russross/meddler"

	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/server/store/datastore/ddl"
)

// datastore is an implementation of a model.Store built on top
// of the sql/database driver with a relational database backend.
type datastore struct {
	*sql.DB

	driver string
	config string
}

// Opts are options for a new database connection
type Opts struct {
	Driver string
	Config string
}

// New creates a database connection for the given driver and datasource
// and returns a new Store.
func New(opts *Opts) (store.Store, error) {
	db, err := open(opts.Driver, opts.Config)
	return &datastore{
		DB:     db,
		driver: opts.Driver,
		config: opts.Config,
	}, err
}

// From returns a Store using an existing database connection.
func From(db *sql.DB) store.Store {
	return &datastore{DB: db}
}

// open opens a new database connection with the specified
// driver and connection string and returns a store.
func open(driver, config string) (*sql.DB, error) {
	db, err := sql.Open(driver, config)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %v", err)
	}
	if driver == "mysql" {
		// per issue https://github.com/go-sql-driver/mysql/issues/257
		db.SetMaxIdleConns(0)
	}

	setupMeddler(driver)

	if err := pingDatabase(db); err != nil {
		return nil, fmt.Errorf("database ping attempts failed: %v", err)
	}

	if err := setupDatabase(driver, db); err != nil {
		return nil, fmt.Errorf("database migration failed: %v", err)
	}
	return db, nil
}

// helper function to ping the database with backoff to ensure
// a connection can be established before we proceed with the
// database setup and migration.
func pingDatabase(db *sql.DB) (err error) {
	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			return
		}
		log.Info().Msgf("database ping failed. retry in 1s")
		time.Sleep(time.Second)
	}
	return
}

// helper function to setup the database by performing
// automated database migration steps.
func setupDatabase(driver string, db *sql.DB) error {
	return ddl.Migrate(driver, db)
}

// helper function to setup the meddler default driver
// based on the selected driver name.
func setupMeddler(driver string) {
	switch driver {
	case "sqlite3":
		meddler.Default = meddler.SQLite
	case "mysql":
		meddler.Default = meddler.MySQL
	case "postgres":
		meddler.Default = meddler.PostgreSQL
	}
}
