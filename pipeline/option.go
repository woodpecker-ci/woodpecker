package pipeline

import (
	"context"

	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

// Option configures a runtime option.
type Option func(*Runtime)

// WithEngine returns an option configured with a runtime engine.
func WithEngine(engine backend.Engine) Option {
	return func(r *Runtime) {
		r.engine = engine
	}
}

// WithLogger returns an option configured with a runtime logger.
func WithLogger(logger Logger) Option {
	return func(r *Runtime) {
		r.logger = logger
	}
}

// WithTracer returns an option configured with a runtime tracer.
func WithTracer(tracer Tracer) Option {
	return func(r *Runtime) {
		r.tracer = tracer
	}
}

// WithContext returns an option configured with a context.
func WithContext(ctx context.Context) Option {
	return func(r *Runtime) {
		r.ctx = ctx
	}
}

func WithDescription(desc map[string]string) Option {
	return func(r *Runtime) {
		r.Description = desc
	}
}
