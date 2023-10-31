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

// Engine defines a container orchestration backend and is used
// to create and manage container resources.
type Engine interface {
	// Name returns the name of the backend.
	Name() string

	// IsAvailable check if the backend is available.
	IsAvailable(ctx context.Context) bool

	// Load the backend engine.
	Load(ctx context.Context) error

	// SetupWorkflow the workflow environment.
	SetupWorkflow(ctx context.Context, conf *Config, taskUUID string) error

	// StartStep start the workflow step.
	StartStep(ctx context.Context, step *Step, taskUUID string) error

	// WaitStep for the workflow step to complete and returns
	// the completion results.
	WaitStep(ctx context.Context, step *Step, taskUUID string) (*State, error)

	// TailStep the workflow step logs.
	TailStep(ctx context.Context, step *Step, taskUUID string) (io.ReadCloser, error)

	// StopStep stop the workflow step.
	StopStep(ctx context.Context, step *Step, taskUUID string) error

	// DestroyWorkflow the workflow environment.
	DestroyWorkflow(ctx context.Context, conf *Config, taskUUID string) error
}
