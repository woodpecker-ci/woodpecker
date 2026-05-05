// Copyright 2023 Woodpecker Authors
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

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
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
		assert.Equal(t, tt.code, r.Code)
	}
}
