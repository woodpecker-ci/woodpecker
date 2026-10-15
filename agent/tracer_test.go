// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build test

package agent

import (
	"errors"
	"fmt"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	rpc_mocks "go.woodpecker-ci.org/woodpecker/v3/rpc/mocks"
)

// The tracer tells the server the well known kind of a step error, so the
// server can act on it without parsing the error text.
func TestTracerReportsErrorKind(t *testing.T) {
	for name, tc := range map[string]struct {
		err      error
		wantKind string
	}{
		"no error":     {err: nil, wantKind: ""},
		"plain error":  {err: errors.New("engine exploded"), wantKind: ""},
		"canceled":     {err: pipeline_errors.ErrCancel, wantKind: "canceled"},
		"wrapped kind": {err: fmt.Errorf("step: %w", backend_types.ErrStepCanceled), wantKind: "canceled"},
	} {
		t.Run(name, func(t *testing.T) {
			var got rpc.StepState
			peer := rpc_mocks.NewMockPeer(t)
			peer.On("Update", mock.Anything, mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					s, ok := args.Get(2).(rpc.StepState)
					require.True(t, ok)
					got = s
				}).
				Return(nil)

			runner := NewRunner(peer, rpc.Filter{}, "test-agent", 1, &State{Metadata: map[string]Info{}}, nil)
			trace := runner.createTracer(t.Context(), zerolog.Nop(), dummyWorkflow())

			s := &state.State{CurrStep: dummyWorkflow().Config.Stages[0].Steps[0]}
			s.CurrStepState = backend_types.State{Exited: true, Error: tc.err}
			require.NoError(t, trace(s))

			assert.Equal(t, tc.wantKind, got.ErrorKind)
			if tc.err != nil {
				assert.Equal(t, tc.err.Error(), got.Error)
			}
		})
	}
}
