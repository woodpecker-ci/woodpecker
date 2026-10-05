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

package status

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
)

func TestNewStepState(t *testing.T) {
	step := &backend_types.Step{UUID: "uuid"}

	t.Run("started", func(t *testing.T) {
		got := NewStepState(&state.State{CurrStep: step})
		assert.Equal(t, StepState{StepUUID: "uuid"}, got)
	})

	t.Run("exited with error", func(t *testing.T) {
		got := NewStepState(&state.State{CurrStep: step, CurrStepState: backend_types.State{
			Started: 1, Exited: true, ExitCode: 2, Error: errors.New("boom"),
		}})
		assert.NotZero(t, got.Finished)
		got.Finished = 0
		assert.Equal(t, StepState{StepUUID: "uuid", Started: 1, Exited: true, ExitCode: 2, Error: "boom"}, got)
	})

	t.Run("canceled", func(t *testing.T) {
		got := NewStepState(&state.State{CurrStep: step, CurrStepState: backend_types.State{
			Error: fmt.Errorf("wrapped: %w", pipeline_errors.ErrCancel),
		}})
		assert.True(t, got.Canceled)
		assert.Zero(t, got.Finished, "only an exited step is finished")
	})

	t.Run("skipped", func(t *testing.T) {
		got := NewStepState(&state.State{CurrStep: step, CurrStepState: backend_types.State{Skipped: true}})
		assert.Equal(t, StepState{StepUUID: "uuid", Skipped: true}, got)
	})
}

func TestWorkflowStateFinish(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s := WorkflowState{Started: 1}
		s.Finish(nil)
		assert.NotZero(t, s.Finished)
		assert.Empty(t, s.Error)
		assert.False(t, s.Canceled)
	})

	t.Run("runtime error", func(t *testing.T) {
		s := WorkflowState{Started: 1}
		s.Finish(errors.New("no backend"))
		assert.Equal(t, "no backend", s.Error)
		assert.False(t, s.Canceled)
	})

	t.Run("canceled is no error", func(t *testing.T) {
		s := WorkflowState{Started: 1}
		s.Finish(errors.Join(errors.New("step killed"), pipeline_errors.ErrCancel))
		assert.Empty(t, s.Error)
		assert.True(t, s.Canceled)
	})
}
