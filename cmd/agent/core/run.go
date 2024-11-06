// Copyright 2024 Woodpecker Authors
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
	"context"
	"os"

	// Load config from .env file.
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	backend "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v2/shared/logger"
	"go.woodpecker-ci.org/woodpecker/v2/shared/utils"
	"go.woodpecker-ci.org/woodpecker/v2/version"
)

func RunAgent(ctx context.Context, backends []backend.Backend, metadata *metadata.Metadata) {
	app := &cli.Command{}
	app.Name = "woodpecker-agent"
	app.Version = version.String()
	app.Usage = "woodpecker agent"
	app.Action = runWithRetry(backends, metadata)
	app.Commands = []*cli.Command{
		{
			Name:   "ping",
			Usage:  "ping the agent",
			Action: pinger,
		},
	}
	agentFlags := utils.MergeSlices(flags, logger.GlobalLoggerFlags)
	for _, b := range backends {
		agentFlags = utils.MergeSlices(agentFlags, b.Flags())
	}
	app.Flags = agentFlags

	if err := app.Run(ctx, os.Args); err != nil {
		log.Fatal().Err(err).Msg("error running agent") //nolint:forbidigo
	}
}
