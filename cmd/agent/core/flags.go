// Copyright 2022 Woodpecker Authors
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

package core

import (
	"os"
	"time"

	"github.com/urfave/cli/v3"
)

//nolint:mnd
var flags = []cli.Flag{
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_SERVER"),
		Name:    "server",
		Usage:   "server address",
		Value:   "localhost:9000",
	},
	&cli.StringFlag{
		Name:  "grpc-token",
		Usage: "server-agent shared token",
		Sources: cli.NewValueSourceChain(
			cli.File(os.Getenv("WOODPECKER_AGENT_SECRET_FILE")),
			cli.EnvVar("WOODPECKER_AGENT_SECRET")),
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_GRPC_SECURE"),
		Name:    "grpc-secure",
		Usage:   "should the connection to WOODPECKER_SERVER be made using a secure transport",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_GRPC_VERIFY"),
		Name:    "grpc-skip-insecure",
		Usage:   "should the grpc server certificate be verified, only valid when WOODPECKER_GRPC_SECURE is true",
		Value:   true,
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_HOSTNAME"),
		Name:    "hostname",
		Usage:   "agent hostname",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_AGENT_CONFIG_FILE"),
		Name:    "agent-config",
		Usage:   "agent config file path, if set empty the agent will be stateless and unregister on termination",
		Value:   "/etc/woodpecker/agent.conf",
	},
	&cli.StringSliceFlag{
		Sources: cli.EnvVars("WOODPECKER_AGENT_LABELS", "WOODPECKER_FILTER_LABELS"), // remove WOODPECKER_FILTER_LABELS in v4.x
		Name:    "labels",
		Aliases: []string{"filter"}, // remove in v4.x
		Usage:   "List of labels to filter tasks on. An agent must be assigned every tag listed in a task to be selected.",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("WOODPECKER_MAX_WORKFLOWS", "WOODPECKER_MAX_PROCS"), // cspell:words PROCS
		Name:    "max-workflows",
		Usage:   "agent parallel workflows",
		Value:   1,
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_HEALTHCHECK"),
		Name:    "healthcheck",
		Usage:   "enable healthcheck endpoint",
		Value:   true,
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_HEALTHCHECK_ADDR"),
		Name:    "healthcheck-addr",
		Usage:   "healthcheck endpoint address",
		Value:   ":3000",
	},
	&cli.DurationFlag{
		Sources: cli.EnvVars("WOODPECKER_KEEPALIVE_TIME"),
		Name:    "keepalive-time",
		Usage:   "after a duration of this time of no activity, the agent pings the server to check if the transport is still alive",
	},
	&cli.DurationFlag{
		Sources: cli.EnvVars("WOODPECKER_KEEPALIVE_TIMEOUT"),
		Name:    "keepalive-timeout",
		Usage:   "after pinging for a keepalive check, the agent waits for a duration of this time before closing the connection if no activity",
		Value:   time.Second * 20,
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND"),
		Name:    "backend-engine",
		Usage:   "backend to run pipelines on",
		Value:   "auto-detect",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("WOODPECKER_CONNECT_RETRY_COUNT"),
		Name:    "connect-retry-count",
		Usage:   "number of times to retry connecting to the server",
		Value:   5,
	},
	&cli.DurationFlag{
		Sources: cli.EnvVars("WOODPECKER_CONNECT_RETRY_DELAY"),
		Name:    "connect-retry-delay",
		Usage:   "duration to wait before retrying to connect to the server",
		Value:   time.Second * 2,
	},
}
