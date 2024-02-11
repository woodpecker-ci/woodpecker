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

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/pipeline"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/shared/token"
)

// GetQueueInfo
//
//	@Summary		Get pipeline queue information
//	@Description	TODO: link the InfoT response object - this is blocked, until the `swaggo/swag` tool dependency is v1.18.12 or newer
//	@Router			/queue/info [get]
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Tags			Pipeline queues
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
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
//	@Success	204
//	@Tags		Pipeline queues
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func PauseQueue(c *gin.Context) {
	server.Config.Services.Queue.Pause()
	c.Status(http.StatusNoContent)
}

// ResumeQueue
//
//	@Summary	Resume a pipeline queue
//	@Router		/queue/resume [post]
//	@Produce	plain
//	@Success	204
//	@Tags		Pipeline queues
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func ResumeQueue(c *gin.Context) {
	server.Config.Services.Queue.Resume()
	c.Status(http.StatusNoContent)
}

// BlockTilQueueHasRunningItem
//
//	@Summary	Block til pipeline queue has a running item
//	@Router		/queue/norunningpipelines [get]
//	@Produce	plain
//	@Success	204
//	@Tags		Pipeline queues
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func BlockTilQueueHasRunningItem(c *gin.Context) {
	for {
		info := server.Config.Services.Queue.Info(c)
		if info.Stats.Running == 0 {
			break
		}
	}
	c.Status(http.StatusNoContent)
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
	_forge := server.Config.Services.Forge

	//
	// 1. Parse webhook
	//

	tmpRepo, tmpPipeline, err := _forge.Hook(c, c.Request)
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

	//
	// 2. Get related repo from store and take repo renaming into account
	//

	repo, err := _store.GetRepoNameFallback(tmpRepo.ForgeRemoteID, tmpRepo.FullName)
	if err != nil {
		log.Error().Err(err).Msgf("failure to get repo %s from store", tmpRepo.FullName)
		handleDBError(c, err)
		return
	}
	if !repo.IsActive {
		log.Debug().Msgf("ignoring hook: repo %s is inactive", tmpRepo.FullName)
		c.Status(http.StatusNoContent)
		return
	}
	oldFullName := repo.FullName

	if repo.UserID == 0 {
		log.Warn().Msgf("ignoring hook. repo %s has no owner.", repo.FullName)
		c.Status(http.StatusNoContent)
		return
	}

	user, err := _store.GetUser(repo.UserID)
	if err != nil {
		handleDBError(c, err)
		return
	}
	forge.Refresh(c, _forge, _store, user)

	//
	// 3. Check if the webhook is a valid and authorized one
	//

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

	//
	// 4. Update repo
	//

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

	//
	// 5. Check if pull requests are allowed for this repo
	//

	if (tmpPipeline.Event == model.EventPull || tmpPipeline.Event == model.EventPullClosed) && !repo.AllowPull {
		log.Debug().Str("repo", repo.FullName).Msg("ignoring hook: pull requests are disabled for this repo in woodpecker")
		c.Status(http.StatusNoContent)
		return
	}

	//
	// 6. Finally create a pipeline
	//

	pl, err := pipeline.Create(c, _store, repo, tmpPipeline)
	if err != nil {
		handlePipelineErr(c, err)
	} else {
		c.JSON(http.StatusOK, pl)
	}
}
