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
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/woodpecker-ci/woodpecker/server"
	cronScheduler "github.com/woodpecker-ci/woodpecker/server/cron"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/pipeline"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// GetCron
//
//	@Summary	Get a cron job by id
//	@Router		/repos/{repo_id}/cron/{cron} [get]
//	@Produce	json
//	@Success	200	{object}	Cron
//	@Tags		Repository cron jobs
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		cron			path	string	true	"the cron job id"
func GetCron(c *gin.Context) {
	repo := session.Repo(c)
	id, err := strconv.ParseInt(c.Param("cron"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing cron id. %s", err)
		return
	}

	cron, err := store.FromContext(c).CronFind(repo, id)
	if err != nil {
		handleDbGetError(c, err)
		return
	}
	c.JSON(http.StatusOK, cron)
}

// RunCron
//
//	@Summary	Start a cron job now
//	@Router		/repos/{repo_id}/cron/{cron} [post]
//	@Produce	json
//	@Success	200	{object}	Pipeline
//	@Tags		Repository cron jobs
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		cron			path	string	true	"the cron job id"
func RunCron(c *gin.Context) {
	repo := session.Repo(c)
	_store := store.FromContext(c)
	id, err := strconv.ParseInt(c.Param("cron"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing cron id. %s", err)
		return
	}

	cron, err := _store.CronFind(repo, id)
	if err != nil {
		handleDbGetError(c, err)
		return
	}

	repo, newPipeline, err := cronScheduler.CreatePipeline(c, _store, server.Config.Services.Forge, cron)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error creating pipeline for cron %q. %s", id, err)
		return
	}

	pl, err := pipeline.Create(c, _store, repo, newPipeline)
	if err != nil {
		handlePipelineErr(c, err)
		return
	}

	c.JSON(http.StatusOK, pl)
}

// PostCron
//
//	@Summary	Persist/creat a cron job
//	@Router		/repos/{repo_id}/cron [post]
//	@Produce	json
//	@Success	200	{object}	Cron
//	@Tags		Repository cron jobs
//	@Param		Authorization	header	string		true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		cronJob			body	Cron	true	"the new cron job"
func PostCron(c *gin.Context) {
	repo := session.Repo(c)
	user := session.User(c)
	_store := store.FromContext(c)
	forge := server.Config.Services.Forge

	in := new(model.Cron)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing request. %s", err)
		return
	}
	cron := &model.Cron{
		RepoID:    repo.ID,
		Name:      in.Name,
		CreatorID: user.ID,
		Schedule:  in.Schedule,
		Branch:    in.Branch,
	}
	if err := cron.Validate(); err != nil {
		c.String(http.StatusUnprocessableEntity, "Error inserting cron. validate failed: %s", err)
		return
	}

	nextExec, err := cronScheduler.CalcNewNext(in.Schedule, time.Now())
	if err != nil {
		c.String(http.StatusBadRequest, "Error inserting cron. schedule could not parsed: %s", err)
		return
	}
	cron.NextExec = nextExec.Unix()

	if in.Branch != "" {
		// check if branch exists on forge
		_, err := forge.BranchHead(c, user, repo, in.Branch)
		if err != nil {
			c.String(http.StatusBadRequest, "Error inserting cron. branch not resolved: %s", err)
			return
		}
	}

	if err := _store.CronCreate(cron); err != nil {
		c.String(http.StatusInternalServerError, "Error inserting cron %q. %s", in.Name, err)
		return
	}
	c.JSON(http.StatusOK, cron)
}

// PatchCron
//
//	@Summary	Update a cron job
//	@Router		/repos/{repo_id}/cron/{cron} [patch]
//	@Produce	json
//	@Success	200	{object}	Cron
//	@Tags		Repository cron jobs
//	@Param		Authorization	header	string		true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		cron			path	string		true	"the cron job id"
//	@Param		cronJob			body	Cron	true	"the cron job data"
func PatchCron(c *gin.Context) {
	repo := session.Repo(c)
	user := session.User(c)
	_store := store.FromContext(c)
	forge := server.Config.Services.Forge

	id, err := strconv.ParseInt(c.Param("cron"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing cron id. %s", err)
		return
	}

	in := new(model.Cron)
	err = c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing request. %s", err)
		return
	}

	cron, err := _store.CronFind(repo, id)
	if err != nil {
		handleDbGetError(c, err)
		return
	}
	if in.Branch != "" {
		// check if branch exists on forge
		_, err := forge.BranchHead(c, user, repo, in.Branch)
		if err != nil {
			c.String(http.StatusBadRequest, "Error inserting cron. branch not resolved: %s", err)
			return
		}
		cron.Branch = in.Branch
	}
	if in.Schedule != "" {
		cron.Schedule = in.Schedule
		nextExec, err := cronScheduler.CalcNewNext(in.Schedule, time.Now())
		if err != nil {
			c.String(http.StatusBadRequest, "Error inserting cron. schedule could not parsed: %s", err)
			return
		}
		cron.NextExec = nextExec.Unix()
	}
	if in.Name != "" {
		cron.Name = in.Name
	}
	cron.CreatorID = user.ID

	if err := cron.Validate(); err != nil {
		c.String(http.StatusUnprocessableEntity, "Error inserting cron. validate failed: %s", err)
		return
	}
	if err := _store.CronUpdate(repo, cron); err != nil {
		c.String(http.StatusInternalServerError, "Error updating cron %q. %s", in.Name, err)
		return
	}
	c.JSON(http.StatusOK, cron)
}

// GetCronList
//
//	@Summary	Get the cron job list
//	@Router		/repos/{repo_id}/cron [get]
//	@Produce	json
//	@Success	200	{array}	Cron
//	@Tags		Repository cron jobs
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetCronList(c *gin.Context) {
	repo := session.Repo(c)
	list, err := store.FromContext(c).CronList(repo, session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting cron list. %s", err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// DeleteCron
//
//	@Summary	Delete a cron job by id
//	@Router		/repos/{repo_id}/cron/{cron} [delete]
//	@Produce	plain
//	@Success	200
//	@Tags		Repository cron jobs
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		cron			path	string	true	"the cron job id"
func DeleteCron(c *gin.Context) {
	repo := session.Repo(c)
	id, err := strconv.ParseInt(c.Param("cron"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing cron id. %s", err)
		return
	}
	if err := store.FromContext(c).CronDelete(repo, id); err != nil {
		handleDbGetError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
