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

package datastore

import (
	"time"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// RecoveryStateCreate creates recovery states for all steps in a workflow.
// This is idempotent - if states already exist for the workflow, it does nothing.
func (s storage) RecoveryStateCreate(workflowID string, stepUUIDs []string, agentID, expiresAt int64) error {
	// Check if recovery states already exist for this workflow
	exists, err := s.engine.Where("workflow_id = ?", workflowID).Exist(new(model.StepRecoveryState))
	if err != nil {
		return err
	}
	if exists {
		// Already initialized, nothing to do
		return nil
	}

	// Batch insert all step recovery states
	now := time.Now().Unix()
	states := make([]*model.StepRecoveryState, 0, len(stepUUIDs))
	for _, stepUUID := range stepUUIDs {
		states = append(states, &model.StepRecoveryState{
			WorkflowID: workflowID,
			StepUUID:   stepUUID,
			Status:     0,
			AgentID:    agentID,
			CreatedAt:  now,
			UpdatedAt:  now,
			ExpiresAt:  expiresAt,
		})
	}

	_, err = s.engine.Insert(&states)
	return err
}

// RecoveryStateGetAll retrieves all recovery states for a workflow.
func (s storage) RecoveryStateGetAll(workflowID string) ([]*model.StepRecoveryState, error) {
	var states []*model.StepRecoveryState
	err := s.engine.Where("workflow_id = ?", workflowID).Find(&states)
	return states, err
}

// RecoveryStateUpdate updates a recovery state.
func (s storage) RecoveryStateUpdate(state *model.StepRecoveryState) error {
	_, err := s.engine.Where("workflow_id = ? AND step_uuid = ?", state.WorkflowID, state.StepUUID).
		Cols("status", "exit_code", "started_at", "finished_at", "updated_at").
		Update(state)
	return err
}

// RecoveryStateCleanExpired removes expired recovery states.
func (s storage) RecoveryStateCleanExpired() error {
	_, err := s.engine.Where("expires_at < ?", time.Now().Unix()).Delete(new(model.StepRecoveryState))
	return err
}
