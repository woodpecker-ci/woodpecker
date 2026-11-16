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
	"time"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
)

// StepState is what a runtime reports about a step.
type StepState struct {
	StepUUID string `json:"step_uuid"`
	Started  int64  `json:"started"`
	Finished int64  `json:"finished"`
	Exited   bool   `json:"exited"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error"`
	Canceled bool   `json:"canceled"`
	Skipped  bool   `json:"skipped"`
}

// NewStepState converts the state traced by the pipeline runtime.
func NewStepState(s *state.State) StepState {
	stepState := StepState{
		StepUUID: s.CurrStep.UUID,
		Exited:   s.CurrStepState.Exited,
		ExitCode: s.CurrStepState.ExitCode,
		Started:  s.CurrStepState.Started,
		Canceled: errors.Is(s.CurrStepState.Error, pipeline_errors.ErrCancel),
		Skipped:  s.CurrStepState.Skipped,
	}
	if s.CurrStepState.Error != nil {
		stepState.Error = s.CurrStepState.Error.Error()
	}
	if s.CurrStepState.Exited {
		stepState.Finished = time.Now().Unix()
	}
	return stepState
}

// Step is the part of a step its status is calculated from.
type Step struct {
	State    Value
	Failure  string // what a failure of the step should cause
	Started  int64
	Finished int64
	ExitCode int
	Error    string
}

// Update returns the step after the runtime reported a new state for it.
// If the step failed and is set to cancel the pipeline on failure,
// cancelPipeline is true.
func (s Step) Update(state StepState) (_ Step, cancelPipeline bool, _ error) {
	switch s.State {
	case Pending:
		// Handle skip before anything else — skipped steps never started,
		// so we must not set Started or transition through Running.
		if state.Skipped {
			s.State = Skipped
			if state.Finished != 0 {
				s.Finished = state.Finished
			}
			return s, false, nil
		}

		// Transition from pending to running when started
		if state.Finished == 0 {
			s.State = Running
		}
		s.Started = state.Started
		if s.Started == 0 {
			s.Started = time.Now().Unix()
		}

	case Running:

	default:
		return s, false, fmt.Errorf("step has state %s and does not expect rpc state updates", s.State)
	}

	// The step finished, which can also happen right away if its setup failed
	if state.Exited || state.Error != "" {
		s.Finished = state.Finished
		if s.Finished == 0 {
			s.Finished = time.Now().Unix()
		}
		s.ExitCode = state.ExitCode
		s.Error = state.Error

		if state.ExitCode == 0 && state.Error == "" {
			s.State = Success
		} else {
			s.State = Failure
			cancelPipeline = s.Failure == string(metadata.FailureCancel)
		}
	}

	if state.Canceled && s.State != Killed {
		s.State = Killed
		if s.Finished == 0 {
			s.Finished = time.Now().Unix()
		}
	}

	return s, cancelPipeline, nil
}

// End returns the step after it got the given status without the runtime
// reporting it, e.g. because its workflow is done or was canceled.
// A step that had started counts as successful, as this is how services end.
func (s Step) End(status Value, finished int64) Step {
	s.State = status
	if s.Started != 0 {
		s.State = Success // for daemons that are killed
		s.Finished = finished
	}
	return s
}
