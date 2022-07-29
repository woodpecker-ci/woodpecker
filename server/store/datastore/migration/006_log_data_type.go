package migration

import (
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

var alterTableLogUpdateColumnLogDataType = task{
	name: "alter-table-logs",
	fn: func(sess *xorm.Session) error {
		dialect := sess.Engine().Dialect().URI().DBType
		var sql string

		switch dialect {
		case schemas.POSTGRES:
			sql = "ALTER TABLE logs ALTER COLUMN log_data TYPE LONGBLOB"
		case schemas.MYSQL:
			sql = "ALTER TABLE logs MODIFY COLUMN log_data LONGBLOB"
		case schemas.MSSQL:
			sql = "ALTER TABLE logs MODIFY COLUMN log_data LONGBLOB"
		}

		if sql != "" {
			res, err := sess.Query(sql)
			_ = res

			if err != nil {
				return err
			}
		}

		return nil
	},
}
