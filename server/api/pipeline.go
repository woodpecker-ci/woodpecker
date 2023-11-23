// Copyright 2022 Woodpecker Authors
// Copyright 2021 Informatyka Boguslawski sp. z o.o. sp.k., http://www.ib.pl/
// Copyright 2018 Drone.IO Inc.
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
//
// This file has been modified by Informatyka Boguslawski sp. z o.o. sp.k.

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/server/model"
	"go.woodpecker-ci.org/woodpecker/server/pipeline"
	"go.woodpecker-ci.org/woodpecker/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/server/store"
)

// CreatePipeline
//
//	@Summary	Run/trigger a pipelines
//	@Router		/repos/{repo_id}/pipelines [post]
//	@Produce	json
//	@Success	200	{object}	Pipeline
//	@Tags			Pipelines
//	@Param		Authorization	header	string					true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		options			body	PipelineOptions	true	"the options for the pipeline to run"
func CreatePipeline(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)
	forge := session.Forge(c)

	// parse create options
	var opts model.PipelineOptions
	err := json.NewDecoder(c.Request.Body).Decode(&opts)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user := session.User(c)

	lastCommit, _ := forge.BranchHead(c, user, repo, opts.Branch)

	tmpBuild := createTmpPipeline(model.EventManual, lastCommit, repo, user, &opts)

	pl, err := pipeline.Create(c, _store, repo, tmpBuild)
	if err != nil {
		handlePipelineErr(c, err)
	} else {
		c.JSON(http.StatusOK, pl)
	}
}

func createTmpPipeline(event model.WebhookEvent, commitSHA string, repo *model.Repo, user *model.User, opts *model.PipelineOptions) *model.Pipeline {
	return &model.Pipeline{
		Event:     event,
		Commit:    commitSHA,
		Branch:    opts.Branch,
		Timestamp: time.Now().UTC().Unix(),

		Avatar:  user.Avatar,
		Message: "MANUAL PIPELINE @ " + opts.Branch,

		Ref:                 opts.Branch,
		AdditionalVariables: opts.Variables,

		Author: user.Login,
		Email:  user.Email,

		// TODO: Generate proper URL to commit
		ForgeURL: repo.ForgeURL,
	}
}

// GetPipelines
//
//	@Summary	Get pipelines, current running and past ones
//	@Router		/repos/{repo_id}/pipelines [get]
//	@Produce	json
//	@Success	200	{array}	Pipeline
//	@Tags		Pipelines
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id		path	int		true	"the repository id"
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage		query	int		false	"for response pagination, max items per page"	default(50)
func GetPipelines(c *gin.Context) {
	repo := session.Repo(c)

	pipelines, err := store.FromContext(c).GetPipelineList(repo, session.Pagination(c))
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, pipelines)
}

// GetPipeline
//
//	@Summary	Pipeline information by number
//	@Router		/repos/{repo_id}/pipelines/{number} [get]
//	@Produce	json
//	@Success	200	{object}	Pipeline
//	@Tags		Pipelines
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id		path	int		true	"the repository id"
//	@Param		number		path	int		true	"the number of the pipeline, OR 'latest'"
func GetPipeline(c *gin.Context) {
	_store := store.FromContext(c)
	if c.Param("number") == "latest" {
		GetPipelineLast(c)
		return
	}

	repo := session.Repo(c)
	num, err := strconv.ParseInt(c.Param("number"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	pl, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleDbError(c, err)
		return
	}
	if pl.Workflows, err = _store.WorkflowGetTree(pl); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, pl)
}

func GetPipelineLast(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)
	branch := c.DefaultQuery("branch", repo.Branch)

	pl, err := _store.GetPipelineLast(repo, branch)
	if err != nil {
		handleDbError(c, err)
		return
	}

	if pl.Workflows, err = _store.WorkflowGetTree(pl); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, pl)
}

// GetStepLogs
//
//	@Summary	Log information
//	@Router		/repos/{repo_id}/logs/{number}/{stepID} [get]
//	@Produce	json
//	@Success	200 {array} LogEntry
//	@Tags			Pipeline logs
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id		path	int		true	"the repository id"
//	@Param		number		path	int		true		"the number of the pipeline"
//	@Param		stepID		path	int		true		"the step id"
func GetStepLogs(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)

	// parse the pipeline number and step sequence number from
	// the request parameter.
	num, err := strconv.ParseInt(c.Params.ByName("number"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	pl, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleDbError(c, err)
		return
	}

	stepID, err := strconv.ParseInt(c.Params.ByName("stepId"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	step, err := _store.StepLoad(stepID)
	if err != nil {
		handleDbError(c, err)
		return
	}

	if step.PipelineID != pl.ID {
		// make sure we can not read arbitrary logs by id
		_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("step with id %d is not part of repo %s", stepID, repo.FullName))
		return
	}

	logs, err := _store.LogFind(step)
	if err != nil {
		handleDbError(c, err)
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetPipelineConfig
//
//	@Summary	Pipeline configuration
//	@Router		/repos/{repo_id}/pipelines/{number}/config [get]
//	@Produce	json
//	@Success	200	{array}	Config
//	@Tags			Pipelines
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		number			path	int		true	"the number of the pipeline"
func GetPipelineConfig(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)
	num, err := strconv.ParseInt(c.Param("number"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	pl, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleDbError(c, err)
		return
	}

	configs, err := _store.ConfigsForPipeline(pl.ID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, configs)
}

// CancelPipeline
//
//	@Summary	Cancels a pipeline
//	@Router		/repos/{repo_id}/pipelines/{number}/cancel [post]
//	@Produce	plain
//	@Success	200
//	@Tags		Pipelines
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		number			path	int		true	"the number of the pipeline"
func CancelPipeline(c *gin.Context) {
	_store := store.FromContext(c)
	_forge := session.Forge(c)
	repo := session.Repo(c)
	user := session.User(c)
	num, _ := strconv.ParseInt(c.Params.ByName("number"), 10, 64)

	pl, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleDbError(c, err)
		return
	}

	if err := pipeline.Cancel(c, _forge, _store, repo, user, pl); err != nil {
		handlePipelineErr(c, err)
	} else {
		c.Status(http.StatusNoContent)
	}
}

// PostApproval
//
//	@Summary	Start pipelines in gated repos
//	@Router		/repos/{repo_id}/pipelines/{number}/approve [post]
//	@Produce	json
//	@Success	200	{object}	Pipeline
//	@Tags		Pipelines
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		number			path	int		true	"the number of the pipeline"
func PostApproval(c *gin.Context) {
	var (
		_store = store.FromContext(c)
		repo   = session.Repo(c)
		user   = session.User(c)
		num, _ = strconv.ParseInt(c.Params.ByName("number"), 10, 64)
	)

	pl, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleDbError(c, err)
		return
	}

	newpipeline, err := pipeline.Approve(c, _store, pl, user, repo)
	if err != nil {
		handlePipelineErr(c, err)
	} else {
		c.JSON(http.StatusOK, newpipeline)
	}
}

// PostDecline
//
//	@Summary	Decline pipelines in gated repos
//	@Router		/repos/{repo_id}/pipelines/{number}/decline [post]
//	@Produce	json
//	@Success	200	{object}	Pipeline
//	@Tags		Pipelines
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		number			path	int		true	"the number of the pipeline"
func PostDecline(c *gin.Context) {
	var (
		_store = store.FromContext(c)
		repo   = session.Repo(c)
		user   = session.User(c)
		num, _ = strconv.ParseInt(c.Params.ByName("number"), 10, 64)
	)

	pl, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleDbError(c, err)
		return
	}

	pl, err = pipeline.Decline(c, _store, pl, user, repo)
	if err != nil {
		handlePipelineErr(c, err)
	} else {
		c.JSON(http.StatusOK, pl)
	}
}

// GetPipelineQueue
//
//	@Summary	List pipeline queues
//	@Router		/pipelines [get]
//	@Produce	json
//	@Success	200	{array}	Feed
//	@Tags		Pipeline queues
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func GetPipelineQueue(c *gin.Context) {
	out, err := store.FromContext(c).GetPipelineQueue()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting pipeline queue. %s", err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// PostPipeline
//
//	@Summary		Restart a pipeline
//	@Description	Restarts a pipeline optional with altered event, deploy or environment
//	@Router			/repos/{repo_id}/pipelines/{number} [post]
//	@Produce		json
//	@Success		200	{object}	Pipeline
//	@Tags			Pipelines
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			owner			path	string	true	"the repository owner's name"
//	@Param			name			path	string	true	"the repository name"
//	@Param			number			path	int		true	"the number of the pipeline"
//	@Param			event			query	string	false	"override the event type"
//	@Param			deploy_to		query	string	false	"override the target deploy value"
func PostPipeline(c *gin.Context) {
	_store := store.FromContext(c)
	_forge := session.Forge(c)
	repo := session.Repo(c)

	num, err := strconv.ParseInt(c.Param("number"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := _store.GetUser(repo.UserID)
	if err != nil {
		handleDbError(c, err)
		return
	}

	pl, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleDbError(c, err)
		return
	}

	// refresh the token to make sure, pipeline.Restart can still obtain the pipeline config if necessary again
	refreshUserToken(c, user)

	// make Deploy overridable
	pl.Deploy = c.DefaultQuery("deploy_to", pl.Deploy)

	// make Event overridable
	if event, ok := c.GetQuery("event"); ok {
		pl.Event = model.WebhookEvent(event)

		if err := model.ValidateWebhookEvent(pl.Event); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	}

	// Read query string parameters into pipelineParams, exclude reserved params
	envs := map[string]string{}
	for key, val := range c.Request.URL.Query() {
		switch key {
		// Skip some options of the endpoint
		case "fork", "event", "deploy_to":
			continue
		default:
			// We only accept string literals, because pipeline parameters will be
			// injected as environment variables
			// TODO: sanitize the value
			envs[key] = val[0]
		}
	}

	netrc, err := _forge.Netrc(user, repo)
	if err != nil {
		handlePipelineErr(c, err)
		return
	}

	newpipeline, err := pipeline.Restart(c, _store, pl, user, repo, envs, netrc)
	if err != nil {
		handlePipelineErr(c, err)
	} else {
		c.JSON(http.StatusOK, newpipeline)
	}
}

// DeletePipelineLogs
//
//	@Summary	Deletes log
//	@Router		/repos/{repo_id}/logs/{number} [delete]
//	@Produce	plain
//	@Success	204
//	@Tags		Pipeline logs
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		number			path	int		true	"the number of the pipeline"
func DeletePipelineLogs(c *gin.Context) {
	_store := store.FromContext(c)

	repo := session.Repo(c)
	num, _ := strconv.ParseInt(c.Params.ByName("number"), 10, 64)

	pl, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleDbError(c, err)
		return
	}

	steps, err := _store.StepList(pl)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	switch pl.Status {
	case model.StatusRunning, model.StatusPending:
		c.String(http.StatusUnprocessableEntity, "Cannot delete logs for a pending or running pipeline")
		return
	}

	for _, step := range steps {
		if lErr := _store.LogDelete(step); err != nil {
			err = errors.Join(err, lErr)
		}
	}
	if err != nil {
		c.String(http.StatusInternalServerError, "There was a problem deleting your logs. %s", err)
		return
	}

	c.Status(http.StatusNoContent)
}
