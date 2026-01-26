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

package pipeline

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

type backendStateRecovery interface {
	RecordStepStarted(ctx context.Context, taskUUID string, step *backend.Step) error
	RecordStepCompleted(ctx context.Context, taskUUID string, step *backend.Step, exitCode int) error
	RecordStepSkipped(ctx context.Context, taskUUID string, step *backend.Step) error
}

type backendStepStatusProvider interface {
	GetStepStatus(ctx context.Context, taskUUID, stepUUID string) (backend.StepStatus, error)
}

// backendStateRecoveryEnabled checks if the backend supports state recovery
// and should preserve state during agent shutdown.
type backendStateRecoveryEnabled interface {
	SupportsStateRecovery() bool
}

func (r *Runtime) stateRecovery() backendStateRecovery {
	if sr, ok := r.engine.(backendStateRecovery); ok {
		return sr
	}
	return nil
}

func (r *Runtime) stepStatusProvider() backendStepStatusProvider {
	if sp, ok := r.engine.(backendStepStatusProvider); ok {
		return sp
	}
	return nil
}

func (r *Runtime) supportsStateRecovery() bool {
	if sr, ok := r.engine.(backendStateRecoveryEnabled); ok {
		return sr.SupportsStateRecovery()
	}
	return false
}

func (r *Runtime) getStepStatus(stepUUID string) (backend.StepStatus, error) {
	if sp := r.stepStatusProvider(); sp != nil {
		status, err := sp.GetStepStatus(r.ctx, r.taskUUID, stepUUID)
		if err != nil {
			log.Warn().Err(err).Str("stepUUID", stepUUID).Msg("failed to get step status")
			return backend.StatusUnknown, fmt.Errorf("failed to get step status: %w", err)
		}
		return status, nil
	}
	return backend.StatusUnknown, nil
}

func (r *Runtime) recordStepStarted(step *backend.Step) {
	if sr := r.stateRecovery(); sr != nil {
		if err := sr.RecordStepStarted(r.ctx, r.taskUUID, step); err != nil {
			log.Warn().Err(err).Str("step", step.Name).Msg("failed to record step started")
		}
	}
}

func (r *Runtime) recordStepCompleted(step *backend.Step, exitCode int) {
	if sr := r.stateRecovery(); sr != nil {
		if err := sr.RecordStepCompleted(r.ctx, r.taskUUID, step, exitCode); err != nil {
			log.Warn().Err(err).Str("step", step.Name).Msg("failed to record step completed")
		}
	}
}

func (r *Runtime) recordStepSkipped(step *backend.Step) {
	if sr := r.stateRecovery(); sr != nil {
		if err := sr.RecordStepSkipped(r.ctx, r.taskUUID, step); err != nil {
			log.Warn().Err(err).Str("step", step.Name).Msg("failed to record step skipped")
		}
	}
}

func (r *Runtime) execReconnect(step *backend.Step) (*backend.State, error) {
	statusProvider, ok := r.engine.(backendStepStatusProvider)
	if ok {
		currentStatus, err := statusProvider.GetStepStatus(r.ctx, r.taskUUID, step.UUID)
		if err != nil {
			log.Warn().Err(err).Str("stepUUID", step.UUID).Msg("failed to get step status during reconnect")
		} else {
			switch currentStatus {
			case backend.StatusSuccess:
				log.Info().Str("stepUUID", step.UUID).Msg("step already succeeded, skipping reconnect")
				return &backend.State{ExitCode: 0, Exited: true}, nil
			case backend.StatusFailed:
				log.Info().Str("stepUUID", step.UUID).Msg("step already failed, skipping reconnect")
				return &backend.State{ExitCode: 1, Exited: true}, nil
			case backend.StatusSkipped:
				log.Info().Str("stepUUID", step.UUID).Msg("step was skipped, skipping reconnect")
				return &backend.State{ExitCode: 0, Exited: true}, nil
			}
		}
	}

	return r.tailAndWait(step)
}
