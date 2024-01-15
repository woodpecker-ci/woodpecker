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

package main

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	_ "go.woodpecker-ci.org/woodpecker/v2/cmd/server/docs"
	"go.woodpecker-ci.org/woodpecker/v2/version"
)

func main() {
	app := cli.NewApp()
	app.Name = "woodpecker-server"
	app.Version = version.String()
	app.Usage = "woodpecker server"
	app.Action = run
	app.Commands = []*cli.Command{
		{
			Name:   "ping",
			Usage:  "ping the server",
			Action: pinger,
		},
	}
	app.Flags = flags

	setupSwaggerStaticConfig()

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msgf("error running server") //nolint:forbidigo
	}
}
