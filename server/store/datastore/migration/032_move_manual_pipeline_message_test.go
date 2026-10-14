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

type pipelineV032 struct {
	ID            int64       `xorm:"pk autoincr 'id'"`
	Event         string      `xorm:"event"`
	Branch        string      `xorm:"branch"`
	Commit        *commitV031 `xorm:"json 'commit'"`
	ManualMessage string      `xorm:"TEXT 'manual_message'"`
}

func (pipelineV032) TableName() string { return "pipelines" }

func TestMoveManualPipelineMessage(t *testing.T) {
	engine, closeDB := testDB(t, true)
	defer closeDB()

	require.NoError(t, engine.DropTables("pipelines"))
	defer func() {
		_ = engine.DropTables("pipelines")
	}()

	type pipelineBeforeV032 struct {
		ID     int64       `xorm:"pk autoincr 'id'"`
		Event  string      `xorm:"event"`
		Branch string      `xorm:"branch"`
		Commit *commitV031 `xorm:"json 'commit'"`
	}
	require.NoError(t, engine.Table("pipelines").Sync(new(pipelineBeforeV032)))

	_, err := engine.Table("pipelines").Insert([]*pipelineBeforeV032{
		{ID: 1, Event: "manual", Branch: "main", Commit: &commitV031{SHA: "a", Message: "MANUAL PIPELINE @ main"}},
		{ID: 2, Event: "manual", Branch: "main", Commit: &commitV031{SHA: "b", Message: "MANUAL: redeploy @ prod @ main"}},
		// a push pipeline whose message only looks like a manual one stays untouched
		{ID: 3, Event: "push", Branch: "main", Commit: &commitV031{SHA: "c", Message: "MANUAL PIPELINE @ main"}},
		// a manual message for another branch is no generated one
		{ID: 4, Event: "manual", Branch: "main", Commit: &commitV031{SHA: "d", Message: "MANUAL: other @ dev"}},
	})
	require.NoError(t, err)

	sess := engine.NewSession()
	defer sess.Close()
	require.NoError(t, moveManualPipelineMessage.MigrateSession(sess))
	require.NoError(t, sess.Commit())

	var pipelines []*pipelineV032
	require.NoError(t, engine.Asc("id").Find(&pipelines))
	require.Len(t, pipelines, 4)

	assert.Equal(t, "a", pipelines[0].Commit.SHA)
	assert.Empty(t, pipelines[0].Commit.Message)
	assert.Empty(t, pipelines[0].ManualMessage)

	assert.Equal(t, "b", pipelines[1].Commit.SHA)
	assert.Empty(t, pipelines[1].Commit.Message)
	assert.Equal(t, "redeploy @ prod", pipelines[1].ManualMessage)

	assert.Equal(t, "MANUAL PIPELINE @ main", pipelines[2].Commit.Message)
	assert.Empty(t, pipelines[2].ManualMessage)

	assert.Equal(t, "MANUAL: other @ dev", pipelines[3].Commit.Message)
	assert.Empty(t, pipelines[3].ManualMessage)
}
