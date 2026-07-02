// Copyright 2024 Woodpecker Authors
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

package scheduler

// State is the lifecycle state of a single workflow inside the DAG.
type State int

const (
	// StatePending means the workflow has not yet had its dependencies
	// evaluated.
	StatePending State = iota
	// StateReady means all dependencies have completed successfully and
	// the workflow is eligible to start. A workflow only stays in this
	// state briefly before being moved to StateRunning.
	StateReady
	// StateRunning means the workflow is currently executing.
	StateRunning
	// StateSuccess means the workflow ran to completion without error.
	StateSuccess
	// StateFailure means the workflow's run function returned a non-nil
	// error.
	StateFailure
	// StateBlocked means at least one dependency did not complete
	// successfully, so the workflow was never started. This is distinct
	// from a step-level skip (which comes from a "when:" clause inside
	// the workflow itself).
	StateBlocked
	// StateCanceled means the workflow was still pending or running
	// when the parent context was canceled.
	StateCanceled
)

// String returns a short, lowercase name for the state, suitable for
// logging and rendering.
func (s State) String() string {
	switch s {
	case StatePending:
		return "pending"
	case StateReady:
		return "ready"
	case StateRunning:
		return "running"
	case StateSuccess:
		return "success"
	case StateFailure:
		return "failure"
	case StateBlocked:
		return "blocked"
	case StateCanceled:
		return "canceled"
	}
	return "unknown"
}

// Terminal reports whether the state is a final state that will not
// transition again for this run.
func (s State) Terminal() bool {
	switch s {
	case StateSuccess, StateFailure, StateBlocked, StateCanceled:
		return true
	}
	return false
}

// Event is a workflow-level state transition emitted by the scheduler.
//
// Events are emitted in the order they occur from a single goroutine
// inside the scheduler, so consumers see a consistent sequence. The
// channel is the only synchronization point between the scheduler and
// its observers.
type Event struct {
	// Workflow is the workflow name as set by the builder
	// (Workflow.Name). It is stable across the run and unique within
	// the run.
	Workflow string
	// State is the new state of the workflow at the moment the event
	// was emitted.
	State State
	// Err is set only when State is StateFailure or when State is
	// StateBlocked with a non-nil underlying dependency failure. The
	// scheduler does not wrap the original error; it is the raw error
	// returned by the run function (or BlockedError for blocked
	// workflows).
	Err error
}

// BlockedError is the error value delivered in an Event when a
// workflow is skipped because a dependency did not succeed.
type BlockedError struct {
	// Dependency is the name of the workflow whose non-success caused
	// this workflow to be blocked. When multiple dependencies failed,
	// the scheduler picks the first one it observed failing — this
	// matches the natural ordering of event emission.
	Dependency string
}

// Error implements error.
func (e *BlockedError) Error() string {
	return "blocked: dependency '" + e.Dependency + "' did not succeed"
}
