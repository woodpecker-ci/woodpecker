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

package pipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/types"
)

type mockRecoveryClient struct {
	initResult map[string]*types.RecoveryState
	initErr    error
	updateErr  error

	// Track calls for assertions
	initCalled      bool
	initWorkflowID  string
	initStepUUIDs   []string
	initTimeout     int64
	updateCalls     []updateCall
}

type updateCall struct {
	workflowID string
	stepUUID   string
	status     types.RecoveryStatus
	exitCode   int
}

func (m *mockRecoveryClient) InitWorkflowRecovery(_ context.Context, workflowID string, stepUUIDs []string, timeout int64) (map[string]*types.RecoveryState, error) {
	m.initCalled = true
	m.initWorkflowID = workflowID
	m.initStepUUIDs = stepUUIDs
	m.initTimeout = timeout
	return m.initResult, m.initErr
}

func (m *mockRecoveryClient) UpdateStepRecoveryState(_ context.Context, workflowID, stepUUID string, status types.RecoveryStatus, exitCode int) error {
	m.updateCalls = append(m.updateCalls, updateCall{workflowID, stepUUID, status, exitCode})
	return m.updateErr
}

func TestInitRecoveryState(t *testing.T) {
	t.Run("disabled manager returns nil without calling client", func(t *testing.T) {
		client := &mockRecoveryClient{}
		mgr := NewRecoveryManager(client, "wf-1", false)

		err := mgr.InitRecoveryState(t.Context(), &backend.Config{}, 300)
		require.NoError(t, err)
		assert.False(t, client.initCalled)
	})

	t.Run("enabled manager collects step UUIDs and populates cache", func(t *testing.T) {
		initResult := map[string]*types.RecoveryState{
			"uuid-1": {Status: types.RecoveryStatusPending},
			"uuid-2": {Status: types.RecoveryStatusRunning},
			"uuid-3": {Status: types.RecoveryStatusSuccess, ExitCode: 0},
		}
		client := &mockRecoveryClient{initResult: initResult}
		mgr := NewRecoveryManager(client, "wf-1", true)

		config := &backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{{UUID: "uuid-1"}, {UUID: "uuid-2"}}},
				{Steps: []*backend.Step{{UUID: "uuid-3"}}},
			},
		}

		err := mgr.InitRecoveryState(t.Context(), config, 300)
		require.NoError(t, err)
		assert.True(t, client.initCalled)

		// Verify params forwarded to client
		assert.Equal(t, "wf-1", client.initWorkflowID)
		assert.Equal(t, []string{"uuid-1", "uuid-2", "uuid-3"}, client.initStepUUIDs)
		assert.Equal(t, int64(300), client.initTimeout)

		// Verify cache is populated
		step1 := &backend.Step{UUID: "uuid-1"}
		state := mgr.GetStepState(step1)
		assert.Equal(t, types.RecoveryStatusPending, state.Status)

		step3 := &backend.Step{UUID: "uuid-3"}
		state = mgr.GetStepState(step3)
		assert.Equal(t, types.RecoveryStatusSuccess, state.Status)
	})

	t.Run("empty UUIDs are filtered from collectStepUUIDs", func(t *testing.T) {
		client := &mockRecoveryClient{initResult: map[string]*types.RecoveryState{
			"uuid-1": {Status: types.RecoveryStatusPending},
		}}
		mgr := NewRecoveryManager(client, "wf-1", true)

		config := &backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{{UUID: "uuid-1"}, {UUID: ""}}},
			},
		}

		err := mgr.InitRecoveryState(t.Context(), config, 300)
		require.NoError(t, err)
		assert.Equal(t, []string{"uuid-1"}, client.initStepUUIDs)
	})

	t.Run("client error propagates", func(t *testing.T) {
		client := &mockRecoveryClient{initErr: errors.New("rpc failed")}
		mgr := NewRecoveryManager(client, "wf-1", true)

		config := &backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{{UUID: "uuid-1"}}},
			},
		}

		err := mgr.InitRecoveryState(t.Context(), config, 300)
		require.EqualError(t, err, "rpc failed")
	})
}

func TestShouldSkipStep(t *testing.T) {
	tests := []struct {
		name       string
		status     types.RecoveryStatus
		wantSkip   bool
	}{
		{"Pending", types.RecoveryStatusPending, false},
		{"Running", types.RecoveryStatusRunning, false},
		{"Success", types.RecoveryStatusSuccess, true},
		{"Failed", types.RecoveryStatusFailed, true},
		{"Skipped", types.RecoveryStatusSkipped, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockRecoveryClient{
				initResult: map[string]*types.RecoveryState{
					"step-1": {Status: tt.status},
				},
			}
			mgr := NewRecoveryManager(client, "wf-1", true)

			config := &backend.Config{
				Stages: []*backend.Stage{
					{Steps: []*backend.Step{{UUID: "step-1"}}},
				},
			}
			err := mgr.InitRecoveryState(t.Context(), config, 300)
			require.NoError(t, err)

			skip, state := mgr.ShouldSkipStep(&backend.Step{UUID: "step-1"})
			assert.Equal(t, tt.wantSkip, skip)
			assert.Equal(t, tt.status, state.Status)
		})
	}

	t.Run("disabled manager returns false nil", func(t *testing.T) {
		mgr := NewRecoveryManager(nil, "wf-1", false)
		skip, state := mgr.ShouldSkipStep(&backend.Step{UUID: "step-1"})
		assert.False(t, skip)
		assert.Nil(t, state)
	})
}

func TestShouldReconnect(t *testing.T) {
	mgr := NewRecoveryManager(nil, "wf-1", true)

	assert.False(t, mgr.ShouldReconnect(nil))
	assert.True(t, mgr.ShouldReconnect(&types.RecoveryState{Status: types.RecoveryStatusRunning}))
	assert.False(t, mgr.ShouldReconnect(&types.RecoveryState{Status: types.RecoveryStatusPending}))
	assert.False(t, mgr.ShouldReconnect(&types.RecoveryState{Status: types.RecoveryStatusSuccess}))
	assert.False(t, mgr.ShouldReconnect(&types.RecoveryState{Status: types.RecoveryStatusFailed}))
	assert.False(t, mgr.ShouldReconnect(&types.RecoveryState{Status: types.RecoveryStatusSkipped}))
}

func TestIsRecoverable(t *testing.T) {
	t.Run("active context returns false", func(t *testing.T) {
		mgr := NewRecoveryManager(nil, "wf-1", true)
		assert.False(t, mgr.IsRecoverable(t.Context()))
	})

	t.Run("canceled context with recovery enabled returns true", func(t *testing.T) {
		mgr := NewRecoveryManager(nil, "wf-1", true)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		assert.True(t, mgr.IsRecoverable(ctx))
	})

	t.Run("canceled context with recovery disabled returns false", func(t *testing.T) {
		mgr := NewRecoveryManager(nil, "wf-1", false)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		assert.False(t, mgr.IsRecoverable(ctx))
	})

	t.Run("canceled context with user cancel returns false", func(t *testing.T) {
		mgr := NewRecoveryManager(nil, "wf-1", true)
		mgr.SetCanceled()
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		assert.False(t, mgr.IsRecoverable(ctx))
	})
}

func TestMarkStepMethods(t *testing.T) {
	t.Run("MarkStepRunning calls client with correct args", func(t *testing.T) {
		client := &mockRecoveryClient{}
		mgr := NewRecoveryManager(client, "wf-1", true)
		step := &backend.Step{UUID: "step-1"}

		err := mgr.MarkStepRunning(t.Context(), step)
		require.NoError(t, err)
		require.Len(t, client.updateCalls, 1)
		assert.Equal(t, "wf-1", client.updateCalls[0].workflowID)
		assert.Equal(t, "step-1", client.updateCalls[0].stepUUID)
		assert.Equal(t, types.RecoveryStatusRunning, client.updateCalls[0].status)
		assert.Equal(t, 0, client.updateCalls[0].exitCode)
	})

	t.Run("MarkStepSuccess calls client with correct args", func(t *testing.T) {
		client := &mockRecoveryClient{}
		mgr := NewRecoveryManager(client, "wf-1", true)
		step := &backend.Step{UUID: "step-2"}

		err := mgr.MarkStepSuccess(t.Context(), step)
		require.NoError(t, err)
		require.Len(t, client.updateCalls, 1)
		assert.Equal(t, types.RecoveryStatusSuccess, client.updateCalls[0].status)
		assert.Equal(t, 0, client.updateCalls[0].exitCode)
	})

	t.Run("MarkStepFailed calls client with correct args", func(t *testing.T) {
		client := &mockRecoveryClient{}
		mgr := NewRecoveryManager(client, "wf-1", true)
		step := &backend.Step{UUID: "step-3"}

		err := mgr.MarkStepFailed(t.Context(), step, 137)
		require.NoError(t, err)
		require.Len(t, client.updateCalls, 1)
		assert.Equal(t, types.RecoveryStatusFailed, client.updateCalls[0].status)
		assert.Equal(t, 137, client.updateCalls[0].exitCode)
	})

	t.Run("disabled manager returns nil without calling client", func(t *testing.T) {
		client := &mockRecoveryClient{}
		mgr := NewRecoveryManager(client, "wf-1", false)
		step := &backend.Step{UUID: "step-1"}

		require.NoError(t, mgr.MarkStepRunning(t.Context(), step))
		require.NoError(t, mgr.MarkStepSuccess(t.Context(), step))
		require.NoError(t, mgr.MarkStepFailed(t.Context(), step, 1))
		assert.Empty(t, client.updateCalls)
	})
}

func TestGetStepStateCacheMiss(t *testing.T) {
	client := &mockRecoveryClient{initResult: map[string]*types.RecoveryState{
		"uuid-1": {Status: types.RecoveryStatusSuccess},
	}}
	mgr := NewRecoveryManager(client, "wf-1", true)

	config := &backend.Config{
		Stages: []*backend.Stage{
			{Steps: []*backend.Step{{UUID: "uuid-1"}}},
		},
	}
	err := mgr.InitRecoveryState(t.Context(), config, 300)
	require.NoError(t, err)

	// Unknown UUID returns default Pending state
	state := mgr.GetStepState(&backend.Step{UUID: "unknown-uuid"})
	assert.Equal(t, types.RecoveryStatusPending, state.Status)
}
