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
	"net"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/cache"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/logging"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub/memory"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
	server_rpc "go.woodpecker-ci.org/woodpecker/v3/server/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/scheduler"
	"go.woodpecker-ci.org/woodpecker/v3/server/services"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/permissions"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

const (
	// TestAgentToken is the shared secret used between the in-process server
	// and agent. Hard-coded for tests — not a real secret.
	TestAgentToken = "test-agent-secret-for-integration-tests"

	// TestJWTSecret is used for signing gRPC auth JWTs.
	TestJWTSecret = "test-jwt-secret-for-integration-tests"

	// TestForgeType is the forge type the mock pretends to be.
	TestForgeType = model.ForgeTypeGitea
)

var configLock = sync.Mutex{}

// ServerEnv holds all the pieces of a running test server environment.
type ServerEnv struct {
	GRPCAddr string
	Store    store.Store
	Queue    queue.Queue
	Fixtures *Fixtures
	Forge    *forge_mocks.MockForge
	Manager  services.Manager
}

// StartServer wires up the full in-process server stack:
//   - in-memory sqlite store (fully migrated) with seeded fixtures
//   - in-memory queue, pubsub, and logging
//   - MockForge that serves the provided workflow files
//   - gRPC server on a random TCP port
//
// files must contain at least one entry. Single-workflow scenarios pass one
// file named ".woodpecker.yaml"; multi-workflow scenarios pass multiple files
// named ".woodpecker/foo.yaml" etc. The repo's Config path is set accordingly.
//
// All resources are cleaned up via t.Cleanup.
func StartServer(ctx context.Context, t *testing.T, files []*forge_types.FileMeta) *ServerEnv {
	t.Helper()
	configLock.Lock()
	defer configLock.Unlock()

	memStore := newStore(ctx, t)
	fixtures := seedFixtures(t, memStore)
	mockForge := newMockForge(t, files)

	mgr, err := newTestManager(memStore, mockForge)
	require.NoError(t, err, "create services manager")

	memQueue, err := queue.New(ctx, queue.Config{Backend: queue.TypeMemory})
	require.NoError(t, err, "create queue")

	// Save and restore server.Config around the test. server.Config is a
	// package-level global read by server/pipeline and server/rpc. Tests run
	// sequentially within a package, but we still need to clean up so the next
	// subtest starts from a known-zero state rather than the previous test's values.
	orig := server.Config
	t.Cleanup(func() {
		configLock.Lock()
		defer configLock.Unlock()
		server.Config = orig
	})

	server.Config.Services.Logs = logging.New()
	server.Config.Services.Scheduler = scheduler.NewScheduler(memQueue, memory.New())
	server.Config.Services.Membership = cache.NewMembershipService(memStore)
	server.Config.Services.Manager = mgr
	server.Config.Services.LogStore = memStore

	server.Config.Server.AgentToken = TestAgentToken
	server.Config.Server.Host = "http://localhost"
	server.Config.Server.JWTSecret = TestJWTSecret

	server.Config.Pipeline.DefaultClonePlugin = "docker.io/woodpeckerci/plugin-git:latest"
	server.Config.Pipeline.TrustedClonePlugins = []string{"docker.io/woodpeckerci/plugin-git:latest"}
	server.Config.Pipeline.DefaultApprovalMode = model.RequireApprovalNone
	server.Config.Pipeline.DefaultTimeout = 60
	server.Config.Pipeline.MaxTimeout = 60

	server.Config.Permissions.Open = true
	server.Config.Permissions.Admins = permissions.NewAdmins([]string{})
	server.Config.Permissions.Orgs = permissions.NewOrgs([]string{})
	server.Config.Permissions.OwnersAllowlist = permissions.NewOwnersAllowlist([]string{})

	grpcAddr := startGRPCServer(ctx, t, memStore)

	return &ServerEnv{
		GRPCAddr: grpcAddr,
		Store:    memStore,
		Queue:    memQueue,
		Fixtures: fixtures,
		Forge:    mockForge,
		Manager:  mgr,
	}
}

// newTestManager builds a services.Manager whose SetupForge always returns
// the provided MockForge, bypassing real forge instantiation.
func newTestManager(s store.Store, mockForge *forge_mocks.MockForge) (services.Manager, error) {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			// Config fetch tuning.
			&cli.DurationFlag{Name: "forge-timeout", Value: defaultTimeout},
			&cli.UintFlag{Name: "forge-retry", Value: defaultRetry},
			&cli.StringSliceFlag{Name: "environment"},
			// services.NewManager reads the forge type from a cli flag and
			// drives a type-switch in setupForgeService. We set the gitea flag
			// to satisfy that switch; the actual forge instance is overridden
			// below via the SetupForge hook, so the switch result is unused.
			&cli.BoolFlag{Name: string(TestForgeType), Value: true},
			&cli.StringFlag{Name: "forge-url", Value: "https://forge.example.test"},
		},
	}

	setupForge := services.SetupForge(func(*model.Forge) (forge.Forge, error) {
		return mockForge, nil
	})

	return services.NewManager(cmd, s, setupForge)
}

// startGRPCServer binds to a random TCP port and serves Woodpecker's gRPC
// services via the shared server_rpc.Serve helper. A fresh prometheus.Registry
// is passed so subtests don't collide on metric names.
//
// Shutdown is synchronous: t.Cleanup cancels the serve context (triggering
// GracefulStop inside Serve) and then blocks until Serve has returned.
// Without this wait, the next subtest can start while the previous server's
// goroutines are still live, which races on shared state like server.Config.
func startGRPCServer(ctx context.Context, t *testing.T, s store.Store) string {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "listen on random port for gRPC")
	addr := lis.Addr().String()

	serveCtx, cancel := context.WithCancelCause(ctx)
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		_ = server_rpc.Serve(serveCtx, server_rpc.ServeConfig{
			Listener:         lis,
			Store:            s,
			Scheduler:        server.Config.Services.Scheduler,
			Logger:           server.Config.Services.Logs,
			JWTSecret:        TestJWTSecret,
			AgentToken:       TestAgentToken,
			KeepaliveMinTime: shortTimeout,
			Registerer:       prometheus.NewRegistry(),
		})
	}()

	t.Cleanup(func() {
		cancel(nil)
		<-stopped
	})
	return addr
}
