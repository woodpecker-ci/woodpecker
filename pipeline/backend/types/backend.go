// Copyright 2023 Woodpecker Authors
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

// Package types defines the Backend interface and related types for
// executing Woodpecker CI workflows across different runtime environments.
package types

import (
	"context"
	"io"

	"github.com/urfave/cli/v3"
)

// Backend defines the mechanism for orchestrating workflows and their steps.
//
// A Backend instance is created once per agent and must handle multiple
// workflows concurrently, depending on the configured parallel workflow
// capacity. Each workflow may have multiple steps executing concurrently.
//
// Thread Safety and Isolation:
//
//   - Each workflow must have a unique taskUUID
//   - Backend implementations must use taskUUID to isolate workflow resources
//   - A single Backend instance must safely handle multiple concurrent workflows
//   - Workflow functions may be called concurrently for different workflows
//   - Step functions must be safe to call concurrently for different steps,
//     even across different workflows
//
// Intended execution flow:
//
//  1. Initialization (once per backend instance):
//     - Name() returns backend identifier
//     - IsAvailable() checks environment compatibility
//     - Flags() registers configuration options
//     - Load() initializes the backend instance
//
//  2. Workflow setup (once per workflow, may be called concurrently):
//     - SetupWorkflow() creates isolated environment for the workflow
//
//  3. Step execution (once per step, may run concurrently):
//     - StartStep() launches the step
//     - TailStep() streams logs (async, in background)
//     - WaitStep() blocks until completion
//     - DestroyStep() cleans up step resources
//
//  4. Workflow cleanup (once per workflow, may be called concurrently):
//     - DestroyWorkflow() removes workflow environment
type Backend interface {
	// Name returns the unique identifier of the backend implementation.
	// Examples: "docker", "kubernetes", "local", "dummy"
	Name() string

	// IsAvailable checks if the backend is available and can be used in the
	// current environment. For example, a Docker backend would check if the
	// Docker daemon is accessible.
	IsAvailable(ctx context.Context) bool

	// Flags returns the configuration flags specific to this backend.
	// Are used to configure backend-specific behavior
	// (e.g., Docker socket path, Kubernetes namespace).
	Flags() []cli.Flag

	// Load initializes the backend engine and returns metadata about its
	// capabilities and configuration.
	// This is called once after flags are parsed.
	// The backend must be ready to handle multiple concurrent workflows
	// after Load completes successfully.
	Load(ctx context.Context) (*BackendInfo, error)

	// SetupWorkflow prepares the execution environment for a new workflow.
	// This is called exactly once per workflow, before any steps are started.
	// The taskUUID uniquely identifies this workflow and must be used to
	// isolate this workflow's resources from other concurrent workflows.
	//
	// Implementations should:
	// - Create isolated workspaces, networks, or namespaces
	// - Initialize shared volumes or storage
	// - Ensure the setup doesn't interfere with other running workflows
	//
	// This function may be called concurrently for different workflows.
	// Implementations must be thread-safe and handle concurrent workflow setup.
	SetupWorkflow(ctx context.Context, conf *Config, taskUUID string, trusted TrustedConfiguration) error

	// StartStep set up and begins execution of a workflow step.
	// This may be called concurrently for multiple steps within the same
	// workflow, depending on the dependency graph.
	//
	// Implementations should:
	// - Start the step's container/process/pod
	// - Use taskUUID to associate the step with its workflow
	// - Ensure steps can run independently without blocking each other
	// - Handle different step types (commands, plugins, services, cache, clone)
	//
	// The step's UUID uniquely identifies it within the workflow.
	// This function must be thread-safe for concurrent calls.
	StartStep(ctx context.Context, step *Step, taskUUID string, trusted TrustedConfiguration) error

	// TailStep streams the step's logs back to the caller.
	// This is started in a background goroutine immediately after
	// StartStep, before WaitStep is called.
	//
	// The returned io.ReadCloser should:
	// - Stream logs as they are produced by the step
	// - Remain open until the step completes or is destroyed
	//
	// The reader will be closed by the caller when no longer needed, which
	// may be after WaitStep returns or during DestroyStep.
	// This function must be thread-safe for concurrent calls.
	TailStep(ctx context.Context, step *Step, taskUUID string) (io.ReadCloser, error)

	// WaitStep blocks until the step completes and returns its final state.
	// This is called after StartStep and TailStep while TailStep is
	// streaming logs in the background.
	//
	// Returns:
	// - State.ExitCode: The step's exit code (0 for success, non-zero for failure)
	// - State.Error: Any error that occurred during step execution
	// - State.Exited: Timestamp when the step completed
	//
	// The TailStep reader may be closed either when WaitStep completes or
	// during DestroyStep - implementations should handle both cases.
	// This function must be thread-safe for concurrent calls.
	WaitStep(ctx context.Context, step *Step, taskUUID string) (*State, error)

	// DestroyStep cleans up resources associated with a step.
	// This is called after WaitStep completes, or if the workflow is canceled.
	//
	// Implementations should:
	// - Stop the step if still running
	// - Clean up step-specific resources (containers, processes)
	// - Close any open log streams
	// - Not affect other steps in the same or other workflows
	//
	// Must be safe to call even if StartStep failed or the step was never started.
	// This function must be thread-safe for concurrent calls.
	DestroyStep(ctx context.Context, step *Step, taskUUID string) error

	// DestroyWorkflow cleans up all workflow-level resources.
	//
	// Implementations should:
	// - Destroy steps still running in the background (detached steps and services)
	// - Remove workflow-specific workspaces, networks, or namespaces
	// - Clean up shared volumes or storage
	// - Ensure complete cleanup so the taskUUID can be reused later
	// - Not affect other workflows that may be running concurrently
	//
	// Must be safe to call even if SetupWorkflow failed.
	// This function may be called concurrently for different workflows
	// and must be thread-safe.
	DestroyWorkflow(ctx context.Context, conf *Config, taskUUID string) error
}

// BackendInfo represents the reported information of a loaded backend.
type BackendInfo struct {
	Platform string
}
