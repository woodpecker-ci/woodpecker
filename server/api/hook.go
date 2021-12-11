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
	_remote := server.Config.Services.Remote
	_store := store.FromContext(c)

	tmpRepo, build, err := _remote.Hook(c.Request)
	if err != nil {
		msg := "failure to parse hook"
		log.Debug().Err(err).Msg(msg)
		c.String(http.StatusBadRequest, "%s: %v", msg, err)
		return
	}
	if build == nil {
		c.String(http.StatusOK, "ignoring hook: hook parse result in empty build")
		return
	}
	if tmpRepo == nil {
		c.String(http.StatusBadRequest, "failure to ascertain repo from hook")
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
		msg := fmt.Sprintf("failure to get repo %s/%s from store", tmpRepo.Owner, tmpRepo.Name)
		log.Error().Err(err).Msg(msg)
		c.String(http.StatusNotFound, "%s: %v", msg, err)
		return
	}
	if !repo.IsActive {
		c.String(204, "ignoring hook: %s/%s is inactive.", tmpRepo.Owner, tmpRepo.Name)
		return
	}

	// get the token and verify the hook is authorized
	parsed, err := token.ParseRequest(c.Request, func(_ *token.Token) (string, error) {
		return repo.Hash, nil
	})
	if err != nil {
		msg := fmt.Sprintf("failure to parse token from hook for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		c.String(http.StatusBadRequest, "%s: %v", msg, err)
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
		msg := "ignoring hook: pull requests are disabled on woodpecker for this repo"
		log.Debug().Str("repo", repo.FullName).Msg(msg)
		c.String(http.StatusNoContent, msg)
		return
	}

	repoUser, err := _store.GetUser(repo.UserID)
	if err != nil {
		log.Error().Err(err).Str("repo", repo.FullName).Msgf("failure to find repo owner via id '%d'", repo.UserID)
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// if the remote has a refresh token, the current access token
	// may be stale. Therefore, we should refresh prior to dispatching
	// the build.
	if refresher, ok := _remote.(remote.Refresher); ok {
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
	configFetcher := shared.NewConfigFetcher(_remote, repoUser, repo, build)
	remoteYamlConfigs, err := configFetcher.Fetch(c)
	if err != nil {
		msg := fmt.Sprintf("cannot find '%s' in '%s', context user: '%s'", repo.Config, build.Ref, repoUser.Login)
		log.Debug().Err(err).Str("repo", repo.FullName).Msg(msg)
		c.String(http.StatusNotFound, "%s: %v", msg, err)
		return
	}

	filtered, err := branchFiltered(build, remoteYamlConfigs)
	if err != nil {
		msg := "failure to parse yaml from hook"
		log.Debug().Err(err).Str("repo", repo.FullName).Msg(msg)
		c.String(http.StatusBadRequest, "%s: %v", msg, err)
	}
	if filtered {
		msg := "ignoring hook: branch does not match restrictions defined in yaml"
		log.Debug().Str("repo", repo.FullName).Msg(msg)
		c.String(200, msg)
		return
	}

	if zeroSteps(build, remoteYamlConfigs) {
		msg := "ignoring hook: step conditions yield zero runnable steps"
		log.Debug().Str("repo", repo.FullName).Msg(msg)
		c.String(200, msg)
		return
	}

	// update some build fields
	build.RepoID = repo.ID
	build.Verified = true
	build.Status = model.StatusPending

	// TODO(336) extend gated feature with an allow/block List
	if repo.IsGated && build.Sender != repoUser.Login {
		build.Status = model.StatusBlocked
	}

	err = _store.CreateBuild(build, build.Procs...)
	if err != nil {
		log.Error().Err(err).Msgf("failure to save commit for %s", repo.FullName)
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// persist the build config for historical correctness, restarts, etc
	for _, remoteYamlConfig := range remoteYamlConfigs {
		_, err := findOrPersistPipelineConfig(repo, build, remoteYamlConfig)
		if err != nil {
			log.Error().Err(err).Msgf("failure to find or persist build config for %s", repo.FullName)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	c.JSON(200, build)

	if build.Status == model.StatusBlocked {
		return
	}

	netrc, err := _remote.Netrc(repoUser, repo)
	if err != nil {
		msg := "Failed to generate netrc file"
		log.Error().Err(err).Msg(msg)
		c.String(http.StatusInternalServerError, "%s: %v", msg, err)
		return
	}

	envs := map[string]string{}
	if server.Config.Services.Environ != nil {
		globals, _ := server.Config.Services.Environ.EnvironList(repo)
		for _, global := range globals {
			envs[global.Name] = global.Value
		}
	}

	secs, err := server.Config.Services.Secrets.SecretListBuild(repo, build)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting secrets for %s#%d", repo.FullName, build.Number)
	}

	regs, err := server.Config.Services.Registries.RegistryList(repo)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting registry credentials for %s#%d", repo.FullName, build.Number)
	}

	// get the previous build so that we can send status change notifications
	last, _ := _store.GetBuildLastBefore(repo, build.Branch, build.ID)

	b := shared.ProcBuilder{
		Repo:  repo,
		Curr:  build,
		Last:  last,
		Netrc: netrc,
		Secs:  secs,
		Regs:  regs,
		Envs:  envs,
		Link:  server.Config.Server.Host,
		Yamls: remoteYamlConfigs,
	}
	buildItems, err := b.Build()
	if err != nil {
		if _, err = shared.UpdateToStatusError(_store, *build, err); err != nil {
			log.Error().Err(err).Msgf("Error setting error status of build for %s#%d", repo.FullName, build.Number)
		}
		return
	}
	build = shared.SetBuildStepsOnBuild(b.Curr, buildItems)

	err = _store.ProcCreate(build.Procs)
	if err != nil {
		log.Error().Msgf("error persisting procs %s/%d: %s", repo.FullName, build.Number, err)
	}

	defer func() {
		for _, item := range buildItems {
			uri := fmt.Sprintf("%s/%s/build/%d", server.Config.Server.Host, repo.FullName, build.Number)
			if len(buildItems) > 1 {
				err = _remote.Status(c, repoUser, repo, build, uri, item.Proc)
			} else {
				err = _remote.Status(c, repoUser, repo, build, uri, nil)
			}
			if err != nil {
				log.Error().Err(err).Msgf("error setting commit status for %s/%d", repo.FullName, build.Number)
			}
		}
	}()

	if err := publishToTopic(c, build, repo, model.Enqueued); err != nil {
		log.Error().Err(err).Msg("publishToTopic")
	}
	if err := queueBuild(build, repo, buildItems); err != nil {
		log.Error().Err(err).Msg("queueBuild")
	}
}

// TODO: parse yaml once and not for each filter function
func branchFiltered(build *model.Build, remoteYamlConfigs []*remote.FileMeta) (bool, error) {
	log.Trace().Msgf("hook.branchFiltered(): build branch: '%s' build event: '%s' config count: %d", build.Branch, build.Event, len(remoteYamlConfigs))
	for _, remoteYamlConfig := range remoteYamlConfigs {
		parsedPipelineConfig, err := yaml.ParseString(string(remoteYamlConfig.Data))
		if err != nil {
			log.Trace().Msgf("parse config '%s': %s", remoteYamlConfig.Name, err)
			return false, err
		}
		log.Trace().Msgf("config '%s': %#v", remoteYamlConfig.Name, parsedPipelineConfig)

		if !parsedPipelineConfig.Branches.Match(build.Branch) && build.Event != model.EventTag && build.Event != model.EventDeploy {
		} else {
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

func findOrPersistPipelineConfig(repo *model.Repo, build *model.Build, remoteYamlConfig *remote.FileMeta) (*model.Config, error) {
	sha := shasum(remoteYamlConfig.Data)
	conf, err := server.Config.Storage.Config.ConfigFindIdentical(build.RepoID, sha)
	if err != nil {
		conf = &model.Config{
			RepoID: build.RepoID,
			Data:   remoteYamlConfig.Data,
			Hash:   sha,
			Name:   shared.SanitizePath(remoteYamlConfig.Name),
		}
		err = server.Config.Storage.Config.ConfigCreate(conf)
		if err != nil {
			// retry in case we receive two hooks at the same time
			conf, err = server.Config.Storage.Config.ConfigFindIdentical(build.RepoID, sha)
			if err != nil {
				return nil, err
			}
		}
	}

	buildConfig := &model.BuildConfig{
		ConfigID: conf.ID,
		BuildID:  build.ID,
	}
	err = server.Config.Storage.Config.BuildConfigCreate(buildConfig)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

// publishes message to UI clients
func publishToTopic(c *gin.Context, build *model.Build, repo *model.Repo, event model.EventType) (err error) {
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
		Type:  model.Enqueued,
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
