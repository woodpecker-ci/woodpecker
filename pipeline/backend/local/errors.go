// Copyright 2023 Woodpecker Authors
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

// cSpell:ignore ERRORLEVEL

package local

import (
	"errors"
	"fmt"
)

var (
	ErrUnsupportedStepType   = errors.New("unsupported step type")
	ErrStepReaderNotFound    = errors.New("could not found pipe reader for step")
	ErrWorkflowStateNotFound = errors.New("workflow state not found")
	ErrNoShellSet            = errors.New("no shell was set")
	ErrNoCmdSet              = errors.New("no commands where set")
)

// ErrNoPosixShell indicates that a shell was assumed to be POSIX-compatible but failed the test.
type ErrNoPosixShell struct {
	Shell string
	Err   error
}

func (e *ErrNoPosixShell) Error() string {
	return fmt.Sprintf("Shell %q was assumed as posix shell but test failed: %v\n(if you want support for it, open an issue at woodpecker project)", e.Shell, e.Err)
}

// Unwrap returns the underlying error for errors.Is and errors.As support.
func (e *ErrNoPosixShell) Unwrap() error {
	return e.Err
}

// Is enables errors.Is comparison.
func (e *ErrNoPosixShell) Is(target error) bool {
	_, ok := target.(*ErrNoPosixShell)
	return ok
}
