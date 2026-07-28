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

package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrSkip is used as a return value when container execution should be
	// skipped at runtime. It is not returned as an error by any function.
	ErrSkip = errors.New("Skipped")

	// ErrCancel is used as a return value when the container execution receives
	// a cancellation signal from the context.
	ErrCancel = errors.New("Canceled")
)

// An ExitError reports an unsuccessful exit.
type ExitError struct {
	UUID string
	Code int
}

// Error returns the error message in string format.
func (e *ExitError) Error() string {
	return fmt.Sprintf("uuid=%s: exit code %d", e.UUID, e.Code)
}

// An OomError reports the process received an OOMKill from the kernel.
type OomError struct {
	UUID string
	Code int
}

// Error returns the error message in string format.
func (e *OomError) Error() string {
	return fmt.Sprintf("uuid=%s: received oom kill", e.UUID)
}

// A WorkflowDependencyError reports that a step's cross-workflow dependency
// (depends_on with a 'workflow' key) resolved unsuccessfully: the target
// workflow or step failed, was skipped/canceled, or a required target is
// missing.
type WorkflowDependencyError struct {
	Step string
	Msg  string
}

// Error returns the error message in string format.
func (e *WorkflowDependencyError) Error() string {
	return fmt.Sprintf("step=%s: cross-workflow dependency failed: %s", e.Step, e.Msg)
}

// IsStepFailure reports whether err was caused by a step itself terminating
// unsuccessfully (non-zero exit code, oom kill or a failed cross-workflow
// dependency), as opposed to the runtime or backend failing to execute the
// workflow.
func IsStepFailure(err error) bool {
	var exitErr *ExitError
	var oomErr *OomError
	var depErr *WorkflowDependencyError
	return errors.As(err, &exitErr) || errors.As(err, &oomErr) || errors.As(err, &depErr)
}
