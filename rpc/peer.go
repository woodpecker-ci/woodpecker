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

import "context"

// Peer defines the bidirectional communication interface between Woodpecker agents and servers.
//
// # Architecture and Implementations
//
// The Peer interface is implemented differently on each side of the communication:
//
//   - Agent side: Implemented by agent/rpc/client_grpc.go's client struct, which wraps
//     a gRPC client connection to make RPC calls to the server.
//
//   - Server side: Implemented by server/rpc/rpc.go's RPC struct, which contains the
//     business logic and is wrapped by server/rpc/server.go's WoodpeckerServer struct
//     to handle incoming gRPC requests.
//
// # Thread Safety and Concurrency
//
//   - Implementations must be safe for concurrent calls across different workflows
//   - The same Peer instance may be called concurrently from multiple goroutines
//   - Each workflow is identified by a unique workflowID string
//   - Implementations must properly isolate workflow state using workflowID
//
// # Error Handling Conventions
//
//   - Methods return errors for communication failures, validation errors, or server-side issues
//   - Errors should not be used for bussines logic
//   - Network/transport errors should be retried by the caller when appropriate
//   - Nil error indicates successful operation
//   - Context cancellation should return nil or context.Canceled, not a custom error
//   - Business logic errors (e.g., workflow not found) return specific error types
//
// # Intended Execution Flow
//
//  1. Agent Lifecycle:
//     - Version() checks compatibility with server
//     - RegisterAgent() announces agent availability
//     - ReportHealth() periodically confirms agent is alive
//     - UnregisterAgent() gracefully disconnects agent
//
//  2. Workflow Execution (may happen concurrently for multiple workflows):
//     - Next() blocks until server assigns a workflow
//     - Init() signals workflow execution has started
//     - Wait() (in background goroutine) monitors for cancellation signals
//     - Update() reports step state changes as workflow progresses
//     - EnqueueLog() streams log output from steps
//     - Extend() extends workflow timeout if needed so queue does not reschedule it as retry
//     - Done() signals workflow has completed
//
//  3. Cancellation Flow:
//     - Server can cancel workflow by releasing Wait() with canceled=true
//     - Agent detects cancellation from Wait() return value
//     - Agent stops workflow execution and calls Done() with canceled state
type Peer interface {
	// Version returns the server- & grpc-version.
	Version(c context.Context) (*Version, error)

	// Next blocks until it provides the next workflow to execute from the queue.
	Next(c context.Context, f Filter) (*Workflow, error)

	// Wait blocks until the workflow with the given ID is completed.
	// Also signals via err if workflow got canceled.
	Wait(c context.Context, workflowID string) (canceled bool, err error)

	// Init signals the workflow is initialized.
	Init(c context.Context, workflowID string, state WorkflowState) error

	// Done let agent signal to server the workflow has stopped.
	Done(c context.Context, workflowID string, state WorkflowState) error

	// Extend extends the workflow deadline.
	Extend(c context.Context, workflowID string) error

	// Update let agent updates the step state at the server.
	Update(c context.Context, workflowID string, state StepState) error

	// EnqueueLog queues the step log entry for delayed sending.
	EnqueueLog(logEntry *LogEntry)

	// RegisterAgent register our agent to the server.
	RegisterAgent(ctx context.Context, info AgentInfo) (int64, error)

	// UnregisterAgent unregister our agent from the server.
	UnregisterAgent(ctx context.Context) error

	// ReportHealth reports health status of the agent to the server.
	ReportHealth(c context.Context) error
}
