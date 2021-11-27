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

package main

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/build"
	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/deploy"
	"github.com/woodpecker-ci/woodpecker/cli/exec"
	"github.com/woodpecker-ci/woodpecker/cli/info"
	"github.com/woodpecker-ci/woodpecker/cli/lint"
	"github.com/woodpecker-ci/woodpecker/cli/log"
	"github.com/woodpecker-ci/woodpecker/cli/loglevel"
	"github.com/woodpecker-ci/woodpecker/cli/registry"
	"github.com/woodpecker-ci/woodpecker/cli/repo"
	"github.com/woodpecker-ci/woodpecker/cli/secret"
	"github.com/woodpecker-ci/woodpecker/cli/user"
	"github.com/woodpecker-ci/woodpecker/version"
)

func main() {
	app := cli.NewApp()
	app.Name = "woodpecker-cli"
	app.Version = version.String()
	app.Usage = "command line utility"
	app.EnableBashCompletion = true
	app.Flags = common.GlobalFlags
	app.Commands = []*cli.Command{
		build.Command,
		log.Command,
		deploy.Command,
		exec.Command,
		info.Command,
		registry.Command,
		secret.Command,
		repo.Command,
		user.Command,
		lint.Command,
		loglevel.Command,
	}

	zlog.Logger = zlog.Output(
		zerolog.ConsoleWriter{
			Out: os.Stderr,
		},
	)

	if err := app.Run(os.Args); err != nil {
		zlog.Fatal().Err(err).Msg("error running cli")
	}
}
