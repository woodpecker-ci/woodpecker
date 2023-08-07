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
		EnvVars:  []string{"WOODPECKER_AGENT_SECRET"},
		Name:     "grpc-token",
		Usage:    "server-agent shared token",
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
		EnvVars: []string{"WOODPECKER_AGENT_CONFIG_FILE"},
		Name:    "agent-config",
		Usage:   "agent config file path",
		Value:   "/etc/woodpecker/agent.conf",
	},
	&cli.StringSliceFlag{
		EnvVars: []string{"WOODPECKER_FILTER_LABELS"},
		Name:    "filter",
		Usage:   "List of labels to filter tasks on. An agent must be assigned every tag listed in a task to be selected.",
	},
	&cli.IntFlag{
		EnvVars: []string{"WOODPECKER_MAX_WORKFLOWS", "WOODPECKER_MAX_PROCS"},
		Name:    "max-workflows",
		Usage:   "agent parallel workflows",
		Value:   1,
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_HEALTHCHECK"},
		Name:    "healthcheck",
		Usage:   "enable healthcheck endpoint",
		Value:   true,
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_HEALTHCHECK_ADDR"},
		Name:    "healthcheck-addr",
		Usage:   "healthcheck endpoint address",
		Value:   ":3000",
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

	// backend docker
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_DOCKER_HOST", "DOCKER_HOST"},
		Name:    "backend-docker-host",
		Usage:   "path to docker socket or url to the docker server",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_DOCKER_API_VERSION", "DOCKER_API_VERSION"},
		Name:    "backend-docker-api-version",
		Usage:   "the version of the API to reach, leave empty for latest.",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_DOCKER_CERT_PATH", "DOCKER_CERT_PATH"},
		Name:    "backend-docker-cert",
		Usage:   "path to load the TLS certificates for connecting to docker server",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_DOCKER_TLS_VERIFY", "DOCKER_TLS_VERIFY"},
		Name:    "backend-docker-tls-verify",
		Usage:   "enable or disable TLS verification for connecting to docker server",
		Value:   true,
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_DOCKER_ENABLE_IPV6"},
		Name:    "backend-docker-ipv6",
		Usage:   "backend docker enable IPV6",
		Value:   false,
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_DOCKER_NETWORK"},
		Name:    "backend-docker-network",
		Usage:   "backend docker network",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_DOCKER_VOLUMES"},
		Name:    "backend-docker-volumes",
		Usage:   "backend docker volumes (comma separated)",
	},

	// backend ssh
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_SSH_ADDRESS"},
		Name:    "backend-ssh-address",
		Usage:   "backend ssh address",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_SSH_USER"},
		Name:    "backend-ssh-user",
		Usage:   "backend ssh user",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_SSH_KEY"},
		Name:    "backend-ssh-key",
		Usage:   "backend ssh key file",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_SSH_KEY_PASSWORD"},
		Name:    "backend-ssh-key-password",
		Usage:   "backend ssh key password",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_SSH_PASSWORD"},
		Name:    "backend-ssh-password",
		Usage:   "backend ssh password",
	},

	// backend k8s
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_K8S_NAMESPACE"},
		Name:    "backend-k8s-namespace",
		Usage:   "backend k8s namespace",
		Value:   "woodpecker",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_K8S_VOLUME_SIZE"},
		Name:    "backend-k8s-volume-size",
		Usage:   "backend k8s volume size (default 10G)",
		Value:   "10G",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_K8S_STORAGE_CLASS"},
		Name:    "backend-k8s-storage-class",
		Usage:   "backend k8s storage class",
		Value:   "",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_K8S_STORAGE_RWX"},
		Name:    "backend-k8s-storage-rwx",
		Usage:   "backend k8s storage access mode, should ReadWriteMany (RWX) instead of ReadWriteOnce (RWO) be used? (default: true)",
		Value:   true,
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_K8S_POD_LABELS"},
		Name:    "backend-k8s-pod-labels",
		Usage:   "backend k8s additional worker pod labels",
		Value:   "",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_K8S_POD_ANNOTATIONS"},
		Name:    "backend-k8s-pod-annotations",
		Usage:   "backend k8s additional worker pod annotations",
		Value:   "",
	},
	&cli.IntFlag{
		EnvVars: []string{"WOODPECKER_CONNECT_RETRY_COUNT"},
		Name:    "connect-retry-count",
		Usage:   "number of times to retry connecting to the server",
		Value:   5,
	},
	&cli.DurationFlag{
		EnvVars: []string{"WOODPECKER_CONNECT_RETRY_DELAY"},
		Name:    "connect-retry-delay",
		Usage:   "duration to wait before retrying to connect to the server",
		Value:   time.Second * 2,
	},
}
