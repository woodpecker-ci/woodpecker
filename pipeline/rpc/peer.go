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

	// State defines the workflow state.
	State struct {
		Step     string `json:"step"`
		Exited   bool   `json:"exited"`
		ExitCode int    `json:"exit_code"`
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
)

// Peer defines a peer-to-peer connection.
type Peer interface {
	// Version returns the server- & grpc-version
	Version(c context.Context) (*Version, error)

	// Next returns the next workflow in the queue
	Next(c context.Context, f Filter) (*Workflow, error)

	// Wait blocks until the workflow is complete
	Wait(c context.Context, id string) error

	// Init signals the workflow is initialized
	Init(c context.Context, id string, state State) error

	// Done signals the workflow is complete
	Done(c context.Context, id string, state State) error

	// Extend extends the workflow deadline
	Extend(c context.Context, id string) error

	// Update updates the workflow state
	Update(c context.Context, id string, state State) error

	// Log writes the workflow log entry
	Log(c context.Context, logEntry *LogEntry) error

	// RegisterAgent register our agent to the server
	RegisterAgent(ctx context.Context, platform, backend, version string, capacity int) (int64, error)

	// UnregisterAgent unregister our agent from the server
	UnregisterAgent(ctx context.Context) error

	// ReportHealth reports health status of the agent to the server
	ReportHealth(c context.Context) error

	// TODO: callback api to update pipeline etc...
}
