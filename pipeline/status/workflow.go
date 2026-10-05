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
	"slices"
	"time"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
)

// WorkflowState is what a runtime reports about a workflow.
type WorkflowState struct {
	Started  int64  `json:"started"`
	Finished int64  `json:"finished"`
	Error    string `json:"error"`
	Canceled bool   `json:"canceled"`
}

// Finish marks the workflow as finished with the error its runtime returned.
func (s *WorkflowState) Finish(err error) {
	s.Finished = time.Now().Unix()
	if err == nil {
		return
	}
	if errors.Is(err, pipeline_errors.ErrCancel) {
		// A canceled workflow did not fail. Error is shown as a runtime
		// error, so only report the cancellation itself.
		s.Canceled = true
	} else {
		s.Error = err.Error()
	}
}

// Workflow returns the status of a workflow based on its steps.
func Workflow(steps []Step) Value {
	status := Success
	for _, s := range steps {
		if s.Failure == string(metadata.FailureFail) || !s.State.IsFailing() {
			status = Merge(status, s.State)
		}
	}
	return status
}

// WorkflowDone returns the status of a workflow its runtime reported as done.
func WorkflowDone(steps []Step, state WorkflowState) Value {
	status := Skipped
	if state.Started != 0 {
		status = Workflow(steps)
	}
	if state.Error != "" {
		status = Failure
	}
	if state.Canceled {
		status = Killed
	}
	return status
}

// ShouldRun tells if a workflow should be run or skipped, based on the status
// of its dependencies and the statuses it runs on.
func ShouldRun(runsOn []string, deps []Value) bool {
	onFailure := slices.Contains(runsOn, string(Failure))
	onSuccess := len(runsOn) == 0 || slices.Contains(runsOn, string(Success))
	failed := slices.ContainsFunc(deps, func(s Value) bool { return s != Success })
	succeeded := slices.Contains(deps, Success)

	switch {
	case onFailure && onSuccess:
		return true
	case onSuccess:
		return !failed
	case onFailure:
		return !succeeded
	}
	return false
}
