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

package model

// StepRecoveryState represents the recovery state for a workflow step.
// This is used to track step progress and enable agent restart recovery.
type StepRecoveryState struct {
	ID         int64  `xorm:"pk autoincr 'id'"`
	WorkflowID string `xorm:"VARCHAR(250) UNIQUE(s) INDEX 'workflow_id'"` // Task ID from queue
	StepUUID   string `xorm:"VARCHAR(250) UNIQUE(s) 'step_uuid'"`         // Step UUID within workflow
	Status     int    `xorm:"'status'"`                                   // Recovery status (see rpc.RecoveryStatus)
	ExitCode   int    `xorm:"'exit_code'"`
	StartedAt  int64  `xorm:"'started_at'"`
	FinishedAt int64  `xorm:"'finished_at'"`
	AgentID    int64  `xorm:"'agent_id'"`
	CreatedAt  int64  `xorm:"created 'created_at'"`
	UpdatedAt  int64  `xorm:"updated 'updated_at'"`
	ExpiresAt  int64  `xorm:"INDEX 'expires_at'"` // For cleanup of old states
}

// TableName returns the database table name.
func (StepRecoveryState) TableName() string {
	return "step_recovery_states"
}
