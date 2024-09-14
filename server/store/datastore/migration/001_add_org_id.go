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

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

var addOrgID = xormigrate.Migration{
	ID: "add-org-id",
	MigrateSession: func(sess *xorm.Session) error {
		if err := sess.Sync(new(userV009)); err != nil {
			return fmt.Errorf("sync new models failed: %w", err)
		}

		// get all users
		var users []*userV009
		if err := sess.Find(&users); err != nil {
			return fmt.Errorf("find all repos failed: %w", err)
		}

		for _, user := range users {
			org := &model.Org{}
			has, err := sess.Where("name = ?", user.Login).Get(org)
			if err != nil {
				return fmt.Errorf("getting org failed: %w", err)
			} else if !has {
				org = &model.Org{
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
