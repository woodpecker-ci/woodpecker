// Copyright 2026 Woodpecker Authors
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

//go:build test

package setup

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	"go.woodpecker-ci.org/woodpecker/v3/agent"
	agent_rpc "go.woodpecker-ci.org/woodpecker/v3/agent/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/dummy"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/version"
)

const (
	AgentMaxWorkflows     = 4
	agentAuthRefreshEvery = 30 * time.Minute
)

// AgentEnv holds the running state of one in-process test agent.
// Use AgentID to assert which agent picked up a workflow.
type AgentEnv struct {
	// AgentID is the server-assigned ID after registration.
	// Valid only after WaitForAgentRegistered returns.
	AgentID int64

	// name is used for logging and as the hostname label.
	name string

	// requestedOrgID is applied to the DB record by WaitForAgentRegistered
	// so the server's GetServerLabels returns the right org-id filter.
	// model.IDNotSet (-1) means global (default).
	requestOrgID int64
}

// AgentOption configures an agent before it registers with the server.
type AgentOption func(*agentConfig)

type agentConfig struct {
	// hostname is sent to the server as the agent's hostname metadata and label.
	hostname string

	// customLabels are merged into the agent's filter labels.
	// They are matched against task Labels set in pipeline YAML (labels: key: value).
	customLabels map[string]string

	// orgID pins the agent to a specific organization (-1 = global).
	// Org agents score higher than global agents for tasks in the same org,
	// so they are always preferred by the queue when available.
	orgID int64
}

// WithHostname sets the agent's hostname label (default: "test-agent").
func WithHostname(name string) AgentOption {
	return func(c *agentConfig) { c.hostname = name }
}

// WithCustomLabels merges extra labels into the agent's filter set.
// Use this to test label-based task routing, e.g.:
//
//	setup.StartAgent(ctx, t, addr, setup.WithCustomLabels(map[string]string{"gpu": "true"}))
//
// The pipeline YAML must set a matching label:
//
//	labels:
//	  gpu: "true"
func WithCustomLabels(labels map[string]string) AgentOption {
	return func(c *agentConfig) {
		for k, v := range labels {
			c.customLabels[k] = v
		}
	}
}

// WithOrgID restricts the agent to a specific organization. Org agents score
// 10× higher than global agents (score 1) for tasks from the same org, so the
// queue always prefers them when both are available. Pass model.IDNotSet (-1)
// for a global agent (the default).
func WithOrgID(id int64) AgentOption {
	return func(c *agentConfig) { c.orgID = id }
}

// StartAgent connects an in-process agent using the dummy backend to the gRPC
// server at grpcAddr and returns an *AgentEnv whose AgentID is populated once
// the agent has registered. Pass AgentOption values to configure labels, hostname,
// or org-scoping; multiple agents can be started in the same test.
func StartAgent(t *testing.T, grpcAddr string, opts ...AgentOption) *AgentEnv {
	t.Helper()

	cfg := &agentConfig{
		hostname:     "test-agent",
		customLabels: make(map[string]string),
		orgID:        model.IDNotSet, // global by default
	}
	for _, o := range opts {
		o(cfg)
	}

	env := &AgentEnv{name: cfg.hostname}

	transport := grpc.WithTransportCredentials(insecure.NewCredentials())
	keepaliveOpts := grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:    defaultTimeout,
		Timeout: shortTimeout,
	})

	agentCtx, agentCancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { agentCancel(nil) })

	authConn, err := grpc.NewClient(grpcAddr, transport, keepaliveOpts)
	if err != nil {
		t.Fatalf("StartAgent(%s): create auth gRPC connection: %v", cfg.hostname, err)
	}
	t.Cleanup(func() { authConn.Close() })

	authClient := agent_rpc.NewAuthGrpcClient(authConn, TestAgentToken, -1)
	authInterceptor, err := agent_rpc.NewAuthInterceptor(agentCtx, authClient, agentAuthRefreshEvery)
	if err != nil {
		t.Fatalf("StartAgent(%s): authenticate with server: %v", cfg.hostname, err)
	}

	conn, err := grpc.NewClient(
		grpcAddr,
		transport,
		keepaliveOpts,
		grpc.WithUnaryInterceptor(authInterceptor.Unary()),
		grpc.WithStreamInterceptor(authInterceptor.Stream()),
	)
	if err != nil {
		t.Fatalf("StartAgent(%s): create main gRPC connection: %v", cfg.hostname, err)
	}
	t.Cleanup(func() { conn.Close() })

	client := agent_rpc.NewGrpcClient(agentCtx, conn)

	grpcCtx := metadata.NewOutgoingContext(agentCtx, metadata.Pairs("hostname", cfg.hostname))

	backend := dummy.New()
	if !backend.IsAvailable(agentCtx) {
		t.Fatalf("StartAgent(%s): dummy backend is not available", cfg.hostname)
	}
	engInfo, err := backend.Load(agentCtx)
	if err != nil {
		t.Fatalf("StartAgent(%s): load dummy backend: %v", cfg.hostname, err)
	}

	env.AgentID, err = client.RegisterAgent(grpcCtx, rpc.AgentInfo{
		Version:      version.String(),
		Backend:      backend.Name(),
		Platform:     engInfo.Platform,
		Capacity:     AgentMaxWorkflows,
		CustomLabels: cfg.customLabels,
	})
	require.NoErrorf(t, err, "StartAgent(%s): register with server: %v", cfg.hostname, err)

	// If a non-global org is requested, update the agent's OrgID in the DB so
	// the server's GetServerLabels returns the right org-id filter (score 10).
	if cfg.orgID != model.IDNotSet {
		// The server stores agents; we patch via the store after registration.
		// This is done in WaitForAgentRegistered which the caller must invoke.
		// We stash the requested orgID so the wait helper can apply it.
		env.requestOrgID = cfg.orgID
	}

	t.Cleanup(func() {
		if err := client.UnregisterAgent(grpcCtx); err != nil {
			log.Warn().Err(err).Str("hostname", cfg.hostname).Msg("test agent: unregister failed (expected during teardown)")
		}
	})

	// Build the filter labels the agent advertises to the queue.
	// org-id is handled server-side via GetServerLabels; we only set
	// the labels the agent explicitly provides (platform, backend, repo wildcard,
	// and any custom labels).
	filter := rpc.Filter{
		Labels: map[string]string{
			"hostname": cfg.hostname,
			"platform": engInfo.Platform,
			"backend":  backend.Name(),
			"repo":     "*",
		},
	}
	for k, v := range cfg.customLabels {
		filter.Labels[k] = v
	}

	counter := &agent.State{
		Polling:  AgentMaxWorkflows,
		Metadata: make(map[string]agent.Info),
	}

	for i := range AgentMaxWorkflows {
		go func(slot int) {
			runner := agent.NewRunner(client, filter, cfg.hostname, counter, backend)
			log.Debug().Int("slot", slot).Str("hostname", cfg.hostname).Msg("test agent: runner started")
			for {
				if agentCtx.Err() != nil {
					return
				}
				if err := runner.Run(agentCtx); err != nil {
					if agentCtx.Err() != nil {
						return
					}
					log.Error().Err(err).Int("slot", slot).Str("hostname", cfg.hostname).Msg("test agent: runner error, retrying")
					select {
					case <-agentCtx.Done():
						return
					case <-time.After(500 * time.Millisecond):
					}
				}
			}
		}(i)
	}

	return env
}
