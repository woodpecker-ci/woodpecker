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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	wp_agent "go.woodpecker-ci.org/woodpecker/v3/agent"
	agent_rpc "go.woodpecker-ci.org/woodpecker/v3/agent/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/dummy"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/version"
)

const (
	agentMaxWorkflows     = 4
	agentAuthRefreshEvery = 30 * time.Minute
)

// StartAgent connects an in-process agent using the dummy backend to the
// gRPC server at grpcAddr. The agent runs agentMaxWorkflows concurrent
// polling loops and stops when ctx is cancelled (wired via t.Cleanup).
func StartAgent(ctx context.Context, t *testing.T, grpcAddr string) {
	t.Helper()

	transport := grpc.WithTransportCredentials(insecure.NewCredentials())
	keepaliveOpts := grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:    30 * time.Second,
		Timeout: 10 * time.Second,
	})

	// Separate auth connection: used only to exchange the agent token for a JWT.
	// Runs on a background context so it outlives the test context slightly
	// during cleanup ordering.
	authCtx, authCancel := context.WithCancelCause(context.Background())
	t.Cleanup(func() { authCancel(nil) })

	authConn, err := grpc.NewClient(grpcAddr, transport, keepaliveOpts)
	if err != nil {
		t.Fatalf("StartAgent: create auth gRPC connection: %v", err)
	}
	t.Cleanup(func() { authConn.Close() })

	authClient := agent_rpc.NewAuthGrpcClient(authConn, TestAgentToken, -1)
	authInterceptor, err := agent_rpc.NewAuthInterceptor(authCtx, authClient, agentAuthRefreshEvery)
	if err != nil {
		t.Fatalf("StartAgent: authenticate with server: %v", err)
	}

	// Main orchestration connection: carries all Next/Update/Done/Log calls.
	conn, err := grpc.NewClient(
		grpcAddr,
		transport,
		keepaliveOpts,
		grpc.WithUnaryInterceptor(authInterceptor.Unary()),
		grpc.WithStreamInterceptor(authInterceptor.Stream()),
	)
	if err != nil {
		t.Fatalf("StartAgent: create main gRPC connection: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	client := agent_rpc.NewGrpcClient(ctx, conn)

	// Attach hostname metadata — the server uses this for logging/assignment.
	const hostname = "test-agent"
	grpcCtx := metadata.NewOutgoingContext(authCtx, metadata.Pairs("hostname", hostname))

	// Verify gRPC protocol version compatibility.
	serverVersion, err := client.Version(grpcCtx)
	if err != nil {
		t.Fatalf("StartAgent: get server version: %v", err)
	}
	if serverVersion.GrpcVersion != agent_rpc.ClientGrpcVersion {
		t.Fatalf("StartAgent: gRPC version mismatch: server=%d client=%d",
			serverVersion.GrpcVersion, agent_rpc.ClientGrpcVersion)
	}

	// Load the dummy backend (available only when compiled with -tags test).
	backend := dummy.New()
	if !backend.IsAvailable(ctx) {
		t.Fatal("StartAgent: dummy backend is not available")
	}
	engInfo, err := backend.Load(ctx)
	if err != nil {
		t.Fatalf("StartAgent: load dummy backend: %v", err)
	}

	// Register with the server.
	agentID, err := client.RegisterAgent(grpcCtx, rpc.AgentInfo{
		Version:  version.String(),
		Backend:  backend.Name(),
		Platform: engInfo.Platform,
		Capacity: agentMaxWorkflows,
	})
	if err != nil {
		t.Fatalf("StartAgent: register with server: %v", err)
	}
	log.Debug().Int64("agent_id", agentID).Msg("test agent registered")

	t.Cleanup(func() {
		if err := client.UnregisterAgent(grpcCtx); err != nil {
			log.Warn().Err(err).Msg("test agent: unregister failed (expected during teardown)")
		}
	})

	filter := rpc.Filter{
		Labels: map[string]string{
			"hostname": hostname,
			"platform": engInfo.Platform,
			"backend":  backend.Name(),
			"repo":     "*",
		},
	}

	counter := &wp_agent.State{
		Polling:  agentMaxWorkflows,
		Metadata: make(map[string]wp_agent.Info),
	}

	// Start one polling goroutine per workflow slot.
	for i := range agentMaxWorkflows {
		go func(slot int) {
			runner := wp_agent.NewRunner(client, filter, hostname, counter, backend)
			log.Debug().Int("slot", slot).Msg("test agent: runner started")
			for {
				if ctx.Err() != nil {
					return
				}
				if err := runner.Run(ctx); err != nil {
					if ctx.Err() != nil {
						return
					}
					log.Error().Err(err).Int("slot", slot).Msg("test agent: runner error, retrying")
					select {
					case <-ctx.Done():
						return
					case <-time.After(500 * time.Millisecond):
					}
				}
			}
		}(i)
	}
}
