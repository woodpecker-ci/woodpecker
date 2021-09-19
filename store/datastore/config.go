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
	gosql "database/sql"

	"github.com/russross/meddler"
	"github.com/woodpecker-ci/woodpecker/model"
	"github.com/woodpecker-ci/woodpecker/store/datastore/sql"
)

func (db *datastore) ConfigsForBuild(buildID int64) ([]*model.Config, error) {
	stmt := sql.Lookup(db.driver, "config-find-id")
	var configs = []*model.Config{}
	err := meddler.QueryAll(db, &configs, stmt, buildID)
	return configs, err
}

func (db *datastore) ConfigFindIdentical(repoID int64, hash string) (*model.Config, error) {
	stmt := sql.Lookup(db.driver, "config-find-repo-hash")
	conf := new(model.Config)
	err := meddler.QueryRow(db, conf, stmt, repoID, hash)
	return conf, err
}

func (db *datastore) ConfigFindApproved(config *model.Config) (bool, error) {
	var dest int64
	stmt := sql.Lookup(db.driver, "config-find-approved")
	err := db.DB.QueryRow(stmt, config.RepoID, config.ID).Scan(&dest)
	if err == gosql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (db *datastore) ConfigCreate(config *model.Config) error {
	return meddler.Insert(db, "config", config)
}

func (db *datastore) BuildConfigCreate(buildConfig *model.BuildConfig) error {
	return meddler.Insert(db, "build_config", buildConfig)
}
