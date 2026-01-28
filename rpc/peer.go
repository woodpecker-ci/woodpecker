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
//   - Errors should not be used for business logic
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
	// Version returns the server and gRPC protocol version information.
	//
	// This is typically called once during agent initialization to verify
	// compatibility between agent and server versions.
	//
	// Returns:
	//   - Version with server version string and gRPC protocol version number
	//   - Error if communication fails or server is unreachable
	Version(c context.Context) (*Version, error)

	// Next blocks until the server provides the next workflow to execute from the queue.
	//
	// This is the primary work-polling mechanism. Agents call this repeatedly in a loop,
	// and it blocks until either:
	//   1. A workflow matching the filter becomes available
	//   2. The context is canceled (agent shutdown, network timeout, etc.)
	//
	// The filter allows agents to specify capabilities via labels (e.g., platform,
	// backend type) so the server only assigns compatible workflows.
	//
	// Context Handling:
	//   - This is a long-polling operation that may block for extended periods
	//   - Implementations MUST check context regularly (not just at entry)
	//   - When context is canceled, must return nil workflow and nil error
	//   - Server may send keep-alive signals or periodically return nil to allow reconnection
	//
	// Returns:
	//   - Workflow object with ID, Config, and Timeout if work is available
	//   - nil, nil if context is canceled or no work available (retry expected)
	//   - nil, error if a non-retryable error occurs
	Next(c context.Context, f Filter) (*Workflow, error)

	// Wait blocks until the workflow with the given ID completes or is canceled by the server.
	//
	// This is used by agents to monitor for server-side cancellation signals. Typically
	// called in a background goroutine immediately after Init(), running concurrently
	// with workflow execution.
	//
	// The method serves two purposes:
	//   1. Signals when server wants to cancel workflow (canceled=true)
	//   2. Unblocks when workflow completes normally on agent (canceled=false)
	//
	// Context Handling:
	//   - This is a long-running blocking operation for the workflow duration
	//   - Context cancellation indicates shutdown, not workflow cancellation
	//   - When context is canceled, should return (false, nil) or (false, ctx.Err())
	//   - Must not confuse context cancellation with workflow cancellation signal
	//
	// Cancellation Flow:
	//   - Server releases Wait() with canceled=true → agent should stop workflow
	//   - Agent completes workflow normally → Done() is called → server releases Wait() with canceled=false
	//   - Agent context canceled → Wait() returns immediately, workflow may continue on agent
	//
	// Returns:
	//   - canceled=true, err=nil: Server initiated cancellation, agent should stop workflow
	//   - canceled=false, err=nil: Workflow completed normally (Wait unblocked by Done call)
	//   - canceled=false, err!=nil: Communication error, agent should retry or handle error
	Wait(c context.Context, workflowID string) (canceled bool, err error)

	// Init signals to the server that the workflow has been initialized and execution has started.
	//
	// This is called once per workflow immediately after the agent accepts it from Next()
	// and before starting step execution. It allows the server to track workflow start time
	// and update workflow status to "running".
	//
	// The WorkflowState should have:
	//   - Started: Unix timestamp when execution began
	//   - Finished: 0 (not finished yet)
	//   - Error: empty string (no error yet)
	//   - Canceled: false (not canceled yet)
	//
	// Returns:
	//   - nil on success
	//   - error if communication fails or server rejects the state
	Init(c context.Context, workflowID string, state WorkflowState) error

	// Done signals to the server that the workflow has completed execution.
	//
	// This is called once per workflow after all steps have finished (or workflow was canceled).
	// It provides the final workflow state including completion time, any errors, and
	// cancellation status.
	//
	// The WorkflowState should have:
	//   - Started: Unix timestamp when execution began (same as Init)
	//   - Finished: Unix timestamp when execution completed
	//   - Error: Error message if workflow failed, empty if successful
	//   - Canceled: true if workflow was canceled, false otherwise
	//
	// After Done() is called:
	//   - Server updates final workflow status in database
	//   - Server releases any Wait() calls for this workflow
	//   - Server removes workflow from active queue
	//   - Server notifies forge of workflow completion
	//
	// Context Handling:
	//   - MUST attempt to complete even if workflow context is canceled
	//   - Often called with a shutdown/cleanup context rather than workflow context
	//   - Critical for proper cleanup - should retry on transient failures
	//
	// Returns:
	//   - nil on success
	//   - error if communication fails or server rejects the state
	Done(c context.Context, workflowID string, state WorkflowState) error

	// Extend extends the execution deadline for the workflow with the given ID.
	//
	// Workflows have a timeout (specified in Workflow.Timeout from Next()). Agents should
	// call Extend() periodically (e.g. constant.TaskTimeout / 3) to signal the workflow is still
	// actively executing and prevent premature timeout.
	//
	// This acts as a heartbeat mechanism to detect stuck workflow executions. If an agent dies or
	// becomes unresponsive, the server will eventually timeout the workflow after the
	// deadline expires without extension.
	//
	// Returns:
	//   - nil on success (deadline was extended)
	//   - error if communication fails or workflow is not found
	Extend(c context.Context, workflowID string) error

	// Update reports step state changes to the server as the workflow progresses.
	//
	// This is called multiple times per step:
	//   1. When step starts (Exited=false, Finished=0)
	//   2. When step completes (Exited=true, Finished and ExitCode set)
	//   3. Potentially on progress updates if step has long-running operations
	//
	// The server uses these updates to:
	//   - Track step execution progress
	//   - Update UI with real-time status
	//   - Store step results in database
	//   - Calculate workflow completion
	//
	// Context Handling:
	//   - Failures should be logged but not block workflow execution
	//
	// Returns:
	//   - nil on success
	//   - error if communication fails or server rejects the state
	Update(c context.Context, workflowID string, state StepState) error

	// EnqueueLog queues a log entry for delayed batch sending to the server.
	//
	// Log entries are produced continuously during step execution and need to be
	// transmitted efficiently. This method adds logs to an internal queue that
	// batches and sends them periodically to reduce network overhead.
	//
	// The implementation should:
	//   - Queue the log entry in a memory buffer
	//   - Batch multiple entries together
	//   - Send batches periodically (e.g., every second) or when buffer fills
	//   - Handle backpressure if server is slow or network is congested
	//
	// Unlike other methods, EnqueueLog:
	//   - Does NOT take a context parameter (fire-and-forget)
	//   - Does NOT return an error (never blocks the caller)
	//   - Does NOT guarantee immediate transmission
	//
	// Thread Safety:
	//   - MUST be safe to call concurrently from multiple goroutines
	//   - May be called concurrently from different steps/workflows
	//   - Internal queue must be properly synchronized
	EnqueueLog(logEntry *LogEntry)

	// RegisterAgent announces this agent to the server and returns an agent ID.
	//
	// This is called once during agent startup to:
	//   - Create an agent record in the server database
	//   - Obtain a unique agent ID for subsequent requests
	//   - Declare agent capabilities (platform, backend, capacity, labels)
	//   - Enable server-side agent tracking and monitoring
	//
	// The AgentInfo should specify:
	//   - Version: Agent version string (e.g., "v2.0.0")
	//   - Platform: OS/architecture (e.g., "linux/amd64")
	//   - Backend: Execution backend (e.g., "docker", "kubernetes")
	//   - Capacity: Maximum concurrent workflows (e.g., 2)
	//   - CustomLabels: Additional key-value labels for filtering
	//
	// Context Handling:
	//   - Context cancellation indicates agent is aborting startup
	//   - Should not retry indefinitely - fail fast on persistent errors
	//
	// Returns:
	//   - agentID: Unique identifier for this agent (use in subsequent calls)
	//   - error: If registration fails
	RegisterAgent(ctx context.Context, info AgentInfo) (int64, error)

	// UnregisterAgent removes this agent from the server's registry.
	//
	// This is called during graceful agent shutdown to:
	//   - Mark agent as offline in server database
	//   - Allow server to stop assigning workflows to this agent
	//   - Clean up any agent-specific server resources
	//   - Provide clean shutdown signal to monitoring systems
	//
	// After UnregisterAgent:
	//   - Agent should stop calling Next() for new work
	//   - Agent should complete any in-progress workflows
	//   - Agent may call Done() to finish existing workflows
	//   - Agent should close network connections
	//
	// Context Handling:
	//   - MUST attempt to complete even during forced shutdown
	//   - Often called with a shutdown context (limited time)
	//   - Failure is logged but should not prevent agent exit
	//
	// Returns:
	//   - nil on success
	//   - error if communication fails
	UnregisterAgent(ctx context.Context) error

	// ReportHealth sends a periodic health status update to the server.
	//
	// This is called regularly (e.g., every 30 seconds) during agent operation to:
	//   - Prove agent is still alive and responsive
	//   - Allow server to detect dead or stuck agents
	//   - Update agent's "last seen" timestamp in database
	//   - Provide application-level keepalive beyond network keep-alive signals
	//
	// Health reporting helps the server:
	//   - Mark unresponsive agents as offline
	//   - Redistribute work from dead agents
	//   - Display accurate agent status in UI
	//   - Trigger alerts for infrastructure issues
	//
	// Returns:
	//   - nil on success
	//   - error if communication fails
	ReportHealth(c context.Context) error
}
