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

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/pipeline"
	"go.woodpecker-ci.org/woodpecker/v2/server/pipeline/stepbuilder"
	"go.woodpecker-ci.org/woodpecker/v2/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"
)

// CreatePipeline
//
//	@Summary	Trigger a manual pipeline
//	@Router		/repos/{repo_id}/pipelines [post]
//	@Produce	json
//	@Success	200	{object}	Pipeline
//	@Tags		Pipelines
//	@Param		Authorization	header	string			true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int				true	"the repository id"
//	@Param		options			body	PipelineOptions	true	"the options for the pipeline to run"
func CreatePipeline(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)
	_forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
	if err != nil {
		log.Error().Err(err).Msg("Cannot get forge from repo")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// parse create options
	var opts model.PipelineOptions
	err = json.NewDecoder(c.Request.Body).Decode(&opts)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user := session.User(c)

	lastCommit, err := _forge.BranchHead(c, user, repo, opts.Branch)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("could not fetch branch head: %w", err))
		return
	}

	tmpPipeline := createTmpPipeline(model.EventManual, lastCommit, user, &opts)

	pl, err := pipeline.Create(c, _store, repo, tmpPipeline)
	if err != nil {
		handlePipelineErr(c, err)
	} else {
		c.JSON(http.StatusOK, pl)
	}
}

func createTmpPipeline(event model.WebhookEvent, commit *model.Commit, user *model.User, opts *model.PipelineOptions) *model.Pipeline {
	return &model.Pipeline{
		Event:     event,
		Commit:    commit.SHA,
		Branch:    opts.Branch,
		Timestamp: time.Now().UTC().Unix(),

		Avatar:  user.Avatar,
		Message: "MANUAL PIPELINE @ " + opts.Branch,

		Ref:                 opts.Branch,
		AdditionalVariables: opts.Variables,

		Author: user.Login,
		Email:  user.Email,

		ForgeURL: commit.ForgeURL,
	}
}

// GetPipelines
//
//	@Summary	List repository pipelines
//	@Description	Get a list of pipelines for a repository.
//	@Router		/repos/{repo_id}/pipelines [get]
//	@Produce	json
//	@Success	200	{array}	Pipeline
//	@Tags		Pipelines
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
//	@Param		before			query	string	false	"only return pipelines before this RFC3339 date"
//	@Param		after			query	string	false	"only return pipelines after this RFC3339 date"
func GetPipelines(c *gin.Context) {
	repo := session.Repo(c)
	before := c.Query("before")
	after := c.Query("after")

	filter := new(model.PipelineFilter)

	if before != "" {
		beforeDt, err := time.Parse(time.RFC3339, before)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		filter.Before = beforeDt.Unix()
	}

	if after != "" {
		afterDt, err := time.Parse(time.RFC3339, after)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		filter.After = afterDt.Unix()
	}

	pipelines, err := store.FromContext(c).GetPipelineList(repo, session.Pagination(c), filter)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, pipelines)
}

// DeletePipeline
//
//	@Summary	Delete a pipeline
//	@Router		/repos/{repo_id}/pipelines/{number} [delete]
//	@Produce	plain
//	@Success	204
//	@Tags		Pipelines
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		number			path	int		true	"the number of the pipeline"
func DeletePipeline(c *gin.Context) {
	_store := store.FromContext(c)

	repo := session.Repo(c)
	num, err := strconv.ParseInt(c.Param("number"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	pl, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleDBError(c, err)
		return
	}

	if ok := pipelineDeleteAllowed(pl); !ok {
		c.String(http.StatusUnprocessableEntity, "Cannot delete pipeline with status %s", pl.Status)
		return
	}

	err = store.FromContext(c).DeletePipeline(pl)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error deleting pipeline. %s", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetPipeline
//
//	@Summary	Get a repositories pipeline
//	@Router		/repos/{repo_id}/pipelines/{number} [get]
//	@Produce	json
//	@Success	200	{object}	Pipeline
//	@Tags		Pipelines
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		number			path	int		true	"the number of the pipeline, OR 'latest'"
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
		handleDBError(c, err)
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
		handleDBError(c, err)
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
//	@Summary	Get logs for a pipeline step
//	@Router		/repos/{repo_id}/logs/{number}/{stepID} [get]
//	@Produce	json
//	@Success	200	{array}	LogEntry
//	@Tags		Pipeline logs
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		number			path	int		true	"the number of the pipeline"
//	@Param		stepID			path	int		true	"the step id"
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
		handleDBError(c, err)
		return
	}

	stepID, err := strconv.ParseInt(c.Params.ByName("stepId"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	step, err := _store.StepLoad(stepID)
	if err != nil {
		handleDBError(c, err)
		return
	}

	if step.PipelineID != pl.ID {
		// make sure we cannot read arbitrary logs by id
		_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("step with id %d is not part of repo %s", stepID, repo.FullName))
		return
	}

	logs, err := server.Config.Services.LogStore.LogFind(step)
	if err != nil {
		handleDBError(c, err)
		return
	}

	c.JSON(http.StatusOK, logs)
}

// DeleteStepLogs
//
//	@Summary	Delete step logs of a pipeline
//	@Router		/repos/{repo_id}/logs/{number}/{stepId} [delete]
//	@Produce	plain
//	@Success	204
//	@Tags		Pipeline logs
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		number			path	int		true	"the number of the pipeline"
//	@Param		stepId			path	int		true	"the step id"
func DeleteStepLogs(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)

	pipelineNumber, err := strconv.ParseInt(c.Params.ByName("number"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	_pipeline, err := _store.GetPipelineNumber(repo, pipelineNumber)
	if err != nil {
		handleDBError(c, err)
		return
	}

	stepID, err := strconv.ParseInt(c.Params.ByName("stepId"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	_step, err := _store.StepLoad(stepID)
	if err != nil {
		handleDBError(c, err)
		return
	}

	if _step.PipelineID != _pipeline.ID {
		// make sure we cannot read arbitrary logs by id
		_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("step with id %d is not part of repo %s", stepID, repo.FullName))
		return
	}

	switch _step.State {
	case model.StatusRunning, model.StatusPending:
		c.String(http.StatusUnprocessableEntity, "Cannot delete logs for a pending or running step")
		return
	}

	err = _store.LogDelete(_step)
	if err != nil {
		handleDBError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetPipelineConfig
//
//	@Summary	Get configuration files for a pipeline
//	@Router		/repos/{repo_id}/pipelines/{number}/config [get]
//	@Produce	json
//	@Success	200	{array}	Config
//	@Tags		Pipelines
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
		handleDBError(c, err)
		return
	}

	configs, err := _store.ConfigsForPipeline(pl.ID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, configs)
}

// GetPipelineMetadata
//
//	@Summary	Get metadata for a pipeline or a specific workflow, including previous pipeline info
//	@Router		/repos/{repo_id}/pipelines/{number}/metadata [get]
//	@Produce	json
//	@Success	200	{object}	metadata.Metadata
//	@Tags		Pipelines
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		number			path	int		true	"the number of the pipeline"
func GetPipelineMetadata(c *gin.Context) {
	repo := session.Repo(c)
	num, err := strconv.ParseInt(c.Param("number"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	_store := store.FromContext(c)
	currentPipeline, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleDBError(c, err)
		return
	}

	forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	prevPipeline, err := _store.GetPipelineLastBefore(repo, currentPipeline.Branch, currentPipeline.ID)
	if err != nil && !errors.Is(err, types.RecordNotExist) {
		handleDBError(c, err)
		return
	}

	metadata := stepbuilder.MetadataFromStruct(forge, repo, currentPipeline, prevPipeline, nil, server.Config.Server.Host)
	c.JSON(http.StatusOK, metadata)
}

// CancelPipeline
//
//	@Summary	Cancel a pipeline
//	@Router		/repos/{repo_id}/pipelines/{number}/cancel [post]
//	@Produce	plain
//	@Success	200
//	@Tags		Pipelines
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		number			path	int		true	"the number of the pipeline"
func CancelPipeline(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)
	user := session.User(c)
	_forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
	if err != nil {
		log.Error().Err(err).Msg("Cannot get forge from repo")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	num, _ := strconv.ParseInt(c.Params.ByName("number"), 10, 64)

	pl, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleDBError(c, err)
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
//	@Summary	Approve and start a pipeline
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
		handleDBError(c, err)
		return
	}

	newPipeline, err := pipeline.Approve(c, _store, pl, user, repo)
	if err != nil {
		handlePipelineErr(c, err)
	} else {
		c.JSON(http.StatusOK, newPipeline)
	}
}

// PostDecline
//
//	@Summary	Decline a pipeline
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
		handleDBError(c, err)
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
//	@Summary	List pipelines in queue
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
//	@Param			repo_id			path	int		true	"the repository id"
//	@Param			number			path	int		true	"the number of the pipeline"
//	@Param			event			query	string	false	"override the event type"
//	@Param			deploy_to		query	string	false	"override the target deploy value"
func PostPipeline(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)

	num, err := strconv.ParseInt(c.Param("number"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := _store.GetUser(repo.UserID)
	if err != nil {
		handleDBError(c, err)
		return
	}

	pl, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleDBError(c, err)
		return
	}

	// refresh the token to make sure, pipeline.Restart can still obtain the pipeline config if necessary again
	refreshUserToken(c, user)

	// make Deploy overridable

	// make Deploy task overridable
	pl.DeployTask = c.DefaultQuery("deploy_task", pl.DeployTask)

	// make Event overridable to deploy
	// TODO: refactor to use own proper API for deploy
	if event, ok := c.GetQuery("event"); ok {
		pl.Event = model.WebhookEvent(event)
		if pl.Event != model.EventDeploy {
			_ = c.AbortWithError(http.StatusBadRequest, model.ErrInvalidWebhookEvent)
			return
		}

		if !repo.AllowDeploy {
			_ = c.AbortWithError(http.StatusForbidden, fmt.Errorf("repo does not allow deployments"))
			return
		}

		pl.DeployTo = c.DefaultQuery("deploy_to", pl.DeployTo)
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

	newPipeline, err := pipeline.Restart(c, _store, pl, user, repo, envs)
	if err != nil {
		handlePipelineErr(c, err)
	} else {
		c.JSON(http.StatusOK, newPipeline)
	}
}

// DeletePipelineLogs
//
//	@Summary	Deletes all logs of a pipeline
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
		handleDBError(c, err)
		return
	}

	steps, err := _store.StepList(pl)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if ok := pipelineDeleteAllowed(pl); !ok {
		c.String(http.StatusUnprocessableEntity, "Cannot delete logs for pipeline with status %s", pl.Status)
		return
	}

	for _, step := range steps {
		if lErr := server.Config.Services.LogStore.LogDelete(step); err != nil {
			err = errors.Join(err, lErr)
		}
	}
	if err != nil {
		c.String(http.StatusInternalServerError, "Error deleting pipeline logs. %s", err)
		return
	}

	c.Status(http.StatusNoContent)
}
