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

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	backend "go.woodpecker-ci.org/woodpecker/pipeline/backend/types"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/pipeline/multipart"
)

// TODO: move runtime into "runtime" subpackage

type (
	// State defines the pipeline and process state.
	State struct {
		// Global state of the pipeline.
		Pipeline struct {
			// Pipeline time started
			Time int64 `json:"time"`
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
	engine  backend.Engine
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
	r.taskUUID = uuid.New().String()
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

// Starts the execution of an workflow and waits for it to complete
func (r *Runtime) Run(runnerCtx context.Context) error {
	logger := r.MakeLogger()
	logger.Debug().Msgf("Executing %d stages, in order of:", len(r.spec.Stages))
	for _, stage := range r.spec.Stages {
		steps := []string{}
		for _, step := range stage.Steps {
			steps = append(steps, step.Name)
		}

		logger.Debug().
			Str("Stage", stage.Name).
			Str("Steps", strings.Join(steps, ",")).
			Msg("stage")
	}

	defer func() {
		if err := r.engine.DestroyWorkflow(runnerCtx, r.spec, r.taskUUID); err != nil {
			logger.Error().Err(err).Msg("could not destroy engine")
		}
	}()

	r.started = time.Now().Unix()
	if err := r.engine.SetupWorkflow(r.ctx, r.spec, r.taskUUID); err != nil {
		return err
	}

	for _, stage := range r.spec.Stages {
		select {
		case <-r.ctx.Done():
			return pipeline_errors.ErrCancel
		case err := <-r.execAll(stage.Steps):
			if err != nil {
				r.err = err
			}
		}
	}

	return r.err
}

// Updates the current status of a step
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
	state.Pipeline.Time = r.started
	state.Pipeline.Step = step
	state.Process = processState // empty
	state.Pipeline.Error = r.err

	if traceErr := r.tracer.Trace(state); traceErr != nil {
		return traceErr
	}
	return err
}

// Executes a set of parallel steps
func (r *Runtime) execAll(steps []*backend.Step) <-chan error {
	var g errgroup.Group
	done := make(chan error)
	logger := r.MakeLogger()

	for _, step := range steps {
		// required since otherwise the loop variable
		// will be captured by the function. This will
		// recreate the step "variable"
		step := step
		g.Go(func() error {
			// Case the pipeline was already complete.
			logger.Debug().
				Str("Step", step.Name).
				Msg("Prepare")

			switch {
			case r.err != nil && !step.OnFailure:
				logger.Debug().
					Str("Step", step.Name).
					Err(r.err).
					Msgf("Skipped due to OnFailure=%t", step.OnFailure)
				return nil
			case r.err == nil && !step.OnSuccess:
				logger.Debug().
					Str("Step", step.Name).
					Msgf("Skipped due to OnSuccess=%t", step.OnSuccess)
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
				Str("Step", step.Name).
				Msg("Executing")

			processState, err := r.exec(step)

			logger.Debug().
				Str("Step", step.Name).
				Msg("Complete")

			// if we got a nil process but an error state
			// then we need to log the internal error to the step.
			if r.logger != nil && err != nil && !errors.Is(err, pipeline_errors.ErrCancel) && processState == nil {
				_ = r.logger.Log(step, multipart.New(strings.NewReader(
					"Backend engine error while running step: "+err.Error(),
				)))
			}

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

			if err := r.logger.Log(step, multipart.New(rc)); err != nil {
				logger.Error().Err(err).Msg("process logging failed")
			}
			_ = rc.Close()
		}()
	}

	// nothing else to do, this is a detached process.
	if step.Detached {
		return nil, nil
	}

	// Some pipeline backends, such as local, will close the pipe from Tail on Wait,
	// so first make sure all reading has finished.
	wg.Wait()
	waitState, err := r.engine.WaitStep(r.ctx, step, r.taskUUID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return waitState, pipeline_errors.ErrCancel
		}
		return nil, err
	}

	if err := r.engine.DestroyStep(r.ctx, step, r.taskUUID); err != nil {
		return nil, err
	}

	if waitState.OOMKilled {
		return waitState, &pipeline_errors.OomError{
			Name: step.Name,
			Code: waitState.ExitCode,
		}
	} else if waitState.ExitCode != 0 {
		return waitState, &pipeline_errors.ExitError{
			Name: step.Name,
			Code: waitState.ExitCode,
		}
	}

	return waitState, nil
}
