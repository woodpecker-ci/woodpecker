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
	"fmt"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/shared/logger"
)

var GlobalFlags = append([]cli.Flag{
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_CONFIG"),
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "path to config file",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_SERVER"),
		Name:    "server",
		Aliases: []string{"s"},
		Usage:   "server address",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_TOKEN"),
		Name:    "token",
		Aliases: []string{"t"},
		Usage:   "server auth token",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_DISABLE_UPDATE_CHECK"),
		Name:    "disable-update-check",
		Usage:   "disable update check",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_SKIP_VERIFY"),
		Name:    "skip-verify",
		Usage:   "skip ssl verification",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("SOCKS_PROXY"),
		Name:    "socks-proxy",
		Usage:   "socks proxy address",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("SOCKS_PROXY_OFF"),
		Name:    "socks-proxy-off",
		Usage:   "socks proxy ignored",
	},
}, logger.GlobalLoggerFlags...)

// FormatFlag return format flag with value set based on template
// if hidden value is set, flag will be hidden.
func FormatFlag(tmpl string, deprecated bool, hidden ...bool) *cli.StringFlag {
	usage := "format output"
	if deprecated {
		usage = fmt.Sprintf("%s (deprecated)", usage)
	}

	return &cli.StringFlag{
		Name:   "format",
		Usage:  usage,
		Value:  tmpl,
		Hidden: len(hidden) != 0,
	}
}

// OutputFlags returns a slice of cli.Flag containing output format options.
func OutputFlags(def string) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "output",
			Usage: "output format",
			Value: def,
		},
		&cli.BoolFlag{
			Name:  "output-no-headers",
			Usage: "don't print headers",
		},
	}
}

var RepoFlag = &cli.StringFlag{
	Name:    "repository",
	Aliases: []string{"repo"},
	Usage:   "repository id or full name (e.g. 134 or octocat/hello-world)",
}

var OrgFlag = &cli.StringFlag{
	Name:    "organization",
	Aliases: []string{"org"},
	Usage:   "organization id or full name (e.g. 123 or octocat)",
}
