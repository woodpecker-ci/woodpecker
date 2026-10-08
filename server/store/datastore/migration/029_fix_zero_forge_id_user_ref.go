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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// Users provisioned through the admin web UI or CLI (`admin user add`) used to
// be stored with forge_id=0 when no forge was given, leaving them unresolvable
// for OAuth login (they never match the real forge id, so login is rejected with
// "registration closed"). This heals those rows by clamping them to the default
// forge, mirroring the runtime default (api.defaultForgeID) now enforced on user
// creation and the sibling `replaceZeroForgeIDsInOrgs` migration for orgs.
//
// The orgs update is repeated because a broken user's personal org is created
// with the same forge_id=0 and could have been added after that earlier orgs
// migration already ran.
//
// See https://github.com/woodpecker-ci/woodpecker/issues/6769.
var replaceZeroForgeIDsInUsers = xormigrate.Migration{
	ID: "replace-zero-forge-ids-in-users",
	MigrateSession: func(sess *xorm.Session) (err error) {
		if _, err = sess.Exec("UPDATE users SET forge_id=1 WHERE forge_id=0;"); err != nil {
			return err
		}
		_, err = sess.Exec("UPDATE orgs SET forge_id=1 WHERE forge_id=0;")
		return err
	},
}
