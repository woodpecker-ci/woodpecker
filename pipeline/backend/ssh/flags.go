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

package ssh

import (
	"github.com/urfave/cli/v2"
)

var Flags = []cli.Flag{
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
}
