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
	"errors"
	"time"

	"github.com/rs/zerolog"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/tracing"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
)

func (r *Runner) createTracer(ctxMeta context.Context, logger zerolog.Logger, workflow *rpc.Workflow) tracing.TraceFunc {
	return func(state *state.State) error {
		stepLogger := logger.With().
			Str("image", state.CurrStep.Image).
			Str("workflow_id", workflow.ID).
			Err(state.CurrStepState.Error).
			Int("exit_code", state.CurrStepState.ExitCode).
			Bool("exited", state.CurrStepState.Exited).
			Logger()

		stepState := rpc.StepState{
			StepUUID: state.CurrStep.UUID,
			Exited:   state.CurrStepState.Exited,
			ExitCode: state.CurrStepState.ExitCode,
			Started:  state.CurrStepState.Started,
			Canceled: errors.Is(state.CurrStepState.Error, pipeline_errors.ErrCancel),
			Skipped:  state.CurrStepState.Skipped,
		}
		if state.CurrStepState.Error != nil {
			stepState.Error = state.CurrStepState.Error.Error()
		}
		if state.CurrStepState.Exited {
			stepState.Finished = time.Now().Unix()
		}

		stepLogger.Debug().Msg("update step status")

		if err := r.client.Update(ctxMeta, workflow.ID, stepState); err != nil {
			return err
		}

		stepLogger.Debug().Msg("update step status complete")
		return nil
	}
}
