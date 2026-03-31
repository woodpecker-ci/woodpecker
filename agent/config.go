// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package agent

import "time"

const (
	// DefaultReportHealthInterval is how often the agent pings the server with
	// a health report.
	DefaultReportHealthInterval = time.Second * 10

	// DefaultAuthInterceptorRefreshInterval is how often the auth token is
	// proactively refreshed.
	DefaultAuthInterceptorRefreshInterval = time.Minute * 30

	// DefaultShutdownTimeout is how long the agent waits for in-flight work to
	// finish after receiving a shutdown signal before forcefully terminating.
	DefaultShutdownTimeout = time.Second * 5
)

// Config carries all runtime configuration for the agent.
// It is populated by cmd/agent/core from CLI flags / environment variables
// and passed to Run, keeping the agent package free of CLI dependencies.
type Config struct {
	// Server is the address of the Woodpecker server gRPC endpoint.
	Server string

	// GRPCToken is the shared secret used for agent authentication.
	GRPCToken string

	// GRPCSecure enables TLS for the gRPC connection.
	GRPCSecure bool

	// GRPCVerify controls whether the server TLS certificate is verified.
	// Only relevant when GRPCSecure is true.
	GRPCVerify bool

	// KeepaliveTime is the duration after which the agent pings the server to
	// check whether the transport is still alive.
	KeepaliveTime time.Duration

	// KeepaliveTimeout is how long the agent waits for a keepalive response
	// before closing the connection.
	KeepaliveTimeout time.Duration

	// Hostname is the agent's hostname reported to the server and used as a
	// label for workflow routing.
	Hostname string

	// AgentID is the persisted agent identity. -1 means the agent has not been
	// registered yet (stateless or first run).
	AgentID int64

	// PersistAgentID is an optional callback invoked after successful registration
	// with the server-assigned agent ID. It should write the ID to durable storage
	// (e.g. the on-disk agent config file) so it survives restarts.
	//
	// When nil, or when it returns an error, the agent is treated as stateless and
	// will unregister itself from the server on shutdown.
	//
	// The implementation lives in cmd/agent/core (which owns AgentConfig /
	// writeAgentConfig) so that the agent package stays free of config-file I/O.
	PersistAgentID func(agentID int64) error

	// MaxWorkflows is the number of workflows the agent may run concurrently.
	MaxWorkflows int

	// BackendEngine is the name of the pipeline backend to use (e.g. "docker",
	// "kubernetes", "local", "auto-detect").
	BackendEngine string

	// CustomLabels are extra key=value labels that control which workflows this
	// agent accepts. They are merged on top of the automatically derived labels.
	CustomLabels map[string]string

	// HealthcheckAddr is the address on which the HTTP health endpoint listens.
	// An empty string disables the healthcheck server.
	HealthcheckAddr string

	// AuthInterceptorRefreshInterval controls how often the auth token is
	// proactively refreshed. Defaults to DefaultAuthInterceptorRefreshInterval.
	AuthInterceptorRefreshInterval time.Duration

	// ShutdownTimeout is how long to wait for graceful shutdown before
	// forcefully terminating. Defaults to DefaultShutdownTimeout.
	ShutdownTimeout time.Duration
}
