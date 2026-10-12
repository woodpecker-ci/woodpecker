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

package session

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// Pipeline returns the pipeline resolved by SetPipeline.
func Pipeline(c *gin.Context) *model.Pipeline {
	v, ok := c.Get("pipeline")
	if !ok {
		return nil
	}
	p, ok := v.(*model.Pipeline)
	if !ok {
		return nil
	}
	return p
}

// SetPipeline resolves the `pipeline_number` path param within the repo of the
// request and stores the pipeline in the context. It must run after SetRepo.
func SetPipeline() gin.HandlerFunc {
	return func(c *gin.Context) {
		_store := store.FromContext(c)
		repo := Repo(c)

		number, err := strconv.ParseInt(c.Param("pipeline_number"), 10, 64)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		pipeline, err := _store.GetPipelineNumber(repo, number)
		if err != nil {
			if errors.Is(err, types.ErrRecordNotExist) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Set("pipeline", pipeline)
		c.Next()
	}
}
