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

package common

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func SetupConsoleLogger(c *cli.Context) error {
	level := c.String("log-level")
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		log.Fatal().Msgf("unknown logging level: %s", level)
	}
	zerolog.SetGlobalLevel(lvl)
	if zerolog.GlobalLevel() <= zerolog.DebugLevel {
		log.Logger = log.With().Caller().Logger()
		log.Log().Msgf("LogLevel = %s", zerolog.GlobalLevel().String())
	}
	return nil
}
