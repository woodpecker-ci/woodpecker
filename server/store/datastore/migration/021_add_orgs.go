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
	"strings"

	"xorm.io/builder"
	"xorm.io/xorm"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

type oldSecret021 struct {
	ID     int64  `json:"id"              xorm:"pk autoincr 'secret_id'"`
	Owner  string `json:"-"               xorm:"'secret_owner'"`
	OrgID  int64  `json:"-"               xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'secret_org_id'"`
	RepoID int64  `json:"-"               xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'secret_repo_id'"`
	Name   string `json:"name"            xorm:"NOT NULL UNIQUE(s) INDEX 'secret_name'"`
}

func (oldSecret021) TableName() string {
	return "secrets"
}

var addOrgs = task{
	name:     "add-orgs",
	required: true,
	fn: func(sess *xorm.Session) error {
		if exist, err := sess.IsTableExist("orgs"); exist && err == nil {
			if err := sess.DropTable("orgs"); err != nil {
				return fmt.Errorf("drop old orgs table failed: %w", err)
			}
		}

		if err := sess.Sync(new(model.Org), new(model.Repo), new(model.User)); err != nil {
			return fmt.Errorf("sync new models failed: %w", err)
		}

		// make sure the columns exist before removing them
		if err := sess.Sync(new(oldSecret021)); err != nil {
			return fmt.Errorf("sync old secrets models failed: %w", err)
		}

		// get all org names from repos
		var repos []*model.Repo
		if err := sess.Find(&repos); err != nil {
			return fmt.Errorf("find all repos failed: %w", err)
		}

		orgs := make(map[string]*model.Org)
		users := make(map[string]bool)
		for _, repo := range repos {
			orgName := strings.ToLower(repo.Owner)

			// check if it's a registered user
			if _, ok := users[orgName]; !ok {
				exist, err := sess.Where("user_login = ?", orgName).Exist(new(model.User))
				if err != nil {
					return fmt.Errorf("check if user '%s' exist failed: %w", orgName, err)
				}
				users[orgName] = exist
			}

			// create org if not already created
			if _, ok := orgs[orgName]; !ok {
				org := &model.Org{
					Name:   orgName,
					IsUser: users[orgName],
				}
				if _, err := sess.Insert(org); err != nil {
					return fmt.Errorf("insert org %#v failed: %w", org, err)
				}
				orgs[orgName] = org

				// update org secrets
				var secrets []*oldSecret021
				if err := sess.Where(builder.Eq{"secret_owner": orgName, "secret_repo_id": 0}).Find(&secrets); err != nil {
					return fmt.Errorf("get org secrets failed: %w", err)
				}

				for _, secret := range secrets {
					secret.OrgID = org.ID
					if _, err := sess.ID(secret.ID).Cols("secret_org_id").Update(secret); err != nil {
						return fmt.Errorf("update org secret %d failed: %w", secret.ID, err)
					}
				}
			}

			// update the repo
			repo.OrgID = orgs[orgName].ID
			if _, err := sess.ID(repo.ID).Cols("repo_org_id").Update(repo); err != nil {
				return fmt.Errorf("update repos failed: %w", err)
			}
		}

		return dropTableColumns(sess, "secrets", "secret_owner")
	},
}
