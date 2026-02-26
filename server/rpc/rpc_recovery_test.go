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

package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func agentCtx(t *testing.T, agentID string) context.Context {
	return metadata.NewIncomingContext(
		t.Context(),
		metadata.Pairs("agent_id", agentID),
	)
}

func TestInitWorkflowRecovery(t *testing.T) {
	t.Run("recovery disabled returns error", func(t *testing.T) {
		rpcServer := &RPC{recoveryEnabled: false}

		_, err := rpcServer.InitWorkflowRecovery(t.Context(), "wf-1", []string{"s1"}, 300)
		require.ErrorIs(t, err, ErrRecoveryDisabled)
	})

	t.Run("happy path returns correct state map", func(t *testing.T) {
		storeMock := store_mocks.NewMockStore(t)
		agent := &model.Agent{ID: 42}
		storeMock.On("AgentFind", int64(42)).Return(agent, nil)
		storeMock.On("RecoveryStateCreate", "wf-1", []string{"s1", "s2"}, int64(42), mock.AnythingOfType("int64")).Return(nil)
		storeMock.On("RecoveryStateGetAll", "wf-1").Return([]*model.StepRecoveryState{
			{WorkflowID: "wf-1", StepUUID: "s1", Status: 0, ExitCode: 0},
			{WorkflowID: "wf-1", StepUUID: "s2", Status: 2, ExitCode: 0},
		}, nil)

		rpcServer := &RPC{store: storeMock, recoveryEnabled: true}
		ctx := agentCtx(t, "42")

		result, err := rpcServer.InitWorkflowRecovery(ctx, "wf-1", []string{"s1", "s2"}, 300)
		require.NoError(t, err)
		require.Len(t, result, 2)

		assert.Equal(t, types.RecoveryStatusPending, result["s1"].Status)
		assert.Equal(t, types.RecoveryStatusSuccess, result["s2"].Status)
	})

	t.Run("store error on create propagates", func(t *testing.T) {
		storeMock := store_mocks.NewMockStore(t)
		agent := &model.Agent{ID: 1}
		storeMock.On("AgentFind", int64(1)).Return(agent, nil)
		storeMock.On("RecoveryStateCreate", "wf-1", []string{"s1"}, int64(1), mock.AnythingOfType("int64")).Return(assert.AnError)

		rpcServer := &RPC{store: storeMock, recoveryEnabled: true}
		ctx := agentCtx(t, "1")

		_, err := rpcServer.InitWorkflowRecovery(ctx, "wf-1", []string{"s1"}, 300)
		require.ErrorIs(t, err, assert.AnError)
	})

	t.Run("store error on GetAll propagates", func(t *testing.T) {
		storeMock := store_mocks.NewMockStore(t)
		agent := &model.Agent{ID: 1}
		storeMock.On("AgentFind", int64(1)).Return(agent, nil)
		storeMock.On("RecoveryStateCreate", "wf-1", []string{"s1"}, int64(1), mock.AnythingOfType("int64")).Return(nil)
		storeMock.On("RecoveryStateGetAll", "wf-1").Return(nil, assert.AnError)

		rpcServer := &RPC{store: storeMock, recoveryEnabled: true}
		ctx := agentCtx(t, "1")

		_, err := rpcServer.InitWorkflowRecovery(ctx, "wf-1", []string{"s1"}, 300)
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestUpdateStepRecoveryState(t *testing.T) {
	t.Run("recovery disabled returns error", func(t *testing.T) {
		rpcServer := &RPC{recoveryEnabled: false}

		err := rpcServer.UpdateStepRecoveryState(t.Context(), "wf-1", "s1", types.RecoveryStatusRunning, 0)
		require.ErrorIs(t, err, ErrRecoveryDisabled)
	})

	t.Run("status Pending sets no timestamps", func(t *testing.T) {
		storeMock := store_mocks.NewMockStore(t)
		storeMock.On("RecoveryStateUpdate", mock.MatchedBy(func(s *model.StepRecoveryState) bool {
			return s.WorkflowID == "wf-1" &&
				s.StepUUID == "s1" &&
				s.Status == int(types.RecoveryStatusPending) &&
				s.StartedAt == 0 &&
				s.FinishedAt == 0
		})).Return(nil)

		rpcServer := &RPC{store: storeMock, recoveryEnabled: true}

		err := rpcServer.UpdateStepRecoveryState(t.Context(), "wf-1", "s1", types.RecoveryStatusPending, 0)
		require.NoError(t, err)
	})

	t.Run("status Running sets StartedAt", func(t *testing.T) {
		storeMock := store_mocks.NewMockStore(t)
		storeMock.On("RecoveryStateUpdate", mock.MatchedBy(func(s *model.StepRecoveryState) bool {
			return s.WorkflowID == "wf-1" &&
				s.StepUUID == "s1" &&
				s.Status == int(types.RecoveryStatusRunning) &&
				s.StartedAt > 0 &&
				s.FinishedAt == 0
		})).Return(nil)

		rpcServer := &RPC{store: storeMock, recoveryEnabled: true}

		err := rpcServer.UpdateStepRecoveryState(t.Context(), "wf-1", "s1", types.RecoveryStatusRunning, 0)
		require.NoError(t, err)
	})

	t.Run("status Success sets FinishedAt", func(t *testing.T) {
		storeMock := store_mocks.NewMockStore(t)
		storeMock.On("RecoveryStateUpdate", mock.MatchedBy(func(s *model.StepRecoveryState) bool {
			return s.Status == int(types.RecoveryStatusSuccess) &&
				s.FinishedAt > 0 &&
				s.StartedAt == 0
		})).Return(nil)

		rpcServer := &RPC{store: storeMock, recoveryEnabled: true}

		err := rpcServer.UpdateStepRecoveryState(t.Context(), "wf-1", "s1", types.RecoveryStatusSuccess, 0)
		require.NoError(t, err)
	})

	t.Run("status Failed sets FinishedAt and ExitCode", func(t *testing.T) {
		storeMock := store_mocks.NewMockStore(t)
		storeMock.On("RecoveryStateUpdate", mock.MatchedBy(func(s *model.StepRecoveryState) bool {
			return s.Status == int(types.RecoveryStatusFailed) &&
				s.ExitCode == 137 &&
				s.FinishedAt > 0 &&
				s.StartedAt == 0
		})).Return(nil)

		rpcServer := &RPC{store: storeMock, recoveryEnabled: true}

		err := rpcServer.UpdateStepRecoveryState(t.Context(), "wf-1", "s1", types.RecoveryStatusFailed, 137)
		require.NoError(t, err)
	})
}
