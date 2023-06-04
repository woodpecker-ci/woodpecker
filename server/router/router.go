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
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/woodpecker-ci/woodpecker/cmd/server/docs"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/api"
	"github.com/woodpecker-ci/woodpecker/server/api/metrics"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/header"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/token"
	"github.com/woodpecker-ci/woodpecker/server/web"
)

// Load loads the router
func Load(noRouteHandler http.HandlerFunc, middleware ...gin.HandlerFunc) http.Handler {
	e := gin.New()
	e.UseRawPath = true
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

	e.NoRoute(gin.WrapF(noRouteHandler))

	base := e.Group(server.Config.Server.RootPath)
	{
		base.GET("/web-config.js", web.Config)

		base.GET("/logout", api.GetLogout)
		base.GET("/login", api.HandleLogin)
		auth := base.Group("/authorize")
		{
			auth.GET("", api.HandleAuth)
			auth.POST("", api.HandleAuth)
			auth.POST("/token", api.GetLoginToken)
		}

		base.GET("/metrics", metrics.PromHandler())
		base.GET("/version", api.Version)
		base.GET("/healthz", api.Health)
	}

	apiRoutes(base)
	setupSwaggerConfigAndRoutes(e)

	return e
}

func setupSwaggerConfigAndRoutes(e *gin.Engine) {
	docs.SwaggerInfo.Host = getHost(server.Config.Server.Host)
	docs.SwaggerInfo.BasePath = server.Config.Server.RootPath + "/api"
	e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

func getHost(s string) string {
	parse, _ := url.Parse(s)
	return parse.Host
}
