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
	"github.com/woodpecker-ci/woodpecker/server/model"
	"xorm.io/xorm"
)

type RepoV008 struct {
	RemoteID string `json:"-" xorm:"'remote_id'"`
}

// TableName return database table name for xorm
func (RepoV008) TableName() string {
	return "repos"
}

var alterTableReposAddRemoteIDCol = task{
	name: "alter-table-repos-add-remote-id-col",
	fn: func(sess *xorm.Session) error {
		if err := sess.Sync2(new(RepoV008)); err != nil {
			return err
		}
		return sess.Sync2(new(model.Redirection))
	},
}
