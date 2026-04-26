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

//go:build test

package runtime

import (
	"io"
	"testing"

	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/logging"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
	tracer_mocks "go.woodpecker-ci.org/woodpecker/v3/pipeline/tracing/mocks"
)

// newTestTracer creates a MockTracer that accepts any number of Trace calls.
func newTestTracer(t *testing.T) *tracer_mocks.MockTracer {
	t.Helper()
	tracer := tracer_mocks.NewMockTracer(t)
	tracer.On("Trace", mock.Anything).Return(nil).Maybe()
	return tracer
}

// newTestLogger creates a noop logger.
func newTestLogger(t *testing.T) logging.Logger {
	return func(_ *types.Step, rc io.ReadCloser) error {
		_, _ = io.Copy(io.Discard, rc)
		return rc.Close()
	}
}

// getTracerStates extracts all state.State values passed to Trace() calls
// on a mockery-generated MockTracer. Thread-safe because mock.Mock.Calls
// is append-only and we only read after the workflow completes.
func getTracerStates(tracer *tracer_mocks.MockTracer) []state.State {
	var states []state.State
	for _, call := range tracer.Calls {
		if call.Method == "Trace" {
			s, _ := call.Arguments.Get(0).(*state.State)
			states = append(states, *s)
		}
	}
	return states
}

// indexOfTrace returns the first index where predicate matches, or -1.
func indexOfTrace(traces []state.State, match func(s state.State) bool) int {
	for i := range traces {
		if match(traces[i]) {
			return i
		}
	}
	return -1
}
