// Copyright 2021 Woodpecker Authors
// Copyright 2011 Drone.IO Inc.
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

package rpc

import (
	"context"

	backend "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

type (
	// Filter defines filters for fetching items from the queue.
	Filter struct {
		Labels map[string]string `json:"labels"`
	}

	// StepState defines the step state.
	StepState struct {
		StepUUID string `json:"step_uuid"`
		Started  int64  `json:"started"`
		Finished int64  `json:"finished"`
		Exited   bool   `json:"exited"`
		ExitCode int    `json:"exit_code"`
		Error    string `json:"error"`
	}

	// WorkflowState defines the workflow state.
	WorkflowState struct {
		Started  int64  `json:"started"`
		Finished int64  `json:"finished"`
		Error    string `json:"error"`
	}

	// Workflow defines the workflow execution details.
	Workflow struct {
		ID      string          `json:"id"`
		Config  *backend.Config `json:"config"`
		Timeout int64           `json:"timeout"`
	}

	Version struct {
		GrpcVersion   int32  `json:"grpc_version,omitempty"`
		ServerVersion string `json:"server_version,omitempty"`
	}

	// AgentInfo represents all the metadata that should be known about an agent.
	AgentInfo struct {
		Version      string            `json:"version"`
		Platform     string            `json:"platform"`
		Backend      string            `json:"backend"`
		Capacity     int               `json:"capacity"`
		CustomLabels map[string]string `json:"custom_labels"`
	}
)

//go:generate mockery --name Peer --output mocks --case underscore --note "+build test"

// Peer defines a peer-to-peer connection.
type Peer interface {
	// Version returns the server- & grpc-version
	Version(c context.Context) (*Version, error)

	// Next returns the next workflow in the queue
	Next(c context.Context, f Filter) (*Workflow, error)

	// Wait blocks until the workflow is complete
	Wait(c context.Context, workflowID string) error

	// Init signals the workflow is initialized
	Init(c context.Context, workflowID string, state WorkflowState) error

	// Done signals the workflow is complete
	Done(c context.Context, workflowID string, state WorkflowState) error

	// Extend extends the workflow deadline
	Extend(c context.Context, workflowID string) error

	// Update updates the step state
	Update(c context.Context, workflowID string, state StepState) error

	// EnqueueLog queues the step log entry for delayed sending
	EnqueueLog(logEntry *LogEntry)

	// RegisterAgent register our agent to the server
	RegisterAgent(ctx context.Context, info AgentInfo) (int64, error)

	// UnregisterAgent unregister our agent from the server
	UnregisterAgent(ctx context.Context) error

	// ReportHealth reports health status of the agent to the server
	ReportHealth(c context.Context) error
}
