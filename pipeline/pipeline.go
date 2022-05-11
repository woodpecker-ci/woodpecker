package pipeline

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
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
}

// New returns a new runtime using the specified runtime
// configuration and runtime engine.
func New(spec *backend.Config, opts ...Option) *Runtime {
	r := new(Runtime)
	r.spec = spec
	r.ctx = context.Background()
	for _, opts := range opts {
		opts(r)
	}
	return r
}

// Starts the execution of the pipeline and waits for it to complete
func (r *Runtime) Run() error {
	defer func() {
		if err := r.engine.Destroy(r.ctx, r.spec); err != nil {
			log.Error().Err(err).Msg("could not destroy pipeline")
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

	return r.tracer.Trace(state)
}

// Executes a set of parallel steps
func (r *Runtime) execAll(steps []*backend.Step) <-chan error {
	var g errgroup.Group
	done := make(chan error)

	for _, step := range steps {
		// required since otherwise the loop variable
		// will be captured by the function. This will
		// recreate the step "variable"
		step := step
		g.Go(func() error {
			// Case the pipeline was already complete.
			switch {
			case r.err != nil && !step.OnFailure:
				return nil
			case r.err == nil && !step.OnSuccess:
				return nil
			}

			// Trace started.
			err := r.traceStep(nil, nil, step)
			if err != nil {
				return err
			}

			processState, err := r.exec(step)

			// Return the error after tracing it.
			traceErr := r.traceStep(processState, err, step)
			if traceErr != nil {
				return traceErr
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
	// TODO: using DRONE_ will be deprecated with 0.15.0. remove fallback with following release
	for key, value := range step.Environment {
		if strings.HasPrefix(key, "CI_") {
			step.Environment[strings.Replace(key, "CI_", "DRONE_", 1)] = value
		}
	}

	if err := r.engine.Exec(r.ctx, step); err != nil {
		return nil, err
	}

	if r.logger != nil {
		rc, err := r.engine.Tail(r.ctx, step)
		if err != nil {
			return nil, err
		}

		go func() {
			if err := r.logger.Log(step, multipart.New(rc)); err != nil {
				log.Error().Err(err).Msg("process logging failed")
			}
			_ = rc.Close()
		}()
	}

	// nothing else to do, this is a detached process.
	if step.Detached {
		return nil, nil
	}

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
