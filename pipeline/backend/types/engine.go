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
	IsAvailable(context.Context) bool

	// Load the backend engine.
	Load(context.Context) error

	// SetupWorkflow the workflow environment.
	// TODO: pass a task UUID
	SetupWorkflow(context.Context, *Config) error

	// StartStep start the workflow step.
	// TODO: pass a task UUID
	StartStep(context.Context, *Step) error

	// WaitStep for the workflow step to complete and returns
	// the completion results.
	// TODO: pass a task UUID
	WaitStep(context.Context, *Step) (*State, error)

	// TailStep the workflow step logs.
	// TODO: pass a task UUID
	TailStep(context.Context, *Step) (io.ReadCloser, error)

	// DestroyWorkflow the workflow environment.
	// TODO: pass a task UUID
	DestroyWorkflow(context.Context, *Config) error
}
