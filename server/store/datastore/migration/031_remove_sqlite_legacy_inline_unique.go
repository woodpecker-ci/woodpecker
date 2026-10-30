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
	"fmt"
	"strconv"
	"strings"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

// Frozen copies of the models at the time of this migration. Do not change
// them when the models change.

type userV031 struct {
	ID            int64  `xorm:"pk autoincr 'id'"`
	ForgeID       int64  `xorm:"forge_id UNIQUE(forge_id) UNIQUE(forge_login)"`
	ForgeRemoteID string `xorm:"forge_remote_id UNIQUE(forge_id)"`
	Login         string `xorm:"'login' UNIQUE(forge_login)"`
	AccessToken   string `xorm:"TEXT 'access_token'"`
	RefreshToken  string `xorm:"TEXT 'refresh_token'"`
	Expiry        int64  `xorm:"expiry"`
	Email         string `xorm:" varchar(500) 'email'"`
	Avatar        string `xorm:" varchar(500) 'avatar'"`
	Admin         bool   `xorm:"admin"`
	Hash          string `xorm:"UNIQUE varchar(500) 'hash'"`
	OrgID         int64  `xorm:"org_id"`
}

func (userV031) TableName() string { return "users" }

type configV031 struct {
	ID     int64  `xorm:"pk autoincr 'id'"`
	RepoID int64  `xorm:"UNIQUE(s) 'repo_id'"`
	Hash   string `xorm:"UNIQUE(s) 'hash'"`
	Name   string `xorm:"UNIQUE(s) 'name'"`
	Data   []byte `xorm:"LONGBLOB 'data'"`
}

func (configV031) TableName() string { return "configs" }

type secretV031 struct {
	ID     int64    `xorm:"pk autoincr 'id'"`
	OrgID  int64    `xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'org_id'"`
	RepoID int64    `xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'repo_id'"`
	Name   string   `xorm:"NOT NULL UNIQUE(s) INDEX 'name'"`
	Value  string   `xorm:"TEXT 'value'"`
	Images []string `xorm:"json 'images'"`
	Events []string `xorm:"json 'events'"`
	Note   string   `xorm:"note"`
}

func (secretV031) TableName() string { return "secrets" }

type permV031 struct {
	UserID  int64 `xorm:"UNIQUE(s) INDEX NOT NULL 'user_id'"`
	RepoID  int64 `xorm:"UNIQUE(s) INDEX NOT NULL 'repo_id'"`
	Pull    bool  `xorm:"pull"`
	Push    bool  `xorm:"push"`
	Admin   bool  `xorm:"admin"`
	Synced  int64 `xorm:"synced"`
	Created int64 `xorm:"created"`
	Updated int64 `xorm:"updated"`
}

func (permV031) TableName() string { return "perms" }

// Databases that started on the schema before xorm still carry inline
// UNIQUE table constraints. Sync cannot see them (they have no index SQL),
// so it never drops them, and some are stricter than the models: they block
// the same login on another forge, the same config under another name and the
// same secret name in two orgs. Only SQLite keeps them; there each affected
// table is rebuilt from its model and the rows are copied over.
var removeSQLiteLegacyInlineUnique = xormigrate.Migration{
	ID: "remove-sqlite-legacy-inline-unique",
	MigrateSession: func(sess *xorm.Session) error {
		if sess.Engine().Dialect().URI().DBType != schemas.SQLITE {
			return nil
		}

		if err := sess.Begin(); err != nil {
			return err
		}

		for _, bean := range []interface{ TableName() string }{
			new(userV031), new(configV031), new(secretV031), new(permV031),
		} {
			if err := rebuildSQLiteTableWithoutInlineUnique(sess, bean); err != nil {
				return fmt.Errorf("table %s: %w", bean.TableName(), err)
			}
		}
		return nil
	},
}

func rebuildSQLiteTableWithoutInlineUnique(sess *xorm.Session, bean interface{ TableName() string }) error {
	table := bean.TableName()

	exist, err := sess.IsTableExist(table)
	if err != nil || !exist {
		return err
	}
	inline, err := sqliteInlineUniqueCount(sess, table)
	if err != nil || inline == 0 {
		return err
	}

	seq, err := sess.QueryString("SELECT seq FROM sqlite_sequence WHERE name = ?", table)
	if err != nil {
		return err
	}
	oldCols, err := sqliteColumns(sess, table)
	if err != nil {
		return err
	}

	// index names stay with a renamed table, Sync would collide with them
	indexes, err := sess.QueryString("SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = ? AND sql IS NOT NULL", table)
	if err != nil {
		return err
	}
	for _, idx := range indexes {
		if _, err := sess.Exec("DROP INDEX `" + idx["name"] + "`"); err != nil {
			return err
		}
	}

	legacy := table + "_legacy_unique"
	if _, err := sess.Exec("ALTER TABLE `" + table + "` RENAME TO `" + legacy + "`"); err != nil {
		return err
	}
	if err := sess.Sync(bean); err != nil {
		return err
	}
	newCols, err := sqliteColumns(sess, table)
	if err != nil {
		return err
	}

	var names, values []string
	for _, col := range newCols {
		old, ok := oldCols[col.name]
		if !ok {
			continue
		}
		value := "`" + col.name + "`"
		if col.notNull && !old.notNull {
			value = "COALESCE(" + value + ", " + col.fallback() + ")"
		}
		names = append(names, "`"+col.name+"`")
		values = append(values, value)
	}
	if _, err := sess.Exec("INSERT INTO `" + table + "` (" + strings.Join(names, ", ") + ") SELECT " + strings.Join(values, ", ") + " FROM `" + legacy + "`"); err != nil {
		return err
	}

	copied, err := sqliteRowCount(sess, table)
	if err != nil {
		return err
	}
	total, err := sqliteRowCount(sess, legacy)
	if err != nil {
		return err
	}
	if copied != total {
		return fmt.Errorf("copied %d of %d rows", copied, total)
	}

	if _, err := sess.Exec("DROP TABLE `" + legacy + "`"); err != nil {
		return err
	}

	// keep ids of deleted rows from being handed out again
	if len(seq) == 1 {
		if _, err := sess.Exec("INSERT INTO sqlite_sequence (name, seq) SELECT ?, 0 WHERE NOT EXISTS (SELECT 1 FROM sqlite_sequence WHERE name = ?)", table, table); err != nil {
			return err
		}
		if _, err := sess.Exec("UPDATE sqlite_sequence SET seq = MAX(seq, CAST(? AS INTEGER)) WHERE name = ?", seq[0]["seq"], table); err != nil {
			return err
		}
	}

	if inline, err = sqliteInlineUniqueCount(sess, table); err != nil {
		return err
	} else if inline != 0 {
		return fmt.Errorf("still has %d inline UNIQUE constraints", inline)
	}
	return nil
}

type sqliteColumn struct {
	name     string
	colType  string
	notNull  bool
	defValue string
}

// fallback is the value a NULL gets when copied into a NOT NULL column.
func (c sqliteColumn) fallback() string {
	if c.defValue != "" {
		return c.defValue
	}
	if t := strings.ToUpper(c.colType); strings.Contains(t, "INT") || strings.Contains(t, "BOOL") {
		return "0"
	}
	return "''"
}

func sqliteColumns(sess *xorm.Session, table string) (map[string]sqliteColumn, error) {
	rows, err := sess.QueryString("SELECT name, type, \"notnull\", dflt_value FROM pragma_table_info(?)", table)
	if err != nil {
		return nil, err
	}
	cols := make(map[string]sqliteColumn, len(rows))
	for _, row := range rows {
		cols[row["name"]] = sqliteColumn{
			name:     row["name"],
			colType:  row["type"],
			notNull:  row["notnull"] == "1",
			defValue: row["dflt_value"],
		}
	}
	return cols, nil
}

func sqliteInlineUniqueCount(sess *xorm.Session, table string) (int, error) {
	rows, err := sess.QueryString("SELECT COUNT(*) AS n FROM pragma_index_list(?) WHERE origin = 'u'", table)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(rows[0]["n"])
}

func sqliteRowCount(sess *xorm.Session, table string) (int, error) {
	rows, err := sess.QueryString("SELECT COUNT(*) AS n FROM `" + table + "`")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(rows[0]["n"])
}
