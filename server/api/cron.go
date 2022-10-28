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

// GetCron gets a cron job by id from the database and writes
// to the response in json format.
func GetCron(c *gin.Context) {
	repo := session.Repo(c)
	id, err := strconv.ParseInt(c.Param("cron"), 10, 64)
	if err != nil {
		c.String(400, "Error parsing cron id. %s", err)
		return
	}

	cron, err := store.FromContext(c).CronFind(repo, id)
	if err != nil {
		c.String(404, "Error getting cron %q. %s", id, err)
		return
	}
	c.JSON(200, cron)
}

// RunCron starts a cron job now.
func RunCron(c *gin.Context) {
	repo := session.Repo(c)
	_store := store.FromContext(c)
	id, err := strconv.ParseInt(c.Param("cron"), 10, 64)
	if err != nil {
		c.String(400, "Error parsing cron id. %s", err)
		return
	}

	cron, err := _store.CronFind(repo, id)
	if err != nil {
		c.String(http.StatusNotFound, "Error getting cron %q. %s", id, err)
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

	c.JSON(200, pl)
}

// PostCron persists the cron job to the database.
func PostCron(c *gin.Context) {
	repo := session.Repo(c)
	user := session.User(c)
	store := store.FromContext(c)
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
		c.String(400, "Error inserting cron. validate failed: %s", err)
		return
	}

	nextExec, err := cronScheduler.CalcNewNext(in.Schedule, time.Now())
	if err != nil {
		c.String(400, "Error inserting cron. schedule could not parsed: %s", err)
		return
	}
	cron.NextExec = nextExec.Unix()

	if in.Branch != "" {
		// check if branch exists on forge
		_, err := forge.BranchHead(c, user, repo, in.Branch)
		if err != nil {
			c.String(400, "Error inserting cron. branch not resolved: %s", err)
			return
		}
	}

	if err := store.CronCreate(cron); err != nil {
		c.String(500, "Error inserting cron %q. %s", in.Name, err)
		return
	}
	c.JSON(200, cron)
}

// PatchCron updates the cron job in the database.
func PatchCron(c *gin.Context) {
	repo := session.Repo(c)
	user := session.User(c)
	store := store.FromContext(c)
	forge := server.Config.Services.Forge

	id, err := strconv.ParseInt(c.Param("cron"), 10, 64)
	if err != nil {
		c.String(400, "Error parsing cron id. %s", err)
		return
	}

	in := new(model.Cron)
	err = c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing request. %s", err)
		return
	}

	cron, err := store.CronFind(repo, id)
	if err != nil {
		c.String(404, "Error getting cron %d. %s", id, err)
		return
	}
	if in.Branch != "" {
		// check if branch exists on forge

		_, err := forge.BranchHead(c, user, repo, in.Branch)
		if err != nil {
			c.String(400, "Error inserting cron. branch not resolved: %s", err)
			return
		}
		cron.Branch = in.Branch
	}
	if in.Schedule != "" {
		cron.Schedule = in.Schedule
		nextExec, err := cronScheduler.CalcNewNext(in.Schedule, time.Now())
		if err != nil {
			c.String(400, "Error inserting cron. schedule could not parsed: %s", err)
			return
		}
		cron.NextExec = nextExec.Unix()
	}
	if in.Name != "" {
		cron.Name = in.Name
	}
	cron.CreatorID = user.ID

	if err := cron.Validate(); err != nil {
		c.String(400, "Error inserting cron. validate failed: %s", err)
		return
	}
	if err := store.CronUpdate(repo, cron); err != nil {
		c.String(500, "Error updating cron %q. %s", in.Name, err)
		return
	}
	c.JSON(200, cron)
}

// GetCronList gets the cron job list from the database and writes
// to the response in json format.
func GetCronList(c *gin.Context) {
	repo := session.Repo(c)
	list, err := store.FromContext(c).CronList(repo)
	if err != nil {
		c.String(500, "Error getting cron list. %s", err)
		return
	}
	c.JSON(200, list)
}

// DeleteCron deletes the named cron job from the database.
func DeleteCron(c *gin.Context) {
	repo := session.Repo(c)
	id, err := strconv.ParseInt(c.Param("cron"), 10, 64)
	if err != nil {
		c.String(400, "Error parsing cron id. %s", err)
		return
	}
	if err := store.FromContext(c).CronDelete(repo, id); err != nil {
		c.String(500, "Error deleting cron %d. %s", id, err)
		return
	}
	c.String(204, "")
}
