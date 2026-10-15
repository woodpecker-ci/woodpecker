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

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

var (
	// ErrSkip is used as a return value when container execution should be
	// skipped at runtime. It is not returned as an error by any function.
	ErrSkip = errors.New("Skipped")

	// ErrCancel is used as a return value when the container execution receives
	// a cancellation signal from the context. It is the well known step error,
	// so a canceled step reaches the server as such.
	ErrCancel error = backend_types.ErrStepCanceled
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

// IsStepFailure reports whether err was caused by a step itself terminating
// unsuccessfully (non-zero exit code or oom kill), as opposed to the runtime
// or backend failing to execute the workflow.
func IsStepFailure(err error) bool {
	var exitErr *ExitError
	return errors.As(err, &exitErr) || errors.Is(err, backend_types.ErrStepOOMKilled)
}
