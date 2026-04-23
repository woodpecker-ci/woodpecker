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

package tracing

import (
	"strconv"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
)

// Tracer handles process tracing.
type Tracer interface {
	Trace(*state.State) error
}

// TraceFunc type is an adapter to allow the use of ordinary
// functions as a Tracer.
type TraceFunc func(*state.State) error

// Trace calls f(state).
func (f TraceFunc) Trace(state *state.State) error {
	return f(state)
}

// NoOpTracer provides a tracer that does nothing.
var NoOpTracer = TraceFunc(func(state *state.State) error {
	return nil
})
