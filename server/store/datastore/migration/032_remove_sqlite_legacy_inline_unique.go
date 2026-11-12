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
	"context"
	"fmt"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

// Databases that started on the schema before xorm still carry inline
// UNIQUE table constraints. Sync cannot see them (they have no index SQL),
// so it never drops them, and some are stricter than the models: they block
// the same login on another forge, the same config under another name and the
// same secret name in two orgs. Only SQLite keeps them; there each affected
// table is rebuilt with the same columns and the rows are copied over.
// No model is used, so it gives the same result whenever this migration runs.
// The indexes are dropped with the old table and added back by the sync after
// the migrations.
var removeSQLiteLegacyInlineUnique = xormigrate.Migration{
	ID:   "remove-sqlite-legacy-inline-unique",
	Long: true,
	MigrateSession: func(sess *xorm.Session) error {
		if sess.Engine().Dialect().URI().DBType != schemas.SQLITE {
			return nil
		}

		if err := sess.Begin(); err != nil {
			return err
		}

		tables, err := sess.QueryString("SELECT DISTINCT m.name FROM sqlite_master m, pragma_index_list(m.name) i WHERE m.type = 'table' AND i.origin = 'u'")
		if err != nil {
			return err
		}
		for _, t := range tables {
			if err := rebuildSQLiteTableWithoutInlineUnique(sess, t["name"]); err != nil {
				return fmt.Errorf("table %s: %w", t["name"], err)
			}
		}
		return nil
	},
}

func rebuildSQLiteTableWithoutInlineUnique(sess *xorm.Session, table string) error {
	cols, err := sess.QueryString(`SELECT name, type, "notnull", dflt_value, pk FROM pragma_table_info(?)`, table)
	if err != nil {
		return err
	}
	autoIncr, err := sess.QueryString("SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ? AND sql LIKE '%AUTOINCREMENT%'", table)
	if err != nil {
		return err
	}
	schema := schemas.NewEmptyTable()
	for _, col := range cols {
		schema.AddColumn(&schemas.Column{
			Name:            col["name"],
			SQLType:         schemas.SQLType{Name: col["type"]},
			Nullable:        col["notnull"] == "0",
			Default:         col["dflt_value"],
			DefaultIsEmpty:  col["dflt_value"] == "",
			IsPrimaryKey:    col["pk"] != "0",
			IsAutoIncrement: col["pk"] != "0" && len(autoIncr) > 0,
		})
	}
	create, _, err := sess.Engine().Dialect().CreateTableSQL(context.Background(), sess.Tx(), schema, table)
	if err != nil {
		return err
	}

	legacy := table + "_legacy_unique"
	if _, err := sess.Exec("ALTER TABLE `" + table + "` RENAME TO `" + legacy + "`"); err != nil {
		return err
	}
	if _, err := sess.Exec(create); err != nil {
		return err
	}
	if _, err := sess.Exec("INSERT INTO `" + table + "` SELECT * FROM `" + legacy + "`"); err != nil {
		return err
	}

	// keep the old sequence, so ids of deleted rows are not handed out again
	if _, err := sess.Exec("DELETE FROM sqlite_sequence WHERE name = ?", table); err != nil {
		return err
	}
	if _, err := sess.Exec("UPDATE sqlite_sequence SET name = ? WHERE name = ?", table, legacy); err != nil {
		return err
	}

	_, err = sess.Exec("DROP TABLE `" + legacy + "`")
	return err
}
