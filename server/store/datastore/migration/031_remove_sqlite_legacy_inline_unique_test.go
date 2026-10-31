// Copyright 2026 Woodpecker Authors
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
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

var legacyInlineUniqueTablesV031 = []string{"users", "configs", "secrets", "perms", "steps", "pipelines"}

func skipIfNotSQLite(t *testing.T) {
	t.Helper()
	if testDriver() != "sqlite3" {
		t.Skip("inline UNIQUE constraints only exist in legacy SQLite databases")
	}
}

func inlineUniqueCount(t *testing.T, engine *xorm.Engine, table string) int {
	t.Helper()
	rows, err := engine.QueryString("SELECT COUNT(*) AS n FROM pragma_index_list(?) WHERE origin = 'u'", table)
	require.NoError(t, err)
	return atoiV031(t, rows[0]["n"])
}

func rowCount(t *testing.T, engine *xorm.Engine, table string) int {
	t.Helper()
	rows, err := engine.QueryString("SELECT COUNT(*) AS n FROM `" + table + "`")
	require.NoError(t, err)
	return atoiV031(t, rows[0]["n"])
}

func sequence(t *testing.T, engine *xorm.Engine, table string) int {
	t.Helper()
	rows, err := engine.QueryString("SELECT seq FROM sqlite_sequence WHERE name = ?", table)
	require.NoError(t, err)
	if len(rows) == 0 {
		return 0
	}
	return atoiV031(t, rows[0]["seq"])
}

func atoiV031(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(s)
	require.NoError(t, err)
	return n
}

// tableShape describes columns and indexes the way xorm compares them:
// by type and ordered columns, not by name.
func tableShape(t *testing.T, engine *xorm.Engine, table string) []string {
	t.Helper()
	var shape []string

	cols, err := engine.QueryString("SELECT name, type, \"notnull\", dflt_value, pk FROM pragma_table_info(?)", table)
	require.NoError(t, err)
	for _, c := range cols {
		shape = append(shape, "col "+c["name"]+" "+c["type"]+" notnull="+c["notnull"]+" default="+c["dflt_value"]+" pk="+c["pk"])
	}

	indexes, err := engine.QueryString("SELECT name, \"unique\", origin FROM pragma_index_list(?)", table)
	require.NoError(t, err)
	for _, idx := range indexes {
		idxCols, err := engine.QueryString("SELECT name FROM pragma_index_info(?) ORDER BY seqno", idx["name"])
		require.NoError(t, err)
		names := make([]string, 0, len(idxCols))
		for _, c := range idxCols {
			names = append(names, c["name"])
		}
		shape = append(shape, "index unique="+idx["unique"]+" origin="+idx["origin"]+" ("+strings.Join(names, ",")+")")
	}

	sort.Strings(shape)
	return shape
}

func TestMigrateRemovesSQLiteLegacyInlineUnique(t *testing.T) {
	skipIfNotSQLite(t)

	engine, closeDB := testDB(t, false)
	defer closeDB()

	// the fixture is a legacy database: it still carries the inline constraints
	require.Equal(t, 1, inlineUniqueCount(t, engine, "users"))
	usersBefore := rowCount(t, engine, "users")
	configsBefore := rowCount(t, engine, "config")
	secretsBefore := rowCount(t, engine, "secrets")
	permsBefore := rowCount(t, engine, "perms")
	stepsBefore := rowCount(t, engine, "steps")
	pipelinesBefore := rowCount(t, engine, "pipelines")
	// pretend users with higher ids were deleted: their ids must not come back
	_, err := engine.Exec("UPDATE sqlite_sequence SET seq = 1000 WHERE name = 'users'")
	require.NoError(t, err)

	// a NULL in a column that is NOT NULL in the model must not stop the copy
	_, err = engine.Exec("UPDATE secrets SET secret_repo_id = NULL WHERE secret_id = (SELECT MIN(secret_id) FROM secrets)")
	require.NoError(t, err)

	require.NoError(t, Migrate(t.Context(), engine, true))

	for _, table := range legacyInlineUniqueTablesV031 {
		assert.Zero(t, inlineUniqueCount(t, engine, table), "table %s still has an inline UNIQUE constraint", table)
	}

	// no row is lost and ids are not reused
	assert.Equal(t, usersBefore, rowCount(t, engine, "users"))
	assert.Equal(t, configsBefore, rowCount(t, engine, "configs"))
	assert.Equal(t, secretsBefore, rowCount(t, engine, "secrets"))
	assert.Equal(t, permsBefore, rowCount(t, engine, "perms"))
	assert.Equal(t, stepsBefore, rowCount(t, engine, "steps"))
	assert.Equal(t, pipelinesBefore, rowCount(t, engine, "pipelines"))
	assert.Equal(t, 1000, sequence(t, engine, "users"))
	seqTypes, err := engine.QueryString("SELECT DISTINCT typeof(seq) AS t FROM sqlite_sequence")
	require.NoError(t, err)
	assert.Equal(t, []map[string]string{{"t": "integer"}}, seqTypes)

	nullSecrets, err := engine.QueryString("SELECT COUNT(*) AS n FROM secrets WHERE repo_id IS NULL")
	require.NoError(t, err)
	assert.Equal(t, "0", nullSecrets[0]["n"])

	// the same login on another forge (fixture user 6543 lives on forge 1)
	_, err = engine.Exec("INSERT INTO users (login, forge_id, forge_remote_id, hash) VALUES ('6543', 2, 'remote-6543', 'hash-6543-forge-2')")
	assert.NoError(t, err)
	// the model constraint still holds
	_, err = engine.Exec("INSERT INTO users (login, forge_id, forge_remote_id, hash) VALUES ('6543', 2, 'remote-other', 'hash-6543-again')")
	assert.ErrorContains(t, err, "UNIQUE constraint failed: users.forge_id, users.login")

	// the same config content under another name
	_, err = engine.Exec("INSERT INTO configs (repo_id, hash, name, data) SELECT repo_id, hash, name || '-copy', data FROM configs LIMIT 1")
	assert.NoError(t, err)

	// the same secret name for two orgs
	_, err = engine.Exec("INSERT INTO secrets (org_id, repo_id, name, value) VALUES (90001, 0, 'shared', 'a'), (90002, 0, 'shared', 'b')")
	assert.NoError(t, err)
	_, err = engine.Exec("INSERT INTO secrets (org_id, repo_id, name, value) VALUES (90001, 0, 'shared', 'c')")
	assert.ErrorContains(t, err, "UNIQUE constraint failed: secrets.org_id, secrets.repo_id, secrets.name")

	// the rebuilt tables look exactly like the ones of a new install
	fresh, closeFresh := testDB(t, true)
	defer closeFresh()
	require.NoError(t, Migrate(t.Context(), fresh, true))
	for _, table := range legacyInlineUniqueTablesV031 {
		assert.Equal(t, tableShape(t, fresh, table), tableShape(t, engine, table), "table %s", table)
	}

	// running it again changes nothing
	shapes := make(map[string][]string, len(legacyInlineUniqueTablesV031))
	for _, table := range legacyInlineUniqueTablesV031 {
		shapes[table] = tableShape(t, engine, table)
	}
	sess := engine.NewSession()
	defer sess.Close()
	require.NoError(t, removeSQLiteLegacyInlineUnique.MigrateSession(sess))
	require.NoError(t, sess.Commit())
	for _, table := range legacyInlineUniqueTablesV031 {
		assert.Equal(t, shapes[table], tableShape(t, engine, table), "table %s", table)
	}
}
