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
	"reflect"
	"strings"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

var updatePipelineStructureCommit = xormigrate.Migration{
	ID: "update-pipeline-structure_commit",
	MigrateSession: func(sess *xorm.Session) error {
		// perPage sets the size of the slice to read per page.
		const perPage = 100

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
			ID       int64  `xorm:"pk autoincr 'id'"`
			Author   string `xorm:"INDEX 'author'"`
			ForgeURL string `xorm:"forge_url"`

			// old columns, folded into the commit substruct below
			Commit    string `xorm:"commit"`
			Message   string `xorm:"TEXT 'message'"`
			Timestamp int64  `xorm:"'timestamp'"`
			Email     string `xorm:"varchar(500) email"`

			// new field, temporary column renamed to 'commit' at the end
			CommitNew *commit `xorm:"json 'commit_new'"`
		}

		if err := sess.Sync(new(pipelines)); err != nil {
			return err
		}

		// Postgres maps the `json` column tag to a native json type, so untyped
		// bind parameters inside a CASE need an explicit cast.
		jsonThen := "?"
		if sess.Engine().Dialect().URI().DBType == schemas.POSTGRES {
			jsonThen = "CAST(? AS json)"
		}

		// marshalJSON encodes v as a JSON string, returning a nil interface for
		// nil pointers so the column is written as SQL NULL rather than the
		// literal "null".
		marshalJSON := func(v any) (any, error) {
			if rv := reflect.ValueOf(v); rv.Kind() == reflect.Pointer && rv.IsNil() {
				return nil, nil
			}
			b, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			return string(b), nil
		}

		page := 0
		oldPipelines := make([]*pipelines, 0, perPage)

		for {
			oldPipelines = oldPipelines[:0]

			if err := sess.Limit(perPage, page*perPage).Find(&oldPipelines); err != nil {
				return err
			}
			if len(oldPipelines) == 0 {
				break
			}

			// Build a single bulk UPDATE for the whole page instead of one
			// statement per row: the commit column becomes a CASE keyed by id,
			// collapsing perPage round-trips into one statement.
			var (
				commitCase strings.Builder
				commitArgs []any
				ids        []any
			)

			for _, p := range oldPipelines {
				commitJSON, err := marshalJSON(&commit{
					SHA:      p.Commit,
					Message:  p.Message,
					ForgeURL: p.ForgeURL,
					Author: commitAuthor{
						Name:  p.Author,
						Email: p.Email,
					},
					Timestamp: p.Timestamp,
				})
				if err != nil {
					return err
				}

				ids = append(ids, p.ID)
				commitCase.WriteString(" WHEN ? THEN " + jsonThen)
				commitArgs = append(commitArgs, p.ID, commitJSON)
			}

			placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
			query := fmt.Sprintf(
				"UPDATE `pipelines` SET `commit_new` = CASE `id`%s END WHERE `id` IN (%s)",
				commitCase.String(), placeholders,
			)

			// Argument order must match the textual order of the placeholders
			// above: the column's CASE, then the WHERE IN list.
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

		// the message, timestamp and email columns now live inside the commit
		// substruct; the old string commit column is replaced by the json one.
		if err := dropTableColumns(sess, "pipelines", "message", "timestamp", "email", "commit"); err != nil {
			return err
		}

		return renameColumn(sess, "pipelines", "commit_new", "commit")
	},
}
