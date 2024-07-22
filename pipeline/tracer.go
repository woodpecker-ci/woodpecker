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
	"strconv"
)

// Tracer handles process tracing.
type Tracer interface {
	Trace(*State) error
}

// TraceFunc type is an adapter to allow the use of ordinary
// functions as a Tracer.
type TraceFunc func(*State) error

// Trace calls f(state).
func (f TraceFunc) Trace(state *State) error {
	return f(state)
}

// DefaultTracer provides a tracer that updates the CI_ environment
// variables to include the correct timestamp and status.
// TODO: find either a new home or better name for this.
var DefaultTracer = TraceFunc(func(state *State) error {
	if state.Process.Exited {
		return nil
	}
	if state.Pipeline.Step.Environment == nil {
		return nil
	}
	state.Pipeline.Step.Environment["CI_PIPELINE_STARTED"] = strconv.FormatInt(state.Pipeline.Started, 10)

	state.Pipeline.Step.Environment["CI_STEP_STARTED"] = strconv.FormatInt(state.Pipeline.Started, 10)

	return nil
})
