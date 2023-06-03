// Copyright 2023 Woodpecker Authors
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

type oldStep017 struct {
	ID      int64  `xorm:"pk autoincr 'step_id'"`
	Machine string `xorm:"step_machine"`
}

func (oldStep017) TableName() string {
	return "steps"
}

var removeMachineCol = task{
	name: "remove-machine-col",
	fn: func(sess *xorm.Session) error {
		// make sure step_machine column exists
		if err := sess.Sync(new(oldStep017)); err != nil {
			return err
		}
		return dropTableColumns(sess, "steps", "step_machine")
	},
}
