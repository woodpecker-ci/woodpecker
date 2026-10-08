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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pipelineV032 struct {
	ID        int64  `xorm:"pk autoincr 'id'"`
	Event     string `xorm:"event"`
	Commit    string `xorm:"commit"`
	Message   string `xorm:"TEXT 'message'"`
	Timestamp int64  `xorm:"'timestamp'"`
	Email     string `xorm:"varchar(500) email"`
}

func (pipelineV032) TableName() string { return "pipelines" }

type commitV032 struct {
	SHA       string `json:"sha"`
	Message   string `json:"message"`
	ForgeURL  string `json:"forge_url"`
	Timestamp int64  `json:"timestamp"`
	Author    struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"author"`
}

type pipelineAfterV032 struct {
	ID     int64       `xorm:"pk autoincr 'id'"`
	Commit *commitV032 `xorm:"json 'commit'"`
}

func (pipelineAfterV032) TableName() string { return "pipelines" }

func TestUpdatePipelineStructureCommit(t *testing.T) {
	engine, closeDB := testDB(t, true)
	defer closeDB()

	// start from the flat schema, other tests may have left the new one behind
	require.NoError(t, engine.DropTables("pipelines"))
	defer func() {
		_ = engine.DropTables("pipelines")
	}()
	require.NoError(t, engine.Sync(new(pipelineV032)))

	_, err := engine.Insert([]*pipelineV032{
		{ID: 1, Event: "push", Commit: "abc123", Message: "fix bug\n", Timestamp: 1700000000, Email: "alice@example.com"},
		{ID: 2, Event: "deployment", Commit: "def456", Message: "deploy to prod"},
		// a release whose commit was never resolved still gets a commit object
		{ID: 3, Event: "release"},
	})
	require.NoError(t, err)
	// enough rows with gaps in the ids to need several pages
	more := make([]*pipelineV032, 0, 250)
	for i := range 250 {
		id := int64(10 + 3*i)
		more = append(more, &pipelineV032{ID: id, Event: "push", Commit: fmt.Sprintf("sha%d", id)})
	}
	_, err = engine.Insert(more)
	require.NoError(t, err)

	sess := engine.NewSession()
	defer sess.Close()
	require.NoError(t, updatePipelineStructureCommit.MigrateSession(sess))
	require.NoError(t, sess.Commit())

	for _, column := range []string{"message", "timestamp", "email", "commit_new"} {
		exist, err := engine.Dialect().IsColumnExist(engine.DB(), t.Context(), "pipelines", column)
		require.NoError(t, err)
		assert.Falsef(t, exist, "column %s should be gone", column)
	}

	var pipelines []*pipelineAfterV032
	require.NoError(t, engine.Asc("id").Find(&pipelines))
	require.Len(t, pipelines, 253)

	for _, p := range pipelines {
		require.NotNilf(t, p.Commit, "pipeline %d", p.ID)
		if p.ID >= 10 {
			assert.Equal(t, fmt.Sprintf("sha%d", p.ID), p.Commit.SHA)
		}
	}
	assert.Equal(t, "abc123", pipelines[0].Commit.SHA)
	assert.Equal(t, "fix bug\n", pipelines[0].Commit.Message)
	assert.Equal(t, int64(1700000000), pipelines[0].Commit.Timestamp)
	assert.Equal(t, "alice@example.com", pipelines[0].Commit.Author.Email)
	assert.Empty(t, pipelines[0].Commit.Author.Name)
	assert.Empty(t, pipelines[0].Commit.ForgeURL)

	assert.Equal(t, "def456", pipelines[1].Commit.SHA)
	assert.Equal(t, "deploy to prod", pipelines[1].Commit.Message)

	assert.Equal(t, commitV032{}, *pipelines[2].Commit)
}
