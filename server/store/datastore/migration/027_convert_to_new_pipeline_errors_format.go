// Copyright 2023 Woodpecker Authors
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

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/errors"
)

// perPage027 set the size of the slice to read per page
var perPage027 = 100

type pipeline027 struct {
	ID     int64                   `json:"id"              xorm:"pk autoincr 'pipeline_id'"`
	Error  string                  `json:"error"           xorm:"LONGTEXT 'pipeline_error'"` // old error format
	Errors []*errors.PipelineError `json:"errors"          xorm:"json 'pipeline_errors'"`    // new error format
}

func (pipeline027) TableName() string {
	return "pipelines"
}

type PipelineError027 struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	IsWarning bool   `json:"is_warning"`
	Data      any    `json:"data"`
}

var convertToNewPipelineErrorFormat = xormigrate.Migration{
	ID:   "convert-to-new-pipeline-error-format",
	Long: true,
	MigrateSession: func(sess *xorm.Session) (err error) {
		// make sure pipeline_error column exists
		if err := sess.Sync(new(pipeline027)); err != nil {
			return err
		}

		page := 0
		oldPipelines := make([]*pipeline027, 0, perPage027)

		for {
			oldPipelines = oldPipelines[:0]

			err := sess.Limit(perPage027, page*perPage027).Cols("pipeline_id", "pipeline_error").Where("pipeline_error != ''").Find(&oldPipelines)
			if err != nil {
				return err
			}

			for _, oldPipeline := range oldPipelines {
				var newPipeline pipeline027
				newPipeline.ID = oldPipeline.ID
				newPipeline.Errors = []*errors.PipelineError{{
					Type:    "generic",
					Message: oldPipeline.Error,
				}}

				if _, err := sess.ID(oldPipeline.ID).Cols("pipeline_errors").Update(newPipeline); err != nil {
					return err
				}
			}

			if len(oldPipelines) < perPage027 {
				break
			}

			page++
		}

		return dropTableColumns(sess, "pipelines", "pipeline_error")
	},
}
