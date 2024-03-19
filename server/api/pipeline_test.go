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
		g.It("should parse pipeline filter", func() {
			pipelines := make([]*model.Pipeline, 1)

			mockStore := mocks.NewStore(t)
			mockStore.On("GetPipelineList", mock.Anything, mock.Anything, mock.Anything).Return(pipelines, nil)
			mockStore.On("DeletePipeline", mock.Anything).Return(nil)

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Set("store", mockStore)
			c.Request, _ = http.NewRequest("DELETE", "/?before=2023-01-16T15:00:00Z&after=2023-01-15T15:00:00Z", nil)

			DeletePipelines(c)

			assert.Equal(t, http.StatusNoContent, c.Writer.Status())
		})

		g.It("should parse pipeline filter with tz offset", func() {
			pipelines := make([]*model.Pipeline, 1)

			mockStore := mocks.NewStore(t)
			mockStore.On("GetPipelineList", mock.Anything, mock.Anything, mock.Anything).Return(pipelines, nil)
			mockStore.On("DeletePipeline", mock.Anything).Return(nil)

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Set("store", mockStore)
			c.Request, _ = http.NewRequest("DELETE", "/?before=2023-01-16T15:00:00%2B01:00&after=2023-01-15T15:00:00%2B01:00", nil)

			GetPipelines(c)

			assert.Equal(t, http.StatusOK, c.Writer.Status())
		})
	})
}
