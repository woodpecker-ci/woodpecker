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
	"bufio"
	"io"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline"
	backend "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/log"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
)

const (
	// writeBufferSize = 4 * 1024 * 1024 // 4mb
	flushInterval = 250 * time.Millisecond
)

func (r *Runner) createLogger(_logger zerolog.Logger, uploads *sync.WaitGroup, workflow *rpc.Workflow) pipeline.Logger {
	return func(step *backend.Step, rc io.Reader) error {
		logger := _logger.With().
			Str("image", step.Image).
			Str("workflowID", workflow.ID).
			Logger()

		uploads.Add(1)

		var secrets []string
		for _, secret := range workflow.Config.Secrets {
			secrets = append(secrets, secret.Value)
		}

		logger.Debug().Msg("log stream opened")

		bufReader := bufio.NewReader(rc)
		logStream := log.NewLineWriter(r.client, step.UUID, flushInterval, secrets...)
		defer logStream.Close()
		if _, err := io.Copy(logStream, bufReader); err != nil {
			logger.Error().Err(err).Msg("copy limited logStream part")
		}

		logger.Debug().Msg("log stream copied, close ...")
		uploads.Done()

		return nil
	}
}
