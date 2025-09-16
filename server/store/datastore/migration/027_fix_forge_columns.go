package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

var fixForgeColumns = xormigrate.Migration{
	ID: "fix-forge-columns",
	MigrateSession: func(sess *xorm.Session) (err error) {
		if err := renameColumn(sess, "forges", "o_auth_client_i_d", "oauth_client_id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "forges", "o_auth_client_secret", "oauth_client_secret"); err != nil {
			return err
		}

		// Drop client and client_secret columns if they still exist
		if err := dropTableColumns(sess, "forges", "client", "client_secret"); err != nil {
			return err
		}

		return nil
	},
}
