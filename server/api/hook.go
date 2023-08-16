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
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/forge/types"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/pipeline"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/shared/token"
)

var skipRe = regexp.MustCompile(`\[(?i:ci *skip|skip *ci)\]`)

// GetQueueInfo
//
//	@Summary	Get pipeline queue information
//	@Description	TODO: link the InfoT response object - this is blocked, until the `swaggo/swag` tool dependency is v1.18.12 or newer
//	@Router		/queue/info [get]
//	@Produce	json
//	@Success	200	{object} map[string]string
//	@Tags		Pipeline queues
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func GetQueueInfo(c *gin.Context) {
	c.IndentedJSON(http.StatusOK,
		server.Config.Services.Queue.Info(c),
	)
}

// PauseQueue
//
//	@Summary	Pause a pipeline queue
//	@Router		/queue/pause [post]
//	@Produce	plain
//	@Success	200
//	@Tags		Pipeline queues
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func PauseQueue(c *gin.Context) {
	server.Config.Services.Queue.Pause()
	c.Status(http.StatusOK)
}

// ResumeQueue
//
//	@Summary	Resume a pipeline queue
//	@Router		/queue/resume [post]
//	@Produce	plain
//	@Success	200
//	@Tags		Pipeline queues
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func ResumeQueue(c *gin.Context) {
	server.Config.Services.Queue.Resume()
	c.Status(http.StatusOK)
}

// BlockTilQueueHasRunningItem
//
//	@Summary	Block til pipeline queue has a running item
//	@Router		/queue/norunningpipelines [get]
//	@Produce	plain
//	@Success	200
//	@Tags		Pipeline queues
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func BlockTilQueueHasRunningItem(c *gin.Context) {
	for {
		info := server.Config.Services.Queue.Info(c)
		if info.Stats.Running == 0 {
			break
		}
	}
	c.Status(http.StatusOK)
}

// PostHook
//
//	@Summary	Incoming webhook from forge
//	@Router		/hook [post]
//	@Produce	plain
//	@Success	200
//	@Tags		System
//	@Param		hook	body	object	true	"the webhook payload; forge is automatically detected"
func PostHook(c *gin.Context) {
	_store := store.FromContext(c)
	forge := server.Config.Services.Forge

	tmpRepo, tmpPipeline, err := forge.Hook(c, c.Request)
	if err != nil {
		if errors.Is(err, &types.ErrIgnoreEvent{}) {
			msg := fmt.Sprintf("forge driver: %s", err)
			log.Debug().Err(err).Msg(msg)
			c.String(http.StatusOK, msg)
			return
		}

		msg := "failure to parse hook"
		log.Debug().Err(err).Msg(msg)
		c.String(http.StatusBadRequest, msg)
		return
	}

	if tmpPipeline == nil {
		msg := "ignoring hook: hook parsing resulted in empty pipeline"
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

	// skip the tmpPipeline if any case-insensitive combination of the words "skip" and "ci"
	// wrapped in square brackets appear in the commit message
	skipMatch := skipRe.FindString(tmpPipeline.Message)
	if len(skipMatch) > 0 {
		msg := fmt.Sprintf("ignoring hook: %s found in %s", skipMatch, tmpPipeline.Commit)
		log.Debug().Msg(msg)
		c.String(http.StatusNoContent, msg)
		return
	}

	repo, err := _store.GetRepoNameFallback(tmpRepo.ForgeRemoteID, tmpRepo.FullName)
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

	oldFullName := repo.FullName
	if oldFullName != tmpRepo.FullName {
		// create a redirection
		err = _store.CreateRedirection(&model.Redirection{RepoID: repo.ID, FullName: repo.FullName})
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	repo.Update(tmpRepo)
	err = _store.UpdateRepo(repo)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
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
	verifiedKey := parsed.Text == oldFullName
	if !verifiedKey {
		verifiedKey, err = _store.HasRedirectionForRepo(repo.ID, repo.FullName)
		if err != nil {
			msg := "failure to verify token from hook. Could not check for redirections of the repo"
			log.Error().Err(err).Msg(msg)
			c.String(http.StatusInternalServerError, msg)
			return
		}
	}

	if !verifiedKey {
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

	if tmpPipeline.Event == model.EventPull && !repo.AllowPull {
		msg := "ignoring hook: pull requests are disabled for this repo in woodpecker"
		log.Debug().Str("repo", repo.FullName).Msg(msg)
		c.String(http.StatusNoContent, msg)
		return
	}

	pl, err := pipeline.Create(c, _store, repo, tmpPipeline)
	if err != nil {
		handlePipelineErr(c, err)
	} else {
		c.JSON(http.StatusOK, pl)
	}
}
