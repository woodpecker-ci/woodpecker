// Copyright 2022 Woodpecker Authors
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
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// FileList gets a list file by pipeline.
func FileList(c *gin.Context) {
	_store := store.FromContext(c)
	num, err := strconv.ParseInt(c.Param("number"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	repo := session.Repo(c)
	pipeline, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleDbGetError(c, err)
		return
	}

	files, err := _store.FileList(pipeline, session.Pagination(c))
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(200, files)
}

// FileGet gets a file by process and name
func FileGet(c *gin.Context) {
	var (
		_store = store.FromContext(c)

		repo = session.Repo(c)
		name = strings.TrimPrefix(c.Param("file"), "/")
		raw  = func() bool {
			return c.DefaultQuery("raw", "false") == "true"
		}()
	)

	num, err := strconv.ParseInt(c.Param("number"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	pid, err := strconv.Atoi(c.Param("step"))
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	pipeline, err := _store.GetPipelineNumber(repo, num)
	if err != nil {
		handleDbGetError(c, err)
		return
	}

	step, err := _store.StepFind(pipeline, pid)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	file, err := _store.FileFind(step, name)
	if err != nil {
		c.String(http.StatusNotFound, "Error getting file %q. %s", name, err)
		return
	}

	if !raw {
		c.JSON(http.StatusOK, file)
		return
	}

	rc, err := _store.FileRead(step, file.Name)
	if err != nil {
		c.String(http.StatusNotFound, "Error getting file stream %q. %s", name, err)
		return
	}
	defer rc.Close()

	switch file.Mime {
	case "application/vnd.test+json":
		c.Header("Content-Type", "application/json")
	}

	if _, err := io.Copy(c.Writer, rc); err != nil {
		log.Error().Err(err).Msg("could not copy file to http response")
	}
}
