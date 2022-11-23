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
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/pipeline"
	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/pipeline/multipart"
	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
)

func (r *Runner) createLogger(logger zerolog.Logger, ctxmeta context.Context, uploads *sync.WaitGroup, work *rpc.Pipeline) pipeline.LogFunc {
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

		limitedPart := io.LimitReader(part, maxLogsUpload)
		logStream := rpc.NewLineWriter(r.client, work.ID, step.Alias, secrets...)
		if _, err := io.Copy(logStream, limitedPart); err != nil {
			log.Error().Err(err).Msg("copy limited logStream part")
		}

		loglogger.Debug().Msg("log stream copied")

		data, err := json.Marshal(logStream.Lines())
		if err != nil {
			loglogger.Err(err).Msg("could not marshal logstream")
		}

		file := &rpc.File{
			Mime: "application/json+logs",
			Step: step.Alias,
			Name: "logs.json",
			Data: data,
			Size: len(data),
			Time: time.Now().Unix(),
		}

		loglogger.Debug().Msg("log stream uploading")
		if serr := r.client.Upload(ctxmeta, work.ID, file); serr != nil {
			loglogger.Error().Err(serr).Msg("log stream upload error")
		} else {
			loglogger.Debug().Msg("log stream upload complete")
		}

		defer func() {
			loglogger.Debug().Msg("log stream closed")
			uploads.Done()
		}()

		part, rerr = rc.NextPart()
		if rerr != nil {
			return nil
		}
		// TODO should be configurable
		limitedPart = io.LimitReader(part, maxFileUpload)
		data, err = io.ReadAll(limitedPart)
		if err != nil {
			loglogger.Err(err).Msg("could not read limited part")
		}

		file = &rpc.File{
			Mime: part.Header().Get("Content-Type"),
			Step: step.Alias,
			Name: part.FileName(),
			Data: data,
			Size: len(data),
			Time: time.Now().Unix(),
			Meta: make(map[string]string),
		}
		for key, value := range part.Header() {
			file.Meta[key] = value[0]
		}

		loglogger.Debug().
			Str("file", file.Name).
			Str("mime", file.Mime).
			Msg("file stream uploading")

		if serr := r.client.Upload(ctxmeta, work.ID, file); serr != nil {
			loglogger.Error().
				Err(serr).
				Str("file", file.Name).
				Str("mime", file.Mime).
				Msg("file stream upload error")
		}

		loglogger.Debug().
			Str("file", file.Name).
			Str("mime", file.Mime).
			Msg("file stream upload complete")
		return nil
	}
}
