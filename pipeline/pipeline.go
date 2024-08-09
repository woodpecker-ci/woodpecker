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

package pipeline

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	backend "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
)

// TODO: move runtime into "runtime" subpackage

type (
	// State defines the pipeline and process state.
	State struct {
		// Global state of the pipeline.
		Pipeline struct {
			// Pipeline time started
			Started int64 `json:"time"`
			// Current pipeline step
			Step *backend.Step `json:"step"`
			// Current pipeline error state
			Error error `json:"error"`
		}

		// Current process state.
		Process *backend.State
	}
)

// Runtime is a configuration runtime.
type Runtime struct {
	err     error
	spec    *backend.Config
	engine  backend.Backend
	started int64

	ctx    context.Context
	tracer Tracer
	logger Logger

	taskUUID string

	Description map[string]string // The runtime descriptors.
}

// New returns a new runtime using the specified runtime
// configuration and runtime engine.
func New(spec *backend.Config, opts ...Option) *Runtime {
	r := new(Runtime)
	r.Description = map[string]string{}
	r.spec = spec
	r.ctx = context.Background()
	r.taskUUID = ulid.Make().String()
	for _, opts := range opts {
		opts(r)
	}
	return r
}

func (r *Runtime) MakeLogger() zerolog.Logger {
	logCtx := log.With()
	for key, val := range r.Description {
		logCtx = logCtx.Str(key, val)
	}
	return logCtx.Logger()
}

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
		return err
	}

	for _, stage := range r.spec.Stages {
		select {
		case <-r.ctx.Done():
			return ErrCancel
		case err := <-r.execAll(stage.Steps):
			if err != nil {
				r.err = err
			}
		}
	}

	return r.err
}

// Updates the current status of a step.
func (r *Runtime) traceStep(processState *backend.State, err error, step *backend.Step) error {
	if r.tracer == nil {
		// no tracer nothing to trace :)
		return nil
	}

	if processState == nil {
		processState = new(backend.State)
		if err != nil {
			processState.Error = err
			processState.Exited = true
			processState.OOMKilled = false
			processState.ExitCode = 126 // command invoked cannot be executed.
		}
	}

	state := new(State)
	state.Pipeline.Started = r.started
	state.Pipeline.Step = step
	state.Process = processState // empty
	state.Pipeline.Error = r.err

	if traceErr := r.tracer.Trace(state); traceErr != nil {
		return traceErr
	}
	return err
}

// Executes a set of parallel steps.
func (r *Runtime) execAll(steps []*backend.Step) <-chan error {
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

			processState, err := r.exec(step)

			logger.Debug().
				Str("step", step.Name).
				Msg("complete")

			// Return the error after tracing it.
			err = r.traceStep(processState, err, step)
			if err != nil && step.Failure == metadata.FailureIgnore {
				return nil
			}
			return err
		})
	}

	go func() {
		done <- g.Wait()
		close(done)
	}()
	return done
}

// Executes the step and returns the state and error.
func (r *Runtime) exec(step *backend.Step) (*backend.State, error) {
	if err := r.engine.StartStep(r.ctx, step, r.taskUUID); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	if r.logger != nil {
		rc, err := r.engine.TailStep(r.ctx, step, r.taskUUID)
		if err != nil {
			return nil, err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			logger := r.MakeLogger()

			if err := r.logger(step, rc); err != nil {
				logger.Error().Err(err).Msg("process logging failed")
			}
			_ = rc.Close()
		}()
	}

	// nothing else to do, this is a detached process.
	if step.Detached {
		return nil, nil
	}

	// We wait until all data was logged. (Needed for some backends like local as WaitStep kills the log stream)
	wg.Wait()

	waitState, err := r.engine.WaitStep(r.ctx, step, r.taskUUID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return waitState, ErrCancel
		}
		return nil, err
	}

	if err := r.engine.DestroyStep(r.ctx, step, r.taskUUID); err != nil {
		return nil, err
	}

	if waitState.OOMKilled {
		return waitState, &OomError{
			UUID: step.UUID,
			Code: waitState.ExitCode,
		}
	} else if waitState.ExitCode != 0 {
		return waitState, &ExitError{
			UUID: step.UUID,
			Code: waitState.ExitCode,
		}
	}

	return waitState, nil
}
