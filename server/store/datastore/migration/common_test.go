package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrate(t *testing.T) {

}

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
