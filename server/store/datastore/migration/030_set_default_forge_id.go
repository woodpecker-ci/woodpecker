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

var setForgeID = xormigrate.Migration{
	ID: "set-forge-id",
	MigrateSession: func(sess *xorm.Session) (err error) {
		if err := sess.Sync(new(model.User), new(model.Repo), new(model.Forge)); err != nil {
			return fmt.Errorf("sync new models failed: %w", err)
		}

		_, err = sess.Exec(fmt.Sprintf("UPDATE `%s` SET forge_id=0;", model.User{}.TableName()))
		if err != nil {
			return err
		}

		_, err = sess.Exec(fmt.Sprintf("UPDATE `%s` SET forge_id=0;", model.Repo{}.TableName()))
		return err
	},
}
