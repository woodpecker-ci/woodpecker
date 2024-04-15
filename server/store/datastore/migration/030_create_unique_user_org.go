// Copyright 2024 Woodpecker Authors
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

type Users struct {
	UserID     int    `xorm:"user_id"`
	ForgeID    string `xorm:"forge_remote_id"`
	UserLogin  string `xorm:"user_login"`
	UserToken  int    `xorm:"user_token"`
	UserSecret int    `xorm:"user_secret"`
	UserExpiry int    `xorm:"user_expiry"`
	UserEmail  string `xorm:"user_email"`
	UserAvatar string `xorm:"user_avatar"`
	UserAdmin  int    `xorm:"user_admin"`
	UserHash   string `xorm:"user_hash"`
	UserOrgID  int    `xorm:"user_org_id"`
}

type Orgs struct {
	ID        int    `xorm:"id"`
	Name      string `xorm:"name"`
	IsUser    int    `xorm:"is_user"`
	IsPrivate int    `xorm:"private"`
}

// checks whether the user_org_id (table users) for a given user_id is unique in table orgs
// if it is not unique, a new entry in table orgs is being created and user_org_id in table users is updated accordingly for the given user_id
// Original issue: https://codeberg.org/Codeberg-CI/feedback/issues/149#issuecomment-1546709
var createUniqueUserOrg = xormigrate.Migration{
	ID: "createUniqueUserOrg",
	MigrateSession: func(sess *xorm.Session) error {
		var users []*Users
		if err := sess.Find(&users); err != nil {
			return fmt.Errorf("'import users table' failed: %w", err)
		}

		// insert row to users table with user_org_id = 1
		newUser := &Users{
			UserID:    2,
			UserOrgID: 1,
		}

		_, err := sess.Insert(newUser)
		if err != nil {
			return fmt.Errorf("failed to insert new user: %w", err)
		}

		var orgs []*Orgs
		if err := sess.Find(&orgs); err != nil {
			return fmt.Errorf("find all orgs failed: %w", err)
		}

		for _, user := range users {
			// count the rows of the current user_org_id in table users
			count, _ := sess.Where("user_org_id = ?", user.UserOrgID).Count(&user)

			if count > 1 {
				var maxID int
				_, err := sess.SQL("SELECT MAX(ID) FROM orgs").Get(&maxID)
				if err != nil {
					return err
				}

				nextID := maxID + 1
				_, err = sess.Insert(&Orgs{ID: nextID, IsUser: 1})
				if err != nil {
					return err
				}

				_, err = sess.Exec("UPDATE users SET user_org_id = (SELECT MAX(id) FROM orgs) WHERE user_id = ?", user.UserID)
				if err != nil {
					return err
				}
			}
		}

		// Check that the database is in the expected state
		var count int
		has, err := sess.SQL("SELECT COUNT(*) FROM users GROUP BY user_org_id HAVING COUNT(*) > 1").Get(&count)
		if err != nil {
			return fmt.Errorf("failed to count users with unique user_org_id: %w", err)
		}
		if !has {
			return fmt.Errorf("no rows returned from query")
		}
		if count > 0 {
			return fmt.Errorf("found %d users with more than one org, want 0", count)
		}

		return nil
	},
}
