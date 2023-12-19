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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type oldRepo013 struct {
	ID       int64  `xorm:"pk autoincr 'repo_id'"`
	RemoteID string `xorm:"remote_id"`
}

func (oldRepo013) TableName() string {
	return "repos"
}

var renameRemoteToForge = xormigrate.Migration{
	ID: "rename-remote-to-forge",
	MigrateSession: func(sess *xorm.Session) error {
		if err := renameColumn(sess, "pipelines", "pipeline_remote", "pipeline_clone_url"); err != nil {
			return err
		}

		// make sure the column exist before rename it
		if err := sess.Sync(new(oldRepo013)); err != nil {
			return err
		}

		return renameColumn(sess, "repos", "remote_id", "forge_id")
	},
}
