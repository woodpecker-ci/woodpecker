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

package types

import "errors"

var ErrNoCliContextFound = errors.New("no CliContext in context found")

// StepError is a well known reason a step ended, reported next to its exit
// code. It is a plain string so it passes gRPC as it is and can be asserted
// with errors.Is on both ends.
type StepError string

const (
	// ErrStepCanceled reports that the step was canceled.
	ErrStepCanceled StepError = "canceled"
	// ErrStepOOMKilled reports that the step was killed for running out of memory.
	ErrStepOOMKilled StepError = "oom_killed"
)

func (e StepError) Error() string {
	return string(e)
}
