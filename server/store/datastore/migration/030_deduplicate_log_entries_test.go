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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type logEntryV030 struct {
	ID     int64  `xorm:"pk autoincr 'id'"`
	StepID int64  `xorm:"'step_id'"`
	Line   int    `xorm:"'line'"`
	Data   []byte `xorm:"LONGBLOB"`
}

func (logEntryV030) TableName() string { return "log_entries" }

func TestDeduplicateLogEntries(t *testing.T) {
	engine, closeDB := testDB(t, true)
	defer closeDB()

	require.NoError(t, engine.Sync(new(logEntryV030)))

	_, err := engine.Insert([]*logEntryV030{
		// step 1 had two entries resent, keeping the lowest id of each
		{ID: 1, StepID: 1, Line: 0, Data: []byte("hello")},
		{ID: 2, StepID: 1, Line: 1, Data: []byte("world")},
		{ID: 3, StepID: 1, Line: 0, Data: []byte("hello")},
		{ID: 4, StepID: 1, Line: 1, Data: []byte("world")},
		// a line number repeated three times collapses to one
		{ID: 5, StepID: 1, Line: 2, Data: []byte("thrice")},
		{ID: 6, StepID: 1, Line: 2, Data: []byte("thrice")},
		{ID: 7, StepID: 1, Line: 2, Data: []byte("thrice")},
		// same line numbers under another step must survive
		{ID: 8, StepID: 2, Line: 0, Data: []byte("other")},
		{ID: 9, StepID: 2, Line: 1, Data: []byte("other")},
	})
	require.NoError(t, err)

	sess := engine.NewSession()
	defer sess.Close()
	require.NoError(t, deduplicateLogEntries.MigrateSession(sess))
	require.NoError(t, sess.Commit())

	var remaining []*logEntryV030
	require.NoError(t, engine.Asc("id").Find(&remaining))

	ids := make([]int64, 0, len(remaining))
	for _, entry := range remaining {
		ids = append(ids, entry.ID)
	}

	// the lowest id of every (step_id, line) pair, and nothing else
	assert.Equal(t, []int64{1, 2, 5, 8, 9}, ids)
}

// TestMigrateRemovesDuplicateLogEntries runs the full migration over an
// existing database carrying resent log entries. The cleanup has to happen
// before the UNIQUE(step_id, line) index is created from the model, otherwise
// creating that index fails and the server does not start.
func TestMigrateRemovesDuplicateLogEntries(t *testing.T) {
	engine, closeDB := testDB(t, false)
	defer closeDB()

	// the fixture already holds (step_id 2, line 0); store it a second time
	_, err := engine.Exec(
		"INSERT INTO log_entries (id, step_id, time, line, data, created, type) VALUES (?,?,?,?,?,?,?)",
		900001, 2, 0, 0, []byte("resent"), 1641630525, 0)
	require.NoError(t, err)

	require.NoError(t, Migrate(t.Context(), engine, true))

	res, err := engine.QueryString("SELECT COUNT(*) AS total FROM log_entries WHERE step_id = 2 AND line = 0")
	require.NoError(t, err)
	assert.Equal(t, "1", res[0]["total"], "the resent entry should have been removed")
}
