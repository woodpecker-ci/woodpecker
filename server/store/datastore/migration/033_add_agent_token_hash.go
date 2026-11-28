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

var addAgentTokenHash = xormigrate.Migration{
	ID: "add-agent-token-hash",
	MigrateSession: func(sess *xorm.Session) error {
		type agents struct {
			ID        int64  `xorm:"pk autoincr 'id'"`
			Token     string `xorm:"token"`
			TokenHash string `xorm:"VARCHAR(64) 'token_hash'"`
		}

		if err := sess.Sync(new(agents)); err != nil {
			return fmt.Errorf("sync models failed: %w", err)
		}

		var list []*agents
		if err := sess.Cols("id", "token").Where("token_hash IS NULL OR token_hash = ''").Find(&list); err != nil {
			return err
		}

		for _, agent := range list {
			agent.TokenHash = model.HashAgentToken(agent.Token)
			if _, err := sess.ID(agent.ID).Cols("token_hash").Update(agent); err != nil {
				return err
			}
		}
		return nil
	},
}
