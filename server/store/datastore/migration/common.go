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

package migration

import (
	"fmt"
	"regexp"
	"strings"

	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func renameTable(sess *xorm.Session, old, new string) error {
	dialect := sess.Engine().Dialect().URI().DBType
	switch dialect {
	case schemas.MYSQL:
		_, err := sess.Exec(fmt.Sprintf("RENAME TABLE `%s` TO `%s`;", old, new))
		return err
	case schemas.POSTGRES, schemas.SQLITE:
		_, err := sess.Exec(fmt.Sprintf("ALTER TABLE `%s` RENAME TO `%s`;", old, new))
		return err
	default:
		return fmt.Errorf("dialect '%s' not supported", dialect)
	}
}

// WARNING: YOU MUST COMMIT THE SESSION AT THE END
func dropTableColumns(sess *xorm.Session, tableName string, columnNames ...string) (err error) {
	// Copyright 2017 The Gitea Authors. All rights reserved.
	// Use of this source code is governed by a MIT-style
	// license that can be found in the LICENSE file.

	if tableName == "" || len(columnNames) == 0 {
		return nil
	}
	// TODO: This will not work if there are foreign keys

	dialect := sess.Engine().Dialect().URI().DBType
	switch dialect {
	case schemas.SQLITE:
		// First drop the indexes on the columns
		res, errIndex := sess.Query(fmt.Sprintf("PRAGMA index_list(`%s`)", tableName))
		if errIndex != nil {
			return errIndex
		}
		for _, row := range res {
			indexName := row["name"]
			indexRes, err := sess.Query(fmt.Sprintf("PRAGMA index_info(`%s`)", indexName))
			if err != nil {
				return err
			}
			if len(indexRes) != 1 {
				continue
			}
			indexColumn := string(indexRes[0]["name"])
			for _, name := range columnNames {
				if name == indexColumn {
					_, err := sess.Exec(fmt.Sprintf("DROP INDEX `%s`", indexName))
					if err != nil {
						return err
					}
				}
			}
		}

		// Here we need to get the columns from the original table
		sql := fmt.Sprintf("SELECT sql FROM sqlite_master WHERE tbl_name='%s' and type='table'", tableName)
		res, err := sess.Query(sql)
		if err != nil {
			return err
		}
		tableSQL := string(res[0]["sql"])

		// Separate out the column definitions
		tableSQL = tableSQL[strings.Index(tableSQL, "("):]

		// Remove the required columnNames
		for _, name := range columnNames {
			tableSQL = regexp.MustCompile(regexp.QuoteMeta("`"+name+"`")+"[^`,)]*?[,)]").ReplaceAllString(tableSQL, "")
		}

		// Ensure the query is ended properly
		tableSQL = strings.TrimSpace(tableSQL)
		if tableSQL[len(tableSQL)-1] != ')' {
			if tableSQL[len(tableSQL)-1] == ',' {
				tableSQL = tableSQL[:len(tableSQL)-1]
			}
			tableSQL += ")"
		}

		// Find all the columns in the table
		columns := regexp.MustCompile("`([^`]*)`").FindAllString(tableSQL, -1)

		tableSQL = fmt.Sprintf("CREATE TABLE `new_%s_new` ", tableName) + tableSQL
		if _, err := sess.Exec(tableSQL); err != nil {
			return err
		}

		// Now restore the data
		columnsSeparated := strings.Join(columns, ",")
		insertSQL := fmt.Sprintf("INSERT INTO `new_%s_new` (%s) SELECT %s FROM %s", tableName, columnsSeparated, columnsSeparated, tableName)
		if _, err := sess.Exec(insertSQL); err != nil {
			return err
		}

		// Now drop the old table
		if _, err := sess.Exec(fmt.Sprintf("DROP TABLE `%s`", tableName)); err != nil {
			return err
		}

		// Rename the table
		if _, err := sess.Exec(fmt.Sprintf("ALTER TABLE `new_%s_new` RENAME TO `%s`", tableName, tableName)); err != nil {
			return err
		}
	case schemas.POSTGRES:
		cols := ""
		for _, col := range columnNames {
			if cols != "" {
				cols += ", "
			}
			cols += "DROP COLUMN `" + col + "` CASCADE"
		}
		if _, err := sess.Exec(fmt.Sprintf("ALTER TABLE `%s` %s", tableName, cols)); err != nil {
			return fmt.Errorf("drop table `%s` columns %v: %v", tableName, columnNames, err)
		}
	case schemas.MYSQL:
		// Drop indexes on columns first
		sql := fmt.Sprintf("SHOW INDEX FROM %s WHERE column_name IN ('%s')", tableName, strings.Join(columnNames, "','"))
		res, err := sess.Query(sql)
		if err != nil {
			return err
		}
		for _, index := range res {
			indexName := index["column_name"]
			if len(indexName) > 0 {
				_, err := sess.Exec(fmt.Sprintf("DROP INDEX `%s` ON `%s`", indexName, tableName))
				if err != nil {
					return err
				}
			}
		}

		// Now drop the columns
		cols := ""
		for _, col := range columnNames {
			if cols != "" {
				cols += ", "
			}
			cols += "DROP COLUMN `" + col + "`"
		}
		if _, err := sess.Exec(fmt.Sprintf("ALTER TABLE `%s` %s", tableName, cols)); err != nil {
			return fmt.Errorf("drop table `%s` columns %v: %v", tableName, columnNames, err)
		}
	case schemas.MSSQL:
		cols := ""
		for _, col := range columnNames {
			if cols != "" {
				cols += ", "
			}
			cols += "`" + strings.ToLower(col) + "`"
		}
		sql := fmt.Sprintf("SELECT Name FROM sys.default_constraints WHERE parent_object_id = OBJECT_ID('%[1]s') AND parent_column_id IN (SELECT column_id FROM sys.columns WHERE LOWER(name) IN (%[2]s) AND object_id = OBJECT_ID('%[1]s'))",
			tableName, strings.ReplaceAll(cols, "`", "'"))
		constraints := make([]string, 0)
		if err := sess.SQL(sql).Find(&constraints); err != nil {
			return fmt.Errorf("find constraints: %v", err)
		}
		for _, constraint := range constraints {
			if _, err := sess.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP CONSTRAINT `%s`", tableName, constraint)); err != nil {
				return fmt.Errorf("drop table `%s` default constraint `%s`: %v", tableName, constraint, err)
			}
		}
		sql = fmt.Sprintf("SELECT DISTINCT Name FROM sys.indexes INNER JOIN sys.index_columns ON indexes.index_id = index_columns.index_id AND indexes.object_id = index_columns.object_id WHERE indexes.object_id = OBJECT_ID('%[1]s') AND index_columns.column_id IN (SELECT column_id FROM sys.columns WHERE LOWER(name) IN (%[2]s) AND object_id = OBJECT_ID('%[1]s'))",
			tableName, strings.ReplaceAll(cols, "`", "'"))
		constraints = make([]string, 0)
		if err := sess.SQL(sql).Find(&constraints); err != nil {
			return fmt.Errorf("find constraints: %v", err)
		}
		for _, constraint := range constraints {
			if _, err := sess.Exec(fmt.Sprintf("DROP INDEX `%[2]s` ON `%[1]s`", tableName, constraint)); err != nil {
				return fmt.Errorf("drop index `%[2]s` on `%[1]s`: %v", tableName, constraint, err)
			}
		}

		if _, err := sess.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN %s", tableName, cols)); err != nil {
			return fmt.Errorf("drop table `%s` columns %v: %v", tableName, columnNames, err)
		}
	default:
		return fmt.Errorf("dialect '%s' not supported", dialect)
	}

	return nil
}
