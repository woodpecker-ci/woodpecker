package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/server/pipeline"
)

func TestHandlePipelineError(t *testing.T) {
	tests := []struct {
		err  error
		code int
	}{
		{
			err:  pipeline.ErrFiltered,
			code: http.StatusNoContent,
		},
		{
			err:  &pipeline.ErrNotFound{Msg: "pipeline not found"},
			code: http.StatusNotFound,
		},
		{
			err:  &pipeline.ErrBadRequest{Msg: "bad request error"},
			code: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		r := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(r)
		handlePipelineErr(c, tt.err)
		c.Writer.WriteHeaderNow() // require written header
		if r.Code != tt.code {
			t.Errorf("status code: %d - expected: %d", r.Code, tt.code)
		}
	}
}
