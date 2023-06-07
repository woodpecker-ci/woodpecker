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

type oldPipeline018 struct {
	ID       int64 `xorm:"pk autoincr 'pipeline_id'"`
	Signed   bool  `xorm:"pipeline_signed"`
	Verified bool  `xorm:"pipeline_verified"`
}

func (oldPipeline018) TableName() string {
	return "pipelines"
}

var dropOldCols = task{
	name: "drop-old-col",
	fn: func(sess *xorm.Session) error {
		// make sure columns on pipelines exist
		if err := sess.Sync(new(oldPipeline018)); err != nil {
			return err
		}
		if err := dropTableColumns(sess, "steps", "step_pgid"); err != nil {
			return err
		}

		return dropTableColumns(sess, "pipelines", "pipeline_signed", "pipeline_verified")
	},
}
