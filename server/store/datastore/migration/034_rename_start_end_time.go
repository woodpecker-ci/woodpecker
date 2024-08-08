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
)

type stepV033 struct {
	Finished int64 `xorm:"stopped"`
}

func (stepV033) TableName() string {
	return "steps"
}

type workflowV033 struct {
	Finished int64 `xorm:"stopped"`
}

func (workflowV033) TableName() string {
	return "workflows"
}

var renameStartEndTime = xormigrate.Migration{
	ID: "rename-start-end-time",
	MigrateSession: func(sess *xorm.Session) (err error) {
		if err := sess.Sync(new(stepV033), new(workflowV033)); err != nil {
			return fmt.Errorf("sync models failed: %w", err)
		}

		// Step
		if err := renameColumn(sess, "steps", "stopped", "finished"); err != nil {
			return err
		}

		// Workflow
		if err := renameColumn(sess, "workflows", "stopped", "finished"); err != nil {
			return err
		}

		return nil
	},
}
