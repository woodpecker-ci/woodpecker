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
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/pubsub"
	"github.com/woodpecker-ci/woodpecker/server/queue"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/shared"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/shared/token"
)

var skipRe = regexp.MustCompile(`\[(?i:ci *skip|skip *ci)\]`)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetQueueInfo(c *gin.Context) {
	c.IndentedJSON(200,
		server.Config.Services.Queue.Info(c),
	)
}

func PauseQueue(c *gin.Context) {
	server.Config.Services.Queue.Pause()
	c.Status(http.StatusOK)
}

func ResumeQueue(c *gin.Context) {
	server.Config.Services.Queue.Resume()
	c.Status(http.StatusOK)
}

func BlockTilQueueHasRunningItem(c *gin.Context) {
	for {
		info := server.Config.Services.Queue.Info(c)
		if info.Stats.Running == 0 {
			break
		}
	}
	c.Status(http.StatusOK)
}

func PostHook(c *gin.Context) {
	_store := store.FromContext(c)

	tmpRepo, build, err := server.Config.Services.Remote.Hook(c, c.Request)
	if err != nil {
		msg := "failure to parse hook"
		log.Debug().Err(err).Msg(msg)
		c.String(http.StatusBadRequest, msg)
		return
	}
	if build == nil {
		msg := "ignoring hook: hook parsing resulted in empty build"
		log.Debug().Msg(msg)
		c.String(http.StatusOK, msg)
		return
	}
	if tmpRepo == nil {
		msg := "failure to ascertain repo from hook"
		log.Debug().Msg(msg)
		c.String(http.StatusBadRequest, msg)
		return
	}

	// skip the build if any case-insensitive combination of the words "skip" and "ci"
	// wrapped in square brackets appear in the commit message
	skipMatch := skipRe.FindString(build.Message)
	if len(skipMatch) > 0 {
		msg := fmt.Sprintf("ignoring hook: %s found in %s", skipMatch, build.Commit)
		log.Debug().Msg(msg)
		c.String(http.StatusNoContent, msg)
		return
	}

	repo, err := _store.GetRepoName(tmpRepo.Owner + "/" + tmpRepo.Name)
	if err != nil {
		msg := fmt.Sprintf("failure to get repo %s from store", tmpRepo.FullName)
		log.Error().Err(err).Msg(msg)
		c.String(http.StatusNotFound, msg)
		return
	}
	if !repo.IsActive {
		msg := fmt.Sprintf("ignoring hook: repo %s is inactive", tmpRepo.FullName)
		log.Debug().Msg(msg)
		c.String(http.StatusNoContent, msg)
		return
	}

	// get the token and verify the hook is authorized
	parsed, err := token.ParseRequest(c.Request, func(_ *token.Token) (string, error) {
		return repo.Hash, nil
	})
	if err != nil {
		msg := fmt.Sprintf("failure to parse token from hook for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		c.String(http.StatusBadRequest, msg)
		return
	}
	if parsed.Text != repo.FullName {
		msg := fmt.Sprintf("failure to verify token from hook. Expected %s, got %s", repo.FullName, parsed.Text)
		log.Debug().Msg(msg)
		c.String(http.StatusForbidden, msg)
		return
	}

	if repo.UserID == 0 {
		msg := fmt.Sprintf("ignoring hook. repo %s has no owner.", repo.FullName)
		log.Warn().Msg(msg)
		c.String(http.StatusNoContent, msg)
		return
	}

	if build.Event == model.EventPull && !repo.AllowPull {
		msg := "ignoring hook: pull requests are disabled for this repo in woodpecker"
		log.Debug().Str("repo", repo.FullName).Msg(msg)
		c.String(http.StatusNoContent, msg)
		return
	}

	repoUser, err := _store.GetUser(repo.UserID)
	if err != nil {
		msg := fmt.Sprintf("failure to find repo owner via id '%d'", repo.UserID)
		log.Error().Err(err).Str("repo", repo.FullName).Msg(msg)
		c.String(http.StatusInternalServerError, msg)
		return
	}

	// if the remote has a refresh token, the current access token
	// may be stale. Therefore, we should refresh prior to dispatching
	// the build.
	if refresher, ok := server.Config.Services.Remote.(remote.Refresher); ok {
		refreshed, err := refresher.Refresh(c, repoUser)
		if err != nil {
			log.Error().Err(err).Msgf("failed to refresh oauth2 token for repoUser: %s", repoUser.Login)
		} else if refreshed {
			if err := _store.UpdateUser(repoUser); err != nil {
				log.Error().Err(err).Msgf("error while updating repoUser: %s", repoUser.Login)
				// move forward
			}
		}
	}

	// fetch the build file from the remote
	configFetcher := shared.NewConfigFetcher(server.Config.Services.Remote, repoUser, repo, build)
	remoteYamlConfigs, err := configFetcher.Fetch(c)
	if err != nil {
		msg := fmt.Sprintf("cannot find config '%s' in '%s' with user: '%s'", repo.Config, build.Ref, repoUser.Login)
		log.Debug().Err(err).Str("repo", repo.FullName).Msg(msg)
		c.String(http.StatusNotFound, msg)
		return
	}

	filtered, err := branchFiltered(build, remoteYamlConfigs)
	if err != nil {
		msg := "failure to parse yaml from hook"
		log.Debug().Err(err).Str("repo", repo.FullName).Msg(msg)
		c.String(http.StatusBadRequest, msg)
		// No return here, we create the entry in the database to show a failed run in the woodpecker webui
	}
	if filtered {
		msg := "ignoring hook: branch does not match restrictions defined in yaml"
		log.Debug().Str("repo", repo.FullName).Msg(msg)
		c.String(http.StatusOK, msg)
		return
	}

	if zeroSteps(build, remoteYamlConfigs) {
		msg := "ignoring hook: step conditions yield zero runnable steps"
		log.Debug().Str("repo", repo.FullName).Msg(msg)
		c.String(http.StatusOK, msg)
		return
	}

	// update some build fields
	build.RepoID = repo.ID
	build.Verified = true
	build.Status = model.StatusPending

	// TODO(336) extend gated feature with an allow/block List
	if repo.IsGated {
		build.Status = model.StatusBlocked
	}

	err = _store.CreateBuild(build, build.Procs...)
	if err != nil {
		msg := fmt.Sprintf("failure to save build for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		c.String(http.StatusInternalServerError, msg)
		return
	}

	// persist the build config for historical correctness, restarts, etc
	for _, remoteYamlConfig := range remoteYamlConfigs {
		_, err := findOrPersistPipelineConfig(_store, build, remoteYamlConfig)
		if err != nil {
			msg := fmt.Sprintf("failure to find or persist pipeline config for %s", repo.FullName)
			log.Error().Err(err).Msg(msg)
			c.String(http.StatusInternalServerError, msg)
			return
		}
	}

	build, buildItems, err := createBuildItems(c, _store, build, repoUser, repo, remoteYamlConfigs, nil)
	if err != nil {
		msg := fmt.Sprintf("failure to createBuildItems for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		c.String(http.StatusInternalServerError, msg)
		return
	}

	if build.Status == model.StatusBlocked {
		if err := publishToTopic(c, build, repo); err != nil {
			log.Error().Err(err).Msg("publishToTopic")
		}

		if err := updateBuildStatus(c, build, repo, repoUser); err != nil {
			log.Error().Err(err).Msg("updateBuildStatus")
		}

		c.JSON(http.StatusOK, build)
		return
	}

	build, err = startBuild(c, _store, build, repoUser, repo, buildItems)
	if err != nil {
		msg := fmt.Sprintf("failure to start build for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		c.String(http.StatusInternalServerError, msg)
		return
	}

	c.JSON(http.StatusOK, build)
}

// TODO: parse yaml once and not for each filter function
func branchFiltered(build *model.Build, remoteYamlConfigs []*remote.FileMeta) (bool, error) {
	log.Trace().Msgf("hook.branchFiltered(): build branch: '%s' build event: '%s' config count: %d", build.Branch, build.Event, len(remoteYamlConfigs))

	if build.Event == model.EventTag || build.Event == model.EventDeploy {
		return false, nil
	}

	for _, remoteYamlConfig := range remoteYamlConfigs {
		parsedPipelineConfig, err := yaml.ParseBytes(remoteYamlConfig.Data)
		if err != nil {
			log.Trace().Msgf("parse config '%s': %s", remoteYamlConfig.Name, err)
			return false, err
		}
		log.Trace().Msgf("config '%s': %#v", remoteYamlConfig.Name, parsedPipelineConfig)

		if parsedPipelineConfig.Branches.Match(build.Branch) {
			return false, nil
		}
	}

	return true, nil
}

func zeroSteps(build *model.Build, remoteYamlConfigs []*remote.FileMeta) bool {
	b := shared.ProcBuilder{
		Repo:  &model.Repo{},
		Curr:  build,
		Last:  &model.Build{},
		Netrc: &model.Netrc{},
		Secs:  []*model.Secret{},
		Regs:  []*model.Registry{},
		Link:  "",
		Yamls: remoteYamlConfigs,
	}

	buildItems, err := b.Build()
	if err != nil {
		return false
	}
	if len(buildItems) == 0 {
		return true
	}

	return false
}

func findOrPersistPipelineConfig(store store.Store, build *model.Build, remoteYamlConfig *remote.FileMeta) (*model.Config, error) {
	sha := shasum(remoteYamlConfig.Data)
	conf, err := store.ConfigFindIdentical(build.RepoID, sha)
	if err != nil {
		conf = &model.Config{
			RepoID: build.RepoID,
			Data:   remoteYamlConfig.Data,
			Hash:   sha,
			Name:   shared.SanitizePath(remoteYamlConfig.Name),
		}
		err = store.ConfigCreate(conf)
		if err != nil {
			// retry in case we receive two hooks at the same time
			conf, err = store.ConfigFindIdentical(build.RepoID, sha)
			if err != nil {
				return nil, err
			}
		}
	}

	buildConfig := &model.BuildConfig{
		ConfigID: conf.ID,
		BuildID:  build.ID,
	}
	if err := store.BuildConfigCreate(buildConfig); err != nil {
		return nil, err
	}

	return conf, nil
}

// publishes message to UI clients
func publishToTopic(c context.Context, build *model.Build, repo *model.Repo) (err error) {
	message := pubsub.Message{
		Labels: map[string]string{
			"repo":    repo.FullName,
			"private": strconv.FormatBool(repo.IsSCMPrivate),
		},
	}
	buildCopy := *build
	if buildCopy.Procs, err = model.Tree(buildCopy.Procs); err != nil {
		return err
	}

	message.Data, _ = json.Marshal(model.Event{
		Repo:  *repo,
		Build: buildCopy,
	})
	return server.Config.Services.Pubsub.Publish(c, "topic/events", message)
}

func queueBuild(build *model.Build, repo *model.Repo, buildItems []*shared.BuildItem) error {
	var tasks []*queue.Task
	for _, item := range buildItems {
		if item.Proc.State == model.StatusSkipped {
			continue
		}
		task := new(queue.Task)
		task.ID = fmt.Sprint(item.Proc.ID)
		task.Labels = map[string]string{}
		for k, v := range item.Labels {
			task.Labels[k] = v
		}
		task.Labels["platform"] = item.Platform
		task.Labels["repo"] = repo.FullName
		task.Dependencies = taskIds(item.DependsOn, buildItems)
		task.RunOn = item.RunsOn
		task.DepStatus = make(map[string]string)

		task.Data, _ = json.Marshal(rpc.Pipeline{
			ID:      fmt.Sprint(item.Proc.ID),
			Config:  item.Config,
			Timeout: repo.Timeout,
		})

		if err := server.Config.Services.Logs.Open(context.Background(), task.ID); err != nil {
			return err
		}
		tasks = append(tasks, task)
	}
	return server.Config.Services.Queue.PushAtOnce(context.Background(), tasks)
}

func taskIds(dependsOn []string, buildItems []*shared.BuildItem) (taskIds []string) {
	for _, dep := range dependsOn {
		for _, buildItem := range buildItems {
			if buildItem.Proc.Name == dep {
				taskIds = append(taskIds, fmt.Sprint(buildItem.Proc.ID))
			}
		}
	}
	return
}

func shasum(raw []byte) string {
	sum := sha256.Sum256(raw)
	return fmt.Sprintf("%x", sum)
}
