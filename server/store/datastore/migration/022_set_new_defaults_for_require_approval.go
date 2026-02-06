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

var setNewDefaultsForRequireApproval = xormigrate.Migration{
	ID: "set-new-defaults-for-require-approval",
	MigrateSession: func(sess *xorm.Session) (err error) {
		const (
			RequireApprovalOldNotGated string = "old_not_gated"
			RequireApprovalNone        string = "none"
			RequireApprovalForks       string = "forks"
			RequireApprovalAllEvents   string = "all_events"
		)

		type repos struct {
			RequireApproval string `xorm:"require_approval"`
			Visibility      string `xorm:"varchar(10) 'visibility'"`
		}

		if err := sess.Sync(new(repos)); err != nil {
			return fmt.Errorf("sync new models failed: %w", err)
		}

		// migrate public repos to require approval for forks
		if _, err := sess.Exec(
			builder.Update(builder.Eq{"require_approval": RequireApprovalForks}).
				From("repos").
				Where(builder.Eq{"require_approval": RequireApprovalOldNotGated, "visibility": "public"})); err != nil {
			return err
		}

		// migrate private repos to require no approval
		if _, err := sess.Exec(
			builder.Update(builder.Eq{"require_approval": RequireApprovalNone}).
				From("repos").
				Where(builder.Eq{"require_approval": RequireApprovalOldNotGated}.And(builder.Neq{"visibility": "public"}))); err != nil {
			return err
		}

		return nil
	},
}
