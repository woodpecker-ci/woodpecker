// Copyright 2024 Woodpecker Authors
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

package config

import (
	"context"
	"errors"
	"slices"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

type Config struct {
	ServerURL string `json:"server_url"`
	Token     string `json:"-"`
	LogLevel  string `json:"log_level"`
}

var skipSetupForCommands = []string{"setup", "help", "h", "version", "update", "lint", "exec", "decode-base64", "completion", "", "context", "ctx"}

func Load(ctx context.Context, c *cli.Command) error {
	if firstArg := c.Args().First(); slices.Contains(skipSetupForCommands, firstArg) {
		return nil
	}

	contextConfig, contextErr := GetCurrentContext(ctx, c)
	if contextErr == nil {
		if !c.IsSet("server") {
			err := c.Set("server", contextConfig.ServerURL)
			if err != nil {
				return err
			}
		}
		if !c.IsSet("token") {
			err := c.Set("token", contextConfig.Token)
			if err != nil {
				return err
			}
		}
		if !c.IsSet("log-level") && contextConfig.LogLevel != "" {
			err := c.Set("log-level", contextConfig.LogLevel)
			if err != nil {
				return err
			}
		}
		log.Debug().Any("config", contextConfig).Msg("loaded config from context")
		return nil
	}

	if c.String("server") != "" && c.String("token") != "" {
		return nil
	}

	log.Info().Msg("woodpecker-cli is not set up, run `woodpecker-cli setup` to create a context")
	return errors.New("woodpecker-cli is not configured")
}
