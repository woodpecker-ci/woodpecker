// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"errors"
	"fmt"
	"io"
	"strings"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
)

// signalOutcome tells completeStep how to report a step that sent a signal.
type signalOutcome int

const (
	// The step is reported as successful (exit code 0).
	outcomeSuccess signalOutcome = iota
	// The caller has to run the step again.
	outcomeRestart
)

// errStepRestart is returned by completeStep when the step sent the restart signal.
var errStepRestart = errors.New("step requested restart")

// signalHandler performs the action of a signal sent by a step.
type signalHandler func(r *Runtime, step *backend_types.Step) signalOutcome

// signalHandlers maps every signal a step can send (see backend_types.StepSignal)
// to its action. New signals only need a new entry here.
var signalHandlers = map[backend_types.StepSignal]signalHandler{
	// skip: all steps of the following stages of this workflow are skipped.
	// It takes effect once the current stage has finished (see Run).
	backend_types.StepSignalSkip: func(r *Runtime, _ *backend_types.Step) signalOutcome {
		r.skipRequested.Store(true)
		return outcomeSuccess
	},
	// success: the step counts as successful, the workflow continues as usual.
	backend_types.StepSignalSuccess: func(_ *Runtime, _ *backend_types.Step) signalOutcome {
		return outcomeSuccess
	},
	// cancel: the workflow is canceled, running steps are stopped.
	backend_types.StepSignalCancel: func(r *Runtime, _ *backend_types.Step) signalOutcome {
		r.cancel(pipeline_errors.ErrCancel)
		return outcomeSuccess
	},
	// restart: the step runs again (see runBlockingStep).
	backend_types.StepSignalRestart: func(_ *Runtime, _ *backend_types.Step) signalOutcome {
		return outcomeRestart
	},
}

// logSignal writes a marker line about a signal into the step log, e.g. to
// separate restart attempts.
func (r *Runtime) logSignal(step *backend_types.Step, msg string) {
	if r.logger == nil {
		return
	}
	line := fmt.Sprintf("--- woodpecker: %s ---\n", msg)
	if err := r.logger(step, io.NopCloser(strings.NewReader(line))); err != nil {
		logger := r.makeLogger()
		logger.Error().Err(err).Str("step", step.Name).Msg("could not write signal marker to step log")
	}
}
