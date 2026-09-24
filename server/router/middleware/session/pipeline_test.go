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

//go:build test

package session

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func TestSetPipeline(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &model.Repo{ID: 1}

	newCtx := func(t *testing.T, pipelineNumber string) (*gin.Context, *httptest.ResponseRecorder, *store_mocks.MockStore) {
		t.Helper()
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		mockStore := store_mocks.NewMockStore(t)
		c.Set("store", mockStore)
		c.Set("repo", repo)
		c.Params = gin.Params{{Key: "pipeline_number", Value: pipelineNumber}}
		return c, rec, mockStore
	}

	t.Run("should resolve the pipeline of the repo", func(t *testing.T) {
		pipeline := &model.Pipeline{ID: 3, Number: 2}
		c, _, mockStore := newCtx(t, "2")
		mockStore.On("GetPipelineNumber", repo, int64(2)).Return(pipeline, nil)

		SetPipeline()(c)

		require.False(t, c.IsAborted())
		assert.Equal(t, pipeline, Pipeline(c))
	})

	t.Run("should reject a non numeric pipeline number", func(t *testing.T) {
		c, rec, _ := newCtx(t, "latest")

		SetPipeline()(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Nil(t, Pipeline(c))
	})

	t.Run("should return not found for an unknown pipeline", func(t *testing.T) {
		c, rec, mockStore := newCtx(t, "3")
		mockStore.On("GetPipelineNumber", repo, int64(3)).Return((*model.Pipeline)(nil), types.ErrRecordNotExist)

		SetPipeline()(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Nil(t, Pipeline(c))
	})

	t.Run("should return internal server error on store failure", func(t *testing.T) {
		c, rec, mockStore := newCtx(t, "4")
		mockStore.On("GetPipelineNumber", repo, int64(4)).Return((*model.Pipeline)(nil), errors.New("database on fire"))

		SetPipeline()(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Nil(t, Pipeline(c))
	})
}
