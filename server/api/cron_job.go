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

	"github.com/gin-gonic/gin"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// GetCronJob gets the cron-job by id from the database and writes
// to the response in json format.
func GetCronJob(c *gin.Context) {
	repo := session.Repo(c)
	id, err := strconv.ParseInt(c.Param("cron"), 10, 64)
	if err != nil {
		c.String(400, "Error parsing cron-job id. %s", err)
		return
	}

	cronJob, err := store.FromContext(c).CronFind(repo, id)
	if err != nil {
		c.String(404, "Error getting cron-job %q. %s", id, err)
		return
	}
	c.JSON(200, cronJob)
}

// PostCronJob persists the cron-job to the database.
func PostCronJob(c *gin.Context) {
	repo := session.Repo(c)

	in := new(model.CronJob)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing request. %s", err)
		return
	}
	cronJob := &model.CronJob{
		RepoID:    repo.ID,
		Title:     in.Title,
		CreatorID: session.User(c).ID,
		Schedule:  in.Schedule,
		Branch:    in.Branch,
		NextExec:  0, // immediately run for the first time
	}
	if err := cronJob.Validate(); err != nil {
		c.String(400, "Error inserting cron-job. %s", err)
		return
	}

	// TODO: check if branch exists on remote

	if err := store.FromContext(c).CronCreate(cronJob); err != nil {
		c.String(500, "Error inserting cron-job %q. %s", in.Title, err)
		return
	}
	c.JSON(200, in)
}

// PatchCronJob updates the cron-job in the database.
func PatchCronJob(c *gin.Context) {
	repo := session.Repo(c)
	store := store.FromContext(c)
	id, err := strconv.ParseInt(c.Param("cron"), 10, 64)
	if err != nil {
		c.String(400, "Error parsing cron-job id. %s", err)
		return
	}

	in := new(model.CronJob)
	err = c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing request. %s", err)
		return
	}

	cronJob, err := store.CronFind(repo, id)
	if err != nil {
		c.String(404, "Error getting cron-job %d. %s", id, err)
		return
	}
	cronJob.NextExec = 0
	if in.Branch != "" {
		// TODO: check if branch exists on remote
		cronJob.Branch = in.Branch
	}
	if in.Schedule != "" {
		cronJob.Schedule = in.Schedule
	}
	if in.Title != "" {
		cronJob.Title = in.Title
	}

	if err := cronJob.Validate(); err != nil {
		c.String(400, "Error updating cron-job. %s", err)
		return
	}
	if err := store.CronUpdate(repo, cronJob); err != nil {
		c.String(500, "Error updating cron-job %q. %s", in.Title, err)
		return
	}
	c.JSON(200, in)
}

// GetCronJobList gets the cron-job list from the database and writes
// to the response in json format.
func GetCronJobList(c *gin.Context) {
	repo := session.Repo(c)
	list, err := store.FromContext(c).CronList(repo)
	if err != nil {
		c.String(500, "Error getting cron-job list. %s", err)
		return
	}
	c.JSON(200, list)
}

// DeleteCronJob deletes the named cron-job from the database.
func DeleteCronJob(c *gin.Context) {
	repo := session.Repo(c)
	id, err := strconv.ParseInt(c.Param("cron"), 10, 64)
	if err != nil {
		c.String(400, "Error parsing cron-job id. %s", err)
		return
	}
	if err := store.FromContext(c).CronDelete(repo, id); err != nil {
		c.String(500, "Error deleting cron-job %d. %s", id, err)
		return
	}
	c.String(204, "")
}
