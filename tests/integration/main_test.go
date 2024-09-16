// Copyright 2024 Woodpecker Authors
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
// +build test

package integration

import (
	"context"
	"encoding/base32"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"go.woodpecker-ci.org/woodpecker/v2/cmd/agent/core"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/dummy"
	backend_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc/proto"
	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/cache"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	woodpeckerGrpcServer "go.woodpecker-ci.org/woodpecker/v2/server/grpc"
	"go.woodpecker-ci.org/woodpecker/v2/server/logging"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/pubsub"
	"go.woodpecker-ci.org/woodpecker/v2/server/queue"
	"go.woodpecker-ci.org/woodpecker/v2/server/router"
	"go.woodpecker-ci.org/woodpecker/v2/server/router/middleware"
	"go.woodpecker-ci.org/woodpecker/v2/server/services"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/permissions"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/datastore"
	"go.woodpecker-ci.org/woodpecker/v2/shared/constant"
)

var (
	testStore        store.Store
	mockForge        forge.Forge
	adminUser        = "admin"
	globalAgentToken = "global-agentSecret"
	grpcPort         = ":9020"
	httpPort         = ":8020"
)

func TestMain(m *testing.M) {
	ctx, ctxCancel := context.WithCancelCause(context.Background())
	defer ctxCancel(nil)
	testStore = setupDatabase(ctx)

	// Create mock forge
	mockForge = &fakeForge{}

	// Start Woodpecker server
	serverCtx, serverCancel := context.WithCancelCause(ctx)
	go startServer(serverCtx, testStore, mockForge)
	time.Sleep(time.Second)

	// Start Woodpecker agent with dummy backend
	agentCtx, agentCancel := context.WithCancelCause(ctx)
	go startAgent(agentCtx, globalAgentToken)

	// Run tests
	exitCode := m.Run()

	// Cleanup
	serverCancel(nil)
	agentCancel(nil)

	os.Exit(exitCode)
}

func setupDatabase(ctx context.Context) store.Store {
	testStore, err := datastore.NewEngine(&store.Opts{
		Driver: "sqlite",
		Config: ":memory:",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize store")
	}
	if err != testStore.Migrate(ctx, true) {
		log.Fatal().Err(err).Msg("Failed to migrate store")
	}
	go func() {
		select {
		case <-ctx.Done():
			_ = testStore.Close()
			return
		}
	}()
	return testStore
}

func startServer(ctx context.Context, _store store.Store, _forge forge.Forge) {
	// Set up server configuration
	server.Config.Server.Port = httpPort
	server.Config.Server.Host = "http://localhost" + httpPort

	// Initialize services
	server.Config.Services.Pubsub = pubsub.New()
	server.Config.Services.Queue = queue.New(ctx)
	server.Config.Services.Logs = logging.New()
	server.Config.Services.Membership = cache.NewMembershipService(_store)
	server.Config.Services.LogStore = _store

	// Initialize service manager
	manager, err := services.NewManager(&cli.Command{Flags: []cli.Flag{
		&cli.StringFlag{Name: "forge-oauth-client", Value: "oauth-client"},
		&cli.StringFlag{Name: "forge-oauth-secret", Value: "oauth-secret"},
		&cli.StringFlag{Name: "forge-url", Value: "https://example.forge"},
		&cli.StringFlag{Name: "forge-oauth-host", Value: ""},
		&cli.BoolFlag{Name: "forge-skip-verify", Value: false},
		&cli.StringFlag{Name: "addon-forge", Value: ""},
		&cli.BoolFlag{Name: "github", Value: false},
		&cli.BoolFlag{Name: "gitlab", Value: true}, // we pretend to have connected to gitlab
		&cli.IntFlag{Name: "forge-retry", Value: 1},
		&cli.DurationFlag{Name: "forge-timeout", Value: 100 * time.Millisecond},
		&cli.StringFlag{Name: "config-service-endpoint", Value: ""},
		&cli.StringFlag{Name: "docker-config", Value: ""},
		&cli.StringSliceFlag{Name: "environment", Value: []string{}},
	}}, _store, func(*model.Forge) (forge.Forge, error) {
		return _forge, nil
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create new manager")
	}
	server.Config.Services.Manager = manager

	// Config
	server.Config.Pipeline.DefaultClonePlugin = constant.DefaultClonePlugin
	server.Config.Pipeline.TrustedClonePlugins = constant.TrustedClonePlugins
	server.Config.Pipeline.DefaultCancelPreviousPipelineEvents = []model.WebhookEvent{"push", "pull_request"}
	server.Config.Pipeline.DefaultTimeout = 60
	server.Config.Pipeline.MaxTimeout = 60
	server.Config.Server.JWTSecret = base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
	server.Config.Permissions.Admins = permissions.NewAdmins([]string{adminUser})
	server.Config.Server.AgentToken = globalAgentToken
	server.Config.Server.StatusContext = "ci/woodpecker"
	server.Config.Server.StatusContextFormat = "{{ .context }}/{{ .event }}/{{ .workflow }}{{if not (eq .axis_id 0)}}/{{.axis_id}}{{end}}"
	server.Config.Server.SessionExpires = time.Hour

	startGrpcServer(ctx, _store)

	// Set up router
	gin.SetMode(gin.TestMode)
	handler := router.Load(func(http.ResponseWriter, *http.Request) {},
		gin.Recovery(),
		middleware.Store(_store),
	)

	// Start server
	srv := &http.Server{
		Addr:    server.Config.Server.Port,
		Handler: handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	log.Info().Msg("Woodpecker server started")

	// Wait for context cancellation
	<-ctx.Done()

	// Shutdown server
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatal().Err(err).Msg("Server shutdown failed")
	}

	log.Info().Msg("Woodpecker server stopped")
}

func startGrpcServer(ctx context.Context, _store store.Store) {
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen on grpc-addr")
	}

	jwtManager := woodpeckerGrpcServer.NewJWTManager("jwtSecret")
	authorizer := woodpeckerGrpcServer.NewAuthorizer(jwtManager)
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(authorizer.StreamInterceptor),
		grpc.UnaryInterceptor(authorizer.UnaryInterceptor),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime: time.Nanosecond,
		}),
	)
	woodpeckerServer := woodpeckerGrpcServer.NewWoodpeckerServer(
		server.Config.Services.Queue,
		server.Config.Services.Logs,
		server.Config.Services.Pubsub,
		_store,
	)
	proto.RegisterWoodpeckerServer(grpcServer, woodpeckerServer)
	woodpeckerAuthServer := woodpeckerGrpcServer.NewWoodpeckerAuthServer(jwtManager, globalAgentToken, _store)
	proto.RegisterWoodpeckerAuthServer(grpcServer, woodpeckerAuthServer)

	go func() {
		<-ctx.Done()
		grpcServer.Stop()
	}()

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("failed to start grpc server")
		}
	}()
}

func startAgent(ctx context.Context, token string, filter ...string) {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "agent-config", Value: ""},
			&cli.StringFlag{Name: "hostname", Value: "dymmy-agent"},
			&cli.IntFlag{Name: "max-workflows", Value: 1},
			&cli.BoolFlag{Name: "healthcheck", Value: false},
			&cli.BoolFlag{Name: "grpc-secure", Value: false},
			&cli.StringFlag{Name: "server", Value: "localhost" + grpcPort},
			&cli.DurationFlag{Name: "grpc-keepalive-time", Value: 10 * time.Millisecond},
			&cli.DurationFlag{Name: "grpc-keepalive-timeout", Value: 10 * time.Second},
			&cli.StringFlag{Name: "grpc-token", Value: token},
			&cli.StringFlag{Name: "backend-engine", Value: "dummy"},

			&cli.StringSliceFlag{Name: "filter", Value: filter},
		},
	}
	core.Run(ctx, cmd, []backend_types.Backend{dummy.New()})
}
