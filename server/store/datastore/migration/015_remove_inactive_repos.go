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
	"xorm.io/xorm"
)

var removeInactiveRepos = task{
	name:     "remove-inactive-repos",
	required: true,
	fn: func(sess *xorm.Session) error {
		// If the timeout is 0, the repo was never activated, so we remove it.
		_, err := sess.Table("repos").Where("repo_active = ?", false).Where("repo_timeout != ?", 0).Delete()
		if err != nil {
			return err
		}

		return dropTableColumns(sess, "users", "user_synced")
	},
}
