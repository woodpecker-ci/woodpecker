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

	// Check if the backend is available.
	IsAvailable(context.Context) bool

	// Load the backend engine.
	Load(context.Context) error

	// Setup the workflow environment.
	// TODO: rename to "SetupWorkflow"
	// TODO: pass a task UUID
	Setup(context.Context, *Config) error

	// Exec start the workflow step.
	// TODO: rename to "StartStep" to make
	// TODO: pass a task UUID
	Exec(context.Context, *Step) error

	// Wait for the workflow step to complete and returns
	// the completion results.
	// TODO: rename to "WaitStep" to make
	// TODO: pass a task UUID
	Wait(context.Context, *Step) (*State, error)

	// Tail the workflow step logs.
	// TODO: rename to "TailStep" to make
	// TODO: pass a task UUID
	Tail(context.Context, *Step) (io.ReadCloser, error)

	// Destroy the workflow environment.
	// TODO: rename to "DestroyWorkflow" to make
	// TODO: pass a task UUID
	Destroy(context.Context, *Config) error
}
