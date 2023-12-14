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

package common

import (
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cmd/common"
)

var GlobalFlags = append([]cli.Flag{
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_TOKEN"},
		Name:    "token",
		Aliases: []string{"t"},
		Usage:   "server auth token",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_SERVER"},
		Name:    "server",
		Aliases: []string{"s"},
		Usage:   "server address",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_SKIP_VERIFY"},
		Name:    "skip-verify",
		Usage:   "skip ssl verification",
		Hidden:  true,
	},
	&cli.StringFlag{
		EnvVars: []string{"SOCKS_PROXY"},
		Name:    "socks-proxy",
		Usage:   "socks proxy address",
		Hidden:  true,
	},
	&cli.BoolFlag{
		EnvVars: []string{"SOCKS_PROXY_OFF"},
		Name:    "socks-proxy-off",
		Usage:   "socks proxy ignored",
		Hidden:  true,
	},
}, common.GlobalLoggerFlags...)

// FormatFlag return format flag with value set based on template
// if hidden value is set, flag will be hidden
func FormatFlag(tmpl string, hidden ...bool) *cli.StringFlag {
	return &cli.StringFlag{
		Name:   "format",
		Usage:  "format output",
		Value:  tmpl,
		Hidden: len(hidden) != 0,
	}
}

// specify repository
var RepoFlag = &cli.StringFlag{
	Name:    "repository",
	Aliases: []string{"repo"},
	Usage:   "repository id or full-name (e.g. 134 or octocat/hello-world)",
}

var OrgFlag = &cli.StringFlag{
	Name:    "organization",
	Aliases: []string{"org"},
	Usage:   "organization id or full-name (e.g. 123 or octocat)",
}
