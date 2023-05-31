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

	// Setup the pipeline environment.
	Setup(context.Context, *Config) error

	// Exec start the pipeline step.
	Exec(context.Context, *Step) error

	// Wait for the pipeline step to complete and returns
	// the completion results.
	Wait(context.Context, *Step) (*State, error)

	// Tail the pipeline step logs.
	Tail(context.Context, *Step) (io.ReadCloser, error)

	// Destroy the pipeline environment.
	Destroy(context.Context, *Config) error
}
