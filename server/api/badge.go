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
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/badges"
	"github.com/woodpecker-ci/woodpecker/server/ccmenu"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/server/store/types"
)

// GetBadge
//
//	@Summary	Get status badge, SVG format
//	@Router		/badges/{repo_id}/status.svg [get]
//	@Produce	image/svg+xml
//	@Success	200
//	@Tags		Badges
//	@Param		repo_id			path	int		true	"the repository id"
func GetBadge(c *gin.Context) {
	_store := store.FromContext(c)

	var repo *model.Repo
	var err error

	if c.Param("repo_name") != "" {
		repo, err = _store.GetRepoName(c.Param("repo_id_or_owner") + "/" + c.Param("repo_name"))
	} else {
		var repoID int64
		repoID, err = strconv.ParseInt(c.Param("repo_id_or_owner"), 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		repo, err = _store.GetRepo(repoID)
	}

	if err != nil {
		handleDbError(c, err)
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
		if !errors.Is(err, types.RecordNotExist) {
			log.Warn().Err(err).Msg("could not get last pipeline for badge")
		}
		pipeline = nil
	}

	// we serve an SVG, so set content type appropriately.
	c.Writer.Header().Set("Content-Type", "image/svg+xml")
	c.String(http.StatusOK, badges.Generate(pipeline))
}

// GetCC
//
//	@Summary		Provide pipeline status information to the CCMenu tool
//	@Description	CCMenu displays the pipeline status of projects on a CI server as an item in the Mac's menu bar.
//	@Description	More details on how to install, you can find at http://ccmenu.org/
//	@Description	The response format adheres to CCTray v1 Specification, https://cctray.org/v1/
//	@Router			/badges/{repo_id}/cc.xml [get]
//	@Produce		xml
//	@Success		200
//	@Tags			Badges
//	@Param			owner	path	string	true	"the repository owner's name"
//	@Param			name	path	string	true	"the repository name"
func GetCC(c *gin.Context) {
	_store := store.FromContext(c)
	repo, err := _store.GetRepoName(c.Param("owner") + "/" + c.Param("name"))
	if err != nil {
		handleDbError(c, err)
		return
	}

	pipelines, err := _store.GetPipelineList(repo, &model.ListOptions{Page: 1, PerPage: 1})
	if err != nil && !errors.Is(err, types.RecordNotExist) {
		log.Warn().Err(err).Msg("could not get pipeline list")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if len(pipelines) == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	url := fmt.Sprintf("%s/%s/%d", server.Config.Server.Host, repo.FullName, pipelines[0].Number)
	cc := ccmenu.New(repo, pipelines[0], url)
	c.XML(http.StatusOK, cc)
}
