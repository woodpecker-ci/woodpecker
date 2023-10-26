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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveColumnFromSQLITETableSchema(t *testing.T) {
	schema := "CREATE TABLE repos ( repo_id INTEGER PRIMARY KEY AUTOINCREMENT, repo_user_id INTEGER, repo_owner TEXT, " +
		"repo_name TEXT, repo_full_name TEXT, `repo_avatar` TEXT, repo_branch TEXT, repo_timeout INTEGER, " +
		"repo_allow_pr BOOLEAN, repo_config_path TEXT, repo_visibility TEXT, repo_counter INTEGER, repo_active BOOLEAN, " +
		"repo_fallback BOOLEAN, UNIQUE(repo_full_name) )"

	assert.EqualValues(t, schema, removeColumnFromSQLITETableSchema(schema, ""))

	assert.EqualValues(t, "CREATE TABLE repos ( repo_id INTEGER PRIMARY KEY AUTOINCREMENT, repo_user_id INTEGER, repo_owner TEXT, "+
		"repo_name TEXT, repo_full_name TEXT, repo_branch TEXT, repo_timeout INTEGER, "+
		"repo_allow_pr BOOLEAN, repo_config_path TEXT, repo_visibility TEXT, repo_counter INTEGER, repo_active BOOLEAN, "+
		"repo_fallback BOOLEAN, UNIQUE(repo_full_name) )", removeColumnFromSQLITETableSchema(schema, "repo_avatar"))

	assert.EqualValues(t, "CREATE TABLE repos ( repo_user_id INTEGER, repo_owner TEXT, "+
		"repo_name TEXT, repo_full_name TEXT, `repo_avatar` TEXT, repo_timeout INTEGER, "+
		"repo_allow_pr BOOLEAN, repo_config_path TEXT, repo_visibility TEXT, repo_counter INTEGER, repo_active BOOLEAN, "+
		"repo_fallback BOOLEAN, UNIQUE(repo_full_name) )", removeColumnFromSQLITETableSchema(schema, "repo_id", "repo_branch", "invalid", ""))
}

func TestNormalizeSQLiteTableSchema(t *testing.T) {
	assert.EqualValues(t, "", normalizeSQLiteTableSchema(``))
	assert.EqualValues(t,
		"CREATE TABLE repos ( repo_id INTEGER PRIMARY KEY AUTOINCREMENT, "+
			"repo_user_id INTEGER, repo_owner TEXT, repo_name TEXT, repo_full_name TEXT, "+
			"`repo_avatar` TEXT, repo_link TEXT, repo_clone TEXT, repo_branch TEXT, "+
			"repo_timeout INTEGER, repo_allow_pr BOOLEAN, repo_config_path TEXT, "+
			"repo_visibility TEXT, repo_counter INTEGER, repo_active BOOLEAN, "+
			"repo_fallback BOOLEAN, UNIQUE(repo_full_name) )",
		normalizeSQLiteTableSchema(`CREATE TABLE repos (
 repo_id            INTEGER PRIMARY KEY AUTOINCREMENT
,repo_user_id       INTEGER
,repo_owner         TEXT,
  repo_name         TEXT
,repo_full_name     TEXT
,`+"`"+`repo_avatar`+"`"+`        TEXT
,repo_link          TEXT
,repo_clone         TEXT
,repo_branch        TEXT ,repo_timeout			INTEGER
,repo_allow_pr      BOOLEAN
,repo_config_path   TEXT
, repo_visibility TEXT, repo_counter INTEGER, repo_active BOOLEAN, repo_fallback BOOLEAN,UNIQUE(repo_full_name)
)`))
}
