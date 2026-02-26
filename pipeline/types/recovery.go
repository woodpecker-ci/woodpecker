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

package types

// RecoveryStatus represents the recovery state of a step.
type RecoveryStatus int

// RecoveryState represents the recovery state for a step.
type RecoveryState struct {
	Status   RecoveryStatus `json:"status"`
	ExitCode int            `json:"exit_code"`
}

const (
	RecoveryStatusPending RecoveryStatus = iota
	RecoveryStatusRunning
	RecoveryStatusSuccess
	RecoveryStatusFailed
	RecoveryStatusSkipped
)
