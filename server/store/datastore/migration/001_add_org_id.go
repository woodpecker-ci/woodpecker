// Copyright 2022 Woodpecker Authors
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

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

var addOrgID = xormigrate.Migration{
	ID: "add-org-id",
	MigrateSession: func(sess *xorm.Session) error {
		type users struct {
			ID    int64  `xorm:"pk autoincr 'user_id'"`
			Login string `xorm:"UNIQUE 'user_login'"`
			OrgID int64  `xorm:"user_org_id"`
		}
		type orgs struct {
			ID     int64  `xorm:"pk autoincr 'id'"`
			Name   string `xorm:"UNIQUE 'name'"`
			IsUser bool   `xorm:"is_user"`
		}

		if err := sess.Sync(new(users), new(orgs)); err != nil {
			return fmt.Errorf("sync new models failed: %w", err)
		}

		// get all users
		var us []*users
		if err := sess.Find(&us); err != nil {
			return fmt.Errorf("find all repos failed: %w", err)
		}

		for _, user := range us {
			org := &orgs{}
			has, err := sess.Where("name = ?", user.Login).Get(org)
			if err != nil {
				return fmt.Errorf("getting org failed: %w", err)
			} else if !has {
				org = &orgs{
					Name:   user.Login,
					IsUser: true,
				}
				if _, err := sess.Insert(org); err != nil {
					return fmt.Errorf("inserting org failed: %w", err)
				}
			}
			user.OrgID = org.ID
			if _, err := sess.Cols("user_org_id").Update(user); err != nil {
				return fmt.Errorf("updating user failed: %w", err)
			}
		}

		return nil
	},
}
