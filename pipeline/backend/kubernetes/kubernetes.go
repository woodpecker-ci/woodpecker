package kubernetes

import (
	"context"
	"io"
	"os"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

type kube struct {
	namespace string
	endpoint  string
	token     string
}

// make sure kube implements Engine
var _ types.Engine = &kube{}

// New returns a new Kubernetes Engine.
func New(namespace, endpoint, token string) types.Engine {
	return &kube{
		namespace: namespace,
		endpoint:  endpoint,
		token:     token,
	}
}

func (e *kube) Name() string {
	return "kubernetes"
}

func (e *kube) IsAvailable() bool {
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	return len(host) > 0
}

func (e *kube) Load() error {
	return nil
}

// Setup the pipeline environment.
func (e *kube) Setup(context.Context, *types.Config) error {
	// POST /api/v1/namespaces
	return nil
}

// Exec the pipeline step.
func (e *kube) Exec(context.Context, *types.Step) error {
	// POST /api/v1/namespaces/{namespace}/pods
	return nil
}

// Wait for the pipeline step to complete and returns
// the completion results.
func (e *kube) Wait(context.Context, *types.Step) (*types.State, error) {
	// GET /api/v1/watch/namespaces/{namespace}/pods
	// GET /api/v1/watch/namespaces/{namespace}/pods/{name}
	return nil, nil
}

// Tail the pipeline step logs.
func (e *kube) Tail(context.Context, *types.Step) (io.ReadCloser, error) {
	// GET /api/v1/namespaces/{namespace}/pods/{name}/log
	return nil, nil
}

// Destroy the pipeline environment.
func (e *kube) Destroy(context.Context, *types.Config) error {
	// DELETE /api/v1/namespaces/{name}
	return nil
}
