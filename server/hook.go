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
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Sirupsen/logrus"
	"github.com/laszlocph/drone-oss-08/model"
	"github.com/laszlocph/drone-oss-08/remote"
	"github.com/laszlocph/drone-oss-08/shared/httputil"
	"github.com/laszlocph/drone-oss-08/shared/token"
	"github.com/laszlocph/drone-oss-08/store"

	"github.com/laszlocph/drone-oss-08/cncd/pipeline/pipeline/frontend/yaml"
	"github.com/laszlocph/drone-oss-08/cncd/pipeline/pipeline/rpc"
	"github.com/laszlocph/drone-oss-08/cncd/pubsub"
	"github.com/laszlocph/drone-oss-08/cncd/queue"
)

var skipRe = regexp.MustCompile(`\[(?i:ci *skip|skip *ci)\]`)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetQueueInfo(c *gin.Context) {
	c.IndentedJSON(200,
		Config.Services.Queue.Info(c),
	)
}

func PostHook(c *gin.Context) {
	remote_ := remote.FromContext(c)

	tmprepo, build, err := remote_.Hook(c.Request)
	if err != nil {
		logrus.Errorf("failure to parse hook. %s", err)
		c.AbortWithError(400, err)
		return
	}
	if build == nil {
		c.Writer.WriteHeader(200)
		return
	}
	if tmprepo == nil {
		logrus.Errorf("failure to ascertain repo from hook.")
		c.Writer.WriteHeader(400)
		return
	}

	// skip the build if any case-insensitive combination of the words "skip" and "ci"
	// wrapped in square brackets appear in the commit message
	skipMatch := skipRe.FindString(build.Message)
	if len(skipMatch) > 0 {
		logrus.Infof("ignoring hook. %s found in %s", skipMatch, build.Commit)
		c.Writer.WriteHeader(204)
		return
	}

	repo, err := store.GetRepoOwnerName(c, tmprepo.Owner, tmprepo.Name)
	if err != nil {
		logrus.Errorf("failure to find repo %s/%s from hook. %s", tmprepo.Owner, tmprepo.Name, err)
		c.AbortWithError(404, err)
		return
	}
	if !repo.IsActive {
		logrus.Errorf("ignoring hook. %s/%s is inactive.", tmprepo.Owner, tmprepo.Name)
		c.AbortWithError(204, err)
		return
	}

	// get the token and verify the hook is authorized
	parsed, err := token.ParseRequest(c.Request, func(t *token.Token) (string, error) {
		return repo.Hash, nil
	})
	if err != nil {
		logrus.Errorf("failure to parse token from hook for %s. %s", repo.FullName, err)
		c.AbortWithError(400, err)
		return
	}
	if parsed.Text != repo.FullName {
		logrus.Errorf("failure to verify token from hook. Expected %s, got %s", repo.FullName, parsed.Text)
		c.AbortWithStatus(403)
		return
	}

	if repo.UserID == 0 {
		logrus.Warnf("ignoring hook. repo %s has no owner.", repo.FullName)
		c.Writer.WriteHeader(204)
		return
	}
	var skipped = true
	if (build.Event == model.EventPush && repo.AllowPush) ||
		(build.Event == model.EventPull && repo.AllowPull) ||
		(build.Event == model.EventDeploy && repo.AllowDeploy) ||
		(build.Event == model.EventTag && repo.AllowTag) {
		skipped = false
	}

	if skipped {
		logrus.Infof("ignoring hook. repo %s is disabled for %s events.", repo.FullName, build.Event)
		c.Writer.WriteHeader(204)
		return
	}

	user, err := store.GetUser(c, repo.UserID)
	if err != nil {
		logrus.Errorf("failure to find repo owner %s. %s", repo.FullName, err)
		c.AbortWithError(500, err)
		return
	}

	// if the remote has a refresh token, the current access token
	// may be stale. Therefore, we should refresh prior to dispatching
	// the build.
	if refresher, ok := remote_.(remote.Refresher); ok {
		ok, _ := refresher.Refresh(user)
		if ok {
			store.UpdateUser(c, user)
		}
	}

	// fetch the build file from the remote
	remoteYamlConfig, err := remote.FileBackoff(remote_, user, repo, build, repo.Config)
	if err != nil {
		logrus.Errorf("error: %s: cannot find %s in %s: %s", repo.FullName, repo.Config, build.Ref, err)
		c.AbortWithError(404, err)
		return
	}
	conf, err := findOrPersistPipelineConfig(repo, remoteYamlConfig)
	if err != nil {
		logrus.Errorf("failure to find or persist build config for %s. %s", repo.FullName, err)
		c.AbortWithError(500, err)
		return
	}
	build.ConfigID = conf.ID

	netrc, err := remote_.Netrc(user, repo)
	if err != nil {
		c.String(500, "Failed to generate netrc file. %s", err)
		return
	}

	// verify the branches can be built vs skipped
	parsedPipelineConfig, err := yaml.ParseString(conf.Data)
	if err == nil {
		if !parsedPipelineConfig.Branches.Match(build.Branch) && build.Event != model.EventTag && build.Event != model.EventDeploy {
			c.String(200, "Branch does not match restrictions defined in yaml")
			return
		}
	}

	// update some build fields
	build.RepoID = repo.ID
	build.Verified = true
	build.Status = model.StatusPending

	if repo.IsGated {
		allowed, _ := Config.Services.Senders.SenderAllowed(user, repo, build, conf)
		if !allowed {
			build.Status = model.StatusBlocked
		}
	}

	err = store.CreateBuild(c, build, build.Procs...)
	if err != nil {
		logrus.Errorf("failure to save commit for %s. %s", repo.FullName, err)
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, build)

	if build.Status == model.StatusBlocked {
		return
	}

	envs := map[string]string{}
	if Config.Services.Environ != nil {
		globals, _ := Config.Services.Environ.EnvironList(repo)
		for _, global := range globals {
			envs[global.Name] = global.Value
		}
	}

	secs, err := Config.Services.Secrets.SecretListBuild(repo, build)
	if err != nil {
		logrus.Debugf("Error getting secrets for %s#%d. %s", repo.FullName, build.Number, err)
	}

	regs, err := Config.Services.Registries.RegistryList(repo)
	if err != nil {
		logrus.Debugf("Error getting registry credentials for %s#%d. %s", repo.FullName, build.Number, err)
	}

	// get the previous build so that we can send
	// on status change notifications
	last, _ := store.GetBuildLastBefore(c, repo, build.Branch, build.ID)

	//
	// BELOW: NEW
	//

	defer func() {
		uri := fmt.Sprintf("%s/%s/%d", httputil.GetURL(c.Request), repo.FullName, build.Number)
		err = remote_.Status(user, repo, build, uri)
		if err != nil {
			logrus.Errorf("error setting commit status for %s/%d: %v", repo.FullName, build.Number, err)
		}
	}()

	b := procBuilder{
		Repo:  repo,
		Curr:  build,
		Last:  last,
		Netrc: netrc,
		Secs:  secs,
		Regs:  regs,
		Envs:  envs,
		Link:  httputil.GetURL(c.Request),
		Yaml:  conf.Data,
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

	setBuildProcs(build, buildItems)

	err = store.FromContext(c).ProcCreate(build.Procs)
	if err != nil {
		logrus.Errorf("error persisting procs %s/%d: %s", repo.FullName, build.Number, err)
	}

	publishToTopic(c, build, repo)
	queueBuild(build, repo, buildItems)
}

func findOrPersistPipelineConfig(repo *model.Repo, remoteYamlConfig []byte) (*model.Config, error) {
	sha := shasum(remoteYamlConfig)
	conf, err := Config.Storage.Config.ConfigFind(repo, sha)
	if err != nil {
		conf = &model.Config{
			RepoID: repo.ID,
			Data:   string(remoteYamlConfig),
			Hash:   sha,
		}
		err = Config.Storage.Config.ConfigCreate(conf)
		if err != nil {
			// retry in case we receive two hooks at the same time
			conf, err = Config.Storage.Config.ConfigFind(repo, sha)
			if err != nil {
				return nil, err
			}
		}
	}

	return conf, nil
}

func setBuildProcs(build *model.Build, buildItems []*buildItem) {
	pcounter := len(buildItems)
	for _, item := range buildItems {
		build.Procs = append(build.Procs, item.Proc)
		item.Proc.BuildID = build.ID

		for _, stage := range item.Config.Stages {
			var gid int
			for _, step := range stage.Steps {
				pcounter++
				if gid == 0 {
					gid = pcounter
				}
				proc := &model.Proc{
					BuildID: build.ID,
					Name:    step.Alias,
					PID:     pcounter,
					PPID:    item.Proc.PID,
					PGID:    gid,
					State:   model.StatusPending,
				}
				build.Procs = append(build.Procs, proc)
			}
		}
	}
}

func publishToTopic(c *gin.Context, build *model.Build, repo *model.Repo) {
	message := pubsub.Message{
		Labels: map[string]string{
			"repo":    repo.FullName,
			"private": strconv.FormatBool(repo.IsPrivate),
		},
	}
	buildCopy := *build
	buildCopy.Procs = model.Tree(buildCopy.Procs)
	message.Data, _ = json.Marshal(model.Event{
		Type:  model.Enqueued,
		Repo:  *repo,
		Build: buildCopy,
	})
	Config.Services.Pubsub.Publish(c, "topic/events", message)
}

func queueBuild(build *model.Build, repo *model.Repo, buildItems []*buildItem) {
	for _, item := range buildItems {
		task := new(queue.Task)
		task.ID = fmt.Sprint(item.Proc.ID)
		task.Labels = map[string]string{}
		for k, v := range item.Labels {
			task.Labels[k] = v
		}
		task.Labels["platform"] = item.Platform
		task.Labels["repo"] = repo.FullName

		task.Data, _ = json.Marshal(rpc.Pipeline{
			ID:      fmt.Sprint(item.Proc.ID),
			Config:  item.Config,
			Timeout: repo.Timeout,
		})

		Config.Services.Logs.Open(context.Background(), task.ID)
		Config.Services.Queue.Push(context.Background(), task)
	}
}

func shasum(raw []byte) string {
	sum := sha256.Sum256(raw)
	return fmt.Sprintf("%x", sum)
}
