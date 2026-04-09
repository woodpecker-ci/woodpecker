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
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"go.woodpecker-ci.org/woodpecker/v3/rpc/proto"
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

	s := newStore(ctx, t)
	fixtures := seedFixtures(t, s, files)
	mockForge := newMockForge(t, files)

	mgr, err := newTestManager(s, mockForge)
	require.NoError(t, err, "create services manager")

	q, err := queue.New(ctx, queue.Config{Backend: queue.TypeMemory})
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
	server.Config.Services.Pubsub = memory.New()
	server.Config.Services.Queue = q
	server.Config.Services.Membership = cache.NewMembershipService(s)
	server.Config.Services.Manager = mgr
	server.Config.Services.LogStore = s // datastore implements log.Service

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

	grpcAddr := startGRPCServer(ctx, t, s)

	return &ServerEnv{
		GRPCAddr: grpcAddr,
		Store:    s,
		Queue:    q,
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
			// Extensions (all empty → disabled).
			&cli.StringFlag{Name: "extensions-allowed-hosts"},
			&cli.StringFlag{Name: "secret-extension-endpoint"},
			&cli.BoolFlag{Name: "secret-extension-netrc"},
			&cli.StringFlag{Name: "docker-config"},
			&cli.StringFlag{Name: "registry-extension-endpoint"},
			&cli.BoolFlag{Name: "registry-extension-netrc"},
			&cli.StringFlag{Name: "config-extension-endpoint"},
			&cli.BoolFlag{Name: "config-extension-netrc"},
			&cli.BoolFlag{Name: "config-extension-exclusive"},
			// Config fetch tuning.
			&cli.DurationFlag{Name: "forge-timeout", Value: defaultTimeout},
			&cli.UintFlag{Name: "forge-retry", Value: 3}, //nolint:mnd
			&cli.StringSliceFlag{Name: "environment"},
			// Forge flags — gitea=true satisfies setupForgeService's type switch.
			&cli.BoolFlag{Name: "gitea", Value: true},
			&cli.StringFlag{Name: "forge-url", Value: "https://forge.example.test"},
			&cli.StringFlag{Name: "forge-oauth-client"},
			&cli.StringFlag{Name: "forge-oauth-secret"},
			&cli.BoolFlag{Name: "forge-skip-verify"},
			&cli.StringFlag{Name: "forge-oauth-host"},
			// Other forge type flags (all false).
			&cli.StringFlag{Name: "addon-forge"},
			&cli.BoolFlag{Name: "github"},
			&cli.BoolFlag{Name: "github-merge-ref"},
			&cli.BoolFlag{Name: "github-public-only"},
			&cli.BoolFlag{Name: "gitlab"},
			&cli.BoolFlag{Name: "forgejo"},
			&cli.BoolFlag{Name: "bitbucket"},
			&cli.BoolFlag{Name: "bitbucket-dc"},
			&cli.StringFlag{Name: "bitbucket-dc-git-username"},
			&cli.StringFlag{Name: "bitbucket-dc-git-password"},
			&cli.BoolFlag{Name: "bitbucket-dc-oauth-enable-oauth2-scope-project-admin"},
		},
	}

	setupForge := services.SetupForge(func(_ *model.Forge) (forge.Forge, error) {
		return mockForge, nil
	})

	return services.NewManager(cmd, s, setupForge)
}

// startGRPCServer binds to a random TCP port, registers Woodpecker's gRPC
// services, and starts serving. Shutdown happens via t.Cleanup.
func startGRPCServer(ctx context.Context, t *testing.T, s store.Store) string {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "listen on random port for gRPC")
	addr := lis.Addr().String()

	jwtManager := server_rpc.NewJWTManager(TestJWTSecret)
	authorizer := server_rpc.NewAuthorizer(jwtManager)

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(authorizer.StreamInterceptor),
		grpc.UnaryInterceptor(authorizer.UnaryInterceptor),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime: shortTimeout,
		}),
	)

	proto.RegisterWoodpeckerServer(grpcServer, server_rpc.NewTestWoodpeckerServer(
		server.Config.Services.Queue,
		server.Config.Services.Logs,
		server.Config.Services.Pubsub,
		s,
		prometheus.NewRegistry(),
	))
	proto.RegisterWoodpeckerAuthServer(grpcServer, server_rpc.NewWoodpeckerAuthServer(
		jwtManager,
		TestAgentToken,
		s,
	))

	grpcCtx, grpcCancel := context.WithCancelCause(ctx)
	go func() {
		<-grpcCtx.Done()
		grpcServer.GracefulStop()
	}()
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			grpcCancel(err)
		}
	}()

	t.Cleanup(func() { grpcCancel(nil) })
	return addr
}
