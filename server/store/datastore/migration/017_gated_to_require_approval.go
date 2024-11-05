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
	"xorm.io/builder"
	"xorm.io/xorm"
)

var gatedToRequireApproval = xormigrate.Migration{
	ID: "gated-to-require-approval",
	MigrateSession: func(sess *xorm.Session) (err error) {
		type repos struct {
			ID              int64  `xorm:"pk autoincr 'id'"`
			IsGated         bool   `xorm:"gated"`
			RequireApproval string `xorm:"require_approval"`
			Visibility      string `xorm:"varchar(10) 'visibility'"`
		}

		if err := sess.Sync(new(repos)); err != nil {
			return fmt.Errorf("sync new models failed: %w", err)
		}

		// migrate gated repos
		if _, err := sess.Exec(
			builder.Update(builder.Eq{"require_approval": "all_events"}).
				From(new(repos)).
				Where(builder.Eq{"gated": true})); err != nil {
			return err
		}

		// migrate public repos to new default require approval
		if _, err := sess.Exec(
			builder.Update(builder.Eq{"require_approval": "pull_requests"}).
				From(new(repos)).
				Where(builder.Eq{"gated": false, "visibility": "public"})); err != nil {
			return err
		}

		// migrate private repos to new default require approval
		if _, err := sess.Exec(
			builder.Update(builder.Eq{"require_approval": "none"}).
				From(new(repos)).
				Where(builder.Eq{"gated": false}.And(builder.Neq{"visibility": "public"}))); err != nil {
			return err
		}

		return dropTableColumns(sess, "repos", "gated")
	},
}
