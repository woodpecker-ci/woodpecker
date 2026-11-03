// Copyright 2026 Woodpecker Authors
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
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// PostWorkflowResume
//
//	@Summary	Resume a paused workflow
//	@Router		/repos/{repo_id}/pipelines/{pipeline_number}/workflows/{workflow_id}/resume [post]
//	@Produce	json
//	@Success	200	{object}	Pipeline
//	@Tags		Pipelines
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		pipeline_number	path	int		true	"the number of the pipeline"
//	@Param		workflow_id		path	int		true	"the id of the paused workflow"
func PostWorkflowResume(c *gin.Context) {
	decideWorkflowPause(c, true)
}

// PostWorkflowDecline
//
//	@Summary	Decline a paused workflow, it gets stopped and reported as killed
//	@Router		/repos/{repo_id}/pipelines/{pipeline_number}/workflows/{workflow_id}/decline [post]
//	@Produce	json
//	@Success	200	{object}	Pipeline
//	@Tags		Pipelines
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		pipeline_number	path	int		true	"the number of the pipeline"
//	@Param		workflow_id		path	int		true	"the id of the paused workflow"
func PostWorkflowDecline(c *gin.Context) {
	decideWorkflowPause(c, false)
}

func decideWorkflowPause(c *gin.Context, resume bool) {
	_store := store.FromContext(c)
	repo := session.Repo(c)
	user := session.User(c)

	num, err := strconv.ParseInt(c.Params.ByName("pipeline_number"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid pipeline number")
		return
	}
	workflowID, err := strconv.ParseInt(c.Params.ByName("workflow_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid workflow id")
		return
	}

	pl, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleDBError(c, err)
		return
	}

	workflow, err := _store.WorkflowLoad(workflowID)
	if err != nil {
		handleDBError(c, err)
		return
	}
	if workflow.PipelineID != pl.ID {
		c.String(http.StatusNotFound, "workflow %d does not belong to pipeline %d", workflowID, num)
		return
	}

	if _, err := pipeline.DecidePause(_store, workflow, resume); err != nil {
		handlePipelineErr(c, err)
		return
	}
	log.Debug().Int64("workflow_id", workflowID).Bool("resume", resume).Str("user", user.Login).Msg("paused workflow decided")

	if pl.Workflows, err = _store.WorkflowGetTree(pl); err != nil {
		handleDBError(c, err)
		return
	}
	if err := server.Config.Services.Scheduler.PublishPipelineEvent(c, repo, pl); err != nil {
		log.Error().Err(err).Msg("could not push pipeline status change to pubsub provider")
	}

	c.JSON(http.StatusOK, pl.ToAPIModel())
}
