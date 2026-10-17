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

package errors_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
)

// A step failure is the step itself ending unsuccessfully. It must not turn
// into a workflow runtime error, which is kept for the runtime or backend
// failing to run the workflow.
func TestIsStepFailure(t *testing.T) {
	for name, tc := range map[string]struct {
		err  error
		want bool
	}{
		"nil":            {err: nil, want: false},
		"exit code":      {err: &pipeline_errors.ExitError{UUID: "s", Code: 1}, want: true},
		"oom kill":       {err: backend_types.ErrStepOOMKilled, want: true},
		"wrapped oom":    {err: fmt.Errorf("step: %w", backend_types.ErrStepOOMKilled), want: true},
		"cancel":         {err: pipeline_errors.ErrCancel, want: false},
		"backend error":  {err: errors.New("engine exploded"), want: false},
		"workflow setup": {err: &pipeline_errors.ErrInvalidWorkflowSetup{Err: errors.New("x")}, want: false},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, pipeline_errors.IsStepFailure(tc.err))
		})
	}
}
