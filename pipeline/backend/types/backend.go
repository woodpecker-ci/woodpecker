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

package types

import (
	"context"
	"io"
)

// Backend defines a container orchestration backend and is used
// to create and manage container resources.
type Backend interface {
	// Name returns the name of the backend.
	Name() string

	// IsAvailable check if the backend is available.
	IsAvailable(ctx context.Context) bool

	// Load loads the backend engine.
	Load(ctx context.Context) (*BackendInfo, error)

	// SetupWorkflow sets up the workflow environment.
	SetupWorkflow(ctx context.Context, conf *Config, taskUUID string) error

	// StartStep starts the workflow step.
	StartStep(ctx context.Context, step *Step, taskUUID string) error

	// WaitStep waits for the workflow step to complete and returns
	// the completion results.
	WaitStep(ctx context.Context, step *Step, taskUUID string) (*State, error)

	// TailStep tails the workflow step logs.
	TailStep(ctx context.Context, step *Step, taskUUID string) (io.ReadCloser, error)

	// DestroyStep destroys the workflow step.
	DestroyStep(ctx context.Context, step *Step, taskUUID string) error

	// DestroyWorkflow destroys the workflow environment.
	DestroyWorkflow(ctx context.Context, conf *Config, taskUUID string) error
}

// BackendInfo represents the reported information of a loaded engine
type BackendInfo struct {
	Platform string
}

// BackendOptions defines advanced options for specific backends
type BackendOptions struct {
	Kubernetes KubernetesBackendOptions `json:"kubernetes,omitempty"`
}
