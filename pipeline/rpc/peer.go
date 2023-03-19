package rpc

import (
	"context"

	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

type (
	// Filter defines filters for fetching items from the queue.
	Filter struct {
		Labels map[string]string `json:"labels"`
	}

	// State defines the pipeline state.
	State struct {
		Step     string `json:"step"`
		Exited   bool   `json:"exited"`
		ExitCode int    `json:"exit_code"`
		Started  int64  `json:"started"`
		Finished int64  `json:"finished"`
		Error    string `json:"error"`
	}

	// Pipeline defines the pipeline execution details.
	Pipeline struct {
		ID      string          `json:"id"`
		Config  *backend.Config `json:"config"`
		Timeout int64           `json:"timeout"`
	}

	// File defines a pipeline artifact.
	File struct {
		Name string            `json:"name"`
		Step string            `json:"step"`
		Mime string            `json:"mime"`
		Time int64             `json:"time"`
		Size int               `json:"size"`
		Data []byte            `json:"data"`
		Meta map[string]string `json:"meta"`
	}

	Version struct {
		GrpcVersion   int32  `json:"grpc_version,omitempty"`
		ServerVersion string `json:"server_version,omitempty"`
	}
)

// Peer defines a peer-to-peer connection.
type Peer interface {
	// Version returns the server- & grpc-version
	Version(c context.Context) (*Version, error)

	// Next returns the next pipeline in the queue.
	Next(c context.Context, f Filter) (*Pipeline, error)

	// Wait blocks until the pipeline is complete.
	Wait(c context.Context, id string) error

	// Init signals the pipeline is initialized.
	Init(c context.Context, id string, state State) error

	// Done signals the pipeline is complete.
	Done(c context.Context, id string, state State) error

	// Extend extends the pipeline deadline
	Extend(c context.Context, id string) error

	// Update updates the pipeline state.
	Update(c context.Context, id string, state State) error

	// Upload uploads the pipeline artifact.
	Upload(c context.Context, id string, file *File) error

	// Log writes the pipeline log entry.
	Log(c context.Context, id string, line *Line) error

	// RegisterAgent register our agent to the server
	RegisterAgent(ctx context.Context, platform, backend, version string, capacity int) (int64, error)

	// ReportHealth reports health status of the agent to the server
	ReportHealth(c context.Context) error
}
