//go:build test

// Copyright 2026 Woodpecker Authors
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
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
	tracer_mocks "go.woodpecker-ci.org/woodpecker/v3/pipeline/tracing/mocks"
)

// getTracerStates extracts all state.State values passed to Trace() calls
// on a mockery-generated MockTracer. Thread-safe because mock.Mock.Calls
// is append-only and we only read after the workflow completes.
func getTracerStates(tracer *tracer_mocks.MockTracer) []state.State {
	var states []state.State
	for _, call := range tracer.Calls {
		if call.Method == "Trace" {
			s := call.Arguments.Get(0).(*state.State)
			states = append(states, *s)
		}
	}
	return states
}
