package migration

import (
	"fmt"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

var fixForgeColumns = xormigrate.Migration{
	ID: "fix-forge-columns",
	MigrateSession: func(sess *xorm.Session) (err error) {
		// Define old forge structure with old column names
		type forge struct {
			OAuthClientID     string `xorm:"VARCHAR(250) 'o_auth_client_i_d'"`
			OAuthClientSecret string `xorm:"VARCHAR(250) 'o_auth_client_secret'"`
		}

		// ensure old columns exist before renaming
		if err := sess.Sync(new(forge)); err != nil {
			return fmt.Errorf("sync old forge model failed: %w", err)
		}

		// Rename old columns to new names
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
