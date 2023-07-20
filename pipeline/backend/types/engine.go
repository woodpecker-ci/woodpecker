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
	Setup(context.Context, *Config) error

	// Exec start the workflow step.
	// TODO: rename to "StartStep" to make
	Exec(context.Context, *Step) error

	// Wait for the workflow step to complete and returns
	// the completion results.
	// TODO: rename to "WaitStep" to make
	Wait(context.Context, *Step) (*State, error)

	// Tail the workflow step logs.
	// TODO: rename to "TailStep" to make
	Tail(context.Context, *Step) (io.ReadCloser, error)

	// Destroy the workflow environment.
	// TODO: rename to "DestroyWorkflow" to make
	Destroy(context.Context, *Config) error
}
