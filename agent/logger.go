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
	"crypto/sha256"
	"io"
	"sync"

	hashvalue_replacer "github.com/6543/go-hashvalue-replacer"
	"github.com/rs/zerolog"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline"
	backend "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/log"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
)

func (r *Runner) createLogger(_logger zerolog.Logger, uploads *sync.WaitGroup, workflow *rpc.Workflow) pipeline.Logger {
	return func(step *backend.Step, rc io.ReadCloser) error {
		defer rc.Close()

		logger := _logger.With().
			Str("image", step.Image).
			Logger()

		uploads.Add(1)
		defer uploads.Done()

		logger.Debug().Msg("log stream opened")

		// mask secrets from reader
		maskedReader, err := hashvalue_replacer.NewReader(rc, workflow.Config.SecretMask.Salt, workflow.Config.SecretMask.Hashes, workflow.Config.SecretMask.Lengths, hashvalue_replacer.Options{
			Mask: "********",
			Hash: func(salt, data []byte) []byte {
				h := sha256.New()
				h.Write(salt)
				h.Write([]byte(data))
				return h.Sum(nil)
			},
		})
		if err != nil {
			logger.Error().Err(err).Msg("could not create masked reader")
			return nil
		}

		logStream := log.NewLineWriter(r.client, step.UUID)
		if err := log.CopyLineByLine(logStream, maskedReader, pipeline.MaxLogLineLength); err != nil {
			logger.Error().Err(err).Msg("copy limited logStream part")
		}

		logger.Debug().Msg("log stream copied, close ...")
		return nil
	}
}
