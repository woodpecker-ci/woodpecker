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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/badges"
	"github.com/woodpecker-ci/woodpecker/server/ccmenu"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

func GetBadge(c *gin.Context) {
	_store := store.FromContext(c)
	repo, err := _store.GetRepoName(c.Param("owner") + "/" + c.Param("name"))
	if err != nil {
		c.AbortWithStatus(404)
		return
	}

	// if no commit was found then display
	// the 'none' badge, instead of throwing
	// an error response
	branch := c.Query("branch")
	if len(branch) == 0 {
		branch = repo.Branch
	}

	build, err := _store.GetBuildLast(repo, branch)
	if err != nil {
		log.Warn().Err(err).Msg("")
		build = nil
	}

	// we serve an SVG, so set content type appropriately.
	c.Writer.Header().Set("Content-Type", "image/svg+xml")
	c.String(http.StatusOK, badges.Generate(build))
}

func GetCC(c *gin.Context) {
	_store := store.FromContext(c)
	repo, err := _store.GetRepoName(c.Param("owner") + "/" + c.Param("name"))
	if err != nil {
		c.AbortWithStatus(404)
		return
	}

	builds, err := _store.GetBuildList(repo, 1)
	if err != nil || len(builds) == 0 {
		c.AbortWithStatus(404)
		return
	}

	url := fmt.Sprintf("%s/%s/%d", server.Config.Server.Host, repo.FullName, builds[0].Number)
	cc := ccmenu.New(repo, builds[0], url)
	c.XML(200, cc)
}
