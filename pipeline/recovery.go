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

package pipeline

import (
	"context"
	"sync/atomic"

	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
)

// RecoveryClient defines the interface for recovery state communication.
type RecoveryClient interface {
	InitWorkflowRecovery(ctx context.Context, workflowID string, stepUUIDs []string, timeoutSeconds int64) error
	GetWorkflowRecoveryStates(ctx context.Context, workflowID string) (map[string]*rpc.RecoveryState, error)
	UpdateStepRecoveryState(ctx context.Context, workflowID, stepUUID string, status rpc.RecoveryStatus, exitCode int) error
}

// RecoveryManager manages the recovery state for pipeline steps.
type RecoveryManager struct {
	client     RecoveryClient
	workflowID string
	enabled    bool
	stateCache map[string]*rpc.RecoveryState // step UUID -> state (loaded once)
	canceled   atomic.Bool                   // set when workflow is canceled by user/API
}

// NewRecoveryManager creates a new RecoveryManager.
func NewRecoveryManager(client RecoveryClient, workflowID string, enabled bool) *RecoveryManager {
	return &RecoveryManager{
		client:     client,
		workflowID: workflowID,
		enabled:    enabled,
	}
}

// InitRecoveryState initializes recovery state for all steps in the config.
// On first run, creates recovery states for all steps.
// On agent restart, loads existing states into cache.
func (m *RecoveryManager) InitRecoveryState(ctx context.Context, config *backend.Config, timeoutSeconds int64) error {
	if !m.enabled || m.client == nil {
		return nil
	}

	// Create recovery states (idempotent - skips if already exists)
	stepUUIDs := collectStepUUIDs(config)
	if err := m.client.InitWorkflowRecovery(ctx, m.workflowID, stepUUIDs, timeoutSeconds); err != nil {
		return err
	}

	// Load all states into cache (single RPC call)
	states, err := m.client.GetWorkflowRecoveryStates(ctx, m.workflowID)
	if err != nil {
		return err
	}
	m.stateCache = states
	return nil
}

// GetStepState retrieves the recovery state for a step from cache.
func (m *RecoveryManager) GetStepState(step *backend.Step) *rpc.RecoveryState {
	if !m.enabled || m.stateCache == nil {
		return &rpc.RecoveryState{Status: rpc.RecoveryStatusPending}
	}

	if state, ok := m.stateCache[step.UUID]; ok {
		return state
	}
	return &rpc.RecoveryState{Status: rpc.RecoveryStatusPending}
}

// MarkStepRunning marks a step as running.
func (m *RecoveryManager) MarkStepRunning(ctx context.Context, step *backend.Step) error {
	if !m.enabled || m.client == nil {
		return nil
	}

	return m.client.UpdateStepRecoveryState(ctx, m.workflowID, step.UUID, rpc.RecoveryStatusRunning, 0)
}

// MarkStepSuccess marks a step as successfully completed.
func (m *RecoveryManager) MarkStepSuccess(ctx context.Context, step *backend.Step) error {
	if !m.enabled || m.client == nil {
		return nil
	}

	return m.client.UpdateStepRecoveryState(ctx, m.workflowID, step.UUID, rpc.RecoveryStatusSuccess, 0)
}

// MarkStepFailed marks a step as failed.
func (m *RecoveryManager) MarkStepFailed(ctx context.Context, step *backend.Step, exitCode int) error {
	if !m.enabled || m.client == nil {
		return nil
	}

	return m.client.UpdateStepRecoveryState(ctx, m.workflowID, step.UUID, rpc.RecoveryStatusFailed, exitCode)
}

// MarkStepSkipped marks a step as skipped.
func (m *RecoveryManager) MarkStepSkipped(ctx context.Context, step *backend.Step) error {
	if !m.enabled || m.client == nil {
		return nil
	}

	return m.client.UpdateStepRecoveryState(ctx, m.workflowID, step.UUID, rpc.RecoveryStatusSkipped, 0)
}

// ShouldSkipStep determines if a step should be skipped based on its recovery state.
// Returns true if the step was already completed (success, failed, or skipped).
func (m *RecoveryManager) ShouldSkipStep(step *backend.Step) (bool, *rpc.RecoveryState) {
	if !m.enabled {
		return false, nil
	}

	state := m.GetStepState(step)

	switch state.Status {
	case rpc.RecoveryStatusSuccess, rpc.RecoveryStatusFailed, rpc.RecoveryStatusSkipped:
		return true, state
	default:
		return false, state
	}
}

// ShouldReconnect determines if we should attempt to reconnect to a running step.
// This is only applicable for backends that support reconnection (Docker, Kubernetes).
func (m *RecoveryManager) ShouldReconnect(state *rpc.RecoveryState) bool {
	if state == nil {
		return false
	}
	return state.Status == rpc.RecoveryStatusRunning
}

// Enabled returns whether recovery is enabled.
func (m *RecoveryManager) Enabled() bool {
	return m.enabled
}

// SetCanceled marks the workflow as canceled by user/API.
func (m *RecoveryManager) SetCanceled() {
	m.canceled.Store(true)
}

// WasCanceled returns whether the workflow was canceled by user/API.
func (m *RecoveryManager) WasCanceled() bool {
	return m.canceled.Load()
}

// collectStepUUIDs extracts all step UUIDs from the config.
func collectStepUUIDs(config *backend.Config) []string {
	var uuids []string
	for _, stage := range config.Stages {
		for _, step := range stage.Steps {
			if step.UUID != "" {
				uuids = append(uuids, step.UUID)
			}
		}
	}
	return uuids
}
