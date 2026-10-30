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

package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// GetWorkflowAgent
//
//	@Summary		Get the agent a workflow got assigned to
//	@Description	Returns the snapshot of the agent taken when it got the workflow assigned, so it is also available after the agent got deleted.
//	@Description	It holds no secrets, so every logged in user who can see the pipeline may read it.
//	@Router			/repos/{repo_id}/pipelines/{pipeline_number}/workflows/{workflow_id}/agent [get]
//	@Produce		json
//	@Success		200	{object}	AgentSnapshot
//	@Tags			Pipelines
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			repo_id			path	int		true	"the repository id"
//	@Param			pipeline_number	path	int		true	"the number of the pipeline"
//	@Param			workflow_id		path	int		true	"the workflow id"
func GetWorkflowAgent(c *gin.Context) {
	_store := store.FromContext(c)
	pl := session.Pipeline(c)

	workflowID, err := strconv.ParseInt(c.Param("workflow_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid workflow ID")
		return
	}

	workflow, err := _store.WorkflowLoad(workflowID)
	if err != nil {
		handleDBError(c, err)
		return
	}
	// workflow IDs are global, the permissions were only checked for this pipeline's repo
	if workflow.PipelineID != pl.ID {
		c.String(http.StatusNotFound, "Workflow not found")
		return
	}

	snapshot, err := workflowAgentSnapshot(_store, workflow)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if snapshot == nil {
		c.String(http.StatusNotFound, "No agent recorded for this workflow")
		return
	}

	// the snapshot is shared by all agents with the same metadata, the workflow knows which one
	response := *snapshot
	response.AgentID = workflow.AgentID
	c.JSON(http.StatusOK, response)
}

// workflowAgentSnapshot returns the agent snapshot referenced by the workflow, or nil if
// there is none, e.g. for workflows that ran before snapshots were introduced.
func workflowAgentSnapshot(_store store.Store, workflow *model.Workflow) (*model.AgentSnapshot, error) {
	if workflow.AgentSnapshotID == 0 {
		return nil, nil
	}
	snapshot, err := _store.AgentSnapshotFind(workflow.AgentSnapshotID)
	if errors.Is(err, types.ErrRecordNotExist) {
		return nil, nil
	}
	return snapshot, err
}
