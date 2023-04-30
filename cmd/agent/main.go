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
	"fmt"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/woodpecker-ci/woodpecker/version"
)

const (
	retryCount = 5
	retryDelay = 2 * time.Second
)

func main() {
	app := cli.NewApp()
	app.Name = "woodpecker-agent"
	app.Version = version.String()
	app.Usage = "woodpecker agent"
	app.Action = func(context *cli.Context) error {
		var err error
		for i := 0; i < retryCount; i++ {
			if err = loop(context); err == nil {
				break
			} else if status.Code(err) == codes.Unavailable {
				log.Warn().Err(err).Msg(fmt.Sprintf("cannot connect to server, retrying in %v", retryDelay))
				time.Sleep(retryDelay)
			}
		}
		return err
	}
	app.Commands = []*cli.Command{
		{
			Name:   "ping",
			Usage:  "ping the agent",
			Action: pinger,
		},
	}
	app.Flags = flags

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
