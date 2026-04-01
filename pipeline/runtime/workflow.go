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
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
)

// Run starts the workflow, executes all stages sequentially, and tears down the
// workflow on exit. The runnerCtx must outlive workflow cancellation so that cleanup
// can still reach the backend (e.g. stopping Docker containers).
func (r *Runtime) Run(runnerCtx context.Context) error {
	if err := r.validateConfig(); err != nil {
		return err
	}

	logger := r.makeLogger()
	r.logStages()

	// we make sure cleanup always happens
	defer func() {
		// Skip destroying workflow if recovery is enabled and context was canceled but NOT by user.
		if r.recoveryManager.IsRecoverable(runnerCtx) {
			logger.Info().Msg("skipping workflow destruction, preserving for recovery")
			return
		}

		ctx := runnerCtx //nolint:contextcheck
		if ctx.Err() != nil {
			// runnerCtx itself is done — fall back to a short-lived shutdown context.
			ctx = GetShutdownCtx()
		}
		if err := r.engine.DestroyWorkflow(ctx, r.spec, r.taskUUID); err != nil {
			logger.Error().Err(err).Msg("could not destroy workflow")
		}
	}()

	r.started = time.Now().Unix()

	if err := r.engine.SetupWorkflow(r.ctx, r.spec, r.taskUUID); err != nil { //nolint:contextcheck
		r.traceWorkflowSetupError(err)
		return err
	}

	for _, stage := range r.spec.Stages {
		select {
		case <-r.ctx.Done():
			return pipeline_errors.ErrCancel
		case err := <-r.runStage(runnerCtx, stage.Steps):
			if err != nil {
				r.err.Set(err)
			}
		}
	}

	return r.err.Get()
}

// The validateConfig checks if a dev made a mistake,
// this should be values a user has no control over.
func (r *Runtime) validateConfig() error {
	if r.tracer == nil {
		return fmt.Errorf("runtime misconfiguration: tracer must not be nil")
	}
	if r.logger == nil {
		return fmt.Errorf("runtime misconfiguration: logger must not be nil")
	}
	if r.spec == nil {
		return fmt.Errorf("runtime misconfiguration: backend configuration is missing")
	}
	return nil
}

// logStages logs the ordered list of stages and their steps at debug level.
func (r *Runtime) logStages() {
	logger := r.makeLogger()
	logger.Debug().Msgf("executing %d stages, in order of:", len(r.spec.Stages))
	for stagePos, stage := range r.spec.Stages {
		stepNames := make([]string, 0, len(stage.Steps))
		for _, step := range stage.Steps {
			stepNames = append(stepNames, step.Name)
		}
		logger.Debug().
			Int("StagePos", stagePos).
			Str("Steps", strings.Join(stepNames, ",")).
			Msg("stage")
	}
}

// traceWorkflowSetupError traces an ErrInvalidWorkflowSetup to the tracer.
func (r *Runtime) traceWorkflowSetupError(err error) {
	var stepErr *pipeline_errors.ErrInvalidWorkflowSetup
	if !errors.As(err, &stepErr) {
		return
	}

	s := new(state.State)
	s.CurrStep = stepErr.Step
	s.Workflow.Error = stepErr.Err
	s.CurrStepState = backend_types.State{
		Error:    stepErr.Err,
		Exited:   true,
		ExitCode: 1,
	}

	if traceErr := r.tracer.Trace(s); traceErr != nil {
		logger := r.makeLogger()
		logger.Error().Err(traceErr).Msg("failed to trace workflow setup error")
	}
}

// runStage executes all steps of a stage in parallel.
// It returns a channel that emits the combined error (if any) once all steps finish.
func (r *Runtime) runStage(runnerCtx context.Context, steps []*backend_types.Step) <-chan error {
	var g errgroup.Group
	done := make(chan error)

	for _, step := range steps {
		g.Go(func() error {
			return r.executeStep(runnerCtx, step)
		})
	}

	go func() {
		done <- g.Wait()
		close(done)
	}()

	return done
}
