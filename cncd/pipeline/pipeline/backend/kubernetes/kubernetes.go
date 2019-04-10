package kubernetes

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"time"

	"github.com/laszlocph/drone-oss-08/cncd/pipeline/pipeline/backend"
	"k8s.io/client-go/kubernetes"

	// To authenticate to GCP K8s clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
)

type engine struct {
	logs       *bytes.Buffer
	kubeClient kubernetes.Interface
}

// New returns a new Kubernetes Engine.
func New() (backend.Engine, error) {
	var kubeClient kubernetes.Interface
	_, err := rest.InClusterConfig()
	if err != nil {
		kubeClient, err = getClientOutOfCluster()
	} else {
		kubeClient, err = getClient()
	}

	if err != nil {
		return nil, err
	}

	return &engine{
		logs:       new(bytes.Buffer),
		kubeClient: kubeClient,
	}, nil
}

// Setup the pipeline environment.
func (e *engine) Setup(context.Context, *backend.Config) error {
	// POST /api/v1/namespaces
	e.logs.WriteString("Setting up Kubernetes primitives\n")
	return nil
}

// Start the pipeline step.
func (e *engine) Exec(context.Context, *backend.Step) error {
	// POST /api/v1/namespaces/{namespace}/pods
	e.logs.WriteString("Execing\n")
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

	time.Sleep(2 * time.Second)

	return &backend.State{
		Exited:    true,
		ExitCode:  0,
		OOMKilled: false,
	}, nil
}

// Tail the pipeline step logs.
func (e *engine) Tail(context.Context, *backend.Step) (io.ReadCloser, error) {
	// GET /api/v1/namespaces/{namespace}/pods/{name}/log

	rc := ioutil.NopCloser(bytes.NewReader(e.logs.Bytes()))
	e.logs.Reset()
	return rc, nil
}

// Destroy the pipeline environment.
func (e *engine) Destroy(context.Context, *backend.Config) error {
	// DELETE /api/v1/namespaces/{name}
	return nil
}
