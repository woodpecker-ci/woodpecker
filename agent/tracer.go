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
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
)

func (r *Runner) createTracer(ctxMeta context.Context, uploads *sync.WaitGroup, logger zerolog.Logger, workflow *rpc.Workflow) pipeline.TraceFunc {
	return func(state *pipeline.State) error {
		uploads.Add(1)

		stepLogger := logger.With().
			Str("image", state.Pipeline.Step.Image).
			Str("workflow_id", workflow.ID).
			Err(state.Process.Error).
			Int("exit_code", state.Process.ExitCode).
			Bool("exited", state.Process.Exited).
			Logger()

		stepState := rpc.StepState{
			StepUUID: state.Pipeline.Step.UUID,
			Exited:   state.Process.Exited,
			ExitCode: state.Process.ExitCode,
			Started:  time.Now().Unix(), // TODO: do not do this
			Finished: time.Now().Unix(),
		}
		if state.Process.Error != nil {
			stepState.Error = state.Process.Error.Error()
		}

		defer func() {
			stepLogger.Debug().Msg("update step status")

			if err := r.client.Update(ctxMeta, workflow.ID, stepState); err != nil {
				stepLogger.Debug().
					Err(err).
					Msg("update step status error")
			}

			stepLogger.Debug().Msg("update step status complete")
			uploads.Done()
		}()
		if state.Process.Exited {
			return nil
		}
		if state.Pipeline.Step.Environment == nil {
			state.Pipeline.Step.Environment = map[string]string{}
		}

		// TODO: find better way to update this state and move it to pipeline to have the same env in cli-exec
		state.Pipeline.Step.Environment["CI_MACHINE"] = r.hostname

		state.Pipeline.Step.Environment["CI_PIPELINE_STARTED"] = strconv.FormatInt(state.Pipeline.Started, 10)

		state.Pipeline.Step.Environment["CI_STEP_STARTED"] = strconv.FormatInt(state.Pipeline.Started, 10)

		state.Pipeline.Step.Environment["CI_SYSTEM_PLATFORM"] = runtime.GOOS + "/" + runtime.GOARCH

		return nil
	}
}
