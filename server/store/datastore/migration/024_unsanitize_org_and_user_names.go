// Excerpt from:
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
	"xorm.io/xorm"
)

var unSanitizeOrgAndUserNames = xormigrate.Migration{
	ID: "unsanitize-org-and-user-names",
	MigrateSession: func(sess *xorm.Session) (err error) {
		type user struct {
			Login string `xorm:"TEXT 'login'"`
		}

		type org struct {
			Name string `xorm:"TEXT 'name'"`
		}

		if err := sess.Sync(new(user), new(org)); err != nil {
			return fmt.Errorf("sync new models failed: %w", err)
		}

		// get all users
		var us []*user
		if err := sess.Find(&us); err != nil {
			return fmt.Errorf("find all repos failed: %w", err)
		}

		for _, user := range us {
			userOrg := &org{}
			_, err := sess.Where("name = ?", user.Login).Get(userOrg)
			if err != nil {
				return fmt.Errorf("getting org failed: %w", err)
			}
			if user.Login != userOrg.Name {
				userOrg.Name = user.Login
				if _, err := sess.ID(userOrg.Name).Cols("Name").Update(userOrg); err != nil {
					return fmt.Errorf("updating org name failed: %w", err)
				}
			}
		}
		return nil
	},
}
