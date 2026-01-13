// Copyright 2022 Woodpecker Authors
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
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func handlePipelineErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, &pipeline.ErrNotFound{}):
		c.String(http.StatusNotFound, "%s", err)
	case errors.Is(err, &pipeline.ErrBadRequest{}):
		c.String(http.StatusBadRequest, "%s", err)
	case errors.Is(err, pipeline.ErrFiltered):
		// for debugging purpose we add a header
		c.Writer.Header().Add("Pipeline-Filtered", "true")
		c.Status(http.StatusNoContent)
	default:
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}
}

func handleDBError(c *gin.Context, err error) {
	if errors.Is(err, types.RecordNotExist) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	_ = c.AbortWithError(http.StatusInternalServerError, err)
}

// pipelineDeleteAllowed checks if the given pipeline can be deleted based on its status.
// It returns a bool indicating if delete is allowed, and the pipeline's status.
func pipelineDeleteAllowed(pl *model.Pipeline) bool {
	switch pl.Status {
	case model.StatusRunning, model.StatusPending, model.StatusBlocked:
		return false
	}

	return true
}
