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
	"fmt"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/builder"
	"xorm.io/xorm"
)

var unSanitizeOrgAndUserNames = xormigrate.Migration{
	ID: "unsanitize-org-and-user-names",
	MigrateSession: func(sess *xorm.Session) (err error) {
		type user struct {
			ID      int64  `xorm:"pk autoincr 'id'"`
			Login   string `xorm:"TEXT 'login'"`
			ForgeID int64  `xorm:"forge_id"`
		}

		type org struct {
			ID      int64  `xorm:"pk autoincr 'id'"`
			Name    string `xorm:"TEXT 'name'"`
			ForgeID int64  `xorm:"forge_id"`
		}

		if err := sess.Sync(new(user), new(org)); err != nil {
			return fmt.Errorf("sync new models failed: %w", err)
		}

		// get all users
		var users []*user
		if err := sess.Find(&users); err != nil {
			return fmt.Errorf("find all repos failed: %w", err)
		}

		for _, user := range users {
			userOrg := &org{}
			_, err := sess.Where("name = ? AND forge_id = ?", user.Login, user.ForgeID).Get(userOrg)
			if err != nil {
				return fmt.Errorf("getting org failed: %w", err)
			}

			if user.Login != userOrg.Name {
				userOrg.Name = user.Login
				if _, err := sess.Where(builder.Eq{"id": userOrg.ID}).Cols("Name").Update(userOrg); err != nil {
					return fmt.Errorf("updating org name failed: %w", err)
				}
			}
		}
		return nil
	},
}
