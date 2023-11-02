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
	"github.com/woodpecker-ci/woodpecker/pipeline/errors"
	"xorm.io/xorm"
)

type oldPipeline026 struct {
	ID     int64  `json:"id"              xorm:"pk autoincr 'secret_id'"`
	Errors string `json:"error"                  xorm:"json 'pipeline_error'"`
}

func (oldPipeline026) TableName() string {
	return "pipelines"
}

type PipelineError026 struct {
	Type      string      `json:"type"`
	Message   string      `json:"message"`
	IsWarning bool        `json:"is_warning"`
	Data      interface{} `json:"data"`
}

type newPipeline026 struct {
	ID     int64                   `json:"id"              xorm:"pk autoincr 'secret_id'"`
	Errors []*errors.PipelineError `json:"errors"          xorm:"json 'pipeline_errors'"`
}

func (newPipeline026) TableName() string {
	return "pipelines"
}

var convertToNewPipelineErrorFormat = task{
	name: "convert-to-new-pipeline-error-format",
	fn: func(sess *xorm.Session) (err error) {
		// make sure plugin_only column exists
		if err := sess.Sync(new(oldPipeline026)); err != nil {
			return err
		}

		var oldPipelines []*oldPipeline026
		if err := sess.Find(&oldPipelines); err != nil {
			return err
		}

		for _, oldPipeline := range oldPipelines {
			var pipelineErrors []*PipelineError026

			var newPipeline newPipeline026
			newPipeline.ID = oldPipeline.ID
			for _, pipelineError := range pipelineErrors {
				newPipeline.Errors = []*errors.PipelineError{{
					Type:    "generic",
					Message: pipelineError.Message,
				}}
			}

			if _, err := sess.ID(oldPipeline.ID).Cols("pipeline_errors").Update(&newPipeline); err != nil {
				return err
			}
		}

		return dropTableColumns(sess, "pipelines", "pipeline_error")
	},
}
