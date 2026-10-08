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
		AgentID:      3,
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
		func(s *model.AgentSnapshot) { s.AgentID = 4 },
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
	withEmpty.CustomLabels = map[string]string{}

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
