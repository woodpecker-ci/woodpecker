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

package feedback

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"

	cicd_feedback "github.com/6543/cicd_feedback"
)

func Get(c *gin.Context) {
	_store := store.FromContext(c)

	repo := session.Repo(c)
	num, err := strconv.ParseInt(c.Param("number"), 10, 64)
	if err != nil {
		handleError(c, err)
		return
	}

	pl, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleError(c, err)
		return
	}
	if pl.Workflows, err = _store.WorkflowGetTree(pl); err != nil {
		handleError(c, err)
		return
	}

	resp, err := Convert(pl, pl.Workflows)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func GetStepLog(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)

	num, err := strconv.ParseInt(c.Params.ByName("number"), 10, 64)
	if err != nil {
		handleError(c, err)
		return
	}

	pl, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleError(c, err)
		return
	}

	stepID, err := strconv.ParseInt(c.Params.ByName("stepId"), 10, 64)
	if err != nil {
		handleError(c, err)
		return
	}

	step, err := _store.StepLoad(stepID)
	if err != nil {
		handleError(c, err)
		return
	}

	if step.PipelineID != pl.ID {
		// make sure we cannot read arbitrary logs by id
		err := fmt.Errorf("step with id %d is not part of repo %s", stepID, repo.FullName)
		handleError(c, err)
		return
	}

	logs, err := server.Config.Services.LogStore.LogFind(step)
	if err != nil {
		handleError(c, err)
		return
	}

	feedbackLog, err := convertLogs(logs)
	if err != nil {
		handleError(c, err)
		return
	}

	// TODO: use c.Stream if step is currently running and pipe stream to it
	c.String(http.StatusOK, feedbackLog)
}

func handleError(c *gin.Context, err error) {
	if errors.Is(err, types.RecordNotExist) ||
		errors.Is(err, ErrNonConvertableStatus) ||
		errors.Is(err, ErrStepDepResolve) ||
		errors.Is(err, ErrWorkflowDepResolve) {
		c.JSON(http.StatusInternalServerError, cicd_feedback.ErrorResponse{
			Error:            cicd_feedback.ErrorInternal,
			ErrorDescription: err.Error(),
		})
		return
	}

	c.JSON(http.StatusBadRequest, cicd_feedback.ErrorResponse{
		Error:            cicd_feedback.ErrorOther,
		ErrorDescription: err.Error(),
	})
}
