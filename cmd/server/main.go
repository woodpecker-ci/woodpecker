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

//go:build !generate && !man

package main

import (
	"context"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"

	_ "go.woodpecker-ci.org/woodpecker/v3/cmd/server/openapi"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

func main() {
	ctx := utils.WithContextSigtermCallback(context.Background(), func() {
		log.Info().Msg("termination signal is received, shutting down server")
	})

	app := genApp()

	setupOpenAPIStaticConfig()

	if err := app.Run(ctx, os.Args); err != nil {
		log.Error().Err(err).Msgf("error running server")
	}
}
