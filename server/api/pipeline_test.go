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
			pipelines := make([]*model.Pipeline, 0)
			pipelines = append(pipelines, fakePipeline)

			mockStore := mocks.NewStore(t)
			mockStore.On("GetPipelineList", mock.Anything, mock.Anything, mock.Anything).Return(pipelines, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("store", mockStore)

			GetPipelines(c)

			mockStore.AssertCalled(t, "GetPipelineList", mock.Anything, mock.Anything, mock.Anything)
			assert.Equal(t, http.StatusOK, c.Writer.Status())
		})

		g.It("should parse pipeline filter", func() {
			pipelines := make([]*model.Pipeline, 0)
			pipelines = append(pipelines, fakePipeline)

			mockStore := mocks.NewStore(t)
			mockStore.On("GetPipelineList", mock.Anything, mock.Anything, mock.Anything).Return(pipelines, nil)

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Set("store", mockStore)
			c.Params = gin.Params{{Key: "before", Value: "2023-01-16T15:00:00Z"}, {Key: "after", Value: "2023-01-15T15:00:00Z"}}

			GetPipelines(c)

			assert.Equal(t, http.StatusOK, c.Writer.Status())
		})

		g.It("should parse pipeline filter with tz offset", func() {
			pipelines := make([]*model.Pipeline, 0)
			pipelines = append(pipelines, fakePipeline)

			mockStore := mocks.NewStore(t)
			mockStore.On("GetPipelineList", mock.Anything, mock.Anything, mock.Anything).Return(pipelines, nil)

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Set("store", mockStore)
			c.Params = gin.Params{{Key: "before", Value: "2023-01-16T15:00:00%2B01:00"}, {Key: "after", Value: "2023-01-15T15:00:00%2B01:00"}}

			GetPipelines(c)

			assert.Equal(t, http.StatusOK, c.Writer.Status())
		})
	})
}
