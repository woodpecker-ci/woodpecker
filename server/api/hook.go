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

package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/shared/token"
)

// getAgentName finds an agent's name, utilizing a map as a cache.
func getAgentName(store store.Store, agentNameMap map[int64]string, agentID int64) (string, bool) {
	// 1. Check the cache first.
	name, exists := agentNameMap[agentID]
	if exists {
		return name, true
	}

	// 2. If not in cache, query the store.
	agent, err := store.AgentFind(agentID)
	if err != nil || agent == nil {
		// Agent not found or an error occurred.
		return "", false
	}

	// 3. Found the agent, update the cache and return the name.
	if agent.Name != "" {
		agentNameMap[agentID] = agent.Name
		return agent.Name, true
	}

	return "", false
}

// backgroundPipelineCreationTimeout caps how long a webhook-triggered pipeline
// creation may keep running after the HTTP handler has already responded 202
// Accepted (see PostHook). It bounds the detached background goroutine so a
// stuck creation cannot leak indefinitely.
const backgroundPipelineCreationTimeout = 2 * time.Minute

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

	//
	// 1. Check if the webhook is valid and authorized
	//

	var repo *model.Repo

	_, err := token.ParseRequest([]token.Type{token.HookToken}, c.Request, func(t *token.Token) (string, error) {
		var err error
		repo, err = getRepoFromToken(_store, t)
		if err != nil {
			return "", err
		}

		return repo.WebhookSecret, nil
	})
	if err != nil {
		msg := "failure to parse token from hook"
		log.Error().Err(err).Msg(msg)
		c.String(http.StatusBadRequest, msg)
		return
	}

	if repo == nil {
		msg := "failure to get repo from token"
		log.Error().Msg(msg)
		c.String(http.StatusBadRequest, msg)
		return
	}

	_forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
	if err != nil {
		log.Error().Err(err).Int64("repo-id", repo.ID).Msgf("Cannot get forge with id: %d", repo.ForgeID)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	//
	// 2. Parse the webhook data
	//

	// A forge whose webhooks can ask for several pipelines (forge.MultiHook) returns all of them;
	// any other returns one, or none.
	var repoFromForge *model.Repo
	var pipelinesFromForge []*model.Pipeline
	if multiHook, ok := _forge.(forge.MultiHook); ok {
		repoFromForge, pipelinesFromForge, err = multiHook.HookPipelines(c, c.Request, repo.WebhookSecret)
	} else {
		var pipelineFromForge *model.Pipeline
		repoFromForge, pipelineFromForge, err = _forge.Hook(c, c.Request, repo.WebhookSecret)
		if pipelineFromForge != nil {
			pipelinesFromForge = []*model.Pipeline{pipelineFromForge}
		}
	}
	if err != nil {
		if errors.Is(err, types.ErrInvalidWebhookSignature) {
			log.Warn().Str("repo", repo.FullName).Msg("ignoring hook: invalid webhook signature")
			c.String(http.StatusForbidden, "invalid webhook signature")
			return
		}
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

	if len(pipelinesFromForge) == 0 {
		msg := "ignoring hook: hook parsing resulted in empty pipeline"
		log.Debug().Msg(msg)
		c.String(http.StatusOK, msg)
		return
	}
	if repoFromForge == nil {
		msg := "failure to ascertain repo from hook"
		log.Debug().Msg(msg)
		c.String(http.StatusBadRequest, msg)
		return
	}

	//
	// 3. Check the repo from the token is matching the repo returned by the forge
	//

	if repo.ForgeRemoteID != repoFromForge.ForgeRemoteID {
		log.Warn().Msgf("ignoring hook: repo %s does not match the repo from the token", repo.FullName)
		c.String(http.StatusBadRequest, "failure to parse token from hook")
		return
	}

	//
	// 4. Check if the repo is active and has an owner
	//

	if !repo.IsActive {
		log.Debug().Msgf("ignoring hook: repo %s is inactive", repoFromForge.FullName)
		c.Status(http.StatusNoContent)
		return
	}

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
	// 4. Update the repo
	//

	if repo.FullName != repoFromForge.FullName {
		// create a redirection
		err = _store.CreateRedirection(&model.Redirection{RepoID: repo.ID, FullName: repo.FullName})
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	repo.Update(repoFromForge)
	err = _store.UpdateRepo(repo)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	//
	// 5. Check if pull requests are allowed for this repo
	//

	pipelinesToCreate := make([]*model.Pipeline, 0, len(pipelinesFromForge))
	for _, pipelineFromForge := range pipelinesFromForge {
		if pipelineFromForge.IsPullRequest() && !repo.AllowPull {
			log.Debug().Str("repo", repo.FullName).Msg("ignoring hook: pull requests are disabled for this repo in woodpecker")
			continue
		}
		pipelinesToCreate = append(pipelinesToCreate, pipelineFromForge)
	}
	if len(pipelinesToCreate) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	//
	// 6. Finally create the pipelines
	//
	// Pipeline creation can be slow (forge round-trips, config fetching). To
	// avoid the forge timing out and retrying the webhook delivery, we wait only
	// up to WebhookSyncTimeout for creation to finish. If it completes in time we
	// respond synchronously with the first created pipeline (preserving the old
	// behavior and API response when the webhook asks for one). If it takes longer
	// we respond 202 Accepted and let creation finish in the background.
	//
	// A webhook asking for several pipelines (forge.MultiHook) gets each created in
	// turn, with a creation timeout of its own; one that fails or is filtered out
	// does not stop the others.
	type created struct {
		pipeline *model.Pipeline
		err      error
	}
	// Buffered: once the handler has answered 202 nothing reads from it, and the
	// goroutine must still be able to finish.
	done := make(chan created, 1)

	go func() {
		var first created
		for i, pipelineFromForge := range pipelinesToCreate {
			bgCtx, cancel := context.WithTimeout(context.WithoutCancel(context.Background()), backgroundPipelineCreationTimeout)
			pl, err := pipeline.Create(bgCtx, _store, repo, pipelineFromForge)
			cancel()
			if err != nil {
				// A filtered pipeline is an expected outcome, not a failure.
				if errors.Is(err, pipeline.ErrFiltered) {
					log.Debug().Str("repo", repo.FullName).Str("ref", pipelineFromForge.Ref).Msg("webhook produced no matching workflows; skipped")
				} else {
					log.Error().Err(err).Str("repo", repo.FullName).Str("ref", pipelineFromForge.Ref).Msg("could not create pipeline from webhook")
				}
			}
			if i == 0 {
				first = created{pipeline: pl, err: err}
			}
		}
		done <- first
	}()

	var result created
	syncTimeout := server.Config.Server.WebhookSyncTimeout
	if syncTimeout > 0 {
		select {
		case result = <-done:
			// handle synchronous
		case <-time.After(syncTimeout):
			log.Debug().Str("repo", repo.FullName).Dur("timeout", syncTimeout).Msg("pipeline creation exceeded webhook sync timeout, continuing in background")
			c.JSON(http.StatusAccepted, nil)
			// do async
			return
		}
	} else {
		// we do synchronous
		result = <-done
	}

	if result.err != nil {
		handlePipelineErr(c, result.err)
	} else {
		c.JSON(http.StatusOK, result.pipeline)
	}
}

func getRepoFromToken(store store.Store, t *token.Token) (*model.Repo, error) {
	if t.Get("repo-forge-remote-id") != "" {
		forgeID, err := strconv.ParseInt(t.Get("forge-id"), 10, 64)
		if err != nil {
			return nil, err
		}

		return store.GetRepoForgeID(forgeID, model.ForgeRemoteID(t.Get("repo-forge-remote-id")))
	}

	// get the repo by the repo-id
	// TODO: remove in next major
	repoID, err := strconv.ParseInt(t.Get("repo-id"), 10, 64)
	if err != nil {
		return nil, err
	}
	return store.GetRepo(repoID)
}
