// Copyright 2023 Woodpecker Authors
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

package docker

import (
	"github.com/urfave/cli/v3"
)

var Flags = []cli.Flag{
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_DOCKER_HOST", "DOCKER_HOST"),
		Name:    "backend-docker-host",
		Usage:   "path to docker socket or url to the docker server",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_DOCKER_API_VERSION", "DOCKER_API_VERSION"),
		Name:    "backend-docker-api-version",
		Usage:   "the version of the API to reach, leave empty for latest.",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_DOCKER_CERT_PATH", "DOCKER_CERT_PATH"),
		Name:    "backend-docker-cert",
		Usage:   "path to load the TLS certificates for connecting to docker server",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_DOCKER_TLS_VERIFY", "DOCKER_TLS_VERIFY"),
		Name:    "backend-docker-tls-verify",
		Usage:   "enable or disable TLS verification for connecting to docker server",
		Value:   true,
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_DOCKER_ENABLE_IPV6"),
		Name:    "backend-docker-ipv6",
		Usage:   "backend docker enable IPV6",
		Value:   false,
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_DOCKER_NETWORK"),
		Name:    "backend-docker-network",
		Usage:   "backend docker network",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_DOCKER_VOLUMES"),
		Name:    "backend-docker-volumes",
		Usage:   "backend docker volumes (comma separated)",
	},
	//
	// resource limit parameters
	//
	&cli.IntFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_DOCKER_LIMIT_MEM_SWAP", "WOODPECKER_LIMIT_MEM_SWAP"),
		Name:    "backend-docker-limit-mem-swap",
		Usage:   "maximum memory used for swap in bytes",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_DOCKER_LIMIT_MEM", "WOODPECKER_LIMIT_MEM"),
		Name:    "backend-docker-limit-mem",
		Usage:   "maximum memory allowed in bytes",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_DOCKER_LIMIT_SHM_SIZE", "WOODPECKER_LIMIT_SHM_SIZE"),
		Name:    "backend-docker-limit-shm-size",
		Usage:   "docker /dev/shm allowed in bytes",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_DOCKER_LIMIT_CPU_QUOTA", "WOODPECKER_LIMIT_CPU_QUOTA"),
		Name:    "backend-docker-limit-cpu-quota",
		Usage:   "impose a cpu quota",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_DOCKER_LIMIT_CPU_SHARES", "WOODPECKER_LIMIT_CPU_SHARES"),
		Name:    "backend-docker-limit-cpu-shares",
		Usage:   "change the cpu shares",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_DOCKER_LIMIT_CPU_SET", "WOODPECKER_LIMIT_CPU_SET"),
		Name:    "backend-docker-limit-cpu-set",
		Usage:   "set the cpus allowed to execute containers",
	},
}
