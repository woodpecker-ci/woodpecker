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
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/pipeline"
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

	tmpRepo, tmpBuild, err := server.Config.Services.Remote.Hook(c, c.Request)
	if err != nil {
		msg := "failure to parse hook"
		log.Debug().Err(err).Msg(msg)
		c.String(http.StatusBadRequest, msg)
		return
	}
	if tmpBuild == nil {
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

	// skip the tmpBuild if any case-insensitive combination of the words "skip" and "ci"
	// wrapped in square brackets appear in the commit message
	skipMatch := skipRe.FindString(tmpBuild.Message)
	if len(skipMatch) > 0 {
		msg := fmt.Sprintf("ignoring hook: %s found in %s", skipMatch, tmpBuild.Commit)
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

	if tmpBuild.Event == model.EventPull && !repo.AllowPull {
		msg := "ignoring hook: pull requests are disabled for this repo in woodpecker"
		log.Debug().Str("repo", repo.FullName).Msg(msg)
		c.String(http.StatusNoContent, msg)
		return
	}

	build, err := pipeline.Create(c, _store, repo, tmpBuild)
	if err != nil {
		if pipeline.IsErrNotFound(err) {
			c.String(http.StatusNotFound, "%v", err)
		} else if pipeline.IsErrBadRequest(err) {
			c.String(http.StatusBadRequest, "%v", err)
		} else {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
		}
	} else {
		c.JSON(200, build)
	}
}
