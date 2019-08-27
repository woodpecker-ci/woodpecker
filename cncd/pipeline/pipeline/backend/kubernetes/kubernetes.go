package kubernetes

import (
	"context"
	"io"

	"github.com/laszlocph/woodpecker/cncd/pipeline/pipeline/backend"
)

type engine struct {
	namespace string
	endpoint  string
	token     string
}

// New returns a new Kubernetes Engine.
func New(namespace, endpoint, token string) backend.Engine {
	return &engine{
		namespace: namespace,
		endpoint:  endpoint,
		token:     token,
	}
}

// Setup the pipeline environment.
func (e *engine) Setup(context.Context, *backend.Config) error {
	// POST /api/v1/namespaces
	return nil
}

// Start the pipeline step.
func (e *engine) Exec(context.Context, *backend.Step) error {
	// POST /api/v1/namespaces/{namespace}/pods
	return nil
}

// DEPRECATED
// Kill the pipeline step.
func (e *engine) Kill(context.Context, *backend.Step) error {
	return nil
}

// Wait for the pipeline step to complete and returns
// the completion results.
func (e *engine) Wait(context.Context, *backend.Step) (*backend.State, error) {
	// GET /api/v1/watch/namespaces/{namespace}/pods
	// GET /api/v1/watch/namespaces/{namespace}/pods/{name}
	return nil, nil
}

// Tail the pipeline step logs.
func (e *engine) Tail(context.Context, *backend.Step) (io.ReadCloser, error) {
	// GET /api/v1/namespaces/{namespace}/pods/{name}/log
	return nil, nil
}

// Destroy the pipeline environment.
func (e *engine) Destroy(context.Context, *backend.Config) error {
	// DELETE /api/v1/namespaces/{name}
	return nil
}
