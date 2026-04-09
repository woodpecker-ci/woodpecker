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

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
)

// executeStep is the single entry point called per step from runStage.
// It checks whether the step should be skipped, emits a "started" trace,
// sets up drone-compat env vars, then hands off to blocking or detached execution.
func (r *Runtime) executeStep(runnerCtx context.Context, step *backend_types.Step) error {
	logger := r.makeLogger()
	logger.Debug().Str("step", step.Name).Msg("prepare")

	if r.shouldSkipStep(step) {
		// Trace the skip so the server marks the step as skipped immediately,
		// rather than leaving it in "pending" until workflow Done.
		return r.traceStep(&backend_types.State{Skipped: true}, nil, step)
	}

	// Check recovery state if recovery is enabled
	if r.recoveryManager.Enabled() {
		shouldSkip, recoveryState := r.recoveryManager.ShouldSkipStep(step)
		if shouldSkip {
			logger.Info().
				Str("step", step.Name).
				Int("status", int(recoveryState.Status)).
				Int("exit_code", recoveryState.ExitCode).
				Msg("skipping step due to recovery state")

			processState := &backend_types.State{
				Exited:   true,
				ExitCode: recoveryState.ExitCode,
			}
			if traceErr := r.traceStep(processState, nil, step); traceErr != nil {
				return traceErr
			}
			if recoveryState.ExitCode != 0 {
				return &pipeline_errors.ExitError{
					UUID: step.UUID,
					Code: recoveryState.ExitCode,
				}
			}
			return nil
		} else if r.recoveryManager.ShouldReconnect(recoveryState) {
			reconnectErr := r.engine.Reconnect(r.ctx, step, r.taskUUID) //nolint:contextcheck
			if reconnectErr == nil {
				logger.Info().Str("step", step.Name).Msg("reconnecting to existing step")
				return r.execReconnected(step)
			}
			logger.Debug().Err(reconnectErr).Str("step", step.Name).Msg("cannot reconnect, re-executing step")
		}

		if err := r.recoveryManager.MarkStepRunning(r.ctx, step); err != nil { //nolint:contextcheck
			logger.Warn().Err(err).Str("step", step.Name).Msg("failed to mark step as running")
		}
	}

	// Emit a "step started" trace before doing any real work.
	if err := r.traceStep(nil, nil, step); err != nil {
		return err
	}

	// Add compatibility environment variables for drone-ci plugins.
	if step.Type == backend_types.StepTypePlugin {
		metadata.SetDroneEnviron(step.Environment)
	}

	logger.Debug().Str("step", step.Name).Msg("executing")

	if step.Detached {
		return r.runDetachedStep(runnerCtx, step)
	}
	return r.runBlockingStep(runnerCtx, step)
}

// shouldSkipStep returns true when the step should not run based on the current
// pipeline error state and the step's OnSuccess / OnFailure flags.
// It logs the reason for skipping before returning.
func (r *Runtime) shouldSkipStep(step *backend_types.Step) bool {
	logger := r.makeLogger()
	currentErr := r.err.Get()

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
func (r *Runtime) startStep(step *backend_types.Step) (func(), int64, error) {
	if err := r.engine.StartStep(r.ctx, step, r.taskUUID); err != nil {
		return nil, 0, err
	}
	startTime := time.Now().Unix()

	rc, err := r.engine.TailStep(r.ctx, step, r.taskUUID)
	if err != nil {
		return nil, 0, err
	}

	var wg sync.WaitGroup
	wg.Go(func() {
		logger := r.makeLogger()
		if err := r.logger(step, rc); err != nil {
			logger.Error().Err(err).Str("step", step.Name).Msg("step log streaming failed")
		}
		_ = rc.Close()
	})

	return wg.Wait, startTime, nil
}

// completeStep drains the log stream, waits for the process to exit, destroys
// the container, and maps exit conditions (OOM kill, non-zero exit code, context
// cancellation) to typed errors.
//
// The runnerCtx is intentionally used for DestroyStep so that container cleanup can
// still reach the backend even after the workflow context (r.ctx) is canceled.
func (r *Runtime) completeStep(runnerCtx context.Context, step *backend_types.Step, waitForLogs func(), startTime int64) (*backend_types.State, error) {
	// Drain the log stream before waiting on the process exit.
	waitForLogs()

	waitState, err := r.engine.WaitStep(r.ctx, step, r.taskUUID) //nolint:contextcheck
	if err != nil {
		if errors.Is(err, context.Canceled) {
			if waitState == nil {
				waitState = &backend_types.State{}
			}
			waitState.Error = pipeline_errors.ErrCancel
		} else {
			return nil, err
		}
	}

	// Use runnerCtx here: the workflow context may already be canceled but we
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
func (r *Runtime) runBlockingStep(runnerCtx context.Context, step *backend_types.Step) error {
	logger := r.makeLogger()

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

	recoverable := r.recoveryManager.IsRecoverable(r.ctx) //nolint:contextcheck
	r.updateStepRecoveryState(step, processState, err, recoverable)

	if !recoverable {
		err = r.traceStep(processState, err, step)
	}
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
func (r *Runtime) runDetachedStep(runnerCtx context.Context, step *backend_types.Step) error {
	waitForLogs, startTime, err := r.startStep(step)
	if err != nil {
		// Setup failed before the container was running — treat it like a
		// blocking failure so the pipeline is aware.
		return r.traceStep(nil, err, step)
	}

	// Container is up and logging is streaming — hand off to background.
	go func() {
		logger := r.makeLogger()

		processState, err := r.completeStep(runnerCtx, step, waitForLogs, startTime)
		logger.Debug().Str("step", step.Name).Msg("complete")

		if errors.Is(err, context.Canceled) {
			err = pipeline_errors.ErrCancel
		}
		if err != nil {
			logger.Error().Err(err).Str("step", step.Name).Msg("detached step failed after while running")
		}

		recoverable := r.recoveryManager.IsRecoverable(r.ctx) //nolint:contextcheck
		r.updateStepRecoveryState(step, processState, err, recoverable)

		if !recoverable {
			if traceErr := r.traceStep(processState, err, step); traceErr != nil {
				logger.Error().Err(traceErr).Str("step", step.Name).Msg("failed to trace detached step result")
			}
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
// Always returns err unchanged so callers can write: return r.traceStep(state, err, step).
func (r *Runtime) traceStep(processState *backend_types.State, err error, step *backend_types.Step) error {
	s := new(state.State)
	s.Workflow.Started = r.started
	s.CurrStep = step
	s.Workflow.Error = r.err.Get()

	switch {
	case processState == nil && err != nil:
		// Step failed to start — create an dummy exited process state.
		s.CurrStepState = backend_types.State{
			Error:     err,
			Exited:    true,
			OOMKilled: false,
		}
	case processState != nil:
		s.CurrStepState = *processState
		// processState == nil && err == nil: step just started, leave s.CurrStepState zero-valued.
	}

	if traceErr := r.tracer.Trace(s); traceErr != nil {
		return traceErr
	}
	return err
}

// updateStepRecoveryState updates the recovery state for a step based on its execution result.
func (r *Runtime) updateStepRecoveryState(step *backend_types.Step, processState *backend_types.State, err error, recoverable bool) {
	if !r.recoveryManager.Enabled() {
		return
	}
	logger := r.makeLogger()
	switch {
	case recoverable:
		logger.Debug().Str("step", step.Name).Msg("workflow is recoverable, not updating step state")
	case processState != nil && processState.ExitCode == 0 && err == nil:
		if markErr := r.recoveryManager.MarkStepSuccess(r.ctx, step); markErr != nil {
			logger.Warn().Err(markErr).Str("step", step.Name).Msg("failed to mark step as success")
		}
	default:
		exitCode := 1
		if processState != nil {
			exitCode = processState.ExitCode
		}
		if markErr := r.recoveryManager.MarkStepFailed(r.ctx, step, exitCode); markErr != nil {
			logger.Warn().Err(markErr).Str("step", step.Name).Msg("failed to mark step as failed")
		}
	}
}

// execReconnected handles a reconnected step (waiting for completion without re-executing).
func (r *Runtime) execReconnected(step *backend_types.Step) error {
	logger := r.makeLogger()

	var wg sync.WaitGroup
	rc, err := r.engine.TailStep(r.ctx, step, r.taskUUID)
	if err != nil {
		logger.Warn().Err(err).Str("step", step.Name).Msg("failed to retrieve logs for reconnected step, continuing without logs")
	} else {
		wg.Go(func() {
			if err := r.logger(step, rc); err != nil {
				logger.Error().Err(err).Str("step", step.Name).Msg("process logging failed")
			}
			_ = rc.Close()
		})
	}

	if step.Detached {
		return nil
	}

	wg.Wait()
	waitState, err := r.engine.WaitStep(r.ctx, step, r.taskUUID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return pipeline_errors.ErrCancel
		}
		return err
	}

	if waitState.ExitCode == 0 {
		if markErr := r.recoveryManager.MarkStepSuccess(r.ctx, step); markErr != nil {
			logger.Warn().Err(markErr).Str("step", step.Name).Msg("failed to mark step as success")
		}
	} else {
		if markErr := r.recoveryManager.MarkStepFailed(r.ctx, step, waitState.ExitCode); markErr != nil {
			logger.Warn().Err(markErr).Str("step", step.Name).Msg("failed to mark step as failed")
		}
	}

	if err := r.traceStep(waitState, nil, step); err != nil {
		return err
	}

	if waitState.OOMKilled {
		return &pipeline_errors.OomError{
			UUID: step.UUID,
			Code: waitState.ExitCode,
		}
	}
	if waitState.ExitCode != 0 {
		return &pipeline_errors.ExitError{
			UUID: step.UUID,
			Code: waitState.ExitCode,
		}
	}

	return nil
}
