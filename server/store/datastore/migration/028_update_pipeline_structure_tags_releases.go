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
	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

var updatePipelineStructureTagsReleases = xormigrate.Migration{
	ID: "update-pipeline-structure_tags-releases",
	MigrateSession: func(sess *xorm.Session) error {
		// perPage set the size of the slice to read per page.
		perPage := 100

		type release struct {
			Title        string `json:"title,omitempty"`
			IsPrerelease bool   `json:"is_prerelease,omitempty"`
		}

		type pipelines struct {
			ID           int64  `xorm:"pk autoincr 'id'"`
			Event        string `xorm:"event"`
			Ref          string `xorm:"ref"`
			Message      string `xorm:"TEXT 'message'"`
			IsPrerelease bool   `xorm:"is_prerelease"`

			// new fields
			Release  *release `xorm:"json 'release'"`
			TagTitle string   `xorm:"tag_title"`
		}

		if err := sess.Sync(new(pipelines)); err != nil {
			return err
		}

		dialect := sess.Engine().Dialect().URI().DBType

		// jsonThen is the THEN placeholder used for JSON columns.
		// Postgres maps `json` column tag to a native json type,
		// so untyped bind parameters inside a CASE need an explicit cast.
		jsonThen := "?"
		if dialect == schemas.POSTGRES {
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

			err := sess.Limit(perPage, page*perPage).Where(builder.In("event", model.EventRelease, model.EventTag)).Find(&oldPipelines)
			if err != nil {
				return err
			}
			if len(oldPipelines) == 0 {
				break
			}

			// Build a single bulk UPDATE for the whole page instead of one
			// statement per row. Every target column becomes a CASE keyed by
			// id, collapsing perPage round-trips into one statement.
			var (
				releaseCase strings.Builder
				tagCase     strings.Builder

				releaseArgs []any
				tagArgs     []any
				ids         []any
			)

			// fill new fields with old values
			for _, p := range oldPipelines {
				p.TagTitle = strings.TrimPrefix(p.Ref, "refs/tags/")

				if p.Event == string(model.EventRelease) {
					p.Release = &release{
						Title:        strings.TrimPrefix(p.Message, "created release "),
						IsPrerelease: p.IsPrerelease,
					}
				}

				releaseJSON, err := marshalJSON(p.Release)
				if err != nil {
					return err
				}

				ids = append(ids, p.ID)

				releaseCase.WriteString(" WHEN ? THEN " + jsonThen)
				releaseArgs = append(releaseArgs, p.ID, releaseJSON)
				tagCase.WriteString(" WHEN ? THEN ?")
				tagArgs = append(tagArgs, p.ID, p.TagTitle)
			}

			placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
			query := fmt.Sprintf(
				"UPDATE `pipelines` SET "+
					"`release` = CASE `id`%s END, "+
					"`tag_title` = CASE `id`%s END "+
					"WHERE `id` IN (%s)",
				releaseCase.String(), tagCase.String(), placeholders,
			)

			// Argument order must match the textual order of the placeholders
			// above: each column's CASE, then the WHERE IN list.
			execArgs := make([]any, 0, 1+len(releaseArgs)+len(tagArgs)+len(ids))
			execArgs = append(execArgs, query)
			execArgs = append(execArgs, releaseArgs...)
			execArgs = append(execArgs, tagArgs...)
			execArgs = append(execArgs, ids...)

			if _, err := sess.Exec(execArgs...); err != nil {
				return err
			}

			if len(oldPipelines) < perPage {
				break
			}

			page++
		}
		return nil
	},
}
