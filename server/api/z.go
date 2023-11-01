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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/version"
)

// Health
//
//	@Summary		Health information
//	@Description	If everything is fine, just a 204 will be returned, a 500 signals server state is unhealthy.
//	@Router			/healthz [get]
//	@Produce		plain
//	@Success		204
//	@Failure		500
//	@Tags			System
func Health(c *gin.Context) {
	if err := store.FromContext(c).Ping(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// Version
//
//	@Summary		Get version
//	@Description	Endpoint returns the server version and build information.
//	@Router			/version [get]
//	@Produce		json
//	@Success		200	{object}	string{source=string,version=string}
//	@Tags			System
func Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"source":  "https://github.com/woodpecker-ci/woodpecker",
		"version": version.String(),
	})
}

// LogLevel
//
//	@Summary		Current log level
//	@Description	Endpoint returns the current logging level. Requires admin rights.
//	@Router			/log-level [get]
//	@Produce		json
//	@Success		200	{object}	string{log-level=string}
//	@Tags			System
func LogLevel(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"log-level": zerolog.GlobalLevel().String(),
	})
}

// SetLogLevel
//
//	@Summary		Set log level
//	@Description	Endpoint sets the current logging level. Requires admin rights.
//	@Router			/log-level [post]
//	@Produce		json
//	@Success		200	{object}	string{log-level=string}
//	@Tags			System
//	@Param			Authorization	header	string						true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			log-level		body	string{log-level=string}	true	"the new log level, one of <debug,trace,info,warn,error,fatal,panic,disabled>"
func SetLogLevel(c *gin.Context) {
	logLevel := struct {
		LogLevel string `json:"log-level"`
	}{}
	if err := c.Bind(&logLevel); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	lvl, err := zerolog.ParseLevel(logLevel.LogLevel)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	log.Log().Msgf("log level set to %s", lvl.String())
	zerolog.SetGlobalLevel(lvl)
	c.JSON(http.StatusOK, logLevel)
}
