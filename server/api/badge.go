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

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/badges"
	"github.com/woodpecker-ci/woodpecker/server/ccmenu"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/server/store/types"
)

func GetBadge(c *gin.Context) {
	_store := store.FromContext(c)
	repo, err := _store.GetRepoName(c.Param("owner") + "/" + c.Param("name"))
	if err != nil || !repo.IsActive {
		if err == nil || errors.Is(err, types.RecordNotExist) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// if no commit was found then display
	// the 'none' badge, instead of throwing
	// an error response
	branch := c.Query("branch")
	if len(branch) == 0 {
		branch = repo.Branch
	}

	pipeline, err := _store.GetPipelineLast(repo, branch)
	if err != nil {
		log.Warn().Err(err).Msg("")
		pipeline = nil
	}

	// we serve an SVG, so set content type appropriately.
	c.Writer.Header().Set("Content-Type", "image/svg+xml")
	c.String(http.StatusOK, badges.Generate(pipeline))
}

func GetCC(c *gin.Context) {
	_store := store.FromContext(c)
	repo, err := _store.GetRepoName(c.Param("owner") + "/" + c.Param("name"))
	if err != nil {
		handleDbGetError(c, err)
		return
	}

	pipelines, err := _store.GetPipelineList(repo, &model.ListOptions{Page: 1, PerPage: 1})
	if err != nil || len(pipelines) == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	url := fmt.Sprintf("%s/%s/%d", server.Config.Server.Host, repo.FullName, pipelines[0].Number)
	cc := ccmenu.New(repo, pipelines[0], url)
	c.XML(http.StatusOK, cc)
}
