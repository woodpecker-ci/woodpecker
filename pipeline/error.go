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
	Name string
	Code int
}

// Error returns the error message in string format.
func (e *ExitError) Error() string {
	return fmt.Sprintf("%s : exit code %d", e.Name, e.Code)
}

// An OomError reports the process received an OOMKill from the kernel.
type OomError struct {
	Name string
	Code int
}

// Error returns the error message in string format.
func (e *OomError) Error() string {
	return fmt.Sprintf("%s : received oom kill", e.Name)
}
