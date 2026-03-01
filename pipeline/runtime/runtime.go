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
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/logging"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/tracing"
)

// Runtime represents a workflow state executed by a specific backend.
// Each workflow gets its own Runtime instance.
type Runtime struct {
	// err holds the first error that occurred in the workflow.
	// Always use getErr/setErr to access it — it is read and written from concurrent goroutines.
	errMu sync.RWMutex
	err   error

	spec    *backend.Config
	engine  backend.Backend
	started int64

	// ctx is the context for the current workflow execution.
	// All normal (non-cleanup) step operations must use this context.
	// Cleanup operations should use the runnerCtx passed to Run().
	ctx context.Context

	tracer tracing.Tracer
	logger logging.Logger

	taskUUID    string
	description map[string]string
}

// New returns a new Runtime for the given workflow spec and options.
func New(spec *backend.Config, opts ...Option) *Runtime {
	r := new(Runtime)
	r.description = map[string]string{}
	r.spec = spec
	r.ctx = context.Background()
	r.taskUUID = ulid.Make().String()
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// MakeLogger returns a logger enriched with all runtime description fields.
func (r *Runtime) MakeLogger() zerolog.Logger {
	logCtx := log.With()
	for key, val := range r.description {
		logCtx = logCtx.Str(key, val)
	}
	return logCtx.Logger()
}

func (r *Runtime) getErr() error {
	r.errMu.RLock()
	defer r.errMu.RUnlock()
	return r.err
}

func (r *Runtime) setErr(err error) {
	r.errMu.Lock()
	defer r.errMu.Unlock()
	r.err = err
}
