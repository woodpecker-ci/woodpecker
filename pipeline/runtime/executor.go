// Copyright 2023 Woodpecker Authors
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
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
)

// Run starts the execution of a workflow and waits for it to complete.
func (r *Runtime) Run(runnerCtx context.Context) error {
	logger := r.MakeLogger()
	logger.Debug().Msgf("executing %d stages, in order of:", len(r.spec.Stages))
	for stagePos, stage := range r.spec.Stages {
		stepNames := []string{}
		for _, step := range stage.Steps {
			stepNames = append(stepNames, step.Name)
		}

		logger.Debug().
			Int("StagePos", stagePos).
			Str("Steps", strings.Join(stepNames, ",")).
			Msg("stage")
	}

	defer func() {
		// Skip destroying workflow if recovery is enabled and context was canceled but NOT by user.
		if r.recoveryManager.IsRecoverable(runnerCtx) {
			logger.Info().Msg("skipping workflow destruction, preserving for recovery")
			return
		}

		ctx := runnerCtx //nolint:contextcheck
		if ctx.Err() != nil {
			ctx = GetShutdownCtx()
		}
		if err := r.engine.DestroyWorkflow(ctx, r.spec, r.taskUUID); err != nil {
			logger.Error().Err(err).Msg("could not destroy engine")
		}
	}()

	r.started = time.Now().Unix()
	if err := r.engine.SetupWorkflow(runnerCtx, r.spec, r.taskUUID); err != nil {
		var stepErr *pipeline_errors.ErrInvalidWorkflowSetup
		if errors.As(err, &stepErr) {
			state := new(state.State)
			state.Pipeline.Step = stepErr.Step
			state.Pipeline.Error = stepErr.Err
			state.Process = backend.State{
				Error:    stepErr.Err,
				Exited:   true,
				ExitCode: 1,
			}

			if r.tracer != nil {
				if err := r.tracer.Trace(state); err != nil {
					logger.Error().Err(err).Msg("failed to trace step error")
				}
			}
		}

		return err
	}

	for _, stage := range r.spec.Stages {
		select {
		case <-r.ctx.Done():
			return pipeline_errors.ErrCancel
		case err := <-r.execAll(runnerCtx, stage.Steps):
			if err != nil {
				r.err = err
			}
		}
	}

	return r.err
}

// Updates the current status of a step.
// If processState is nil, we assume the step did not start.
// If step did not started and err exists, it's a step start issue and step is done.
func (r *Runtime) traceStep(processState *backend.State, err error, step *backend.Step) error {
	if r.tracer == nil {
		// no tracer nothing to trace :)
		return nil
	}

	state := new(state.State)
	state.Pipeline.Started = r.started
	state.Pipeline.Step = step
	state.Pipeline.Error = r.err

	// We have an error while starting the step
	if processState == nil && err != nil {
		state.Process = backend.State{
			Error:     err,
			Exited:    true,
			OOMKilled: false,
		}
	} else if processState != nil {
		state.Process = *processState
	}

	if traceErr := r.tracer.Trace(state); traceErr != nil {
		return traceErr
	}
	return err
}

// Executes a set of parallel steps.
func (r *Runtime) execAll(runnerCtx context.Context, steps []*backend.Step) <-chan error {
	var g errgroup.Group
	done := make(chan error)
	logger := r.MakeLogger()

	for _, step := range steps {
		// Required since otherwise the loop variable
		// will be captured by the function. This will
		// recreate the step "variable"
		step := step
		g.Go(func() error {
			// Case the pipeline was already complete.
			logger.Debug().
				Str("step", step.Name).
				Msg("prepare")

			switch {
			case r.err != nil && !step.OnFailure:
				logger.Debug().
					Str("step", step.Name).
					Err(r.err).
					Msgf("skipped due to OnFailure=%t", step.OnFailure)
				return nil
			case r.err == nil && !step.OnSuccess:
				logger.Debug().
					Str("step", step.Name).
					Msgf("skipped due to OnSuccess=%t", step.OnSuccess)
				return nil
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

					// Trace the recovered state
					processState := &backend.State{
						Exited:   true,
						ExitCode: recoveryState.ExitCode,
					}
					if traceErr := r.traceStep(processState, nil, step); traceErr != nil {
						return traceErr
					}

					// Return error if step was failed
					if recoveryState.ExitCode != 0 {
						return &pipeline_errors.ExitError{
							UUID: step.UUID,
							Code: recoveryState.ExitCode,
						}
					}
					return nil
				} else if r.recoveryManager.ShouldReconnect(recoveryState) {
					// Attempt to reconnect to a running step
					reconnectErr := r.engine.Reconnect(r.ctx, step, r.taskUUID) //nolint:contextcheck
					if reconnectErr == nil {
						logger.Info().Str("step", step.Name).Msg("reconnecting to existing step")
						return r.execReconnected(step)
					}
					logger.Debug().Err(reconnectErr).Str("step", step.Name).Msg("cannot reconnect, re-executing step")
				}

				// Mark step as running in recovery state
				if err := r.recoveryManager.MarkStepRunning(r.ctx, step); err != nil { //nolint:contextcheck
					logger.Warn().Err(err).Str("step", step.Name).Msg("failed to mark step as running")
				}
			}

			// Trace started.
			err := r.traceStep(nil, nil, step)
			if err != nil {
				return err
			}

			// add compatibility for drone-ci plugins
			metadata.SetDroneEnviron(step.Environment)

			logger.Debug().
				Str("step", step.Name).
				Msg("executing")

			// setup exec func in a way it can be detached if needed
			// wg will signal once
			execAndTrace := func(wg *sync.WaitGroup) error {
				processState, err := r.exec(runnerCtx, step, wg)

				logger.Debug().
					Str("step", step.Name).
					Msg("complete")

				// normalize context cancel error
				if errors.Is(err, context.Canceled) {
					err = pipeline_errors.ErrCancel
				}

				// Check if workflow is recoverable
				recoverable := r.recoveryManager.IsRecoverable(r.ctx) //nolint:contextcheck

				// Update recovery state based on step result
				if r.recoveryManager.Enabled() {
					switch {
					case recoverable:
						logger.Debug().Str("step", step.Name).Msg("workflow is recoverable, not updating step state")
					case processState != nil && processState.ExitCode == 0 && err == nil:
						if markErr := r.recoveryManager.MarkStepSuccess(r.ctx, step); markErr != nil { //nolint:contextcheck
							logger.Warn().Err(markErr).Str("step", step.Name).Msg("failed to mark step as success")
						}
					default:
						exitCode := 1
						if processState != nil {
							exitCode = processState.ExitCode
						}
						if markErr := r.recoveryManager.MarkStepFailed(r.ctx, step, exitCode); markErr != nil { //nolint:contextcheck
							logger.Warn().Err(markErr).Str("step", step.Name).Msg("failed to mark step as failed")
						}
					}
				}

				// Skip tracing if workflow is recoverable
				if !recoverable {
					err = r.traceStep(processState, err, step)
				}
				if err != nil && step.Failure == metadata.FailureIgnore {
					return nil
				}
				return err
			}

			// Report all errors until the setup happened.
			// Afterwards errors will be dropped.
			if step.Detached {
				var wg sync.WaitGroup
				wg.Add(1)
				var setupErr error
				go func() {
					setupErr = execAndTrace(&wg)
				}()
				wg.Wait()
				return setupErr
			}

			// run blocking
			return execAndTrace(nil)
		})
	}

	go func() {
		done <- g.Wait()
		close(done)
	}()

	return done
}

// execReconnected handles a reconnected step (waiting for completion without re-executing).
func (r *Runtime) execReconnected(step *backend.Step) error {
	logger := r.MakeLogger()

	var wg sync.WaitGroup
	if r.logger != nil {
		rc, err := r.engine.TailStep(r.ctx, step, r.taskUUID)
		if err != nil {
			logger.Warn().Err(err).Str("step", step.Name).Msg("failed to retrieve logs for reconnected step, continuing without logs")
		} else {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := r.logger(step, rc); err != nil {
					logger.Error().Err(err).Msg("process logging failed")
				}
				_ = rc.Close()
			}()
		}
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

	// Trace the result
	if err := r.traceStep(waitState, nil, step); err != nil {
		return err
	}

	return exitError(step, waitState)
}

// exitError returns an OomError or ExitError based on the wait state, or nil if the step succeeded.
func exitError(step *backend.Step, waitState *backend.State) error {
	if waitState.OOMKilled {
		return &pipeline_errors.OomError{
			UUID: step.UUID,
			Code: waitState.ExitCode,
		}
	} else if waitState.ExitCode != 0 {
		return &pipeline_errors.ExitError{
			UUID: step.UUID,
			Code: waitState.ExitCode,
		}
	}

	return nil
}

// Executes the step and returns the state and error.
func (r *Runtime) exec(runnerCtx context.Context, step *backend.Step, setupWg *sync.WaitGroup) (*backend.State, error) {
	defer func() {
		if setupWg != nil {
			setupWg.Done()
		}
	}()

	if err := r.engine.StartStep(r.ctx, step, r.taskUUID); err != nil { //nolint:contextcheck
		return nil, err
	}
	startTime := time.Now().Unix()
	logger := r.MakeLogger()

	var wg sync.WaitGroup
	if r.logger != nil {
		rc, err := r.engine.TailStep(r.ctx, step, r.taskUUID) //nolint:contextcheck
		if err != nil {
			return nil, err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := r.logger(step, rc); err != nil {
				logger.Error().Err(err).Msg("process logging failed")
			}
			_ = rc.Close()
		}()
	}

	// nothing else to block for detached process.
	if setupWg != nil {
		setupWg.Done()
		// set to nil so the setupWg.Done in defer does not call it a second time
		setupWg = nil
	}

	// We wait until all data was logged. (Needed for some backends like local as WaitStep kills the log stream)
	wg.Wait()

	waitState, err := r.engine.WaitStep(r.ctx, step, r.taskUUID) //nolint:contextcheck
	if err != nil {
		if errors.Is(err, context.Canceled) {
			waitState.Error = pipeline_errors.ErrCancel
		} else {
			return nil, err
		}
	}

	// It is important to use the runnerCtx here because
	// in case the workflow was canceled we still have the docker daemon to stop the container.
	if err := r.engine.DestroyStep(runnerCtx, step, r.taskUUID); err != nil {
		return nil, err
	}

	// we update with our start time here
	waitState.Started = startTime

	// we handle cancel case
	if ctxErr := r.ctx.Err(); ctxErr != nil && errors.Is(ctxErr, context.Canceled) {
		waitState.Error = pipeline_errors.ErrCancel
	}

	return waitState, exitError(step, waitState)
}
