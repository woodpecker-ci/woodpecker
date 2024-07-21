package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/franela/goblin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/mocks"
)

var fakePipeline = &model.Pipeline{
	Status: model.StatusSuccess,
}

func TestGetPipelines(t *testing.T) {
	gin.SetMode(gin.TestMode)

	g := goblin.Goblin(t)
	g.Describe("Pipeline", func() {
		g.It("should get pipelines", func() {
			pipelines := []*model.Pipeline{fakePipeline}

			mockStore := mocks.NewStore(t)
			mockStore.On("GetPipelineList", mock.Anything, mock.Anything, mock.Anything).Return(pipelines, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("store", mockStore)

			GetPipelines(c)

			mockStore.AssertCalled(t, "GetPipelineList", mock.Anything, mock.Anything, mock.Anything)
			assert.Equal(t, http.StatusOK, c.Writer.Status())
		})

		g.It("should not parse pipeline filter", func() {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request, _ = http.NewRequest(http.MethodDelete, "/?before=2023-01-16&after=2023-01-15", nil)

			GetPipelines(c)

			assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
		})

		g.It("should parse pipeline filter", func() {
			pipelines := []*model.Pipeline{fakePipeline}

			mockStore := mocks.NewStore(t)
			mockStore.On("GetPipelineList", mock.Anything, mock.Anything, mock.Anything).Return(pipelines, nil)

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Set("store", mockStore)
			c.Request, _ = http.NewRequest(http.MethodDelete, "/?2023-01-16T15:00:00Z&after=2023-01-15T15:00:00Z", nil)

			GetPipelines(c)

			assert.Equal(t, http.StatusOK, c.Writer.Status())
		})

		g.It("should parse pipeline filter with tz offset", func() {
			pipelines := []*model.Pipeline{fakePipeline}

			mockStore := mocks.NewStore(t)
			mockStore.On("GetPipelineList", mock.Anything, mock.Anything, mock.Anything).Return(pipelines, nil)

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Set("store", mockStore)
			c.Request, _ = http.NewRequest(http.MethodDelete, "/?before=2023-01-16T15:00:00%2B01:00&after=2023-01-15T15:00:00%2B01:00", nil)

			GetPipelines(c)

			assert.Equal(t, http.StatusOK, c.Writer.Status())
		})
	})
}

func TestDeletePipeline(t *testing.T) {
	gin.SetMode(gin.TestMode)

	g := goblin.Goblin(t)
	g.Describe("Pipeline", func() {
		g.It("should delete pipeline", func() {
			mockStore := mocks.NewStore(t)
			mockStore.On("GetPipelineNumber", mock.Anything, mock.Anything).Return(fakePipeline, nil)
			mockStore.On("DeletePipeline", mock.Anything).Return(nil)

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Set("store", mockStore)
			c.Params = gin.Params{{Key: "number", Value: "1"}}

			DeletePipeline(c)

			mockStore.AssertCalled(t, "GetPipelineNumber", mock.Anything, mock.Anything)
			mockStore.AssertCalled(t, "DeletePipeline", mock.Anything)
			assert.Equal(t, http.StatusNoContent, c.Writer.Status())
		})

		g.It("should not delete without pipeline number", func() {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())

			DeletePipeline(c)

			assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
		})

		g.It("should not delete pending", func() {
			fakePipeline.Status = model.StatusPending

			mockStore := mocks.NewStore(t)
			mockStore.On("GetPipelineNumber", mock.Anything, mock.Anything).Return(fakePipeline, nil)

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Set("store", mockStore)
			c.Params = gin.Params{{Key: "number", Value: "1"}}

			DeletePipeline(c)

			mockStore.AssertCalled(t, "GetPipelineNumber", mock.Anything, mock.Anything)
			mockStore.AssertNotCalled(t, "DeletePipeline", mock.Anything)
			assert.Equal(t, http.StatusUnprocessableEntity, c.Writer.Status())
		})
	})
}
