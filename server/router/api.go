// Copyright 2021 Woodpecker Authors
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
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/woodpecker-ci/woodpecker/server/api"
	"github.com/woodpecker-ci/woodpecker/server/api/debug"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
)

func apiRoutes(e *gin.Engine) {
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

	orgBase := e.Group("/api/orgs/:owner")
	{
		orgBase.GET("/permissions", api.GetOrgPermissions)

		org := orgBase.Group("")
		{
			org.Use(session.MustOrgMember(true))
			org.GET("/secrets", api.GetOrgSecretList)
			org.POST("/secrets", api.PostOrgSecret)
			org.GET("/secrets/:secret", api.GetOrgSecret)
			org.PATCH("/secrets/:secret", api.PatchOrgSecret)
			org.DELETE("/secrets/:secret", api.DeleteOrgSecret)
		}
	}

	repoBase := e.Group("/api/repos/:owner/:name")
	{
		repoBase.Use(session.SetRepo())
		repoBase.Use(session.SetPerm())

		repoBase.GET("/permissions", api.GetRepoPermissions)

		repo := repoBase.Group("")
		{
			repo.Use(session.MustPull)

			repo.POST("", session.MustRepoAdmin(), api.PostRepo)
			repo.GET("", api.GetRepo)

			repo.GET("/branches", api.GetRepoBranches)

			repo.GET("/builds", api.GetBuilds)
			repo.POST("/builds", api.CreateBuild)
			repo.GET("/builds/:number", api.GetBuild)
			repo.GET("/builds/:number/config", api.GetBuildConfig)

			// requires push permissions
			repo.POST("/builds/:number", session.MustPush, api.PostBuild)
			repo.DELETE("/builds/:number", session.MustPush, api.DeleteBuild)
			repo.POST("/builds/:number/approve", session.MustPush, api.PostApproval)
			repo.POST("/builds/:number/decline", session.MustPush, api.PostDecline)
			repo.DELETE("/builds/:number/:job", session.MustPush, api.DeleteBuild)

			repo.GET("/logs/:number/:pid", api.GetProcLogs)
			repo.GET("/logs/:number/:pid/:proc", api.GetBuildLogs)

			// requires push permissions
			repo.DELETE("/logs/:number", session.MustPush, api.DeleteBuildLogs)

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

			// requires push permissions
			repo.GET("/cron", session.MustPush, api.GetCronList)
			repo.POST("/cron", session.MustPush, api.PostCron)
			repo.GET("/cron/:cron", session.MustPush, api.GetCron)
			repo.PATCH("/cron/:cron", session.MustPush, api.PatchCron)
			repo.DELETE("/cron/:cron", session.MustPush, api.DeleteCron)

			// requires admin permissions
			repo.PATCH("", session.MustRepoAdmin(), api.PatchRepo)
			repo.DELETE("", session.MustRepoAdmin(), api.DeleteRepo)
			repo.POST("/chown", session.MustRepoAdmin(), api.ChownRepo)
			repo.POST("/repair", session.MustRepoAdmin(), api.RepairRepo)
			repo.POST("/move", session.MustRepoAdmin(), api.MoveRepo)
		}
	}

	badges := e.Group("/api/badges/:owner/:name")
	{
		badges.GET("/status.svg", api.GetBadge)
		badges.GET("/cc.xml", api.GetCC)
	}

	builds := e.Group("/api/builds")
	{
		builds.Use(session.MustAdmin())
		builds.GET("", api.GetBuildQueue)
	}

	queue := e.Group("/api/queue")
	{
		queue.Use(session.MustAdmin())
		queue.GET("/info", api.GetQueueInfo)
		queue.GET("/pause", api.PauseQueue)
		queue.GET("/resume", api.ResumeQueue)
		queue.GET("/norunningbuilds", api.BlockTilQueueHasRunningItem)
	}

	secrets := e.Group("/api/secrets")
	{
		secrets.Use(session.MustAdmin())
		secrets.GET("", api.GetGlobalSecretList)
		secrets.POST("", api.PostGlobalSecret)
		secrets.GET("/:secret", api.GetGlobalSecret)
		secrets.PATCH("/:secret", api.PatchGlobalSecret)
		secrets.DELETE("/:secret", api.DeleteGlobalSecret)
	}

	if zerolog.GlobalLevel() <= zerolog.DebugLevel {
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
	}

	logLevel := e.Group("/api/log-level")
	{
		logLevel.Use(session.MustAdmin())
		logLevel.GET("", api.LogLevel)
		logLevel.POST("", api.SetLogLevel)
	}

	e.GET("/api/signature/public-key", session.MustUser(), api.GetSignaturePublicKey)

	// TODO: remove /hook in favor of /api/hook
	e.POST("/hook", api.PostHook)
	e.POST("/api/hook", api.PostHook)

	// TODO: move to /api/stream
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
}
