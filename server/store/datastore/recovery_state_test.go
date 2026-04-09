// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datastore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestRecoveryStateCreateAndGetAll(t *testing.T) {
	store, closer := newTestStore(t, new(model.StepRecoveryState))
	defer closer()

	workflowID := "workflow-123"
	stepUUIDs := []string{"step-a", "step-b", "step-c"}
	agentID := int64(42)
	expiresAt := time.Now().Add(time.Hour).Unix()

	err := store.RecoveryStateCreate(workflowID, stepUUIDs, agentID, expiresAt)
	require.NoError(t, err)

	states, err := store.RecoveryStateGetAll(workflowID)
	require.NoError(t, err)
	require.Len(t, states, 3)

	uuids := make(map[string]bool)
	for _, s := range states {
		assert.Equal(t, workflowID, s.WorkflowID)
		assert.Equal(t, 0, s.Status)
		assert.Equal(t, agentID, s.AgentID)
		assert.Equal(t, expiresAt, s.ExpiresAt)
		assert.Greater(t, s.CreatedAt, int64(0), "CreatedAt should be auto-populated")
		assert.Greater(t, s.UpdatedAt, int64(0), "UpdatedAt should be auto-populated")
		uuids[s.StepUUID] = true
	}
	for _, uuid := range stepUUIDs {
		assert.True(t, uuids[uuid], "expected step UUID %s", uuid)
	}
}

func TestRecoveryStateCreateIdempotent(t *testing.T) {
	store, closer := newTestStore(t, new(model.StepRecoveryState))
	defer closer()

	workflowID := "workflow-456"
	expiresAt := time.Now().Add(time.Hour).Unix()

	err := store.RecoveryStateCreate(workflowID, []string{"step-1", "step-2"}, 1, expiresAt)
	require.NoError(t, err)

	// Second call with different steps should be a no-op
	err = store.RecoveryStateCreate(workflowID, []string{"step-3", "step-4"}, 2, expiresAt)
	require.NoError(t, err)

	states, err := store.RecoveryStateGetAll(workflowID)
	require.NoError(t, err)
	require.Len(t, states, 2)

	uuids := make(map[string]bool)
	for _, s := range states {
		uuids[s.StepUUID] = true
	}
	assert.True(t, uuids["step-1"])
	assert.True(t, uuids["step-2"])
	assert.False(t, uuids["step-3"])
	assert.False(t, uuids["step-4"])
}

func TestRecoveryStateUpdate(t *testing.T) {
	store, closer := newTestStore(t, new(model.StepRecoveryState))
	defer closer()

	workflowID := "workflow-789"
	expiresAt := time.Now().Add(time.Hour).Unix()

	err := store.RecoveryStateCreate(workflowID, []string{"step-x", "step-y"}, 1, expiresAt)
	require.NoError(t, err)

	now := time.Now().Unix()
	err = store.RecoveryStateUpdate(&model.StepRecoveryState{
		WorkflowID: workflowID,
		StepUUID:   "step-x",
		Status:     2, // Success
		ExitCode:   0,
		FinishedAt: now,
	})
	require.NoError(t, err)

	states, err := store.RecoveryStateGetAll(workflowID)
	require.NoError(t, err)
	require.Len(t, states, 2)

	for _, s := range states {
		if s.StepUUID == "step-x" {
			assert.Equal(t, 2, s.Status)
			assert.Equal(t, 0, s.ExitCode)
			assert.Equal(t, now, s.FinishedAt)
		} else {
			assert.Equal(t, "step-y", s.StepUUID)
			assert.Equal(t, 0, s.Status)
		}
	}
}

func TestRecoveryStateCleanExpired(t *testing.T) {
	store, closer := newTestStore(t, new(model.StepRecoveryState))
	defer closer()

	pastExpiry := time.Now().Add(-time.Hour).Unix()
	futureExpiry := time.Now().Add(time.Hour).Unix()

	err := store.RecoveryStateCreate("expired-wf", []string{"s1", "s2"}, 1, pastExpiry)
	require.NoError(t, err)

	err = store.RecoveryStateCreate("active-wf", []string{"s3", "s4"}, 2, futureExpiry)
	require.NoError(t, err)

	err = store.RecoveryStateCleanExpired()
	require.NoError(t, err)

	expiredStates, err := store.RecoveryStateGetAll("expired-wf")
	require.NoError(t, err)
	assert.Empty(t, expiredStates)

	activeStates, err := store.RecoveryStateGetAll("active-wf")
	require.NoError(t, err)
	assert.Len(t, activeStates, 2)
}

func TestRecoveryStateGetAllNonExistent(t *testing.T) {
	store, closer := newTestStore(t, new(model.StepRecoveryState))
	defer closer()

	states, err := store.RecoveryStateGetAll("does-not-exist")
	require.NoError(t, err)
	assert.Empty(t, states)
}
