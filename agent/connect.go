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

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	grpc_credentials "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	agent_rpc "go.woodpecker-ci.org/woodpecker/v3/agent/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
)

// grpcTransport returns the appropriate gRPC dial option for the given TLS config.
func grpcTransport(secure, skipVerify bool) grpc.DialOption {
	if secure {
		log.Trace().Msg("use ssl for grpc")
		return grpc.WithTransportCredentials(grpc_credentials.NewTLS(&tls.Config{InsecureSkipVerify: skipVerify})) //nolint:gosec
	}
	return grpc.WithTransportCredentials(insecure.NewCredentials())
}

// GRPCConnections holds both gRPC client connections used by the agent.
type GRPCConnections struct {
	// AuthConn is used exclusively for the auth interceptor handshake.
	AuthConn *grpc.ClientConn
	// Conn is the main orchestration connection carrying interceptors.
	Conn *grpc.ClientConn
}

// Close closes both connections, logging any errors.
func (c *GRPCConnections) Close() {
	if err := c.AuthConn.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close auth gRPC connection")
	}
	if err := c.Conn.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close gRPC connection")
	}
}

// ConnectGRPC establishes both gRPC connections and wires the auth interceptor.
// The caller is responsible for calling GRPCConnections.Close() when done.
func ConnectGRPC(ctx context.Context, cfg Config) (*GRPCConnections, *agent_rpc.AuthInterceptor, error) {
	transport := grpcTransport(cfg.GRPCSecure, !cfg.GRPCVerify)

	keepaliveParams := keepalive.ClientParameters{
		Time:    cfg.KeepaliveTime,
		Timeout: cfg.KeepaliveTimeout,
	}

	// Auth connection — used only for token refresh.
	authConn, err := grpc.NewClient(
		cfg.Server,
		transport,
		grpc.WithKeepaliveParams(keepaliveParams),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create gRPC auth channel: %w", err)
	}

	authClient := agent_rpc.NewAuthGrpcClient(authConn, cfg.GRPCToken, cfg.AgentID)
	authInterceptor, err := agent_rpc.NewAuthInterceptor(ctx, authClient, cfg.AuthInterceptorRefreshInterval)
	if err != nil {
		authConn.Close()
		return nil, nil, fmt.Errorf("agent could not authenticate: %w", err)
	}

	// Main orchestration connection with auth interceptors attached.
	conn, err := grpc.NewClient(
		cfg.Server,
		transport,
		grpc.WithKeepaliveParams(keepaliveParams),
		grpc.WithUnaryInterceptor(authInterceptor.Unary()),
		grpc.WithStreamInterceptor(authInterceptor.Stream()),
	)
	if err != nil {
		authConn.Close()
		return nil, nil, fmt.Errorf("could not create gRPC orchestration channel: %w", err)
	}

	return &GRPCConnections{AuthConn: authConn, Conn: conn}, authInterceptor, nil
}

// CheckGRPCVersion verifies that the server speaks the same gRPC protocol
// version as this agent. Returns an error if they are incompatible.
func CheckGRPCVersion(ctx context.Context, client rpc.Peer) error {
	serverVersion, err := client.Version(ctx)
	if err != nil {
		return fmt.Errorf("could not get gRPC server version: %w", err)
	}
	if serverVersion.GrpcVersion != agent_rpc.ClientGrpcVersion {
		return fmt.Errorf("%w: server %s reports gRPC version %d but agent requires %d",
			errors.New("gRPC version mismatch"),
			serverVersion.ServerVersion,
			serverVersion.GrpcVersion,
			agent_rpc.ClientGrpcVersion,
		)
	}
	return nil
}
