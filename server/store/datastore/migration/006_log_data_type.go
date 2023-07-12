package migration

import (
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

var alterTableLogUpdateColumnLogDataType = task{
	name: "alter-table-logs-update-type-of-data",
	fn: func(sess *xorm.Session) (err error) {
		dialect := sess.Engine().Dialect().URI().DBType

		switch dialect {
		case schemas.POSTGRES:
			_, err = sess.Exec("ALTER TABLE logs ALTER COLUMN log_data TYPE BYTEA")
		case schemas.MYSQL:
			_, err = sess.Exec("ALTER TABLE logs MODIFY COLUMN log_data LONGBLOB")
		default:
			// sqlite does only know BLOB in all cases
			return nil
		}

		return err
	},
}
