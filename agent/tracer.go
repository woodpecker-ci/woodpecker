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
	"time"

	"github.com/rs/zerolog"

	"github.com/woodpecker-ci/woodpecker/pipeline"
	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
)

func (r *Runner) createTracer(ctxmeta context.Context, logger zerolog.Logger, work *rpc.Pipeline) pipeline.TraceFunc {
	return func(state *pipeline.State) error {
		steplogger := logger.With().
			Str("image", state.Pipeline.Step.Image).
			Str("stage", state.Pipeline.Step.Alias).
			Err(state.Process.Error).
			Int("exit_code", state.Process.ExitCode).
			Bool("exited", state.Process.Exited).
			Logger()

		stepState := rpc.State{
			Step:     state.Pipeline.Step.Alias,
			Exited:   state.Process.Exited,
			ExitCode: state.Process.ExitCode,
			Started:  time.Now().Unix(), // TODO do not do this
			Finished: time.Now().Unix(),
		}
		if state.Process.Error != nil {
			stepState.Error = state.Process.Error.Error()
		}

		defer func() {
			steplogger.Debug().Msg("update step status")

			if uerr := r.client.Update(ctxmeta, work.ID, stepState); uerr != nil {
				steplogger.Debug().
					Err(uerr).
					Msg("update step status error")
			}

			steplogger.Debug().Msg("update step status complete")
		}()
		if state.Process.Exited {
			return nil
		}
		if state.Pipeline.Step.Environment == nil {
			state.Pipeline.Step.Environment = map[string]string{}
		}

		// TODO: find better way to update this state and move it to pipeline to have the same env in cli-exec
		state.Pipeline.Step.Environment["CI_MACHINE"] = r.hostname

		state.Pipeline.Step.Environment["CI_PIPELINE_STATUS"] = "success"
		state.Pipeline.Step.Environment["CI_PIPELINE_STARTED"] = strconv.FormatInt(state.Pipeline.Time, 10)
		state.Pipeline.Step.Environment["CI_PIPELINE_FINISHED"] = strconv.FormatInt(time.Now().Unix(), 10)

		state.Pipeline.Step.Environment["CI_STEP_STATUS"] = "success"
		state.Pipeline.Step.Environment["CI_STEP_STARTED"] = strconv.FormatInt(state.Pipeline.Time, 10)
		state.Pipeline.Step.Environment["CI_STEP_FINISHED"] = strconv.FormatInt(time.Now().Unix(), 10)

		state.Pipeline.Step.Environment["CI_SYSTEM_ARCH"] = runtime.GOOS + "/" + runtime.GOARCH

		// DEPRECATED
		state.Pipeline.Step.Environment["CI_BUILD_STATUS"] = "success"
		state.Pipeline.Step.Environment["CI_BUILD_STARTED"] = strconv.FormatInt(state.Pipeline.Time, 10)
		state.Pipeline.Step.Environment["CI_BUILD_FINISHED"] = strconv.FormatInt(time.Now().Unix(), 10)
		state.Pipeline.Step.Environment["CI_JOB_STATUS"] = "success"
		state.Pipeline.Step.Environment["CI_JOB_STARTED"] = strconv.FormatInt(state.Pipeline.Time, 10)
		state.Pipeline.Step.Environment["CI_JOB_FINISHED"] = strconv.FormatInt(time.Now().Unix(), 10)

		if state.Pipeline.Error != nil {
			state.Pipeline.Step.Environment["CI_PIPELINE_STATUS"] = "failure"
			state.Pipeline.Step.Environment["CI_STEP_STATUS"] = "failure"

			// DEPRECATED
			state.Pipeline.Step.Environment["CI_BUILD_STATUS"] = "failure"
			state.Pipeline.Step.Environment["CI_JOB_STATUS"] = "failure"
		}

		return nil
	}
}
