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
	"io"

	"github.com/rs/zerolog"

	"go.woodpecker-ci.org/woodpecker/v3/agent/log"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/logging"
	pipeline_utils "go.woodpecker-ci.org/woodpecker/v3/pipeline/utils"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
)

func (r *Runner) createLogger(_logger zerolog.Logger, workflow *rpc.Workflow) logging.Logger {
	return func(step *backend_types.Step, rc io.ReadCloser) error {
		defer rc.Close()

		logger := _logger.With().
			Str("image", step.Image).
			Logger()

		var secrets []string
		for _, secret := range workflow.Config.Secrets {
			secrets = append(secrets, secret.Value)
		}

		logger.Debug().Msg("log stream opened")

		logStream := log.NewLineWriter(r.client, step.UUID, secrets...)
		if err := pipeline_utils.CopyLineByLine(logStream, rc, pipeline.MaxLogLineLength); err != nil {
			logger.Error().Err(err).Msg("copy limited logStream part")
		}

		logger.Debug().Msg("log stream copied, close ...")
		return nil
	}
}
