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
	"context"
	"encoding/json"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// perPage005 set the size of the slice to read per page.
var perPage005 = 100

var convertToNewPipelineErrorFormat = xormigrate.Migration{
	ID: "convert-to-new-pipeline-error-format",
	MigrateSession: func(sess *xorm.Session) (err error) {
		type pipelineError struct {
			Type      string `json:"type"`
			Message   string `json:"message"`
			IsWarning bool   `json:"is_warning"`
			Data      any    `json:"data"`
		}

		type pipelines struct {
			ID     int64            `json:"id"     xorm:"pk autoincr 'pipeline_id'"`
			Error  string           `json:"error"  xorm:"LONGTEXT 'pipeline_error'"` // old error format
			Errors []*pipelineError `json:"errors" xorm:"json 'pipeline_errors'"`    // new error format
		}

		// This migration was Long, so it can run before or after migration
		// unify-columns-tables renamed pipeline_id and pipeline_errors.
		dialect := sess.Engine().Dialect()
		oldNames, err := dialect.IsColumnExist(sess.Engine().DB(), context.Background(), "pipelines", "pipeline_id")
		if err != nil {
			return err
		}
		idCol, errorsCol := "id", "errors"
		if oldNames {
			idCol, errorsCol = "pipeline_id", "pipeline_errors"
			// make sure pipeline_error column exists
			if err := sess.Sync(new(pipelines)); err != nil {
				return err
			}
		} else if exist, err := dialect.IsColumnExist(sess.Engine().DB(), context.Background(), "pipelines", "pipeline_error"); err != nil || !exist {
			return err
		}

		page := 0
		for {
			var oldPipelines []struct {
				ID    int64  `xorm:"'id'"`
				Error string `xorm:"'pipeline_error'"`
			}
			err := sess.SQL("SELECT "+idCol+" AS id, pipeline_error FROM pipelines WHERE pipeline_error != '' ORDER BY "+idCol+" LIMIT ? OFFSET ?", perPage005, page*perPage005).Find(&oldPipelines)
			if err != nil {
				return err
			}

			for _, oldPipeline := range oldPipelines {
				newErrors, err := json.Marshal([]*pipelineError{{
					Type:    "generic",
					Message: oldPipeline.Error,
				}})
				if err != nil {
					return err
				}

				if _, err := sess.Exec("UPDATE pipelines SET "+errorsCol+" = ? WHERE "+idCol+" = ?", string(newErrors), oldPipeline.ID); err != nil {
					return err
				}
			}

			if len(oldPipelines) < perPage005 {
				break
			}

			page++
		}

		return dropTableColumns(sess, "pipelines", "pipeline_error")
	},
}
