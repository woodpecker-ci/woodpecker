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
	"log/slog"
	"runtime"
	"strconv"
	"time"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/shared/logger/errorattr"
)

func (r *Runner) createTracer(ctxmeta context.Context, logger *slog.Logger, workflow *rpc.Workflow) pipeline.TraceFunc {
	return func(state *pipeline.State) error {
		steplogger := logger.With(
			slog.String("image", state.Pipeline.Step.Image),
			slog.String("workflowID", workflow.ID),
			errorattr.Default(state.Process.Error),
			slog.Int("exit_code", state.Process.ExitCode),
			slog.Bool("exited", state.Process.Exited))

		stepState := rpc.State{
			StepUUID: state.Pipeline.Step.UUID,
			Exited:   state.Process.Exited,
			ExitCode: state.Process.ExitCode,
			Started:  time.Now().Unix(), // TODO do not do this
			Finished: time.Now().Unix(),
		}
		if state.Process.Error != nil {
			stepState.Error = state.Process.Error.Error()
		}

		defer func() {
			steplogger.Debug("update step status")

			if uerr := r.client.Update(ctxmeta, workflow.ID, stepState); uerr != nil {
				steplogger.Debug("update step status error", errorattr.Default(uerr))
			}

			steplogger.Debug("update step status complete")
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

		state.Pipeline.Step.Environment["CI_SYSTEM_PLATFORM"] = runtime.GOOS + "/" + runtime.GOARCH

		if state.Pipeline.Error != nil {
			state.Pipeline.Step.Environment["CI_PIPELINE_STATUS"] = "failure"
			state.Pipeline.Step.Environment["CI_STEP_STATUS"] = "failure"
		}

		return nil
	}
}
