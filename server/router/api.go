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

	"go.woodpecker-ci.org/woodpecker/v2/server/api"
	"go.woodpecker-ci.org/woodpecker/v2/server/api/debug"
	"go.woodpecker-ci.org/woodpecker/v2/server/router/middleware/session"
)

func apiRoutes(e *gin.RouterGroup) {
	apiBase := e.Group("/api")
	{
		user := apiBase.Group("/user")
		{
			user.Use(session.MustUser())
			user.GET("", api.GetSelf)
			user.GET("/feed", api.GetFeed)
			user.GET("/repos", api.GetRepos)
			user.POST("/token", api.PostToken)
			user.DELETE("/token", api.DeleteToken)
		}

		users := apiBase.Group("/users")
		{
			users.Use(session.MustAdmin())
			users.GET("", api.GetUsers)
			users.POST("", api.PostUser)
			users.GET("/:login", api.GetUser)
			users.PATCH("/:login", api.PatchUser)
			users.DELETE("/:login", api.DeleteUser)
		}

		orgs := apiBase.Group("/orgs")
		{
			orgs.GET("", session.MustAdmin(), api.GetOrgs)
			orgs.GET("/lookup/*org_full_name", api.LookupOrg)
			orgBase := orgs.Group("/:org_id")
			{
				orgBase.GET("/permissions", api.GetOrgPermissions)

				org := orgBase.Group("")
				{
					org.Use(session.MustOrgMember(true))
					org.DELETE("", session.MustAdmin(), api.DeleteOrg)
					org.GET("", api.GetOrg)

					org.GET("/secrets", api.GetOrgSecretList)
					org.POST("/secrets", api.PostOrgSecret)
					org.GET("/secrets/:secret", api.GetOrgSecret)
					org.PATCH("/secrets/:secret", api.PatchOrgSecret)
					org.DELETE("/secrets/:secret", api.DeleteOrgSecret)

					org.GET("/registries", api.GetOrgRegistryList)
					org.POST("/registries", api.PostOrgRegistry)
					org.GET("/registries/:registry", api.GetOrgRegistry)
					org.PATCH("/registries/:registry", api.PatchOrgRegistry)
					org.DELETE("/registries/:registry", api.DeleteOrgRegistry)
				}
			}
		}

		repo := apiBase.Group("/repos")
		{
			repo.GET("/lookup/*repo_full_name", session.SetRepo(), session.SetPerm(), session.MustPull, api.LookupRepo)
			repo.POST("", session.MustUser(), api.PostRepo)
			repo.GET("", session.MustAdmin(), api.GetAllRepos)
			repo.POST("/repair", session.MustAdmin(), api.RepairAllRepos)
			repoBase := repo.Group("/:repo_id")
			{
				repoBase.Use(session.SetRepo())
				repoBase.Use(session.SetPerm())

				repoBase.GET("/permissions", api.GetRepoPermissions)

				repo := repoBase.Group("")
				{
					repo.Use(session.MustPull)

					repo.GET("", api.GetRepo)

					repo.GET("/branches", api.GetRepoBranches)
					repo.GET("/pull_requests", api.GetRepoPullRequests)

					repo.GET("/pipelines", api.GetPipelines)
					repo.POST("/pipelines", session.MustPush, api.CreatePipeline)
					repo.DELETE("/pipelines/:number", session.MustRepoAdmin(), api.DeletePipeline)
					repo.GET("/pipelines/:number", api.GetPipeline)
					repo.GET("/pipelines/:number/config", api.GetPipelineConfig)

					// requires push permissions
					repo.POST("/pipelines/:number", session.MustPush, api.PostPipeline)
					repo.POST("/pipelines/:number/cancel", session.MustPush, api.CancelPipeline)
					repo.POST("/pipelines/:number/approve", session.MustPush, api.PostApproval)
					repo.POST("/pipelines/:number/decline", session.MustPush, api.PostDecline)

					repo.GET("/logs/:number/:stepId", api.GetStepLogs)
					repo.DELETE("/logs/:number/:stepId", session.MustPush, api.DeleteStepLogs)

					// requires push permissions
					repo.DELETE("/logs/:number", session.MustPush, api.DeletePipelineLogs)

					// requires push permissions
					repo.GET("/secrets", session.MustPush, api.GetSecretList)
					repo.POST("/secrets", session.MustPush, api.PostSecret)
					repo.GET("/secrets/:secret", session.MustPush, api.GetSecret)
					repo.PATCH("/secrets/:secret", session.MustPush, api.PatchSecret)
					repo.DELETE("/secrets/:secret", session.MustPush, api.DeleteSecret)

					// requires push permissions
					repo.GET("/registries", session.MustPush, api.GetRegistryList)
					repo.POST("/registries", session.MustPush, api.PostRegistry)
					repo.GET("/registries/:registry", session.MustPush, api.GetRegistry)
					repo.PATCH("/registries/:registry", session.MustPush, api.PatchRegistry)
					repo.DELETE("/registries/:registry", session.MustPush, api.DeleteRegistry)

					// requires push permissions
					repo.GET("/cron", session.MustPush, api.GetCronList)
					repo.POST("/cron", session.MustPush, api.PostCron)
					repo.GET("/cron/:cron", session.MustPush, api.GetCron)
					repo.POST("/cron/:cron", session.MustPush, api.RunCron)
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
		}

		badges := apiBase.Group("/badges/:repo_id_or_owner")
		{
			badges.GET("/status.svg", api.GetBadge)
			badges.GET("/cc.xml", api.GetCC)
		}

		_badges := apiBase.Group("/badges/:repo_id_or_owner/:repo_name")
		{
			_badges.GET("/status.svg", api.GetBadge)
			_badges.GET("/cc.xml", api.GetCC)
		}

		pipelines := apiBase.Group("/pipelines")
		{
			pipelines.Use(session.MustAdmin())
			pipelines.GET("", api.GetPipelineQueue)
		}

		queue := apiBase.Group("/queue")
		{
			queue.Use(session.MustAdmin())
			queue.GET("/info", api.GetQueueInfo)
			queue.POST("/pause", api.PauseQueue)
			queue.POST("/resume", api.ResumeQueue)
			queue.GET("/norunningpipelines", api.BlockTilQueueHasRunningItem)
		}

		// global secrets can be read without actual values by any user
		readGlobalSecrets := apiBase.Group("/secrets")
		{
			readGlobalSecrets.Use(session.MustUser())
			readGlobalSecrets.GET("", api.GetGlobalSecretList)
			readGlobalSecrets.GET("/:secret", api.GetGlobalSecret)
		}
		secrets := apiBase.Group("/secrets")
		{
			secrets.Use(session.MustAdmin())
			secrets.POST("", api.PostGlobalSecret)
			secrets.PATCH("/:secret", api.PatchGlobalSecret)
			secrets.DELETE("/:secret", api.DeleteGlobalSecret)
		}

		// global registries can be read without actual values by any user
		readGlobalRegistries := apiBase.Group("/registries")
		{
			readGlobalRegistries.Use(session.MustUser())
			readGlobalRegistries.GET("", api.GetGlobalRegistryList)
			readGlobalRegistries.GET("/:registry", api.GetGlobalRegistry)
		}
		registries := apiBase.Group("/registries")
		{
			registries.Use(session.MustAdmin())
			registries.POST("", api.PostGlobalRegistry)
			registries.PATCH("/:registry", api.PatchGlobalRegistry)
			registries.DELETE("/:registry", api.DeleteGlobalRegistry)
		}

		logLevel := apiBase.Group("/log-level")
		{
			logLevel.Use(session.MustAdmin())
			logLevel.GET("", api.LogLevel)
			logLevel.POST("", api.SetLogLevel)
		}

		agentBase := apiBase.Group("/agents")
		{
			agentBase.Use(session.MustAdmin())
			agentBase.GET("", api.GetAgents)
			agentBase.POST("", api.PostAgent)
			agentBase.GET("/:agent", api.GetAgent)
			agentBase.GET("/:agent/tasks", api.GetAgentTasks)
			agentBase.PATCH("/:agent", api.PatchAgent)
			agentBase.DELETE("/:agent", api.DeleteAgent)
		}

		apiBase.GET("/forges", api.GetForges)
		apiBase.GET("/forges/:forgeId", api.GetForge)
		forgeBase := apiBase.Group("/forges")
		{
			forgeBase.Use(session.MustAdmin())
			forgeBase.POST("", api.PostForge)
			forgeBase.PATCH("/:forgeId", api.PatchForge)
			forgeBase.DELETE("/:forgeId", api.DeleteForge)
		}

		apiBase.GET("/signature/public-key", session.MustUser(), api.GetSignaturePublicKey)

		apiBase.POST("/hook", api.PostHook)

		stream := apiBase.Group("/stream")
		{
			stream.GET("/logs/:repo_id/:pipeline/:stepId",
				session.SetRepo(),
				session.SetPerm(),
				session.MustPull,
				api.LogStreamSSE)
			stream.GET("/events", api.EventStreamSSE)
		}

		if zerolog.GlobalLevel() <= zerolog.DebugLevel {
			debugger := apiBase.Group("/debug")
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
	}
}
