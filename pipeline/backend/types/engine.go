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

	// DestroyWorkflow the workflow environment.
	DestroyWorkflow(ctx context.Context, conf *Config, taskUUID string) error
}
