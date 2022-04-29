// Copyright 2019 Laszlo Fogas
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

package main

import (
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

var flags = []cli.Flag{
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_SERVER"},
		Name:    "server",
		Usage:   "server address",
		Value:   "localhost:9000",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_USERNAME"},
		Name:    "grpc-username",
		Usage:   "auth username",
		Value:   "x-oauth-basic",
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_AGENT_SECRET"},
		Name:     "grpc-password",
		Usage:    "server-agent shared password",
		FilePath: os.Getenv("WOODPECKER_AGENT_SECRET_FILE"),
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GRPC_SECURE"},
		Name:    "grpc-secure",
		Usage:   "should the connection to WOODPECKER_SERVER be made using a secure transport",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GRPC_VERIFY"},
		Name:    "grpc-skip-insecure",
		Usage:   "should the grpc server certificate be verified, only valid when WOODPECKER_GRPC_SECURE is true",
		Value:   true,
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_LOG_LEVEL"},
		Name:    "log-level",
		Usage:   "set logging level",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_DEBUG_PRETTY"},
		Name:    "pretty",
		Usage:   "enable pretty-printed debug output",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_DEBUG_NOCOLOR"},
		Name:    "nocolor",
		Usage:   "disable colored debug output",
		Value:   true,
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_HOSTNAME"},
		Name:    "hostname",
		Usage:   "agent hostname",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_FILTER"},
		Name:    "filter",
		Usage:   "filter expression to restrict builds by label",
	},
	&cli.IntFlag{
		EnvVars: []string{"WOODPECKER_MAX_PROCS"},
		Name:    "max-procs",
		Usage:   "agent parallel builds",
		Value:   1,
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_HEALTHCHECK"},
		Name:    "healthcheck",
		Usage:   "enable healthcheck endpoint",
		Value:   true,
	},
	&cli.DurationFlag{
		EnvVars: []string{"WOODPECKER_KEEPALIVE_TIME"},
		Name:    "keepalive-time",
		Usage:   "after a duration of this time of no activity, the agent pings the server to check if the transport is still alive",
	},
	&cli.DurationFlag{
		EnvVars: []string{"WOODPECKER_KEEPALIVE_TIMEOUT"},
		Name:    "keepalive-timeout",
		Usage:   "after pinging for a keepalive check, the agent waits for a duration of this time before closing the connection if no activity",
		Value:   time.Second * 20,
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND"},
		Name:    "backend-engine",
		Usage:   "backend engine to run pipelines on",
		Value:   "auto-detect",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_DOCKER_NETWORK"},
		Name:    "backend-docker-network",
		Usage:   "attach pipeline containers (steps) to an existing docker network",
		Value:   "",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_DOCKER_ENABLE_IPV6"},
		Name:    "backend-docker-ipv6",
		Usage:   "enable ipv6 for pipeline containers (steps)",
		Value:   false,
	},
}
