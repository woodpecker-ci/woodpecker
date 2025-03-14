// Copyright 2025 Woodpecker Authors
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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

var replaceZeroForgeIDsInOrgs = xormigrate.Migration{
	ID: "replace-zero-forge-ids-in-orgs",
	MigrateSession: func(sess *xorm.Session) (err error) {
		_, err = sess.Exec("UPDATE orgs SET forge_id=1 WHERE forge_id=0;")
		return err
	},
}
