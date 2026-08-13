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
	"encoding/base64"
	"fmt"
	"os"
	"slices"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/shared/logger"
	"go.woodpecker-ci.org/woodpecker/v3/version"
)

func base64Decoder(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() != 1 {
		return fmt.Errorf("expected exactly one base64 argument")
	}

	decoded, err := base64.StdEncoding.DecodeString(c.Args().First())
	if err != nil {
		return fmt.Errorf("decode base64: %w", err)
	}

	_, err = os.Stdout.Write(decoded)
	return err
}

func GenApp(backends []backend_types.Backend) *cli.Command {
	app := &cli.Command{}
	app.Name = "woodpecker-agent"
	app.Version = version.String()
	app.Usage = "woodpecker agent"
	app.Action = runWithRetry(backends)
	app.Commands = []*cli.Command{
		{
			Name:   "ping",
			Usage:  "ping the agent",
			Action: pinger,
		},
		{
			Name:   "decode-base64",
			Usage:  "decodes a base64 string",
			Hidden: true,
			Action: base64Decoder,
		},
	}
	agentFlags := slices.Concat(flags, logger.GlobalLoggerFlags)
	for _, b := range backends {
		agentFlags = slices.Concat(agentFlags, b.Flags())
	}
	app.Flags = agentFlags
	return app
}

func RunAgent(ctx context.Context, backends []backend_types.Backend) {
	app := GenApp(backends)

	if err := app.Run(ctx, os.Args); err != nil {
		log.Fatal().Err(err).Msg("error running agent") //nolint:forbidigo
	}
}
