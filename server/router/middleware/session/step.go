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

// Step returns the step resolved by SetStep.
func Step(c *gin.Context) *model.Step {
	v, ok := c.Get("step")
	if !ok {
		return nil
	}
	s, ok := v.(*model.Step)
	if !ok {
		return nil
	}
	return s
}

// SetStep resolves the `step_id` path param within the pipeline of the request
// and stores the step in the context. It must run after SetPipeline.
func SetStep() gin.HandlerFunc {
	return func(c *gin.Context) {
		_store := store.FromContext(c)
		pipeline := Pipeline(c)

		stepID, err := strconv.ParseInt(c.Param("step_id"), 10, 64)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		step, err := _store.StepLoad(pipeline.ID, stepID)
		if err != nil {
			if errors.Is(err, types.ErrRecordNotExist) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Set("step", step)
		c.Next()
	}
}
