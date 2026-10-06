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

// Package status holds the rules that turn what a runtime reports about its
// steps and workflows into the status shown to users. The server applies them
// to the reports of its agents, the cli to its local runs.
package status

// Value is the status of a pipeline, workflow or step.
type Value string

const (
	Skipped  Value = "skipped"  // skipped as per condition of current workflow failed/success state
	Pending  Value = "pending"  // pending to be executed
	Running  Value = "running"  // currently running
	Success  Value = "success"  // successfully finished
	Failure  Value = "failure"  // failed to finish (exit code != 0)
	Killed   Value = "killed"   // killed by user
	Canceled Value = "canceled" // canceled but hasn't been started
	Error    Value = "error"    // error with the config / while parsing / some other system problem
	Blocked  Value = "blocked"  // waiting for approval
	Declined Value = "declined" // blocked and declined
	Created  Value = "created"  // created / internal use only
)

// IsActive returns true if the status is pending or running.
func (v Value) IsActive() bool {
	return v == Pending || v == Running
}

// IsFailing returns true if the status is failure, killed or error.
func (v Value) IsFailing() bool {
	return v == Error || v == Killed || v == Failure
}

// list of statuses by their priority. Most important is on top.
var priorityOrder = []Value{
	// blocked, declined and created cannot appear in the
	// same workflow/pipeline at the same time
	Declined,
	Blocked,
	Created,

	// errors have highest priority.
	Error,

	// skipped and killed cannot appear together with running/pending.
	Killed,
	Canceled,

	// running states
	Running,
	Pending,

	// finished states
	Failure,
	Success,

	// skipped due to status condition
	Skipped,
}

var priority = func() map[Value]int {
	m := map[Value]int{}
	for i, s := range priorityOrder {
		m[s] = i
	}
	return m
}()

// Merge returns the status two parts of a workflow or pipeline have together.
func Merge(s, t Value) Value {
	// both are skipped due to cancellation -> canceled
	if s == Canceled && t == Canceled {
		return Canceled
	}
	// if only one was skipped -> use killed as the workflow/pipeline was running once already
	if s == Canceled {
		s = Killed
	}
	if t == Canceled {
		t = Killed
	}
	return priorityOrder[min(priority[s], priority[t])]
}
