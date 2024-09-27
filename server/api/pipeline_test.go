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
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v2/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v2/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	mocks_manager "go.woodpecker-ci.org/woodpecker/v2/server/services/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v2/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"
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

		mockStore := store_mocks.NewStore(t)
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

		mockStore := store_mocks.NewStore(t)
		mockStore.On("GetPipelineList", mock.Anything, mock.Anything, mock.Anything).Return(pipelines, nil)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("store", mockStore)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/?2023-01-16T15:00:00Z&after=2023-01-15T15:00:00Z", nil)

		GetPipelines(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
	})

	t.Run("should parse pipeline filter with tz offset", func(t *testing.T) {
		pipelines := []*model.Pipeline{fakePipeline}

		mockStore := store_mocks.NewStore(t)
		mockStore.On("GetPipelineList", mock.Anything, mock.Anything, mock.Anything).Return(pipelines, nil)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("store", mockStore)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/?before=2023-01-16T15:00:00%2B01:00&after=2023-01-15T15:00:00%2B01:00", nil)

		GetPipelines(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
	})
}

func TestDeletePipeline(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should delete pipeline", func(t *testing.T) {
		mockStore := store_mocks.NewStore(t)
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

		mockStore := store_mocks.NewStore(t)
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

	mockForge := forge_mocks.NewForge(t)
	mockForge.On("Name").Return("mock")
	mockForge.On("URL").Return("https://codeberg.org")

	mockManager := mocks_manager.NewManager(t)
	mockManager.On("ForgeFromRepo", fakeRepo).Return(mockForge, nil)
	server.Config.Services.Manager = mockManager

	mockStore := store_mocks.NewStore(t)
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
			mockStore := store_mocks.NewStore(t)
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
