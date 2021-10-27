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

// Health endpoint returns a 500 if the server state is unhealthy.
func Health(c *gin.Context) {
	if err := store.FromContext(c).Ping(); err != nil {
		c.String(500, err.Error())
		return
	}
	c.String(200, "")
}

// Version endpoint returns the server version and build information.
func Version(c *gin.Context) {
	c.JSON(200, gin.H{
		"source":  "https://github.com/woodpecker-ci/woodpecker",
		"version": version.String(),
	})
}

// LogLevel endpoint returns the current logging level
func LogLevel(c *gin.Context) {
	c.JSON(200, gin.H{
		"log-level": zerolog.GlobalLevel().String(),
	})
}

// SetLogLevel endpoint allows setting the logging level via API
func SetLogLevel(c *gin.Context) {
	logLevel := struct {
		LogLevel string `json:"log-level"`
	}{}
	if err := c.Bind(&logLevel); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	lvl, err := zerolog.ParseLevel(logLevel.LogLevel)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	log.Log().Msgf("log level set to %s", lvl.String())
	zerolog.SetGlobalLevel(lvl)
	c.JSON(200, logLevel)
}
