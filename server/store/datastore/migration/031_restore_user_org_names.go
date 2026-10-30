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
	"fmt"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

var restoreUserOrgNames = xormigrate.Migration{
	ID: "restore-user-org-names",
	MigrateSession: func(sess *xorm.Session) error {
		type users struct {
			ID    int64  `xorm:"pk autoincr 'id'"`
			Login string `xorm:"'login'"`
			OrgID int64  `xorm:"org_id"`
		}

		type orgs struct {
			ID     int64  `xorm:"pk autoincr 'id'"`
			Name   string `xorm:"'name'"`
			IsUser bool   `xorm:"is_user"`
		}

		var userRows []*users
		if err := sess.Find(&userRows); err != nil {
			return fmt.Errorf("find users failed: %w", err)
		}

		for _, user := range userRows {
			if user.OrgID == 0 {
				continue
			}

			org := new(orgs)
			found, err := sess.ID(user.OrgID).Get(org)
			if err != nil {
				return fmt.Errorf("find personal org for user %d failed: %w", user.ID, err)
			}
			if !found || !org.IsUser || org.Name == user.Login {
				continue
			}

			org.Name = user.Login
			if _, err := sess.ID(org.ID).Cols("Name").Update(org); err != nil {
				return fmt.Errorf("restore personal org name for user %d failed: %w", user.ID, err)
			}
		}

		return nil
	},
}
