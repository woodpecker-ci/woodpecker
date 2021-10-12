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
	"github.com/woodpecker-ci/woodpecker/server/api/debug"
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

	e.GET("/logout", api.GetLogout)
	e.GET("/login", api.HandleLogin)

	user := e.Group("/api/user")
	{
		user.Use(session.MustUser())
		user.GET("", api.GetSelf)
		user.GET("/feed", api.GetFeed)
		user.GET("/repos", api.GetRepos)
		user.POST("/token", api.PostToken)
		user.DELETE("/token", api.DeleteToken)
	}

	users := e.Group("/api/users")
	{
		users.Use(session.MustAdmin())
		users.GET("", api.GetUsers)
		users.POST("", api.PostUser)
		users.GET("/:login", api.GetUser)
		users.PATCH("/:login", api.PatchUser)
		users.DELETE("/:login", api.DeleteUser)
	}

	repo := e.Group("/api/repos/:owner/:name")
	{
		repo.Use(session.SetRepo())
		repo.Use(session.SetPerm())
		repo.Use(session.MustPull)

		repo.POST("", session.MustRepoAdmin(), api.PostRepo)
		repo.GET("", api.GetRepo)
		repo.GET("/builds", api.GetBuilds)
		repo.GET("/builds/:number", api.GetBuild)
		repo.GET("/logs/:number/:pid", api.GetProcLogs)
		repo.GET("/logs/:number/:pid/:proc", api.GetBuildLogs)

		repo.GET("/files/:number", api.FileList)
		repo.GET("/files/:number/:proc/*file", api.FileGet)

		// requires push permissions
		repo.GET("/secrets", session.MustPush, api.GetSecretList)
		repo.POST("/secrets", session.MustPush, api.PostSecret)
		repo.GET("/secrets/:secret", session.MustPush, api.GetSecret)
		repo.PATCH("/secrets/:secret", session.MustPush, api.PatchSecret)
		repo.DELETE("/secrets/:secret", session.MustPush, api.DeleteSecret)

		// requires push permissions
		repo.GET("/registry", session.MustPush, api.GetRegistryList)
		repo.POST("/registry", session.MustPush, api.PostRegistry)
		repo.GET("/registry/:registry", session.MustPush, api.GetRegistry)
		repo.PATCH("/registry/:registry", session.MustPush, api.PatchRegistry)
		repo.DELETE("/registry/:registry", session.MustPush, api.DeleteRegistry)

		// requires admin permissions
		repo.PATCH("", session.MustRepoAdmin(), api.PatchRepo)
		repo.DELETE("", session.MustRepoAdmin(), api.DeleteRepo)
		repo.POST("/chown", session.MustRepoAdmin(), api.ChownRepo)
		repo.POST("/repair", session.MustRepoAdmin(), api.RepairRepo)
		repo.POST("/move", session.MustRepoAdmin(), api.MoveRepo)

		repo.POST("/builds/:number", session.MustPush, api.PostBuild)
		repo.DELETE("/builds/:number", session.MustPush, api.DeleteBuild)
		repo.POST("/builds/:number/approve", session.MustPush, api.PostApproval)
		repo.POST("/builds/:number/decline", session.MustPush, api.PostDecline)
		repo.DELETE("/builds/:number/:job", session.MustPush, api.DeleteBuild)
		repo.DELETE("/logs/:number", session.MustPush, api.DeleteBuildLogs)
	}

	badges := e.Group("/api/badges/:owner/:name")
	{
		badges.GET("/status.svg", api.GetBadge)
		badges.GET("/cc.xml", api.GetCC)
	}

	e.POST("/hook", api.PostHook)
	e.POST("/api/hook", api.PostHook)

	sse := e.Group("/stream")
	{
		sse.GET("/events", api.EventStreamSSE)
		sse.GET("/logs/:owner/:name/:build/:number",
			session.SetRepo(),
			session.SetPerm(),
			session.MustPull,
			api.LogStreamSSE,
		)
	}

	queue := e.Group("/api/queue")
	{
		queue.GET("/info",
			session.MustAdmin(),
			api.GetQueueInfo,
		)
		queue.GET("/pause",
			session.MustAdmin(),
			api.PauseQueue,
		)
		queue.GET("/resume",
			session.MustAdmin(),
			api.ResumeQueue,
		)
		queue.GET("/norunningbuilds",
			session.MustAdmin(),
			api.BlockTilQueueHasRunningItem,
		)
	}

	auth := e.Group("/authorize")
	{
		auth.GET("", api.HandleAuth)
		auth.POST("", api.HandleAuth)
		auth.POST("/token", api.GetLoginToken)
	}

	builds := e.Group("/api/builds")
	{
		builds.Use(session.MustAdmin())
		builds.GET("", api.GetBuildQueue)
	}

	debugger := e.Group("/api/debug")
	{
		debugger.Use(session.MustAdmin())
		debugger.GET("/pprof/", debug.IndexHandler())
		debugger.GET("/pprof/heap", debug.HeapHandler())
		debugger.GET("/pprof/goroutine", debug.GoroutineHandler())
		debugger.GET("/pprof/block", debug.BlockHandler())
		debugger.GET("/pprof/threadcreate", debug.ThreadCreateHandler())
		debugger.GET("/pprof/cmdline", debug.CmdlineHandler())
		debugger.GET("/pprof/profile", debug.ProfileHandler())
		debugger.GET("/pprof/symbol", debug.SymbolHandler())
		debugger.POST("/pprof/symbol", debug.SymbolHandler())
		debugger.GET("/pprof/trace", debug.TraceHandler())
	}

	monitor := e.Group("/metrics")
	{
		monitor.GET("", metrics.PromHandler())
	}

	e.GET("/version", api.Version)
	e.GET("/healthz", api.Health)

	return e
}
