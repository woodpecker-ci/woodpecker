// Copyright 2024 Woodpecker Authors
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

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

var correctPotentialCorruptOrgsUsersRelation = xormigrate.Migration{
	ID: "correct-potential-corrupt-orgs-users-relation",
	MigrateSession: func(sess *xorm.Session) error {
		type users struct {
			ID      int64  `xorm:"pk autoincr 'id'"`
			ForgeID int64  `xorm:"forge_id"`
			Login   string `xorm:"UNIQUE 'login'"`
			OrgID   int64  `xorm:"org_id"`
		}

		type orgs struct {
			ID      int64  `xorm:"pk autoincr 'id'"`
			ForgeID int64  `xorm:"forge_id"`
			Name    string `xorm:"UNIQUE 'name'"`
		}

		if err := sess.Sync(new(users), new(orgs)); err != nil {
			return fmt.Errorf("sync new models failed: %w", err)
		}

		dialect := sess.Engine().Dialect().URI().DBType
		var err error
		switch dialect {
		case schemas.MYSQL:
			_, err = sess.Exec(`UPDATE users u JOIN orgs o ON o.name = u.login AND o.forge_id = u.forge_id SET u.org_id = o.id;`)
		case schemas.POSTGRES:
			_, err = sess.Exec(`UPDATE users u SET org_id = o.id FROM orgs o WHERE o.name = u.login AND o.forge_id = u.forge_id;`)
		case schemas.SQLITE:
			_, err = sess.Exec(`UPDATE users SET org_id = ( SELECT orgs.id FROM orgs WHERE orgs.name = users.login AND orgs.forge_id = users.forge_id ) WHERE users.login IN (SELECT orgs.name FROM orgs);`)
		default:
			err = fmt.Errorf("dialect '%s' not supported", dialect)
		}
		return err
	},
}
