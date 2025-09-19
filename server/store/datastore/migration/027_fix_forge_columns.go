// Copyright 2025 Woodpecker Authors
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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

var fixForgeColumns = xormigrate.Migration{
	ID: "fix-forge-columns",
	MigrateSession: func(sess *xorm.Session) (err error) {
		type forge struct {
			OAuthClientID     string `xorm:"VARCHAR(250) 'o_auth_client_i_d'"`
			OAuthClientSecret string `xorm:"VARCHAR(250) 'o_auth_client_secret'"`
		}

		// Ensure columns to rename exist
		if err := sess.Sync(new(forge)); err != nil {
			return err
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
