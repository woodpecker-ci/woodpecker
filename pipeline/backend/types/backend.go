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
// A Backend instance may be created multiple times per agent, depending on the
// configured parallel workflow capacity. Each instance handles one workflow at
// a time, but multiple steps within that workflow may execute concurrently.
//
// Thread Safety and Isolation:
//
//   - Each workflow must have a unique taskUUID
//   - Backend implementations must use taskUUID to isolate workflow resources
//   - Expect to have multiple Backend instances running on the same host simultaneously
//   - Workflow functions affect only one workflow
//   - Step functions must be safe to call concurrently for different steps within an workflow
//
// Intended execution flow:
//
//  1. Initialization (once per backend instance):
//     - Name() returns backend identifier
//     - IsAvailable() checks environment compatibility
//     - Flags() registers configuration options
//     - Load() initializes an backend instance
//
//  2. Workflow setup (once per workflow):
//     - SetupWorkflow() creates isolated environment
//
//  3. Step execution (once per step, may run concurrently):
//     - StartStep() launches the step
//     - TailStep() streams logs (async, in background)
//     - WaitStep() blocks until completion
//     - DestroyStep() cleans up step resources
//
//  4. Workflow cleanup (once per workflow):
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
	// This is called once per backend instance after flags are parsed.
	Load(ctx context.Context) (*BackendInfo, error)

	// SetupWorkflow prepares the execution environment for a new workflow.
	// This is called exactly once per workflow, before any steps are started.
	// The taskUUID uniquely identifies this workflow and must be used to
	// isolate this workflow's resources from other concurrent workflows on
	// the same backend instance or host.
	//
	// Implementations should:
	// - Create isolated workspaces, networks, or namespaces
	// - Initialize shared volumes or storage
	// - Ensure the setup doesn't interfere with other running workflows
	//
	// Note: Only one workflow is executed at a time per Backend instance,
	// but multiple Backend instances may run on the same host.
	SetupWorkflow(ctx context.Context, conf *Config, taskUUID string) error

	// StartStep set up and begins execution of a workflow step.
	// This may be called concurrently for multiple but unique steps within
	// the same workflow, depending on the dependency graph.
	//
	// Implementations should:
	// - Start the step's container/process/pod
	// - Use taskUUID to associate the step with its workflow
	// - Ensure steps can run independently without blocking each other
	// - Handle different step types (commands, plugins, services, cache, clone)
	//
	// The step's UUID uniquely identifies it within the workflow.
	StartStep(ctx context.Context, step *Step, taskUUID string) error

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
	WaitStep(ctx context.Context, step *Step, taskUUID string) (*State, error)

	// DestroyStep cleans up resources associated with a step.
	// This is called after WaitStep completes, or if the workflow is canceled.
	//
	// Implementations should:
	// - Stop the step if still running
	// - Clean up step-specific resources (containers, processes)
	// - Close any open log streams
	// - Not affect other steps in the same workflow
	//
	// Must be safe to call even if StartStep failed or the step was never started.
	DestroyStep(ctx context.Context, step *Step, taskUUID string) error

	// DestroyWorkflow cleans up all workflow-level resources.
	//
	// Implementations should:
	// - Destroy steps still running in the background (detached steps and services) 
	// - Remove workflow-specific workspaces, networks, or namespaces
	// - Clean up shared volumes or storage
	// - Ensure complete cleanup so the taskUUID can be reused later
	// - Not affect other workflows that may be running on the host
	//
	// Must be safe to call even if SetupWorkflow failed.
	DestroyWorkflow(ctx context.Context, conf *Config, taskUUID string) error
}

// BackendInfo represents the reported information of a loaded backend.
type BackendInfo struct {
	Platform string
}
