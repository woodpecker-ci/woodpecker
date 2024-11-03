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

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type agentV017 struct {
	ID     int64 `xorm:"pk autoincr 'id'"`
	RepoID int64 `xorm:"INDEX 'repo_id'"`
}

func (agentV017) TableName() string {
	return "agents"
}

var addRepoAgents = xormigrate.Migration{
	ID: "add-repo-agents",
	MigrateSession: func(sess *xorm.Session) (err error) {
		if err := sess.Sync(new(agentV017)); err != nil {
			return fmt.Errorf("sync models failed: %w", err)
		}

		// Update all existing agents to be global agents
		_, err = sess.Cols("repo_id").Update(&model.Agent{
			OrgID: model.IDNotSet,
		})
		return err
	},
}
