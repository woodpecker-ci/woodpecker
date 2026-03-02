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
	"context"
	"errors"
	"sync"
	"time"

	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
)

// executeStep is the single entry point called per step from runStage.
// It checks whether the step should be skipped, emits a "started" trace,
// sets up drone-compat env vars, then hands off to blocking or detached execution.
func (r *Runtime) executeStep(runnerCtx context.Context, step *backend.Step) error {
	logger := r.MakeLogger()
	logger.Debug().Str("step", step.Name).Msg("prepare")

	if r.shouldSkipStep(step) {
		return nil
	}

	// Emit a "step started" trace before doing any real work.
	if err := r.traceStep(nil, nil, step); err != nil {
		return err
	}

	// Add compatibility environment variables for drone-ci plugins.
	metadata.SetDroneEnviron(step.Environment)

	logger.Debug().Str("step", step.Name).Msg("executing")

	if step.Detached {
		return r.runDetachedStep(runnerCtx, step)
	}
	return r.runBlockingStep(runnerCtx, step)
}

// shouldSkipStep returns true when the step should not run based on the current
// pipeline error state and the step's OnSuccess / OnFailure flags.
// It logs the reason for skipping before returning.
func (r *Runtime) shouldSkipStep(step *backend.Step) bool {
	logger := r.MakeLogger()
	currentErr := r.getErr()

	if currentErr != nil && !step.OnFailure {
		logger.Debug().
			Str("step", step.Name).
			Err(currentErr).
			Msgf("skipped due to OnFailure=%t", step.OnFailure)
		return true
	}

	if currentErr == nil && !step.OnSuccess {
		logger.Debug().
			Str("step", step.Name).
			Msgf("skipped due to OnSuccess=%t", step.OnSuccess)
		return true
	}

	return false
}

// startStep starts the step container and spawns a goroutine to stream its logs.
// It returns:
//   - waitForLogs: must be called before WaitStep — it blocks until the log stream
//     is fully drained. Some backends (e.g. local) close the log stream when
//     WaitStep is called, so draining first is required.
//   - startTime: unix timestamp recorded right after the container started, used
//     later to fill waitState.Started.
//
// If StartStep or TailStep fail, startStep returns a non-nil error and the caller
// must not call waitForLogs.
func (r *Runtime) startStep(step *backend.Step) (waitForLogs func(), startTime int64, err error) {
	if err := r.engine.StartStep(r.ctx, step, r.taskUUID); err != nil { //nolint:contextcheck
		return nil, 0, err
	}
	startTime = time.Now().Unix()

	var wg sync.WaitGroup

	if r.logger != nil {
		rc, err := r.engine.TailStep(r.ctx, step, r.taskUUID) //nolint:contextcheck
		if err != nil {
			return nil, 0, err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			logger := r.MakeLogger()
			if err := r.logger(step, rc); err != nil {
				logger.Error().Err(err).Str("step", step.Name).Msg("step log streaming failed")
			}
			_ = rc.Close()
		}()
	}

	return wg.Wait, startTime, nil
}

// completeStep drains the log stream, waits for the process to exit, destroys
// the container, and maps exit conditions (OOM kill, non-zero exit code, context
// cancellation) to typed errors.
//
// runnerCtx is intentionally used for DestroyStep so that container cleanup can
// still reach the backend even after the workflow context (r.ctx) is cancelled.
func (r *Runtime) completeStep(runnerCtx context.Context, step *backend.Step, waitForLogs func(), startTime int64) (*backend.State, error) {
	// Drain the log stream before waiting on the process exit.
	waitForLogs()

	waitState, err := r.engine.WaitStep(r.ctx, step, r.taskUUID) //nolint:contextcheck
	if err != nil {
		if errors.Is(err, context.Canceled) {
			waitState.Error = pipeline_errors.ErrCancel
		} else {
			return nil, err
		}
	}

	// Use runnerCtx here: the workflow context may already be cancelled but we
	// still need to reach the backend to stop/remove the container.
	if err := r.engine.DestroyStep(runnerCtx, step, r.taskUUID); err != nil {
		return nil, err
	}

	waitState.Started = startTime

	// Re-check context cancellation: the wait may have raced with cancellation.
	if ctxErr := r.ctx.Err(); ctxErr != nil && errors.Is(ctxErr, context.Canceled) {
		waitState.Error = pipeline_errors.ErrCancel
	}

	if waitState.OOMKilled {
		return waitState, &pipeline_errors.OomError{
			UUID: step.UUID,
			Code: waitState.ExitCode,
		}
	}
	if waitState.ExitCode != 0 {
		return waitState, &pipeline_errors.ExitError{
			UUID: step.UUID,
			Code: waitState.ExitCode,
		}
	}

	return waitState, nil
}

// runBlockingStep starts the step and blocks until it fully completes.
// The error is traced and returned to runStage, which feeds it into the
// stage error group.
func (r *Runtime) runBlockingStep(runnerCtx context.Context, step *backend.Step) error {
	logger := r.MakeLogger()

	waitForLogs, startTime, err := r.startStep(step)
	if err != nil {
		// The step never ran — trace the start failure and surface it.
		return r.traceStep(nil, err, step)
	}

	processState, err := r.completeStep(runnerCtx, step, waitForLogs, startTime)
	logger.Debug().Str("step", step.Name).Msg("complete")

	if errors.Is(err, context.Canceled) {
		err = pipeline_errors.ErrCancel
	}

	err = r.traceStep(processState, err, step)
	if err != nil && step.Failure == metadata.FailureIgnore {
		return nil
	}
	return err
}

// runDetachedStep starts the step and returns as soon as the container is running
// and log streaming is set up. The rest of the step lifecycle runs in the background.
//
// Any error that occurs after setup is logged but not propagated — it cannot
// influence the pipeline outcome at that point.
func (r *Runtime) runDetachedStep(runnerCtx context.Context, step *backend.Step) error {
	waitForLogs, startTime, err := r.startStep(step)
	if err != nil {
		// Setup failed before the container was running — treat it like a
		// blocking failure so the pipeline is aware.
		return r.traceStep(nil, err, step)
	}

	// Container is up and logging is streaming — hand off to background.
	go func() {
		logger := r.MakeLogger()

		processState, err := r.completeStep(runnerCtx, step, waitForLogs, startTime)
		logger.Debug().Str("step", step.Name).Msg("complete")

		if errors.Is(err, context.Canceled) {
			err = pipeline_errors.ErrCancel
		}
		if err != nil {
			logger.Error().Err(err).Str("step", step.Name).Msg("detached step failed after setup")
		}

		if traceErr := r.traceStep(processState, err, step); traceErr != nil {
			logger.Error().Err(traceErr).Str("step", step.Name).Msg("failed to trace detached step result")
		}
	}()

	return nil
}

// traceStep reports the current state of a step to the tracer.
//
//   - processState == nil, err == nil  →  step is being marked as started
//   - processState == nil, err != nil  →  step failed to start
//   - processState != nil              →  step has finished (err may or may not be set)
//
// Always returns err unchanged so callers can write: return r.traceStep(state, err, step)
func (r *Runtime) traceStep(processState *backend.State, err error, step *backend.Step) error {
	if r.tracer == nil {
		return err
	}

	s := new(state.State)
	s.Pipeline.Started = r.started
	s.Pipeline.Step = step
	s.Pipeline.Error = r.getErr()

	switch {
	case processState == nil && err != nil:
		// Step failed to start — synthesise an exited process state.
		s.Process = backend.State{
			Error:     err,
			Exited:    true,
			OOMKilled: false,
		}
	case processState != nil:
		s.Process = *processState
		// processState == nil && err == nil: step just started, leave s.Process zero-valued.
	}

	if traceErr := r.tracer.Trace(s); traceErr != nil {
		return traceErr
	}
	return err
}
