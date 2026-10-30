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

package datastore

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm/contexts"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func testAgentSnapshot() *model.AgentSnapshot {
	return &model.AgentSnapshot{
		OrgID:        model.IDNotSet,
		Name:         "builder-03",
		Platform:     "linux/amd64",
		Backend:      "docker",
		CustomLabels: map[string]string{"zone": "eu", "gpu": "false"},
	}
}

func TestAgentSnapshotPersist(t *testing.T) {
	store, closer := newTestStore(t, new(model.AgentSnapshot))
	defer closer()

	first, err := store.AgentSnapshotPersist(testAgentSnapshot())
	require.NoError(t, err)
	assert.NotZero(t, first.ID)
	assert.NotEmpty(t, first.Hash)

	// identical content is stored once and referenced again
	second, err := store.AgentSnapshotPersist(testAgentSnapshot())
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)

	// every field is part of the identity
	changes := []func(*model.AgentSnapshot){
		func(s *model.AgentSnapshot) { s.OrgID = 5 },
		func(s *model.AgentSnapshot) { s.Name = "builder-04" },
		func(s *model.AgentSnapshot) { s.Platform = "linux/arm64" },
		func(s *model.AgentSnapshot) { s.Backend = "kubernetes" },
		func(s *model.AgentSnapshot) { s.CustomLabels["zone"] = "us" },
	}
	ids := map[int64]bool{first.ID: true}
	for _, change := range changes {
		snapshot := testAgentSnapshot()
		change(snapshot)
		persisted, err := store.AgentSnapshotPersist(snapshot)
		require.NoError(t, err)
		assert.False(t, ids[persisted.ID], "changed snapshot %+v must not reuse an existing one", snapshot)
		ids[persisted.ID] = true
	}

	count, err := store.engine.Count(new(model.AgentSnapshot))
	require.NoError(t, err)
	assert.EqualValues(t, len(changes)+1, count)

	loaded, err := store.AgentSnapshotFind(first.ID)
	require.NoError(t, err)
	assert.Equal(t, first, loaded)

	_, err = store.AgentSnapshotFind(9999)
	assert.ErrorIs(t, err, types.ErrRecordNotExist)
}

func TestAgentSnapshotPersistTreatsNoLabelsAlike(t *testing.T) {
	store, closer := newTestStore(t, new(model.AgentSnapshot))
	defer closer()

	withNil := testAgentSnapshot()
	withNil.CustomLabels = nil
	withEmpty := testAgentSnapshot()
	withEmpty.CustomLabels = make(map[string]string)

	first, err := store.AgentSnapshotPersist(withNil)
	require.NoError(t, err)
	second, err := store.AgentSnapshotPersist(withEmpty)
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
}

// concurrentAgentSnapshotInsertHook simulates a concurrent request that persists the
// same snapshot after AgentSnapshotPersist looked it up but before it inserts its own.
type concurrentAgentSnapshotInsertHook struct {
	store    *storage
	snapshot model.AgentSnapshot
	fired    bool
	err      error
}

var agentSnapshotInsertRe = regexp.MustCompile("(?i)^INSERT INTO [`\"]?agent_snapshots[`\"]? ")

func (h *concurrentAgentSnapshotInsertHook) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	if !h.fired && agentSnapshotInsertRe.MatchString(c.SQL) {
		h.fired = true
		snapshot := h.snapshot
		// bounded, so an AgentSnapshotPersist that holds the only connection fails instead of hanging
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, h.err = h.store.engine.Context(ctx).Insert(&snapshot)
	}
	return c.Ctx, nil
}

func (h *concurrentAgentSnapshotInsertHook) AfterProcess(*contexts.ContextHook) error {
	return nil
}

// agents poll concurrently, so the same snapshot can be persisted twice at once
// (same race as https://github.com/woodpecker-ci/woodpecker/issues/7173 for configs).
func TestAgentSnapshotPersistConcurrentDuplicate(t *testing.T) {
	store, closer := newTestStore(t, new(model.AgentSnapshot))
	defer closer()

	// what the concurrent request stores: same content, so same hash
	reference, err := store.AgentSnapshotPersist(testAgentSnapshot())
	require.NoError(t, err)
	_, err = store.engine.ID(reference.ID).Delete(new(model.AgentSnapshot))
	require.NoError(t, err)

	hook := &concurrentAgentSnapshotInsertHook{store: store, snapshot: *testAgentSnapshot()}
	hook.snapshot.Hash = reference.Hash
	store.engine.AddHook(hook)

	snapshot, err := store.AgentSnapshotPersist(testAgentSnapshot())
	require.True(t, hook.fired)
	require.NoError(t, hook.err)
	require.NoError(t, err)

	count, err := store.engine.Count(new(model.AgentSnapshot))
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)
	assert.NotZero(t, snapshot.ID)
	existing, err := store.AgentSnapshotFind(snapshot.ID)
	require.NoError(t, err)
	assert.Equal(t, reference.Hash, existing.Hash)
}

// Autoscaled agents with the same metadata must not create a row each.
func TestAgentSnapshotPersistSharedByAgents(t *testing.T) {
	store, closer := newTestStore(t, new(model.AgentSnapshot))
	defer closer()

	first := testAgentSnapshot()
	first.AgentID = 3
	second := testAgentSnapshot()
	second.AgentID = 4

	persistedFirst, err := store.AgentSnapshotPersist(first)
	require.NoError(t, err)
	persistedSecond, err := store.AgentSnapshotPersist(second)
	require.NoError(t, err)

	assert.Equal(t, persistedFirst.ID, persistedSecond.ID)
	count, err := store.engine.Count(new(model.AgentSnapshot))
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)

	// the agent id is not stored, whether the row was created or found
	assert.Zero(t, persistedFirst.AgentID)
	assert.Zero(t, persistedSecond.AgentID)
	loaded, err := store.AgentSnapshotFind(persistedFirst.ID)
	require.NoError(t, err)
	assert.Zero(t, loaded.AgentID)
}

func TestAgentSnapshotCleanup(t *testing.T) {
	tests := []struct {
		name   string
		remove func(*storage, int64) error
		ids    []int64
	}{
		{
			name: "delete pipelines",
			remove: func(store *storage, id int64) error {
				return store.DeletePipeline(&model.Pipeline{ID: id})
			},
			ids: []int64{1, 2, 3},
		},
		{
			name: "delete repositories",
			remove: func(store *storage, id int64) error {
				return store.DeleteRepo(&model.Repo{ID: id})
			},
			ids: []int64{1, 2},
		},
		{
			name: "delete organizations",
			remove: func(store *storage, id int64) error {
				return store.OrgDelete(id)
			},
			ids: []int64{1, 2},
		},
		{
			name: "replace workflows",
			remove: func(store *storage, id int64) error {
				return store.WorkflowsReplace(&model.Pipeline{ID: id}, []*model.Workflow{
					{PipelineID: id, PID: 1, Name: "replacement"},
				})
			},
			ids: []int64{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, closer := newTestStore(t, new(model.AgentSnapshot), new(model.Workflow),
				new(model.Pipeline), new(model.Step), new(model.LogEntry), new(model.PipelineConfig),
				new(model.Config), new(model.Repo), new(model.Perm), new(model.Secret),
				new(model.Registry), new(model.Redirection), new(model.Org))
			defer closer()

			shared, err := store.AgentSnapshotPersist(testAgentSnapshot())
			require.NoError(t, err)
			unique := testAgentSnapshot()
			unique.Name = "unique-builder"
			unique, err = store.AgentSnapshotPersist(unique)
			require.NoError(t, err)

			_, err = store.engine.Insert(
				&model.Org{ID: 1, Name: "one"},
				&model.Org{ID: 2, Name: "two"},
				&model.Repo{ID: 1, OrgID: 1, ForgeRemoteID: "1", Owner: "org", Name: "one", FullName: "org/one"},
				&model.Repo{ID: 2, OrgID: 2, ForgeRemoteID: "2", Owner: "org", Name: "two", FullName: "org/two"},
				&model.Pipeline{ID: 1, RepoID: 1, Number: 1},
				&model.Pipeline{ID: 2, RepoID: 1, Number: 2},
				&model.Pipeline{ID: 3, RepoID: 2, Number: 1},
			)
			require.NoError(t, err)
			require.NoError(t, store.WorkflowsCreate([]*model.Workflow{
				{PipelineID: 1, PID: 1, AgentID: 1, AgentSnapshotID: shared.ID},
				{PipelineID: 1, PID: 2, AgentID: 2, AgentSnapshotID: shared.ID},
				{PipelineID: 1, PID: 3, AgentID: 3, AgentSnapshotID: unique.ID},
				// Workflows from before snapshots were introduced have no reference.
				{PipelineID: 1, PID: 4, AgentID: 4},
				{PipelineID: 2, PID: 1, AgentID: 2, AgentSnapshotID: shared.ID},
				{PipelineID: 3, PID: 1, AgentID: 5, AgentSnapshotID: shared.ID},
			}))

			require.NoError(t, tt.remove(store, tt.ids[0]))
			_, err = store.AgentSnapshotFind(unique.ID)
			assert.ErrorIs(t, err, types.ErrRecordNotExist)
			loaded, err := store.AgentSnapshotFind(shared.ID)
			require.NoError(t, err, "another pipeline still references this snapshot")
			assert.Equal(t, shared, loaded)

			for _, id := range tt.ids[1:] {
				require.NoError(t, tt.remove(store, id))
			}
			_, err = store.AgentSnapshotFind(shared.ID)
			assert.ErrorIs(t, err, types.ErrRecordNotExist, "the final reference was removed")
		})
	}
}

func TestAgentSnapshotCleanupRollback(t *testing.T) {
	store, closer := newTestStore(t, new(model.AgentSnapshot), new(model.Workflow), new(model.Step))
	defer closer()

	snapshot, err := store.AgentSnapshotPersist(testAgentSnapshot())
	require.NoError(t, err)
	pipeline := &model.Pipeline{ID: 1}
	require.NoError(t, store.WorkflowsCreate([]*model.Workflow{
		{PipelineID: pipeline.ID, PID: 1, AgentID: 1, AgentSnapshotID: snapshot.ID},
	}))

	// A failed replacement must restore both workflows and their snapshots.
	replacement := &model.Workflow{PipelineID: pipeline.ID, PID: 1}
	require.Error(t, store.WorkflowsReplace(pipeline, []*model.Workflow{replacement, replacement}))
	loaded, err := store.AgentSnapshotFind(snapshot.ID)
	require.NoError(t, err)
	assert.Equal(t, snapshot, loaded)
	workflows, err := store.WorkflowList(pipeline)
	require.NoError(t, err)
	require.Len(t, workflows, 1)
	assert.Equal(t, snapshot.ID, workflows[0].AgentSnapshotID)
}
