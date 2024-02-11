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

package debug

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

// IndexHandler
//
//	@Summary		List available pprof profiles (HTML)
//	@Description	Only available, when server was started with WOODPECKER_LOG_LEVEL=debug
//	@Router			/debug/pprof [get]
//	@Produce		html
//	@Success		200
//	@Tags			Process profiling and debugging
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func IndexHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Index(c.Writer, c.Request)
	}
}

// HeapHandler
//
//	@Summary		Get pprof heap dump, a sampling of memory allocations of live objects
//	@Description	Only available, when server was started with WOODPECKER_LOG_LEVEL=debug
//	@Router			/debug/pprof/heap [get]
//	@Produce		plain
//	@Success		200
//	@Tags			Process profiling and debugging
//	@Param			Authorization	header	string	true	"Insert your personal access token"									default(Bearer <personal access token>)
//	@Param			gc				query	string	false	"You can specify gc=heap to run GC before taking the heap sample"	default()
func HeapHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Handler("heap").ServeHTTP(c.Writer, c.Request)
	}
}

// GoroutineHandler
//
//	@Summary		Get pprof stack traces of all current goroutines
//	@Description	Only available, when server was started with WOODPECKER_LOG_LEVEL=debug
//	@Router			/debug/pprof/goroutine [get]
//	@Produce		plain
//	@Success		200
//	@Tags			Process profiling and debugging
//	@Param			Authorization	header	string	true	"Insert your personal access token"															default(Bearer <personal access token>)
//	@Param			debug			query	int		false	"Use debug=2 as a query parameter to export in the same format as an un-recovered panic"	default(1)
func GoroutineHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Handler("goroutine").ServeHTTP(c.Writer, c.Request)
	}
}

// BlockHandler
//
//	@Summary		Get pprof stack traces that led to blocking on synchronization primitives
//	@Description	Only available, when server was started with WOODPECKER_LOG_LEVEL=debug
//	@Router			/debug/pprof/block [get]
//	@Produce		plain
//	@Success		200
//	@Tags			Process profiling and debugging
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func BlockHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Handler("block").ServeHTTP(c.Writer, c.Request)
	}
}

// ThreadCreateHandler
//
//	@Summary		Get pprof stack traces that led to the creation of new OS threads
//	@Description	Only available, when server was started with WOODPECKER_LOG_LEVEL=debug
//	@Router			/debug/pprof/threadcreate [get]
//	@Produce		plain
//	@Success		200
//	@Tags			Process profiling and debugging
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func ThreadCreateHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Handler("threadcreate").ServeHTTP(c.Writer, c.Request)
	}
}

// CmdlineHandler
//
//	@Summary		Get the command line invocation of the current program
//	@Description	Only available, when server was started with WOODPECKER_LOG_LEVEL=debug
//	@Router			/debug/pprof/cmdline [get]
//	@Produce		plain
//	@Success		200
//	@Tags			Process profiling and debugging
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func CmdlineHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Cmdline(c.Writer, c.Request)
	}
}

// ProfileHandler
//
//	@Summary		Get pprof CPU profile
//	@Description	Only available, when server was started with WOODPECKER_LOG_LEVEL=debug
//	@Description	After you get the profile file, use the go tool pprof command to investigate the profile.
//	@Router			/debug/pprof/profile [get]
//	@Produce		plain
//	@Success		200
//	@Tags			Process profiling and debugging
//	@Param			Authorization	header	string	true	"Insert your personal access token"								default(Bearer <personal access token>)
//	@Param			seconds			query	int		true	"You can specify the duration in the seconds GET parameter."	default	(30)
func ProfileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Profile(c.Writer, c.Request)
	}
}

// SymbolHandler
//
//	@Summary		Get pprof program counters mapping to function names
//	@Description	Only available, when server was started with WOODPECKER_LOG_LEVEL=debug
//	@Description	Looks up the program counters listed in the request,
//	@Description	responding with a table mapping program counters to function names.
//	@Description	The requested program counters can be provided via GET + query parameters,
//	@Description	or POST + body parameters. Program counters shall be space delimited.
//	@Router			/debug/pprof/symbol [get]
//	@Router			/debug/pprof/symbol [post]
//	@Produce		plain
//	@Success		200
//	@Tags			Process profiling and debugging
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func SymbolHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Symbol(c.Writer, c.Request)
	}
}

// TraceHandler
//
//	@Summary		Get a trace of execution of the current program
//	@Description	Only available, when server was started with WOODPECKER_LOG_LEVEL=debug
//	@Description	After you get the profile file, use the go tool pprof command to investigate the profile.
//	@Router			/debug/pprof/trace [get]
//	@Produce		plain
//	@Success		200
//	@Tags			Process profiling and debugging
//	@Param			Authorization	header	string	true	"Insert your personal access token"								default(Bearer <personal access token>)
//	@Param			seconds			query	int		true	"You can specify the duration in the seconds GET parameter."	default	(30)
func TraceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pprof.Trace(c.Writer, c.Request)
	}
}
