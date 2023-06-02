// Copyright 2022 Woodpecker Authors
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

package agent

import (
	"context"
	"io"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/pipeline"
	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/pipeline/multipart"
	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
)

func (r *Runner) createLogger(_ context.Context, logger zerolog.Logger, uploads *sync.WaitGroup, work *rpc.Pipeline) pipeline.LogFunc {
	return func(step *backend.Step, rc multipart.Reader) error {
		loglogger := logger.With().
			Str("image", step.Image).
			Str("stage", step.Alias).
			Logger()

		part, rerr := rc.NextPart()
		if rerr != nil {
			return rerr
		}
		uploads.Add(1)

		var secrets []string
		for _, secret := range work.Config.Secrets {
			if secret.Mask {
				secrets = append(secrets, secret.Value)
			}
		}

		loglogger.Debug().Msg("log stream opened")

		logStream := rpc.NewLineWriter(r.client, work.ID, step.Alias, secrets...)
		if _, err := io.Copy(logStream, part); err != nil {
			log.Error().Err(err).Msg("copy limited logStream part")
		}

		loglogger.Debug().Msg("log stream copied")

		defer func() {
			loglogger.Debug().Msg("log stream closed")
			uploads.Done()
		}()

		return nil
	}
}
