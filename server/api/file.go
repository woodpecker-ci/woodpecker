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

// FileList gets a list file by build.
func FileList(c *gin.Context) {
	store_ := store.FromContext(c)
	num, err := strconv.ParseInt(c.Param("number"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	repo := session.Repo(c)
	build, err := store_.GetBuildNumber(repo, num)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	files, err := store_.FileList(build)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(200, files)
}

// FileGet gets a file by process and name
func FileGet(c *gin.Context) {
	var (
		store_ = store.FromContext(c)

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

	pid, err := strconv.Atoi(c.Param("proc"))
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	build, err := store_.GetBuildNumber(repo, num)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	proc, err := store_.ProcFind(build, pid)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	file, err := store_.FileFind(proc, name)
	if err != nil {
		c.String(404, "Error getting file %q. %s", name, err)
		return
	}

	if !raw {
		c.JSON(200, file)
		return
	}

	rc, err := store_.FileRead(proc, file.Name)
	if err != nil {
		c.String(404, "Error getting file stream %q. %s", name, err)
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
