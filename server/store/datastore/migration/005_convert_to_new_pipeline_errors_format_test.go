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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func TestConvertToNewPipelineErrorFormat(t *testing.T) {
	skipIfNotSQLite(t)

	// the fixture is a legacy database with errors in the old format
	oldErrors := func(t *testing.T, engine *xorm.Engine) []map[string]string {
		rows, err := engine.QueryString("SELECT pipeline_id AS id, pipeline_error FROM pipelines WHERE pipeline_error != ''")
		require.NoError(t, err)
		require.NotEmpty(t, rows)
		return rows
	}

	assertConverted := func(t *testing.T, engine *xorm.Engine, old []map[string]string) {
		exist, err := engine.Dialect().IsColumnExist(engine.DB(), t.Context(), "pipelines", "pipeline_error")
		require.NoError(t, err)
		assert.False(t, exist)

		for _, row := range old {
			rows, err := engine.QueryString("SELECT errors FROM pipelines WHERE id = ?", row["id"])
			require.NoError(t, err)
			require.Len(t, rows, 1)
			want, err := json.Marshal([]map[string]any{{"type": "generic", "message": row["pipeline_error"], "is_warning": false, "data": nil}})
			require.NoError(t, err)
			assert.JSONEq(t, string(want), rows[0]["errors"], "pipeline %s", row["id"])
		}
	}

	t.Run("in order", func(t *testing.T) {
		engine, closeDB := testDB(t, false)
		defer closeDB()
		old := oldErrors(t, engine)

		// it runs without allowing long migrations
		require.NoError(t, Migrate(t.Context(), engine, false))
		assertConverted(t, engine, old)
	})

	t.Run("without the old column", func(t *testing.T) {
		engine, closeDB := testDB(t, false)
		defer closeDB()
		_, err := engine.Exec("ALTER TABLE pipelines DROP COLUMN pipeline_error")
		require.NoError(t, err)

		require.NoError(t, Migrate(t.Context(), engine, false))
		assertConverted(t, engine, nil)
	})

	t.Run("after the columns were renamed", func(t *testing.T) {
		engine, closeDB := testDB(t, false)
		defer closeDB()
		old := oldErrors(t, engine)

		// it was Long before, so older instances skipped it and renamed the columns meanwhile
		convertToNewPipelineErrorFormat.Long = true
		defer func() { convertToNewPipelineErrorFormat.Long = false }()
		require.NoError(t, Migrate(t.Context(), engine, false))
		convertToNewPipelineErrorFormat.Long = false

		require.NoError(t, Migrate(t.Context(), engine, false))
		assertConverted(t, engine, old)
	})
}
