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

package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/laszlocph/drone-oss-08/cncd/queue"
	"github.com/laszlocph/drone-oss-08/remote"
	"github.com/laszlocph/drone-oss-08/shared/httputil"
	"github.com/laszlocph/drone-oss-08/store"

	"github.com/laszlocph/drone-oss-08/model"
	"github.com/laszlocph/drone-oss-08/router/middleware/session"
)

func GetBuilds(c *gin.Context) {
	repo := session.Repo(c)
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	builds, err := store.GetBuildList(c, repo, page)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, builds)
}

func GetBuild(c *gin.Context) {
	if c.Param("number") == "latest" {
		GetBuildLast(c)
		return
	}

	repo := session.Repo(c)
	num, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	build, err := store.GetBuildNumber(c, repo, num)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	files, _ := store.FromContext(c).FileList(build)
	procs, _ := store.FromContext(c).ProcList(build)
	build.Procs = model.Tree(procs)
	build.Files = files

	c.JSON(http.StatusOK, build)
}

func GetBuildLast(c *gin.Context) {
	repo := session.Repo(c)
	branch := c.DefaultQuery("branch", repo.Branch)

	build, err := store.GetBuildLast(c, repo, branch)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	procs, _ := store.FromContext(c).ProcList(build)
	build.Procs = model.Tree(procs)
	c.JSON(http.StatusOK, build)
}

func GetBuildLogs(c *gin.Context) {
	repo := session.Repo(c)

	// parse the build number and job sequence number from
	// the repquest parameter.
	num, _ := strconv.Atoi(c.Params.ByName("number"))
	ppid, _ := strconv.Atoi(c.Params.ByName("pid"))
	name := c.Params.ByName("proc")

	build, err := store.GetBuildNumber(c, repo, num)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	proc, err := store.FromContext(c).ProcChild(build, ppid, name)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	rc, err := store.FromContext(c).LogFind(proc)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	defer rc.Close()

	c.Header("Content-Type", "application/json")
	io.Copy(c.Writer, rc)
}

func GetProcLogs(c *gin.Context) {
	repo := session.Repo(c)

	// parse the build number and job sequence number from
	// the repquest parameter.
	num, _ := strconv.Atoi(c.Params.ByName("number"))
	pid, _ := strconv.Atoi(c.Params.ByName("pid"))

	build, err := store.GetBuildNumber(c, repo, num)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	proc, err := store.FromContext(c).ProcFind(build, pid)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	rc, err := store.FromContext(c).LogFind(proc)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	defer rc.Close()

	c.Header("Content-Type", "application/json")
	io.Copy(c.Writer, rc)
}

func DeleteBuild(c *gin.Context) {
	repo := session.Repo(c)

	// parse the build number and job sequence number from
	// the repquest parameter.
	num, _ := strconv.Atoi(c.Params.ByName("number"))
	seq, _ := strconv.Atoi(c.Params.ByName("job"))

	build, err := store.GetBuildNumber(c, repo, num)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	proc, err := store.FromContext(c).ProcFind(build, seq)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	if proc.State != model.StatusRunning {
		c.String(400, "Cannot cancel a non-running build")
		return
	}

	proc.State = model.StatusKilled
	proc.Stopped = time.Now().Unix()
	if proc.Started == 0 {
		proc.Started = proc.Stopped
	}
	proc.ExitCode = 137
	// TODO cancel child procs
	store.FromContext(c).ProcUpdate(proc)

	Config.Services.Queue.Error(context.Background(), fmt.Sprint(proc.ID), queue.ErrCancel)
	c.String(204, "")
}

// ZombieKill kills zombie processes stuck in an infinite pending
// or running state. This can only be invoked by administrators and
// may have negative effects.
func ZombieKill(c *gin.Context) {
	repo := session.Repo(c)

	// parse the build number and job sequence number from
	// the repquest parameter.
	num, _ := strconv.Atoi(c.Params.ByName("number"))

	build, err := store.GetBuildNumber(c, repo, num)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	procs, err := store.FromContext(c).ProcList(build)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	if build.Status != model.StatusRunning {
		c.String(400, "Cannot force cancel a non-running build")
		return
	}

	for _, proc := range procs {
		if proc.Running() {
			proc.State = model.StatusKilled
			proc.ExitCode = 137
			proc.Stopped = time.Now().Unix()
			if proc.Started == 0 {
				proc.Started = proc.Stopped
			}
		}
	}

	for _, proc := range procs {
		store.FromContext(c).ProcUpdate(proc)
		Config.Services.Queue.Error(context.Background(), fmt.Sprint(proc.ID), queue.ErrCancel)
	}

	build.Status = model.StatusKilled
	build.Finished = time.Now().Unix()
	store.FromContext(c).UpdateBuild(build)

	c.String(204, "")
}

func PostApproval(c *gin.Context) {
	var (
		remote_ = remote.FromContext(c)
		repo    = session.Repo(c)
		user    = session.User(c)
		num, _  = strconv.Atoi(
			c.Params.ByName("number"),
		)
	)

	build, err := store.GetBuildNumber(c, repo, num)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}
	if build.Status != model.StatusBlocked {
		c.String(500, "cannot decline a build with status %s", build.Status)
		return
	}
	build.Status = model.StatusPending
	build.Reviewed = time.Now().Unix()
	build.Reviewer = user.Login

	// fetch the build file from the database
	configs, err := Config.Storage.Config.ConfigLoad(build.ID)
	if err != nil {
		logrus.Errorf("failure to get build config for %s. %s", repo.FullName, err)
		c.AbortWithError(404, err)
		return
	}

	netrc, err := remote_.Netrc(user, repo)
	if err != nil {
		c.String(500, "Failed to generate netrc file. %s", err)
		return
	}

	if uerr := store.UpdateBuild(c, build); err != nil {
		c.String(500, "error updating build. %s", uerr)
		return
	}

	c.JSON(200, build)

	// get the previous build so that we can send
	// on status change notifications
	last, _ := store.GetBuildLastBefore(c, repo, build.Branch, build.ID)
	secs, err := Config.Services.Secrets.SecretListBuild(repo, build)
	if err != nil {
		logrus.Debugf("Error getting secrets for %s#%d. %s", repo.FullName, build.Number, err)
	}
	regs, err := Config.Services.Registries.RegistryList(repo)
	if err != nil {
		logrus.Debugf("Error getting registry credentials for %s#%d. %s", repo.FullName, build.Number, err)
	}
	envs := map[string]string{}
	if Config.Services.Environ != nil {
		globals, _ := Config.Services.Environ.EnvironList(repo)
		for _, global := range globals {
			envs[global.Name] = global.Value
		}
	}

	defer func() {
		uri := fmt.Sprintf("%s/%s/%d", httputil.GetURL(c.Request), repo.FullName, build.Number)
		err = remote_.Status(user, repo, build, uri)
		if err != nil {
			logrus.Errorf("error setting commit status for %s/%d: %v", repo.FullName, build.Number, err)
		}
	}()

	var yamls []string
	for _, y := range configs {
		yamls = append(yamls, string(y.Data))
	}

	b := procBuilder{
		Repo:  repo,
		Curr:  build,
		Last:  last,
		Netrc: netrc,
		Secs:  secs,
		Regs:  regs,
		Link:  httputil.GetURL(c.Request),
		Yamls: yamls,
		Envs:  envs,
	}
	buildItems, err := b.Build()
	if err != nil {
		build.Status = model.StatusError
		build.Started = time.Now().Unix()
		build.Finished = build.Started
		build.Error = err.Error()
		store.UpdateBuild(c, build)
		return
	}

	err = store.FromContext(c).ProcCreate(build.Procs)
	if err != nil {
		logrus.Errorf("error persisting procs %s/%d: %s", repo.FullName, build.Number, err)
	}

	publishToTopic(c, build, repo)
	queueBuild(build, repo, buildItems)
}

func PostDecline(c *gin.Context) {
	var (
		remote_ = remote.FromContext(c)
		repo    = session.Repo(c)
		user    = session.User(c)
		num, _  = strconv.Atoi(
			c.Params.ByName("number"),
		)
	)

	build, err := store.GetBuildNumber(c, repo, num)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}
	if build.Status != model.StatusBlocked {
		c.String(500, "cannot decline a build with status %s", build.Status)
		return
	}
	build.Status = model.StatusDeclined
	build.Reviewed = time.Now().Unix()
	build.Reviewer = user.Login

	err = store.UpdateBuild(c, build)
	if err != nil {
		c.String(500, "error updating build. %s", err)
		return
	}

	uri := fmt.Sprintf("%s/%s/%d", httputil.GetURL(c.Request), repo.FullName, build.Number)
	err = remote_.Status(user, repo, build, uri)
	if err != nil {
		logrus.Errorf("error setting commit status for %s/%d: %v", repo.FullName, build.Number, err)
	}

	c.JSON(200, build)
}

func GetBuildQueue(c *gin.Context) {
	out, err := store.GetBuildQueue(c)
	if err != nil {
		c.String(500, "Error getting build queue. %s", err)
		return
	}
	c.JSON(200, out)
}

// PostBuild restarts a build
func PostBuild(c *gin.Context) {
	remote_ := remote.FromContext(c)
	repo := session.Repo(c)

	num, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := store.GetUser(c, repo.UserID)
	if err != nil {
		logrus.Errorf("failure to find repo owner %s. %s", repo.FullName, err)
		c.AbortWithError(500, err)
		return
	}

	build, err := store.GetBuildNumber(c, repo, num)
	if err != nil {
		logrus.Errorf("failure to get build %d. %s", num, err)
		c.AbortWithError(404, err)
		return
	}

	switch build.Status {
	case model.StatusDeclined,
		model.StatusBlocked:
		c.String(500, "cannot restart a build with status %s", build.Status)
		return
	}

	// if the remote has a refresh token, the current access token
	// may be stale. Therefore, we should refresh prior to dispatching
	// the job.
	if refresher, ok := remote_.(remote.Refresher); ok {
		ok, _ := refresher.Refresh(user)
		if ok {
			store.UpdateUser(c, user)
		}
	}

	// fetch the .drone.yml file from the database
	configs, err := Config.Storage.Config.ConfigLoad(build.ID)
	if err != nil {
		logrus.Errorf("failure to get build config for %s. %s", repo.FullName, err)
		c.AbortWithError(404, err)
		return
	}

	netrc, err := remote_.Netrc(user, repo)
	if err != nil {
		logrus.Errorf("failure to generate netrc for %s. %s", repo.FullName, err)
		c.AbortWithError(500, err)
		return
	}

	build.ID = 0
	build.Number = 0
	build.Parent = num
	build.Status = model.StatusPending
	build.Started = 0
	build.Finished = 0
	build.Enqueued = time.Now().UTC().Unix()
	build.Error = ""
	build.Deploy = c.DefaultQuery("deploy_to", build.Deploy)

	event := c.DefaultQuery("event", build.Event)
	if event == model.EventPush ||
		event == model.EventPull ||
		event == model.EventTag ||
		event == model.EventDeploy {
		build.Event = event
	}

	err = store.CreateBuild(c, build)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	// Read query string parameters into buildParams, exclude reserved params
	var buildParams = map[string]string{}
	for key, val := range c.Request.URL.Query() {
		switch key {
		case "fork", "event", "deploy_to":
		default:
			// We only accept string literals, because build parameters will be
			// injected as environment variables
			buildParams[key] = val[0]
		}
	}

	// get the previous build so that we can send
	// on status change notifications
	last, _ := store.GetBuildLastBefore(c, repo, build.Branch, build.ID)
	secs, err := Config.Services.Secrets.SecretListBuild(repo, build)
	if err != nil {
		logrus.Debugf("Error getting secrets for %s#%d. %s", repo.FullName, build.Number, err)
	}
	regs, err := Config.Services.Registries.RegistryList(repo)
	if err != nil {
		logrus.Debugf("Error getting registry credentials for %s#%d. %s", repo.FullName, build.Number, err)
	}
	if Config.Services.Environ != nil {
		globals, _ := Config.Services.Environ.EnvironList(repo)
		for _, global := range globals {
			buildParams[global.Name] = global.Value
		}
	}

	var yamls []string
	for _, y := range configs {
		yamls = append(yamls, string(y.Data))
	}

	b := procBuilder{
		Repo:  repo,
		Curr:  build,
		Last:  last,
		Netrc: netrc,
		Secs:  secs,
		Regs:  regs,
		Link:  httputil.GetURL(c.Request),
		Yamls: yamls,
		Envs:  buildParams,
	}
	buildItems, err := b.Build()
	if err != nil {
		build.Status = model.StatusError
		build.Started = time.Now().Unix()
		build.Finished = build.Started
		build.Error = err.Error()
		c.JSON(500, build)
		return
	}

	err = store.FromContext(c).ProcCreate(build.Procs)
	if err != nil {
		logrus.Errorf("cannot restart %s#%d: %s", repo.FullName, build.Number, err)
		build.Status = model.StatusError
		build.Started = time.Now().Unix()
		build.Finished = build.Started
		build.Error = err.Error()
		c.JSON(500, build)
		return
	}
	c.JSON(202, build)

	publishToTopic(c, build, repo)
	queueBuild(build, repo, buildItems)
}

func DeleteBuildLogs(c *gin.Context) {
	repo := session.Repo(c)
	user := session.User(c)
	num, _ := strconv.Atoi(c.Params.ByName("number"))

	build, err := store.GetBuildNumber(c, repo, num)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	procs, err := store.FromContext(c).ProcList(build)
	if err != nil {
		c.AbortWithError(404, err)
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
		lerr := store.FromContext(c).LogSave(proc, buf)
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
