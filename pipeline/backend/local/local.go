package local

import (
	"context"
	"io"
	"os"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

type local struct {}

// make sure local implements Engine
var _ types.Engine = &local{}

// New returns a new local Engine.
func New() types.Engine {
	return &local{}
}

func (e *local) Name() string {
	return "local"
}

func (e *local) IsAvailable() bool {
	return true
}

func (e *local) Load() error {
	return nil
}

// Setup the pipeline environment.
func (e *local) Setup(context.Context, *types.Config) error {
	return nil
}

// Exec the pipeline step.
func (e *local) Exec(context.Context, *types.Step) error {
	return nil
}

// Wait for the pipeline step to complete and returns
// the completion results.
func (e *local) Wait(context.Context, *types.Step) (*types.State, error) {
	return nil, nil
}

// Tail the pipeline step logs.
func (e *local) Tail(context.Context, *types.Step) (io.ReadCloser, error) {
	return nil, nil
}

// Destroy the pipeline environment.
func (e *local) Destroy(context.Context, *types.Config) error {
	return nil
}
