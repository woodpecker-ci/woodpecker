// Copyright 2024 Woodpecker Authors
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

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
	queue_mocks "go.woodpecker-ci.org/woodpecker/v3/server/queue/mocks"
	mocks_manager "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

var fakeAgent = &model.Agent{
	ID:         1,
	Name:       "test-agent",
	OwnerID:    1,
	NoSchedule: false,
}

func TestGetAgents(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should get agents", func(t *testing.T) {
		agents := []*model.Agent{fakeAgent}

		mockStore := store_mocks.NewStore(t)
		mockStore.On("AgentList", mock.Anything).Return(agents, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)

		GetAgents(c)
		c.Writer.WriteHeaderNow()

		mockStore.AssertCalled(t, "AgentList", mock.Anything)
		assert.Equal(t, http.StatusOK, w.Code)

		var response []*model.Agent
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, agents, response)
	})
}

func TestGetAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should get agent", func(t *testing.T) {
		mockStore := store_mocks.NewStore(t)
		mockStore.On("AgentFind", int64(1)).Return(fakeAgent, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Params = gin.Params{{Key: "agent_id", Value: "1"}}

		GetAgent(c)
		c.Writer.WriteHeaderNow()

		mockStore.AssertCalled(t, "AgentFind", int64(1))
		assert.Equal(t, http.StatusOK, w.Code)

		var response model.Agent
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, fakeAgent, &response)
	})

	t.Run("should return bad request for invalid agent id", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "agent_id", Value: "invalid"}}

		GetAgent(c)
		c.Writer.WriteHeaderNow()

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return not found for non-existent agent", func(t *testing.T) {
		mockStore := store_mocks.NewStore(t)
		mockStore.On("AgentFind", int64(2)).Return((*model.Agent)(nil), types.RecordNotExist)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Params = gin.Params{{Key: "agent_id", Value: "2"}}

		GetAgent(c)
		c.Writer.WriteHeaderNow()

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestPatchAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should update agent", func(t *testing.T) {
		updatedAgent := *fakeAgent
		updatedAgent.Name = "updated-agent"

		mockStore := store_mocks.NewStore(t)
		mockStore.On("AgentFind", int64(1)).Return(fakeAgent, nil)
		mockStore.On("AgentUpdate", mock.AnythingOfType("*model.Agent")).Return(nil)

		mockManager := mocks_manager.NewManager(t)
		server.Config.Services.Manager = mockManager

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Params = gin.Params{{Key: "agent_id", Value: "1"}}
		c.Request, _ = http.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"name":"updated-agent"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		PatchAgent(c)
		c.Writer.WriteHeaderNow()

		mockStore.AssertCalled(t, "AgentFind", int64(1))
		mockStore.AssertCalled(t, "AgentUpdate", mock.AnythingOfType("*model.Agent"))
		assert.Equal(t, http.StatusOK, w.Code)

		var response model.Agent
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "updated-agent", response.Name)
	})
}

func TestPostAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should create agent", func(t *testing.T) {
		newAgent := &model.Agent{
			Name:       "new-agent",
			NoSchedule: false,
		}

		mockStore := store_mocks.NewStore(t)
		mockStore.On("AgentCreate", mock.AnythingOfType("*model.Agent")).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Set("user", &model.User{ID: 1})
		c.Request, _ = http.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"new-agent"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		PostAgent(c)
		c.Writer.WriteHeaderNow()

		mockStore.AssertCalled(t, "AgentCreate", mock.AnythingOfType("*model.Agent"))
		assert.Equal(t, http.StatusOK, w.Code)

		var response model.Agent
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, newAgent.Name, response.Name)
		assert.NotEmpty(t, response.Token)
	})
}

func TestDeleteAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should delete agent", func(t *testing.T) {
		mockStore := store_mocks.NewStore(t)
		mockStore.On("AgentFind", int64(1)).Return(fakeAgent, nil)
		mockStore.On("AgentDelete", mock.AnythingOfType("*model.Agent")).Return(nil)

		mockManager := mocks_manager.NewManager(t)
		server.Config.Services.Manager = mockManager

		mockQueue := queue_mocks.NewQueue(t)
		mockQueue.On("Info", mock.Anything).Return(queue.InfoT{})
		mockQueue.On("KickAgentWorkers", int64(1)).Return()
		server.Config.Services.Queue = mockQueue

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Params = gin.Params{{Key: "agent_id", Value: "1"}}

		DeleteAgent(c)
		c.Writer.WriteHeaderNow()

		mockStore.AssertCalled(t, "AgentFind", int64(1))
		mockStore.AssertCalled(t, "AgentDelete", mock.AnythingOfType("*model.Agent"))
		mockQueue.AssertCalled(t, "KickAgentWorkers", int64(1))
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("should not delete agent with running tasks", func(t *testing.T) {
		mockStore := store_mocks.NewStore(t)
		mockStore.On("AgentFind", int64(1)).Return(fakeAgent, nil)

		mockManager := mocks_manager.NewManager(t)
		server.Config.Services.Manager = mockManager

		mockQueue := queue_mocks.NewQueue(t)
		mockQueue.On("Info", mock.Anything).Return(queue.InfoT{
			Running: []*model.Task{{AgentID: 1}},
		})
		server.Config.Services.Queue = mockQueue

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Params = gin.Params{{Key: "agent_id", Value: "1"}}

		DeleteAgent(c)
		c.Writer.WriteHeaderNow()

		mockStore.AssertCalled(t, "AgentFind", int64(1))
		mockStore.AssertNotCalled(t, "AgentDelete", mock.Anything)
		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestPostOrgAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("create org agent should succeed", func(t *testing.T) {
		mockStore := store_mocks.NewStore(t)
		mockStore.On("AgentCreate", mock.AnythingOfType("*model.Agent")).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)

		// Set up a non-admin user
		c.Set("user", &model.User{
			ID:    1,
			Admin: false,
		})

		c.Params = gin.Params{{Key: "org_id", Value: "1"}}
		c.Request, _ = http.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"new-agent"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		PostOrgAgent(c)
		c.Writer.WriteHeaderNow()

		assert.Equal(t, http.StatusOK, w.Code)

		// Ensure an agent was created
		mockStore.AssertCalled(t, "AgentCreate", mock.AnythingOfType("*model.Agent"))
	})
}
