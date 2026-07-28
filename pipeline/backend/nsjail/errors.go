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

package nsjail

import (
	"errors"
	"fmt"
)

var (
	ErrUnsupportedStepType   = errors.New("nsjail: unsupported step type")
	ErrNsjailNotFound        = errors.New("nsjail binary not found")
	ErrNsjailNotAvailable    = errors.New("nsjail requires Linux")
	ErrNoCmdSet              = errors.New("nsjail: no commands provided")
	ErrNoShellSet            = errors.New("nsjail: no shell set")
	ErrStepReaderNotFound    = errors.New("nsjail: could not found pipe reader for step")
	ErrWorkflowStateNotFound = errors.New("nsjail: workflow state not found")
	ErrStepStateNotFound     = errors.New("nsjail: step state not found")
)

// ErrNoPosixShell indicates that a shell was assumed to be POSIX-compatible but failed the test.
type ErrNoPosixShell struct {
	Shell string
	Err   error
}

func (e *ErrNoPosixShell) Error() string {
	return fmt.Sprintf("nsjail: shell %q is not POSIX compatible: %v", e.Shell, e.Err)
}

func (e *ErrNoPosixShell) Unwrap() error {
	return e.Err
}
