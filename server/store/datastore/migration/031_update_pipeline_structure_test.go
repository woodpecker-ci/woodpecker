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

type pipelineV031 struct {
	ID         int64  `xorm:"pk autoincr 'id'"`
	Event      string `xorm:"event"`
	Author     string `xorm:"INDEX 'author'"`
	ForgeURL   string `xorm:"forge_url"`
	Ref        string `xorm:"ref"`
	Commit     string `xorm:"commit"`
	Title      string `xorm:"title"`
	Message    string `xorm:"TEXT 'message'"`
	Timestamp  int64  `xorm:"'timestamp'"`
	Email      string `xorm:"varchar(500) email"`
	Avatar     string `xorm:"varchar(500) avatar"`
	Sender     string `xorm:"sender"`
	DeployTo   string `xorm:"deploy"`
	DeployTask string `xorm:"deploy_task"`
}

func (pipelineV031) TableName() string { return "pipelines" }

type commitV031 struct {
	SHA       string `json:"sha"`
	Message   string `json:"message"`
	ForgeURL  string `json:"forge_url"`
	Timestamp int64  `json:"timestamp"`
	Author    struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"author"`
}

type pipelineAfterV031 struct {
	ID     int64       `xorm:"pk autoincr 'id'"`
	Event  string      `xorm:"event"`
	Commit *commitV031 `xorm:"json 'commit'"`
}

func (pipelineAfterV031) TableName() string { return "pipelines" }

func TestUpdatePipelineStructure(t *testing.T) {
	engine, closeDB := testDB(t, true)
	defer closeDB()

	// start from the old schema, other tests may have left the new one behind
	require.NoError(t, engine.DropTables("pipelines"))
	defer func() {
		_ = engine.DropTables("pipelines")
	}()
	require.NoError(t, engine.Sync(new(pipelineV031)))

	_, err := engine.Insert([]*pipelineV031{
		{ID: 1, Event: "push", Author: "alice", Commit: "a1", Message: "fix bug", Timestamp: 1700000000, Email: "alice@example.com", ForgeURL: "https://forge/commit/a1"},
		{ID: 2, Event: "tag", Author: "alice", Commit: "b2", Message: "release it", Ref: "refs/tags/v1"},
		{ID: 3, Event: "pull_request", Author: "bob", Commit: "c3", Message: "add feature", Title: "Add feature", Ref: "refs/pull/7/head"},
		{ID: 4, Event: "deployment", Author: "bob", Commit: "d4", Message: "deploy prod", DeployTo: "prod"},
		{ID: 5, Event: "cron", Commit: "e5", Message: "nightly"},
		{ID: 6, Event: "manual", Author: "carol", Commit: "f6", Message: "MANUAL PIPELINE @ main"},
	})
	require.NoError(t, err)

	sess := engine.NewSession()
	defer sess.Close()
	require.NoError(t, updatePipelineStructure.MigrateSession(sess))
	require.NoError(t, sess.Commit())

	var pipelines []*pipelineAfterV031
	require.NoError(t, engine.Asc("id").Find(&pipelines))
	require.Len(t, pipelines, 6)

	wantSHA := []string{"a1", "b2", "c3", "d4", "e5", "f6"}
	wantMessage := []string{"fix bug", "release it", "add feature", "deploy prod", "nightly", "MANUAL PIPELINE @ main"}
	for i, p := range pipelines {
		if assert.NotNilf(t, p.Commit, "commit of %s pipeline", p.Event) {
			assert.Equalf(t, wantSHA[i], p.Commit.SHA, "sha of %s pipeline", p.Event)
			assert.Equalf(t, wantMessage[i], p.Commit.Message, "message of %s pipeline", p.Event)
		}
	}
	assert.Equal(t, int64(1700000000), pipelines[0].Commit.Timestamp)
	assert.Equal(t, "alice@example.com", pipelines[0].Commit.Author.Email)
}
