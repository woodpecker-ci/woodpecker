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

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

type orgV031 struct {
	ID            int64  `xorm:"pk autoincr 'id'"`
	ForgeID       int64  `xorm:"forge_id UNIQUE(s) UNIQUE(forge)"`
	ForgeRemoteID string `xorm:"UNIQUE(forge) forge_remote_id"`
	Name          string `xorm:"'name' UNIQUE(s)"`
}

func (orgV031) TableName() string { return "orgs" }

// Orgs are now found by their forge remote id. The personal org of a user has
// the remote id of that user, so it can be set right away. All other orgs get
// their remote id from the forge the next time they are used.
var setOrgForgeRemoteIDFromUsers = xormigrate.Migration{
	ID: "set-org-forge-remote-id-from-users",
	MigrateSession: func(sess *xorm.Session) error {
		if err := sess.Sync(new(orgV031)); err != nil {
			return fmt.Errorf("sync new models failed: %w", err)
		}

		var users []*struct {
			ForgeRemoteID model.ForgeRemoteID `xorm:"forge_remote_id"`
			OrgID         int64               `xorm:"org_id"`
		}
		if err := sess.Table("users").Where("org_id > 0").Find(&users); err != nil {
			return err
		}

		for _, user := range users {
			if !user.ForgeRemoteID.IsValid() {
				continue
			}
			if _, err := sess.Table("orgs").
				Where("id = ? AND (forge_remote_id IS NULL OR forge_remote_id = '')", user.OrgID).
				Update(map[string]any{"forge_remote_id": user.ForgeRemoteID}); err != nil {
				return err
			}
		}
		return nil
	},
}
