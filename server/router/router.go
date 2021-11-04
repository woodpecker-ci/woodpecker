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

package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/api"
	"github.com/woodpecker-ci/woodpecker/server/api/metrics"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/header"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/token"
	"github.com/woodpecker-ci/woodpecker/server/web"
)

// Load loads the router
func Load(serveHTTP func(w http.ResponseWriter, r *http.Request), middleware ...gin.HandlerFunc) http.Handler {
	e := gin.New()
	e.Use(gin.Recovery())

	e.Use(func(c *gin.Context) {
		log.Trace().Msgf("[%s] %s", c.Request.Method, c.Request.URL.String())
		c.Next()
	})

	e.Use(header.NoCache)
	e.Use(header.Options)
	e.Use(header.Secure)
	e.Use(middleware...)
	e.Use(session.SetUser())
	e.Use(token.Refresh)

	e.NoRoute(func(c *gin.Context) {
		req := c.Request.WithContext(
			web.WithUser(
				c.Request.Context(),
				session.User(c),
			),
		)
		serveHTTP(c.Writer, req)
	})

	e.GET("/web-config.js", web.WebConfig)

	e.GET("/logout", api.GetLogout)
	e.GET("/login", api.HandleLogin)
	auth := e.Group("/authorize")
	{
		auth.GET("", api.HandleAuth)
		auth.POST("", api.HandleAuth)
		auth.POST("/token", api.GetLoginToken)
	}

	e.GET("/metrics", metrics.PromHandler())
	e.GET("/version", api.Version)
	e.GET("/healthz", api.Health)

	apiRoutes(e)

	return e
}
