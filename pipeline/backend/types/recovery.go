// Copyright 2026 Woodpecker Authors
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

import "context"

// StepStatus represents the state of a step during recovery.
type StepStatus int8

const (
	StatusPending StepStatus = 0
	StatusRunning StepStatus = 1
	StatusSuccess StepStatus = 2
	StatusFailed  StepStatus = 3
	StatusSkipped StepStatus = 4
	StatusUnknown StepStatus = -1
)

func (s StepStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusRunning:
		return "running"
	case StatusSuccess:
		return "success"
	case StatusFailed:
		return "failed"
	case StatusSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// BackendRecovery is an optional interface that backends can implement to support
// workflow state recovery after agent restarts or failures.
type BackendRecovery interface {
	// GetStepStatus retrieves the current status of a step from persistent state.
	GetStepStatus(ctx context.Context, taskUUID, stepUUID string) (StepStatus, error)

	// CleanupExpiredStates removes expired state from previous workflow runs.
	// This is called on agent startup to clean up state that accumulated while
	// the agent was offline or from workflows that exceeded their timeout.
	CleanupExpiredStates(ctx context.Context)
}
