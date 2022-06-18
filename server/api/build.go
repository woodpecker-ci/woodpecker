// Copyright 2018 Drone.IO Inc.
// Copyright 2021 Informatyka Boguslawski sp. z o.o. sp.k., http://www.ib.pl/
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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/pipeline"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

func GetBuilds(c *gin.Context) {
	repo := session.Repo(c)
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	builds, err := store.FromContext(c).GetBuildList(repo, page)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, builds)
}

func GetBuild(c *gin.Context) {
	_store := store.FromContext(c)
	if c.Param("number") == "latest" {
		GetBuildLast(c)
		return
	}

	repo := session.Repo(c)
	num, err := strconv.ParseInt(c.Param("number"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	build, err := _store.GetBuildNumber(repo, num)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	files, _ := _store.FileList(build)
	procs, _ := _store.ProcList(build)
	if build.Procs, err = model.Tree(procs); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	build.Files = files

	c.JSON(http.StatusOK, build)
}

func GetBuildLast(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)
	branch := c.DefaultQuery("branch", repo.Branch)

	build, err := _store.GetBuildLast(repo, branch)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	procs, err := _store.ProcList(build)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if build.Procs, err = model.Tree(procs); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, build)
}

func GetBuildLogs(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)

	// parse the build number and job sequence number from
	// the request parameter.
	num, _ := strconv.ParseInt(c.Params.ByName("number"), 10, 64)
	ppid, _ := strconv.Atoi(c.Params.ByName("pid"))
	name := c.Params.ByName("proc")

	build, err := _store.GetBuildNumber(repo, num)
	if err != nil {
		_ = c.AbortWithError(404, err)
		return
	}

	proc, err := _store.ProcChild(build, ppid, name)
	if err != nil {
		_ = c.AbortWithError(404, err)
		return
	}

	rc, err := _store.LogFind(proc)
	if err != nil {
		_ = c.AbortWithError(404, err)
		return
	}

	defer rc.Close()

	c.Header("Content-Type", "application/json")
	if _, err := io.Copy(c.Writer, rc); err != nil {
		log.Error().Err(err).Msg("could not copy log to http response")
	}
}

func GetProcLogs(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)

	// parse the build number and job sequence number from
	// the request parameter.
	num, _ := strconv.ParseInt(c.Params.ByName("number"), 10, 64)
	pid, _ := strconv.Atoi(c.Params.ByName("pid"))

	build, err := _store.GetBuildNumber(repo, num)
	if err != nil {
		_ = c.AbortWithError(http.StatusNotFound, err)
		return
	}

	proc, err := _store.ProcFind(build, pid)
	if err != nil {
		_ = c.AbortWithError(http.StatusNotFound, err)
		return
	}

	rc, err := _store.LogFind(proc)
	if err != nil {
		_ = c.AbortWithError(http.StatusNotFound, err)
		return
	}

	defer rc.Close()

	c.Header("Content-Type", "application/json")
	if _, err := io.Copy(c.Writer, rc); err != nil {
		log.Error().Err(err).Msg("could not copy log to http response")
	}
}

func GetBuildConfig(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)
	num, err := strconv.ParseInt(c.Param("number"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	build, err := _store.GetBuildNumber(repo, num)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	configs, err := _store.ConfigsForBuild(build.ID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, configs)
}

// DeleteBuild cancels a build
func DeleteBuild(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)
	num, _ := strconv.ParseInt(c.Params.ByName("number"), 10, 64)

	build, err := _store.GetBuildNumber(repo, num)
	if err != nil {
		_ = c.AbortWithError(http.StatusNotFound, err)
		return
	}

	if err := pipeline.Cancel(c, _store, repo, build); err != nil {
		handlePipelineErr(c, err)
	} else {
		c.Status(http.StatusNoContent)
	}
}

// PostApproval start pipelines in gated repos
func PostApproval(c *gin.Context) {
	var (
		_store = store.FromContext(c)
		repo   = session.Repo(c)
		user   = session.User(c)
		num, _ = strconv.ParseInt(c.Params.ByName("number"), 10, 64)
	)

	build, err := _store.GetBuildNumber(repo, num)
	if err != nil {
		_ = c.AbortWithError(404, err)
		return
	}

	newBuild, err := pipeline.Approve(c, _store, build, user, repo)
	if err != nil {
		handlePipelineErr(c, err)
	} else {
		c.JSON(200, newBuild)
	}
}

// PostDecline decline pipelines in gated repos
func PostDecline(c *gin.Context) {
	var (
		_store = store.FromContext(c)
		repo   = session.Repo(c)
		user   = session.User(c)
		num, _ = strconv.ParseInt(c.Params.ByName("number"), 10, 64)
	)

	build, err := _store.GetBuildNumber(repo, num)
	if err != nil {
		c.String(http.StatusNotFound, "%v", err)
		return
	}

	build, err = pipeline.Decline(c, _store, build, user, repo)
	if err != nil {
		handlePipelineErr(c, err)
	} else {
		c.JSON(200, build)
	}
}

func GetBuildQueue(c *gin.Context) {
	out, err := store.FromContext(c).GetBuildQueue()
	if err != nil {
		c.String(500, "Error getting build queue. %s", err)
		return
	}
	c.JSON(200, out)
}

// PostBuild restarts a build optional with altered event, deploy or environment
func PostBuild(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)

	num, err := strconv.ParseInt(c.Param("number"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := _store.GetUser(repo.UserID)
	if err != nil {
		log.Error().Msgf("failure to find repo owner %s. %s", repo.FullName, err)
		_ = c.AbortWithError(500, err)
		return
	}

	build, err := _store.GetBuildNumber(repo, num)
	if err != nil {
		log.Error().Msgf("failure to get build %d. %s", num, err)
		_ = c.AbortWithError(404, err)
		return
	}

	// refresh the token to make sure, pipeline.ReStart can still obtain the pipeline config if nessessary again
	refreshUserToken(c, user)

	// make Deploy overridable
	build.Deploy = c.DefaultQuery("deploy_to", build.Deploy)

	// make Event overridable
	if event, ok := c.GetQuery("event"); ok {
		build.Event = model.WebhookEvent(event)

		if !model.ValidateWebhookEvent(build.Event) {
			msg := fmt.Sprintf("build event '%s' is invalid", event)
			c.String(http.StatusBadRequest, msg)
			return
		}
	}

	// Read query string parameters into buildParams, exclude reserved params
	envs := map[string]string{}
	for key, val := range c.Request.URL.Query() {
		switch key {
		// Skip some options of the endpoint
		case "fork", "event", "deploy_to":
			continue
		default:
			// We only accept string literals, because build parameters will be
			// injected as environment variables
			// TODO: sanitize the value
			envs[key] = val[0]
		}
	}

	newBuild, err := pipeline.Restart(c, _store, build, user, repo, envs)
	if err != nil {
		handlePipelineErr(c, err)
	} else {
		c.JSON(200, newBuild)
	}
}

func DeleteBuildLogs(c *gin.Context) {
	_store := store.FromContext(c)

	repo := session.Repo(c)
	user := session.User(c)
	num, _ := strconv.ParseInt(c.Params.ByName("number"), 10, 64)

	build, err := _store.GetBuildNumber(repo, num)
	if err != nil {
		_ = c.AbortWithError(404, err)
		return
	}

	procs, err := _store.ProcList(build)
	if err != nil {
		_ = c.AbortWithError(404, err)
		return
	}

	switch build.Status {
	case model.StatusRunning, model.StatusPending:
		c.String(400, "Cannot delete logs for a pending or running build")
		return
	}

	for _, proc := range procs {
		t := time.Now().UTC()
		buf := bytes.NewBufferString(fmt.Sprintf(deleteStr, proc.Name, user.Login, t.Format(time.UnixDate)))
		lerr := _store.LogSave(proc, buf)
		if lerr != nil {
			err = lerr
		}
	}
	if err != nil {
		c.String(400, "There was a problem deleting your logs. %s", err)
		return
	}

	c.String(204, "")
}

var deleteStr = `[
	{
		"proc": %q,
		"pos": 0,
		"out": "logs purged by %s on %s\n"
	}
]`
