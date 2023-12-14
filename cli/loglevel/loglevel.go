// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package loglevel

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

// Command exports the log-level command used to change the servers log-level.
var Command = &cli.Command{
	Name:      "log-level",
	ArgsUsage: "[level]",
	Usage:     "get the logging level of the server, or set it with [level]",
	Action:    logLevel,
}

func logLevel(c *cli.Context) error {
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	var ll *woodpecker.LogLevel
	arg := c.Args().First()
	if arg != "" {
		lvl, err := zerolog.ParseLevel(arg)
		if err != nil {
			return err
		}
		ll, err = client.SetLogLevel(&woodpecker.LogLevel{
			Level: lvl.String(),
		})
		if err != nil {
			return err
		}
	} else {
		ll, err = client.LogLevel()
		if err != nil {
			return err
		}
	}

	log.Info().Msgf("Logging level: %s", ll.Level)
	return nil
}
