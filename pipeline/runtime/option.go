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

	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/logging"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/tracing"
)

// Option configures a Runtime.
type Option func(*Runtime)

// WithBackend sets the backend engine used to run steps.
func WithBackend(backend backend.Backend) Option {
	return func(r *Runtime) {
		r.engine = backend
	}
}

// WithLogger sets the function used to stream step logs.
func WithLogger(logger logging.Logger) Option {
	return func(r *Runtime) {
		r.logger = logger
	}
}

// WithTracer sets the tracer used to report step state changes.
func WithTracer(tracer tracing.Tracer) Option {
	return func(r *Runtime) {
		r.tracer = tracer
	}
}

// WithContext sets the workflow execution context.
func WithContext(ctx context.Context) Option {
	return func(r *Runtime) {
		r.ctx = ctx
	}
}

// WithDescription sets the descriptive key-value pairs attached to every log line.
func WithDescription(desc map[string]string) Option {
	return func(r *Runtime) {
		r.description = desc
	}
}

// WithTaskUUID sets a specific task UUID instead of the auto-generated one.
func WithTaskUUID(uuid string) Option {
	return func(r *Runtime) {
		r.taskUUID = uuid
	}
}
