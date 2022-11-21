package pipeline

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend"
	"github.com/woodpecker-ci/woodpecker/pipeline/multipart"
)

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

	Description map[string]string // The runtime descriptors.
}

// New returns a new runtime using the specified runtime
// configuration and runtime engine.
func New(spec *backend.Config, opts ...Option) *Runtime {
	r := new(Runtime)
	r.Description = map[string]string{}
	r.spec = spec
	r.ctx = context.Background()
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

// Starts the execution of the pipeline and waits for it to complete
func (r *Runtime) Run() error {
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
		if err := r.engine.Destroy(r.ctx, r.spec); err != nil {
			logger.Error().Err(err).Msg("could not destroy engine")
		}
	}()

	r.started = time.Now().Unix()
	if err := r.engine.Setup(r.ctx, r.spec); err != nil {
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

			SetDroneEnviron(step.Environment)

			logger.Debug().
				Str("Step", step.Name).
				Msgf("environment after SetDroneEnviron(): %s", step.Environment)

			processState, err := r.exec(step)

			logger.Debug().
				Str("Step", step.Name).
				Msg("Complete")

			// if we got a nil process but an error state
			// then we need to log the internal error to the step.
			if r.logger != nil && err != nil && processState == nil {
				_ = r.logger.Log(step, multipart.New(strings.NewReader(
					"Backend engine error while running step: "+err.Error(),
				)))
			}

			// Return the error after tracing it.
			err = r.traceStep(processState, err, step)
			if err != nil && step.Failure == frontend.FailureIgnore {
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
	if err := r.engine.Exec(r.ctx, step); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	if r.logger != nil {
		rc, err := r.engine.Tail(r.ctx, step)
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
	waitState, err := r.engine.Wait(r.ctx, step)
	if err != nil {
		return nil, err
	}

	if waitState.OOMKilled {
		return waitState, &OomError{
			Name: step.Name,
			Code: waitState.ExitCode,
		}
	} else if waitState.ExitCode != 0 {
		return waitState, &ExitError{
			Name: step.Name,
			Code: waitState.ExitCode,
		}
	}

	return waitState, nil
}
