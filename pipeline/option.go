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

func WithTaskUUID(uuid string) Option {
	return func(r *Runtime) {
		r.taskUUID = uuid
	}
}
