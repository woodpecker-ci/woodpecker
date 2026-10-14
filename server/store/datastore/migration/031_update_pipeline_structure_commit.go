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
	"encoding/json"
	"fmt"
	"strings"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

var updatePipelineStructureCommit = xormigrate.Migration{
	ID: "update-pipeline-structure_commit",
	MigrateSession: func(sess *xorm.Session) error {
		// perPage set the size of the slice to read per page.
		perPage := 100

		type commitAuthor struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		type commit struct {
			SHA       string       `json:"sha"`
			Message   string       `json:"message"`
			ForgeURL  string       `json:"forge_url"`
			Author    commitAuthor `json:"author"`
			Timestamp int64        `json:"timestamp"`
		}

		type pipelines struct {
			ID        int64  `xorm:"pk autoincr 'id'"`
			Commit    string `xorm:"commit"`
			Message   string `xorm:"TEXT 'message'"`
			Timestamp int64  `xorm:"'timestamp'"`
			Email     string `xorm:"varchar(500) email"`

			// new field, renamed to commit once the flat columns are gone
			CommitNew *commit `xorm:"json 'commit_new'"`
		}

		if err := sess.Sync(new(pipelines)); err != nil {
			return err
		}

		// jsonThen is the THEN placeholder used for JSON columns.
		// Postgres maps `json` column tag to a native json type,
		// so untyped bind parameters inside a CASE need an explicit cast.
		jsonThen := "?"
		if sess.Engine().Dialect().URI().DBType == schemas.POSTGRES {
			jsonThen = "CAST(? AS json)"
		}

		page := 0
		oldPipelines := make([]*pipelines, 0, perPage)

		for {
			oldPipelines = oldPipelines[:0]

			err := sess.Limit(perPage, page*perPage).
				OrderBy("id ASC").
				Find(&oldPipelines)
			if err != nil {
				return err
			}
			if len(oldPipelines) == 0 {
				break
			}

			// Build a single bulk UPDATE for the whole page instead of one
			// statement per row, the column becomes a CASE keyed by id.
			var (
				commitCase strings.Builder
				commitArgs []any
				ids        []any
			)

			// every pipeline gets a commit object, even without a commit sha
			for _, p := range oldPipelines {
				commitJSON, err := json.Marshal(&commit{
					SHA:       p.Commit,
					Message:   p.Message,
					Timestamp: p.Timestamp,
					Author: commitAuthor{
						Email: p.Email,
					},
				})
				if err != nil {
					return err
				}

				ids = append(ids, p.ID)
				commitCase.WriteString(" WHEN ? THEN " + jsonThen)
				commitArgs = append(commitArgs, p.ID, string(commitJSON))
			}

			placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
			query := fmt.Sprintf(
				"UPDATE `pipelines` SET `commit_new` = CASE `id`%s END WHERE `id` IN (%s)",
				commitCase.String(), placeholders,
			)

			// Argument order must match the textual order of the placeholders
			// above: the CASE, then the WHERE IN list.
			execArgs := make([]any, 0, 1+len(commitArgs)+len(ids))
			execArgs = append(execArgs, query)
			execArgs = append(execArgs, commitArgs...)
			execArgs = append(execArgs, ids...)

			if _, err := sess.Exec(execArgs...); err != nil {
				return err
			}

			if len(oldPipelines) < perPage {
				break
			}

			page++
		}

		if err := dropTableColumns(sess, "pipelines", "commit", "message", "timestamp", "email"); err != nil {
			return err
		}

		return renameColumn(sess, "pipelines", "commit_new", "commit")
	},
}
