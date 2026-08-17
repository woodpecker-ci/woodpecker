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

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

type workflowV030 struct {
	ID    int64  `xorm:"pk autoincr 'id'"`
	Name  string `xorm:"'name'"`
	Phase string `xorm:"phase"`
}

func (workflowV030) TableName() string { return "workflows" }

type pipelineConfigV030 struct {
	ConfigID   int64 `xorm:"UNIQUE(s) NOT NULL 'config_id'"`
	PipelineID int64 `xorm:"UNIQUE(s) NOT NULL 'pipeline_id'"`
	Source     bool  `xorm:"NOT NULL DEFAULT false 'source'"`
	Effective  bool  `xorm:"NOT NULL DEFAULT false 'effective'"`
}

func (pipelineConfigV030) TableName() string { return "pipeline_configs" }

// The name has to match `-run TestMigrate` so this runs in the serial
// migration pass of `make test-server-datastore`. MySQL and Postgres hand
// every package in `server/store/...` the same database, and the second pass
// runs the datastore package alongside this one.
func TestMigrateAddCompilePhase(t *testing.T) {
	engine, closeDB := testDB(t, true)
	defer closeDB()

	// testDB only resets the database for Postgres, and the datastore tests
	// sync onto whatever tables they find without truncating them. A leftover
	// (config_id, pipeline_id) pair therefore trips the unique constraint over
	// there rather than here.
	defer func() {
		require.NoError(t, engine.DropTables("pipeline_configs", "workflows"))
	}()

	// the pre-migration shape: neither column exists yet
	type workflowsBefore struct {
		ID   int64  `xorm:"pk autoincr 'id'"`
		Name string `xorm:"'name'"`
	}
	type pipelineConfigsBefore struct {
		ConfigID   int64 `xorm:"UNIQUE(s) NOT NULL 'config_id'"`
		PipelineID int64 `xorm:"UNIQUE(s) NOT NULL 'pipeline_id'"`
	}
	require.NoError(t, engine.Table("workflows").Sync(new(workflowsBefore)))
	require.NoError(t, engine.Table("pipeline_configs").Sync(new(pipelineConfigsBefore)))

	_, err := engine.Table("workflows").Insert(&workflowsBefore{Name: "build"})
	require.NoError(t, err)
	_, err = engine.Table("pipeline_configs").Insert(&pipelineConfigsBefore{ConfigID: 1, PipelineID: 1})
	require.NoError(t, err)

	sess := engine.NewSession()
	defer sess.Close()
	require.NoError(t, addCompilePhase.MigrateSession(sess))
	require.NoError(t, sess.Commit())

	// Every workflow that predates the compile phase is an ordinary one. Left
	// empty they would read as "no phase", and the merge would have to guess.
	workflow := new(workflowV030)
	found, err := engine.Where("name = ?", "build").Get(workflow)
	require.NoError(t, err)
	require.True(t, found)
	assert.EqualValues(t, model.WorkflowPhaseRun, workflow.Phase)

	// No pipeline that predates the compile phase had one, so what it was built
	// from is also what it ran. Without the backfill every existing pipeline's
	// config view would come back empty.
	link := new(pipelineConfigV030)
	found, err = engine.Where("pipeline_id = ?", 1).Get(link)
	require.NoError(t, err)
	require.True(t, found)
	assert.True(t, link.Source)
	assert.True(t, link.Effective)
}
