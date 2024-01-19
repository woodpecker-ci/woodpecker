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
	"log/slog"
	"sync"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline"
	backend "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/shared/logger/errorattr"
)

func (r *Runner) createLogger(logger *slog.Logger, uploads *sync.WaitGroup, workflow *rpc.Workflow) pipeline.Logger {
	return func(step *backend.Step, rc io.Reader) error {
		loglogger := logger.With(slog.String("image", step.Image), slog.String("workflowID", workflow.ID))

		uploads.Add(1)

		var secrets []string
		for _, secret := range workflow.Config.Secrets {
			if secret.Mask {
				secrets = append(secrets, secret.Value)
			}
		}

		loglogger.Debug("log stream opened")

		logStream := rpc.NewLineWriter(r.client, step.UUID, secrets...)
		if _, err := io.Copy(logStream, rc); err != nil {
			slog.Error("copy limited logStream part", errorattr.Default(err))
		}

		loglogger.Debug("log stream copied, close ...")
		uploads.Done()

		return nil
	}
}
