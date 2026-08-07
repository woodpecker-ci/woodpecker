// Copyright 2022 Woodpecker Authors
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

package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestRegisterAgent(t *testing.T) {
	t.Run("When existing agent Name is empty it should update Name with hostname from metadata", func(t *testing.T) {
		store := store_mocks.NewMockStore(t)
		storeAgent := new(model.Agent)
		storeAgent.ID = 1337
		updatedAgent := model.Agent{
			ID:          1337,
			Created:     0,
			Updated:     0,
			Name:        "hostname",
			OwnerID:     0,
			Token:       "",
			LastContact: 0,
			Platform:    "platform",
			Backend:     "backend",
			Capacity:    2,
			Version:     "version",
			NoSchedule:  false,
		}

		store.On("AgentFind", int64(1337)).Once().Return(storeAgent, nil)
		store.On("AgentUpdate", &updatedAgent).Once().Return(nil)
		grpc := RPC{
			store: store,
		}
		ctx := metadata.NewIncomingContext(
			t.Context(),
			metadata.Pairs("hostname", "hostname"),
		)
		ctx = context.WithValue(ctx, agentIDKey, int64(1337))
		agentID, err := grpc.RegisterAgent(ctx, rpc.AgentInfo{
			Version:  "version",
			Platform: "platform",
			Backend:  "backend",
			Capacity: 2,
		})
		require.NoError(t, err)

		assert.EqualValues(t, 1337, agentID)
	})

	t.Run("When existing agent hostname is present it should not update the hostname", func(t *testing.T) {
		store := store_mocks.NewMockStore(t)
		storeAgent := new(model.Agent)
		storeAgent.ID = 1337
		storeAgent.Name = "originalHostname"
		updatedAgent := model.Agent{
			ID:          1337,
			Created:     0,
			Updated:     0,
			Name:        "originalHostname",
			OwnerID:     0,
			Token:       "",
			LastContact: 0,
			Platform:    "platform",
			Backend:     "backend",
			Capacity:    2,
			Version:     "version",
			NoSchedule:  false,
		}

		store.On("AgentFind", int64(1337)).Once().Return(storeAgent, nil)
		store.On("AgentUpdate", &updatedAgent).Once().Return(nil)
		grpc := RPC{
			store: store,
		}
		ctx := metadata.NewIncomingContext(
			t.Context(),
			metadata.Pairs("hostname", "newHostname"),
		)
		ctx = context.WithValue(ctx, agentIDKey, int64(1337))
		agentID, err := grpc.RegisterAgent(ctx, rpc.AgentInfo{
			Version:  "version",
			Platform: "platform",
			Backend:  "backend",
			Capacity: 2,
		})
		require.NoError(t, err)

		assert.EqualValues(t, 1337, agentID)
	})
}

func TestCompleteChildrenIfParentCompleted(t *testing.T) {
	t.Run("When a service step is still running it should update state so workflow finishes as success", func(t *testing.T) {
		successStep := &model.Step{
			ID:      1,
			State:   model.StatusSuccess,
			Started: 1234567800,
		}
		runningService := &model.Step{
			ID:      2,
			State:   model.StatusRunning,
			Started: 1234567800,
		}
		workflow := model.Workflow{
			ID:       7,
			State:    model.StatusRunning,
			Children: []*model.Step{successStep, runningService},
		}

		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("StepUpdate", mock.Anything).Return(nil)
		mockStore.On("WorkflowUpdate", mock.Anything).Return(nil)

		s := RPC{store: mockStore}
		s.completeChildrenIfParentCompleted(&workflow, 1234567900)

		assert.Equal(t, model.StatusSuccess, runningService.State)
		assert.Equal(t, int64(1234567900), runningService.Finished)

		result, err := pipeline.UpdateWorkflowStatusToDone(mockStore, workflow, rpc.WorkflowState{
			Started:  1234567800,
			Finished: 1234567900,
		})
		require.NoError(t, err)
		assert.Equal(t, model.StatusSuccess, result.State)
	})
}

func TestUpdateAgentLastWork(t *testing.T) {
	t.Run("When last work was never updated it should update last work timestamp", func(t *testing.T) {
		agent := model.Agent{
			LastWork: 0,
		}
		store := store_mocks.NewMockStore(t)
		rpc := RPC{
			store: store,
		}
		store.On("AgentUpdate", mock.Anything).Once().Return(nil)

		err := rpc.updateAgentLastWork(&agent)
		assert.NoError(t, err)

		assert.NotZero(t, agent.LastWork)
	})

	t.Run("When last work was updated over a minute ago it should update last work timestamp", func(t *testing.T) {
		lastWork := time.Now().Add(-time.Hour).Unix()
		agent := model.Agent{
			LastWork: lastWork,
		}
		store := store_mocks.NewMockStore(t)
		rpc := RPC{
			store: store,
		}
		store.On("AgentUpdate", mock.Anything).Once().Return(nil)

		err := rpc.updateAgentLastWork(&agent)
		assert.NoError(t, err)

		assert.NotEqual(t, lastWork, agent.LastWork)
	})

	t.Run("When last work was updated in the last minute it should not update last work timestamp again", func(t *testing.T) {
		lastWork := time.Now().Add(-time.Second * 30).Unix()
		agent := model.Agent{
			LastWork: lastWork,
		}
		rpc := RPC{}

		err := rpc.updateAgentLastWork(&agent)
		assert.NoError(t, err)

		assert.Equal(t, lastWork, agent.LastWork)
	})
}

func TestWaitWorkflow(t *testing.T) {
	agent := &model.Agent{ID: 5, OwnerID: model.IDNotSet, OrgID: model.IDNotSet}
	callerWorkflow := &model.Workflow{ID: 10, PipelineID: 100, AgentID: 5, Name: "base"}
	currentPipeline := &model.Pipeline{ID: 100, RepoID: 200}
	repo := &model.Repo{ID: 200}

	newRPC := func(t *testing.T) (*RPC, *store_mocks.MockStore, context.Context) {
		store := store_mocks.NewMockStore(t)
		store.On("AgentFind", int64(5)).Return(agent, nil)
		store.On("WorkflowLoad", int64(10)).Return(callerWorkflow, nil)
		store.On("GetPipeline", int64(100)).Return(currentPipeline, nil)
		store.On("GetRepo", int64(200)).Return(repo, nil)
		ctx := context.WithValue(t.Context(), agentIDKey, int64(5))
		return &RPC{store: store}, store, ctx
	}

	t.Run("terminal target returns its status", func(t *testing.T) {
		grpc, store, ctx := newRPC(t)
		store.On("WorkflowGetTree", currentPipeline).Return([]*model.Workflow{
			callerWorkflow,
			{ID: 11, Name: "auxiliaries", State: model.StatusSuccess},
		}, nil)

		result, err := grpc.WaitWorkflow(ctx, "10", "auxiliaries", "")
		require.NoError(t, err)
		assert.True(t, result.Found)
		assert.Equal(t, string(model.StatusSuccess), result.Status)
	})

	t.Run("unknown target reports not found", func(t *testing.T) {
		grpc, store, ctx := newRPC(t)
		store.On("WorkflowGetTree", currentPipeline).Return([]*model.Workflow{callerWorkflow}, nil)

		result, err := grpc.WaitWorkflow(ctx, "10", "auxiliaries", "")
		require.NoError(t, err)
		assert.False(t, result.Found)
	})

	t.Run("step target returns the step status while the workflow still runs", func(t *testing.T) {
		grpc, store, ctx := newRPC(t)
		store.On("WorkflowGetTree", currentPipeline).Return([]*model.Workflow{
			callerWorkflow,
			{ID: 11, Name: "auxiliaries", State: model.StatusRunning, Children: []*model.Step{
				{Name: "resolve-pins", State: model.StatusSuccess},
				{Name: "publish", State: model.StatusRunning},
			}},
		}, nil)

		result, err := grpc.WaitWorkflow(ctx, "10", "auxiliaries", "resolve-pins")
		require.NoError(t, err)
		assert.True(t, result.Found)
		assert.Equal(t, string(model.StatusSuccess), result.Status)
	})

	t.Run("unknown step target reports not found", func(t *testing.T) {
		grpc, store, ctx := newRPC(t)
		store.On("WorkflowGetTree", currentPipeline).Return([]*model.Workflow{
			callerWorkflow,
			{ID: 11, Name: "auxiliaries", State: model.StatusRunning, Children: []*model.Step{
				{Name: "publish", State: model.StatusRunning},
			}},
		}, nil)

		result, err := grpc.WaitWorkflow(ctx, "10", "auxiliaries", "missing")
		require.NoError(t, err)
		assert.False(t, result.Found)
	})

	t.Run("waits until the target is terminal", func(t *testing.T) {
		grpc, store, ctx := newRPC(t)
		store.On("WorkflowGetTree", currentPipeline).Once().Return([]*model.Workflow{
			callerWorkflow,
			{ID: 11, Name: "auxiliaries", State: model.StatusRunning},
		}, nil)
		store.On("WorkflowGetTree", currentPipeline).Return([]*model.Workflow{
			callerWorkflow,
			{ID: 11, Name: "auxiliaries", State: model.StatusFailure},
		}, nil)

		result, err := grpc.WaitWorkflow(ctx, "10", "auxiliaries", "")
		require.NoError(t, err)
		assert.True(t, result.Found)
		assert.Equal(t, string(model.StatusFailure), result.Status)
	})

	t.Run("matrix instances merge worst-wins and skipped stays visible", func(t *testing.T) {
		grpc, store, ctx := newRPC(t)
		store.On("WorkflowGetTree", currentPipeline).Return([]*model.Workflow{
			callerWorkflow,
			{ID: 11, Name: "auxiliaries", State: model.StatusSuccess},
			{ID: 12, Name: "auxiliaries", State: model.StatusSkipped},
		}, nil)

		result, err := grpc.WaitWorkflow(ctx, "10", "auxiliaries", "")
		require.NoError(t, err)
		assert.True(t, result.Found)
		// a skipped matrix sibling must not disappear behind a successful
		// one: any non-success outcome surfaces to the dependent step
		assert.Equal(t, string(model.StatusSkipped), result.Status)
	})

	t.Run("agent not owning the calling workflow is rejected", func(t *testing.T) {
		store := store_mocks.NewMockStore(t)
		store.On("AgentFind", int64(5)).Return(agent, nil)
		store.On("WorkflowLoad", int64(10)).Return(&model.Workflow{ID: 10, PipelineID: 100, AgentID: 6}, nil)
		store.On("GetPipeline", int64(100)).Return(currentPipeline, nil)
		store.On("GetRepo", int64(200)).Return(repo, nil)
		ctx := context.WithValue(t.Context(), agentIDKey, int64(5))
		grpc := &RPC{store: store}

		_, err := grpc.WaitWorkflow(ctx, "10", "auxiliaries", "")
		assert.ErrorIs(t, err, ErrAgentIllegalWorkflowAgentID)
	})
}
