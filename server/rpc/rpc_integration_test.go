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

package rpc

import (
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/logging"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub"
	queueMocks "go.woodpecker-ci.org/woodpecker/v3/server/queue/mocks"
	logMocks "go.woodpecker-ci.org/woodpecker/v3/server/services/log/mocks"
	storeMocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

// newTestRPC creates an RPC instance with common test infrastructure.
func newTestRPC(t *testing.T, mockStore *storeMocks.MockStore) RPC {
	t.Helper()

	pipelineTime := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "woodpecker_test",
		Name:      "pipeline_time_" + t.Name(),
	}, []string{"repo", "branch", "status", "pipeline"})
	pipelineCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "woodpecker_test",
		Name:      "pipeline_count_" + t.Name(),
	}, []string{"repo", "branch", "status", "pipeline"})

	return RPC{
		store:         mockStore,
		pubsub:        pubsub.New(),
		logger:        logging.New(),
		pipelineTime:  pipelineTime,
		pipelineCount: pipelineCount,
	}
}

// defaultAgent returns a system agent (OrgID=-1) that can access any repo.
func defaultAgent() *model.Agent {
	return &model.Agent{
		ID:    1,
		Name:  "test-agent",
		OrgID: model.IDNotSet,
	}
}

// orgAgent returns an agent scoped to a specific org.
func orgAgent(orgID int64) *model.Agent {
	return &model.Agent{
		ID:    2,
		Name:  "org-agent",
		OrgID: orgID,
	}
}

func defaultRepo() *model.Repo {
	return &model.Repo{
		ID:       10,
		OrgID:    100,
		FullName: "test-org/test-repo",
	}
}

func defaultPipeline(status model.StatusValue) *model.Pipeline {
	return &model.Pipeline{
		ID:     20,
		RepoID: 10,
		Status: status,
		Branch: "main",
	}
}

func defaultWorkflow(state model.StatusValue) *model.Workflow {
	return &model.Workflow{
		ID:         30,
		PipelineID: 20,
		State:      state,
		Name:       "test-workflow",
	}
}

func defaultStep(state model.StatusValue) *model.Step {
	return &model.Step{
		ID:         40,
		UUID:       "step-uuid-123",
		PipelineID: 20,
		State:      state,
	}
}

func TestRPCUpdate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		mockLogStore := logMocks.NewMockService(t)
		origLogStore := server.Config.Services.LogStore
		server.Config.Services.LogStore = mockLogStore
		t.Cleanup(func() { server.Config.Services.LogStore = origLogStore })

		agent := defaultAgent()
		repo := defaultRepo()
		pipeline := defaultPipeline(model.StatusRunning)
		workflow := defaultWorkflow(model.StatusRunning)
		step := defaultStep(model.StatusRunning)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("StepByUUID", "step-uuid-123").Return(step, nil)
		mockStore.On("GetRepo", int64(10)).Return(repo, nil)
		// pipeline.UpdateStepStatus calls StepUpdate
		mockStore.On("StepUpdate", mock.Anything).Return(nil)
		mockStore.On("WorkflowGetTree", mock.Anything).Return([]*model.Workflow{workflow}, nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Update(ctx, "30", rpc.StepState{
			StepUUID: "step-uuid-123",
			Started:  100,
			Exited:   false,
		})
		assert.NoError(t, err)
	})

	t.Run("reject pipeline already succeeded", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusSuccess)
		workflow := defaultWorkflow(model.StatusRunning)
		step := defaultStep(model.StatusRunning)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("StepByUUID", "step-uuid-123").Return(step, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Update(ctx, "30", rpc.StepState{StepUUID: "step-uuid-123"})
		assert.ErrorIs(t, err, ErrAgentIllegalPipelineWorkflowReRunStateChange)
	})

	t.Run("reject pipeline already failed", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusFailure)
		workflow := defaultWorkflow(model.StatusRunning)
		step := defaultStep(model.StatusRunning)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("StepByUUID", "step-uuid-123").Return(step, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Update(ctx, "30", rpc.StepState{StepUUID: "step-uuid-123"})
		assert.ErrorIs(t, err, ErrAgentIllegalPipelineWorkflowReRunStateChange)
	})

	t.Run("reject pipeline blocked", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusBlocked)
		workflow := defaultWorkflow(model.StatusRunning)
		step := defaultStep(model.StatusRunning)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("StepByUUID", "step-uuid-123").Return(step, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Update(ctx, "30", rpc.StepState{StepUUID: "step-uuid-123"})
		assert.ErrorIs(t, err, ErrAgentIllegalPipelineWorkflowRun)
	})

	t.Run("reject workflow already finished", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusRunning)
		workflow := defaultWorkflow(model.StatusSuccess) // finished
		step := defaultStep(model.StatusRunning)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("StepByUUID", "step-uuid-123").Return(step, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Update(ctx, "30", rpc.StepState{StepUUID: "step-uuid-123"})
		assert.ErrorIs(t, err, ErrAgentIllegalWorkflowReRunStateChange)
	})

	t.Run("reject step already finished", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusRunning)
		workflow := defaultWorkflow(model.StatusRunning)
		step := defaultStep(model.StatusSuccess) // finished

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("StepByUUID", "step-uuid-123").Return(step, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Update(ctx, "30", rpc.StepState{StepUUID: "step-uuid-123"})
		assert.ErrorIs(t, err, ErrAgentIllegalStepReRunStateChange)
	})

	t.Run("reject step belongs to different pipeline", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusRunning)
		workflow := defaultWorkflow(model.StatusRunning)
		step := &model.Step{
			ID:         40,
			UUID:       "step-uuid-123",
			PipelineID: 999, // different pipeline!
			State:      model.StatusRunning,
		}

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("StepByUUID", "step-uuid-123").Return(step, nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Update(ctx, "30", rpc.StepState{StepUUID: "step-uuid-123"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not belong to current pipeline")
	})

	t.Run("reject agent from wrong org", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := orgAgent(999) // org 999
		repo := defaultRepo()  // org 100
		pipeline := defaultPipeline(model.StatusRunning)
		workflow := defaultWorkflow(model.StatusRunning)
		step := defaultStep(model.StatusRunning)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("AgentFind", int64(2)).Return(agent, nil)
		mockStore.On("StepByUUID", "step-uuid-123").Return(step, nil)
		mockStore.On("GetRepo", int64(10)).Return(repo, nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "2"))

		err := rpcInst.Update(ctx, "30", rpc.StepState{StepUUID: "step-uuid-123"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed to interact")
	})

	t.Run("reject invalid workflow ID", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Update(ctx, "not-a-number", rpc.StepState{StepUUID: "step-uuid-123"})
		assert.Error(t, err)
	})

	t.Run("reject nonexistent workflow", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("WorkflowLoad", int64(999)).Return(nil, errors.New("not found"))

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Update(ctx, "999", rpc.StepState{StepUUID: "step-uuid-123"})
		assert.Error(t, err)
	})

	t.Run("reject nonexistent step UUID", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusRunning)
		workflow := defaultWorkflow(model.StatusRunning)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("StepByUUID", "nonexistent").Return(nil, errors.New("not found"))

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Update(ctx, "30", rpc.StepState{StepUUID: "nonexistent"})
		assert.Error(t, err)
	})

	t.Run("reject missing agent metadata", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		pipeline := defaultPipeline(model.StatusRunning)
		workflow := defaultWorkflow(model.StatusRunning)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)

		rpcInst := newTestRPC(t, mockStore)
		// no agent_id in metadata
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs())

		err := rpcInst.Update(ctx, "30", rpc.StepState{StepUUID: "step-uuid-123"})
		assert.Error(t, err)
	})
}

func TestRPCInit(t *testing.T) {
	t.Run("happy path - pending pipeline", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		repo := defaultRepo()
		pipeline := defaultPipeline(model.StatusPending)
		workflow := defaultWorkflow(model.StatusPending)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(repo, nil)
		// pipeline.UpdateToStatusRunning -> UpdatePipeline
		mockStore.On("UpdatePipeline", mock.Anything).Return(nil)
		// updateForgeStatus -> GetUser returns error so forge interaction is skipped
		mockStore.On("GetUser", mock.Anything).Return(nil, errors.New("user not found"))
		// pipeline.UpdateWorkflowStatusToRunning -> WorkflowUpdate
		mockStore.On("WorkflowUpdate", mock.Anything).Return(nil)
		// pubsub deferred -> WorkflowGetTree
		mockStore.On("WorkflowGetTree", mock.Anything).Return([]*model.Workflow{workflow}, nil)
		// updateAgentLastWork -> AgentUpdate
		mockStore.On("AgentUpdate", mock.Anything).Return(nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Init(ctx, "30", rpc.WorkflowState{Started: 100})
		assert.NoError(t, err)
	})

	t.Run("happy path - already running pipeline", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		repo := defaultRepo()
		pipeline := defaultPipeline(model.StatusRunning) // another workflow already started it
		workflow := defaultWorkflow(model.StatusPending)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(repo, nil)
		// updateForgeStatus -> GetUser returns error so forge interaction is skipped
		mockStore.On("GetUser", mock.Anything).Return(nil, errors.New("user not found"))
		mockStore.On("WorkflowUpdate", mock.Anything).Return(nil)
		mockStore.On("WorkflowGetTree", mock.Anything).Return([]*model.Workflow{workflow}, nil)
		mockStore.On("AgentUpdate", mock.Anything).Return(nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Init(ctx, "30", rpc.WorkflowState{Started: 100})
		assert.NoError(t, err)
	})

	t.Run("reject pipeline already succeeded", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusSuccess)
		workflow := defaultWorkflow(model.StatusPending)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Init(ctx, "30", rpc.WorkflowState{Started: 100})
		assert.ErrorIs(t, err, ErrAgentIllegalPipelineWorkflowReRunStateChange)
	})

	t.Run("reject pipeline blocked", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusBlocked)
		workflow := defaultWorkflow(model.StatusPending)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Init(ctx, "30", rpc.WorkflowState{Started: 100})
		assert.ErrorIs(t, err, ErrAgentIllegalPipelineWorkflowRun)
	})

	t.Run("reject workflow already finished", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusRunning)
		workflow := defaultWorkflow(model.StatusSuccess)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Init(ctx, "30", rpc.WorkflowState{Started: 100})
		assert.ErrorIs(t, err, ErrAgentIllegalWorkflowReRunStateChange)
	})

	t.Run("reject workflow blocked", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusRunning)
		workflow := defaultWorkflow(model.StatusBlocked)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Init(ctx, "30", rpc.WorkflowState{Started: 100})
		assert.ErrorIs(t, err, ErrAgentIllegalWorkflowRun)
	})

	t.Run("reject agent wrong org", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := orgAgent(999)
		pipeline := defaultPipeline(model.StatusRunning)
		workflow := defaultWorkflow(model.StatusPending)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("AgentFind", int64(2)).Return(agent, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "2"))

		err := rpcInst.Init(ctx, "30", rpc.WorkflowState{Started: 100})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed to interact")
	})

	t.Run("reject invalid workflow ID", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Init(ctx, "not-a-number", rpc.WorkflowState{})
		assert.Error(t, err)
	})
}

func TestRPCDone(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		mockQueue := queueMocks.NewMockQueue(t)
		mockLogStore := logMocks.NewMockService(t)
		origLogStore := server.Config.Services.LogStore
		server.Config.Services.LogStore = mockLogStore
		t.Cleanup(func() { server.Config.Services.LogStore = origLogStore })

		agent := defaultAgent()
		repo := defaultRepo()
		pipeline := defaultPipeline(model.StatusRunning)
		workflow := defaultWorkflow(model.StatusRunning)
		workflow.Children = []*model.Step{}

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("StepListFromWorkflowFind", mock.Anything).Return([]*model.Step{}, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(repo, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("WorkflowUpdate", mock.Anything).Return(nil)
		mockStore.On("WorkflowGetTree", mock.Anything).Return([]*model.Workflow{}, nil)
		mockStore.On("UpdatePipeline", mock.Anything).Return(nil)
		mockStore.On("GetUser", mock.Anything).Return(nil, errors.New("user not found"))
		mockStore.On("AgentUpdate", mock.Anything).Return(nil)
		mockQueue.On("Done", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		rpcInst := newTestRPC(t, mockStore)
		rpcInst.queue = mockQueue
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Done(ctx, "30", rpc.WorkflowState{Started: 100, Finished: 200})
		assert.NoError(t, err)
	})

	t.Run("reject pipeline already succeeded", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusSuccess)
		workflow := defaultWorkflow(model.StatusRunning)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("StepListFromWorkflowFind", mock.Anything).Return([]*model.Step{}, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Done(ctx, "30", rpc.WorkflowState{Finished: 200})
		assert.ErrorIs(t, err, ErrAgentIllegalPipelineWorkflowReRunStateChange)
	})

	t.Run("reject pipeline killed", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusKilled)
		workflow := defaultWorkflow(model.StatusRunning)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("StepListFromWorkflowFind", mock.Anything).Return([]*model.Step{}, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Done(ctx, "30", rpc.WorkflowState{Finished: 200})
		assert.ErrorIs(t, err, ErrAgentIllegalPipelineWorkflowReRunStateChange)
	})

	t.Run("reject pipeline blocked", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusBlocked)
		workflow := defaultWorkflow(model.StatusRunning)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("StepListFromWorkflowFind", mock.Anything).Return([]*model.Step{}, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Done(ctx, "30", rpc.WorkflowState{Finished: 200})
		assert.ErrorIs(t, err, ErrAgentIllegalPipelineWorkflowRun)
	})

	t.Run("reject workflow already finished", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusRunning)
		workflow := defaultWorkflow(model.StatusSuccess)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("StepListFromWorkflowFind", mock.Anything).Return([]*model.Step{}, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Done(ctx, "30", rpc.WorkflowState{Finished: 200})
		assert.ErrorIs(t, err, ErrAgentIllegalWorkflowReRunStateChange)
	})

	t.Run("reject agent wrong org", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := orgAgent(999)
		pipeline := defaultPipeline(model.StatusRunning)
		workflow := defaultWorkflow(model.StatusRunning)

		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("StepListFromWorkflowFind", mock.Anything).Return([]*model.Step{}, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)
		mockStore.On("AgentFind", int64(2)).Return(agent, nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "2"))

		err := rpcInst.Done(ctx, "30", rpc.WorkflowState{Finished: 200})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed to interact")
	})

	t.Run("reject invalid workflow ID", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Done(ctx, "invalid", rpc.WorkflowState{})
		assert.Error(t, err)
	})
}

func TestRPCLog(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		mockLogStore := logMocks.NewMockService(t)
		origLogStore := server.Config.Services.LogStore
		server.Config.Services.LogStore = mockLogStore
		t.Cleanup(func() { server.Config.Services.LogStore = origLogStore })

		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusRunning)
		step := defaultStep(model.StatusRunning)

		mockStore.On("StepByUUID", "step-uuid-123").Return(step, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)
		mockStore.On("AgentUpdate", mock.Anything).Return(nil)
		mockLogStore.On("LogAppend", mock.Anything, mock.Anything).Return(nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		entries := []*rpc.LogEntry{
			{StepUUID: "step-uuid-123", Line: 0, Data: []byte("hello")},
			{StepUUID: "step-uuid-123", Line: 1, Data: []byte("world")},
		}
		err := rpcInst.Log(ctx, "step-uuid-123", entries)
		assert.NoError(t, err)
	})

	t.Run("reject pipeline already finished", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusSuccess)
		step := defaultStep(model.StatusRunning)

		mockStore.On("StepByUUID", "step-uuid-123").Return(step, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Log(ctx, "step-uuid-123", []*rpc.LogEntry{
			{StepUUID: "step-uuid-123", Data: []byte("test")},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can not alter logs")
		assert.ErrorIs(t, err, ErrAgentIllegalPipelineWorkflowReRunStateChange)
	})

	t.Run("reject step pending (not running)", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusRunning)
		step := defaultStep(model.StatusPending)

		mockStore.On("StepByUUID", "step-uuid-123").Return(step, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Log(ctx, "step-uuid-123", []*rpc.LogEntry{
			{StepUUID: "step-uuid-123", Data: []byte("test")},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can not alter logs")
		assert.ErrorIs(t, err, ErrAgentIllegalLogStreaming)
	})

	t.Run("reject step already succeeded", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusRunning)
		step := defaultStep(model.StatusSuccess)

		mockStore.On("StepByUUID", "step-uuid-123").Return(step, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Log(ctx, "step-uuid-123", []*rpc.LogEntry{
			{StepUUID: "step-uuid-123", Data: []byte("test")},
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAgentIllegalLogStreaming)
	})

	t.Run("reject step killed", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusRunning)
		step := defaultStep(model.StatusKilled)

		mockStore.On("StepByUUID", "step-uuid-123").Return(step, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Log(ctx, "step-uuid-123", []*rpc.LogEntry{
			{StepUUID: "step-uuid-123", Data: []byte("test")},
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAgentIllegalLogStreaming)
	})

	t.Run("reject mismatched step UUID in log entry", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		mockLogStore := logMocks.NewMockService(t)
		origLogStore := server.Config.Services.LogStore
		server.Config.Services.LogStore = mockLogStore
		t.Cleanup(func() { server.Config.Services.LogStore = origLogStore })

		agent := defaultAgent()
		pipeline := defaultPipeline(model.StatusRunning)
		step := defaultStep(model.StatusRunning)

		mockStore.On("StepByUUID", "step-uuid-123").Return(step, nil)
		mockStore.On("AgentFind", int64(1)).Return(agent, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)
		mockStore.On("AgentUpdate", mock.Anything).Return(nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		// one entry has a different UUID — rogue agent trying to inject logs into another step
		entries := []*rpc.LogEntry{
			{StepUUID: "step-uuid-123", Line: 0, Data: []byte("ok")},
			{StepUUID: "DIFFERENT-UUID", Line: 1, Data: []byte("injected!")},
		}
		err := rpcInst.Log(ctx, "step-uuid-123", entries)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected step UUID")
	})

	t.Run("reject agent wrong org", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := orgAgent(999)
		pipeline := defaultPipeline(model.StatusRunning)
		step := defaultStep(model.StatusRunning)

		mockStore.On("StepByUUID", "step-uuid-123").Return(step, nil)
		mockStore.On("AgentFind", int64(2)).Return(agent, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "2"))

		err := rpcInst.Log(ctx, "step-uuid-123", []*rpc.LogEntry{
			{StepUUID: "step-uuid-123", Data: []byte("test")},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed to interact")
	})

	t.Run("reject nonexistent step UUID", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("StepByUUID", "nonexistent").Return(nil, errors.New("not found"))

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "1"))

		err := rpcInst.Log(ctx, "nonexistent", []*rpc.LogEntry{
			{StepUUID: "nonexistent", Data: []byte("test")},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not find step")
	})
}

func TestRPCExtend(t *testing.T) {
	t.Run("reject agent wrong org via permission check", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := orgAgent(999)
		workflow := defaultWorkflow(model.StatusRunning)
		pipeline := defaultPipeline(model.StatusRunning)

		mockStore.On("AgentFind", int64(2)).Return(agent, nil)
		mockStore.On("AgentUpdate", mock.Anything).Return(nil)
		// checkAgentPermissionByWorkflow with nil pipeline/repo -> loads from store
		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "2"))

		err := rpcInst.Extend(ctx, "30")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed to interact")
	})
}

func TestRPCWait(t *testing.T) {
	t.Run("reject agent wrong org", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		agent := orgAgent(999)
		workflow := defaultWorkflow(model.StatusRunning)
		pipeline := defaultPipeline(model.StatusRunning)

		mockStore.On("AgentFind", int64(2)).Return(agent, nil)
		// checkAgentPermissionByWorkflow loads from store
		mockStore.On("WorkflowLoad", int64(30)).Return(workflow, nil)
		mockStore.On("GetPipeline", int64(20)).Return(pipeline, nil)
		mockStore.On("GetRepo", int64(10)).Return(defaultRepo(), nil)

		rpcInst := newTestRPC(t, mockStore)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("agent_id", "2"))

		_, err := rpcInst.Wait(ctx, "30")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed to interact")
	})
}
