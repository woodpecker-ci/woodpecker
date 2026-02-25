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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub"
	queue_mocks "go.woodpecker-ci.org/woodpecker/v3/server/queue/mocks"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

var fakePipeline = &model.Pipeline{
	ID:     2,
	Number: 2,
	Status: model.StatusSuccess,
}

func TestGetPipelines(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should get pipelines", func(t *testing.T) {
		pipelines := []*model.Pipeline{fakePipeline}

		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("GetPipelineList", mock.Anything, mock.Anything, mock.Anything).Return(pipelines, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)

		GetPipelines(c)

		mockStore.AssertCalled(t, "GetPipelineList", mock.Anything, mock.Anything, mock.Anything)
		assert.Equal(t, http.StatusOK, c.Writer.Status())
	})

	t.Run("should not parse pipeline filter", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request, _ = http.NewRequest(http.MethodDelete, "/?before=2023-01-16&after=2023-01-15", nil)

		GetPipelines(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})

	t.Run("should parse pipeline filter", func(t *testing.T) {
		pipelines := []*model.Pipeline{fakePipeline}

		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("GetPipelineList", mock.Anything, mock.Anything, mock.Anything).Return(pipelines, nil)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("store", mockStore)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/?2023-01-16T15:00:00Z&after=2023-01-15T15:00:00Z", nil)

		GetPipelines(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
	})

	t.Run("should parse pipeline filter with tz offset", func(t *testing.T) {
		pipelines := []*model.Pipeline{fakePipeline}

		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("GetPipelineList", mock.Anything, mock.Anything, mock.Anything).Return(pipelines, nil)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("store", mockStore)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/?before=2023-01-16T15:00:00%2B01:00&after=2023-01-15T15:00:00%2B01:00", nil)

		GetPipelines(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
	})

	t.Run("should filter pipelines by events", func(t *testing.T) {
		pipelines := []*model.Pipeline{fakePipeline}
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("GetPipelineList", mock.Anything, mock.Anything, mock.Anything).Return(pipelines, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Request, _ = http.NewRequest(http.MethodGet, "/?event=push,pull_request", nil)

		GetPipelines(c)

		mockStore.AssertCalled(t, "GetPipelineList", mock.Anything, mock.Anything, &model.PipelineFilter{
			Events: model.WebhookEventList{model.EventPush, model.EventPull},
		})
		assert.Equal(t, http.StatusOK, c.Writer.Status())
	})
}

func TestDeletePipeline(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should delete pipeline", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("GetPipelineNumber", mock.Anything, mock.Anything).Return(fakePipeline, nil)
		mockStore.On("DeletePipeline", mock.Anything).Return(nil)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("store", mockStore)
		c.Params = gin.Params{{Key: "number", Value: "2"}}

		DeletePipeline(c)

		mockStore.AssertCalled(t, "GetPipelineNumber", mock.Anything, mock.Anything)
		mockStore.AssertCalled(t, "DeletePipeline", mock.Anything)
		assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	})

	t.Run("should not delete without pipeline number", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		DeletePipeline(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})

	t.Run("should not delete pending", func(t *testing.T) {
		fakePipeline := *fakePipeline
		fakePipeline.Status = model.StatusPending

		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("GetPipelineNumber", mock.Anything, mock.Anything).Return(&fakePipeline, nil)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("store", mockStore)
		c.Params = gin.Params{{Key: "number", Value: "2"}}

		DeletePipeline(c)

		mockStore.AssertCalled(t, "GetPipelineNumber", mock.Anything, mock.Anything)
		mockStore.AssertNotCalled(t, "DeletePipeline", mock.Anything)
		assert.Equal(t, http.StatusUnprocessableEntity, c.Writer.Status())
	})
}

func TestGetPipelineMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	prevPipeline := &model.Pipeline{
		ID:     1,
		Number: 1,
		Status: model.StatusFailure,
	}

	fakeRepo := &model.Repo{ID: 1}

	mockForge := forge_mocks.NewMockForge(t)
	mockForge.On("Name").Return("mock")
	mockForge.On("URL").Return("https://codeberg.org")

	mockManager := manager_mocks.NewMockManager(t)
	mockManager.On("ForgeFromRepo", fakeRepo).Return(mockForge, nil)
	server.Config.Services.Manager = mockManager

	mockStore := store_mocks.NewMockStore(t)
	mockStore.On("GetPipelineNumber", mock.Anything, int64(2)).Return(fakePipeline, nil)
	mockStore.On("GetPipelineLastBefore", mock.Anything, mock.Anything, int64(2)).Return(prevPipeline, nil)

	t.Run("PipelineMetadata", func(t *testing.T) {
		t.Run("should get pipeline metadata", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "number", Value: "2"}}
			c.Set("store", mockStore)
			c.Set("forge", mockForge)
			c.Set("repo", fakeRepo)

			GetPipelineMetadata(c)

			assert.Equal(t, http.StatusOK, w.Code)

			var response metadata.Metadata
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, int64(1), response.Repo.ID)
			assert.Equal(t, int64(2), response.Curr.Number)
			assert.Equal(t, int64(1), response.Prev.Number)
		})

		t.Run("should return bad request for invalid pipeline number", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "number", Value: "invalid"}}

			GetPipelineMetadata(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("should return not found for non-existent pipeline", func(t *testing.T) {
			mockStore := store_mocks.NewMockStore(t)
			mockStore.On("GetPipelineNumber", mock.Anything, int64(3)).Return((*model.Pipeline)(nil), types.RecordNotExist)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "number", Value: "3"}}
			c.Set("store", mockStore)
			c.Set("repo", fakeRepo)

			GetPipelineMetadata(c)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})
}

func TestCancelPipeline(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should cancel running pipeline", func(t *testing.T) {
		runningPipeline := &model.Pipeline{
			ID:     2,
			Number: 2,
			Status: model.StatusRunning,
		}

		fakeRepo := &model.Repo{ID: 1}
		fakeUser := &model.User{Login: "testuser"}

		mockForge := forge_mocks.NewMockForge(t)
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("GetPipelineNumber", fakeRepo, int64(2)).Return(runningPipeline, nil)
		mockStore.On("WorkflowGetTree", mock.Anything).Return([]*model.Workflow{}, nil)
		mockStore.On("UpdatePipeline", mock.Anything).Return(nil)

		mockManager := manager_mocks.NewMockManager(t)
		mockManager.On("ForgeFromRepo", fakeRepo).Return(mockForge, nil)
		server.Config.Services.Manager = mockManager
		server.Config.Services.Pubsub = pubsub.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Set("repo", fakeRepo)
		c.Set("user", fakeUser)
		c.Params = gin.Params{{Key: "number", Value: "2"}}

		CancelPipeline(c)

		assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	})
}

func TestCreatePipeline(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. normal: config fetch succeeds (no error, returns config) -> success
	t.Run("normal workflow - config can be read", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockConfigService := config_mocks.NewMockService(t)
		mockSecretService := secret_mocks.NewMockService(t)
		mockRegistryService := registry_mocks.NewMockService(t)

		fakeRepo := &model.Repo{ID: 1, UserID: 1, FullName: "test/repo"}
		fakeUser := &model.User{ID: 1, Login: "testuser", Email: "test@example.com", Avatar: "avatar.png", Hash: "hash123"}
		fakeCommit := &model.Commit{SHA: "abc123", ForgeURL: "https://example.com/commit/abc123"}

		mockForge := forge_mocks.NewMockForge(t)
		mockForge.On("Name").Return("mock").Maybe()
		mockForge.On("URL").Return("https://example.com").Maybe()
		mockForge.On("BranchHead", mock.Anything, fakeUser, fakeRepo, "main").Return(fakeCommit, nil)
		mockForge.On("Netrc", fakeUser, fakeRepo).Return(&model.Netrc{
			Machine:  "example.com",
			Login:    "testuser",
			Password: "testpass",
		}, nil).Maybe()
		mockForge.On("Status", mock.Anything, fakeUser, fakeRepo, mock.Anything, mock.Anything).Return(nil).Maybe()

		mockSecretService.On("SecretListPipeline", fakeRepo, mock.Anything).Return([]*model.Secret{}, nil).Maybe()
		mockRegistryService.On("RegistryListPipeline", fakeRepo, mock.Anything).Return([]*model.Registry{}, nil).Maybe()

		mockManager := manager_mocks.NewMockManager(t)
		mockManager.On("ForgeFromRepo", fakeRepo).Return(mockForge, nil)
		mockManager.On("ConfigServiceFromRepo", fakeRepo).Return(mockConfigService)
		mockManager.On("SecretServiceFromRepo", fakeRepo).Return(mockSecretService).Maybe()
		mockManager.On("RegistryServiceFromRepo", fakeRepo).Return(mockRegistryService).Maybe()
		mockManager.On("EnvironmentService").Return(nil).Maybe()
		server.Config.Services.Manager = mockManager

		server.Config.Services.Pubsub = pubsub.New()
		mockQueue := queue_mocks.NewMockQueue(t)
		mockQueue.On("Push", mock.Anything, mock.Anything).Return(nil).Maybe()
		mockQueue.On("PushAtOnce", mock.Anything, mock.Anything).Return(nil).Maybe()
		server.Config.Services.Queue = mockQueue

		// mimic the valid config data
		configData := []*forge_types.FileMeta{
			{Name: ".woodpecker.yml", Data: []byte("when:\n  event: manual\nsteps:\n  test:\n    image: alpine:latest\n    commands:\n      - echo test")},
		}
		mockConfigService.On("Fetch", mock.Anything, mockForge, fakeUser, fakeRepo, mock.Anything, mock.Anything, false).Return(configData, nil)

		mockStore.On("GetUser", int64(1)).Return(fakeUser, nil)
		mockStore.On("CreatePipeline", mock.Anything).Return(nil)
		mockStore.On("GetPipelineLastBefore", fakeRepo, "main", mock.Anything).Return(nil, nil).Maybe()
		mockStore.On("ConfigPersist", mock.Anything).Return(&model.Config{ID: 1}, nil).Maybe()
		mockStore.On("ConfigFindIdentical", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
		mockStore.On("PipelineConfigCreate", mock.Anything).Return(nil).Maybe()
		mockStore.On("WorkflowsCreate", mock.Anything).Return(nil).Maybe()
		mockStore.On("UpdatePipeline", mock.Anything).Return(nil).Maybe()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Set("repo", fakeRepo)
		c.Set("user", fakeUser)

		c.Request, _ = http.NewRequest(http.MethodPost, "", io.NopCloser(bytes.NewBufferString(`{"branch": "main"}`)))
		c.Request.Header.Set("Content-Type", "application/json")

		CreatePipeline(c)

		// verify the config service was called successfully (no error, returns config)
		mockConfigService.AssertCalled(t, "Fetch", mock.Anything, mockForge, fakeUser, fakeRepo, mock.Anything, mock.Anything, false)
		mockForge.AssertCalled(t, "BranchHead", mock.Anything, fakeUser, fakeRepo, "main")
		mockStore.AssertCalled(t, "GetUser", int64(1))
		mockStore.AssertCalled(t, "CreatePipeline", mock.Anything)
	})

	// 2. abnormal with oldconfig: config fetch fails but returns config data (error + non-nil config) -> continues with fallback
	t.Run("abnormal workflow - cannot read config but has oldconfig", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockConfigService := config_mocks.NewMockService(t)
		mockSecretService := secret_mocks.NewMockService(t)
		mockRegistryService := registry_mocks.NewMockService(t)

		fakeRepo := &model.Repo{ID: 1, UserID: 1, FullName: "test/repo"}
		fakeUser := &model.User{ID: 1, Login: "testuser", Email: "test@example.com", Avatar: "avatar.png", Hash: "hash123"}
		fakeCommit := &model.Commit{SHA: "abc123", ForgeURL: "https://example.com/commit/abc123"}

		mockForge := forge_mocks.NewMockForge(t)
		mockForge.On("Name").Return("mock").Maybe()
		mockForge.On("URL").Return("https://example.com").Maybe()
		mockForge.On("BranchHead", mock.Anything, fakeUser, fakeRepo, "main").Return(fakeCommit, nil)
		// mock the netrc for parse config
		mockForge.On("Netrc", fakeUser, fakeRepo).Return(&model.Netrc{
			Machine:  "example.com",
			Login:    "testuser",
			Password: "testpass",
		}, nil).Maybe()

		mockForge.On("Status", mock.Anything, fakeUser, fakeRepo, mock.Anything, mock.Anything).Return(nil).Maybe()
		mockSecretService.On("SecretListPipeline", fakeRepo, mock.Anything).Return([]*model.Secret{}, nil).Maybe()
		mockRegistryService.On("RegistryListPipeline", fakeRepo, mock.Anything).Return([]*model.Registry{}, nil).Maybe()

		mockManager := manager_mocks.NewMockManager(t)
		mockManager.On("ForgeFromRepo", fakeRepo).Return(mockForge, nil)
		mockManager.On("ConfigServiceFromRepo", fakeRepo).Return(mockConfigService)
		mockManager.On("SecretServiceFromRepo", fakeRepo).Return(mockSecretService).Maybe()
		mockManager.On("RegistryServiceFromRepo", fakeRepo).Return(mockRegistryService).Maybe()
		mockManager.On("EnvironmentService").Return(nil).Maybe()
		server.Config.Services.Manager = mockManager

		server.Config.Services.Pubsub = pubsub.New()
		mockQueue := queue_mocks.NewMockQueue(t)
		mockQueue.On("Push", mock.Anything, mock.Anything).Return(nil).Maybe()
		mockQueue.On("PushAtOnce", mock.Anything, mock.Anything).Return(nil).Maybe()
		server.Config.Services.Queue = mockQueue

		// mimic the old config data
		oldConfigData := []*forge_types.FileMeta{
			{Name: ".woodpecker.yml", Data: []byte("when:\n  event: manual\nsteps:\n  test:\n    image: alpine:latest\n    commands:\n      - echo test")},
		}
		mockConfigService.On("Fetch", mock.Anything, mockForge, fakeUser, fakeRepo, mock.Anything, mock.Anything, false).Return(oldConfigData, http.ErrHandlerTimeout)

		mockStore.On("GetUser", int64(1)).Return(fakeUser, nil)
		mockStore.On("CreatePipeline", mock.Anything).Return(nil)
		mockStore.On("GetPipelineLastBefore", fakeRepo, "main", mock.Anything).Return(nil, nil).Maybe()
		mockStore.On("ConfigPersist", mock.Anything).Return(&model.Config{ID: 1}, nil).Maybe()
		mockStore.On("ConfigFindIdentical", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
		mockStore.On("PipelineConfigCreate", mock.Anything).Return(nil).Maybe()
		mockStore.On("WorkflowsCreate", mock.Anything).Return(nil).Maybe()
		mockStore.On("UpdatePipeline", mock.Anything).Return(nil).Maybe()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Set("repo", fakeRepo)
		c.Set("user", fakeUser)

		c.Request, _ = http.NewRequest(http.MethodPost, "", io.NopCloser(bytes.NewBufferString(`{"branch": "main"}`)))
		c.Request.Header.Set("Content-Type", "application/json")

		CreatePipeline(c)

		// verify the config service returned error + old config (fallback scenario)
		mockConfigService.AssertCalled(t, "Fetch", mock.Anything, mockForge, fakeUser, fakeRepo, mock.Anything, mock.Anything, false)
		mockStore.AssertCalled(t, "GetUser", int64(1))
		mockStore.AssertCalled(t, "CreatePipeline", mock.Anything)
	})

	// 3. abnormal without oldconfig: config fetch fails without config data (error + nil config) -> fails immediately
	t.Run("abnormal workflow - cannot read config and no oldconfig", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockConfigService := config_mocks.NewMockService(t)

		fakeRepo := &model.Repo{ID: 1, UserID: 1, FullName: "test/repo"}
		fakeUser := &model.User{ID: 1, Login: "testuser", Email: "test@example.com", Avatar: "avatar.png", Hash: "hash123"}
		fakeCommit := &model.Commit{SHA: "abc123", ForgeURL: "https://example.com/commit/abc123"}

		mockForge := forge_mocks.NewMockForge(t)
		mockForge.On("BranchHead", mock.Anything, fakeUser, fakeRepo, "main").Return(fakeCommit, nil)
		mockForge.On("Netrc", fakeUser, fakeRepo).Return(nil, nil).Maybe()
		mockForge.On("Status", mock.Anything, fakeUser, fakeRepo, mock.Anything, mock.Anything).Return(nil).Maybe()

		mockManager := manager_mocks.NewMockManager(t)
		mockManager.On("ForgeFromRepo", fakeRepo).Return(mockForge, nil)
		mockManager.On("ConfigServiceFromRepo", fakeRepo).Return(mockConfigService)
		server.Config.Services.Manager = mockManager
		server.Config.Services.Pubsub = pubsub.New()

		// return nil config with error
		mockConfigService.On("Fetch", mock.Anything, mockForge, fakeUser, fakeRepo, mock.Anything, mock.Anything, false).Return(nil, http.ErrHandlerTimeout)

		mockStore.On("GetUser", int64(1)).Return(fakeUser, nil)
		mockStore.On("CreatePipeline", mock.Anything).Return(nil)
		mockStore.On("UpdatePipeline", mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Set("repo", fakeRepo)
		c.Set("user", fakeUser)

		c.Request, _ = http.NewRequest(http.MethodPost, "", io.NopCloser(bytes.NewBufferString(`{"branch": "main"}`)))
		c.Request.Header.Set("Content-Type", "application/json")

		CreatePipeline(c)

		// verify the config service returned error without any config data
		mockConfigService.AssertCalled(t, "Fetch", mock.Anything, mockForge, fakeUser, fakeRepo, mock.Anything, mock.Anything, false)
		mockStore.AssertCalled(t, "GetUser", int64(1))
		mockStore.AssertCalled(t, "CreatePipeline", mock.Anything)
		mockStore.AssertCalled(t, "UpdatePipeline", mock.Anything)
	})
}
