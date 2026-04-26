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

package rpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"google.golang.org/grpc"
	grpc_credentials "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// DialConfig bundles everything Dial needs. Callers build this from a
// *cli.Command (production) or with literals (tests).
type DialConfig struct {
	ServerAddr       string
	AgentToken       string
	AgentID          int64
	Secure           bool
	SkipTLSVerify    bool
	KeepaliveTime    time.Duration
	KeepaliveTimeout time.Duration
	AuthRefreshEvery time.Duration
}

// AgentConn holds the two gRPC connections and the auth interceptor an agent
// needs. Callers are responsible for closing both connections and canceling
// the auth context passed to Dial when the agent shuts down.
type AgentConn struct {
	AuthConn        *grpc.ClientConn
	MainConn        *grpc.ClientConn
	AuthInterceptor *AuthInterceptor
}

// Close closes both connections. Safe to call even if one or both are nil.
func (c *AgentConn) Close() {
	if c.MainConn != nil {
		_ = c.MainConn.Close()
	}
	if c.AuthConn != nil {
		_ = c.AuthConn.Close()
	}
}

// Dial builds the auth gRPC connection, authenticates, then builds the
// authenticated main gRPC connection.
//
// The authCtx parameter governs the lifetime of the token-refresh goroutine
// inside the interceptor; callers typically want a context separate from
// request ctx so the interceptor survives long-running polls.
func Dial(authCtx context.Context, cfg DialConfig) (*AgentConn, error) {
	var transport grpc.DialOption
	if cfg.Secure {
		transport = grpc.WithTransportCredentials(grpc_credentials.NewTLS(
			&tls.Config{InsecureSkipVerify: cfg.SkipTLSVerify}, //nolint:gosec // user-opt-in via DialConfig.SkipTLSVerify
		))
	} else {
		transport = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	keepaliveOpts := grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:    cfg.KeepaliveTime,
		Timeout: cfg.KeepaliveTimeout,
	})

	authConn, err := grpc.NewClient(cfg.ServerAddr, transport, keepaliveOpts)
	if err != nil {
		return nil, fmt.Errorf("create auth gRPC connection: %w", err)
	}

	authClient := NewAuthGrpcClient(authConn, cfg.AgentToken, cfg.AgentID)
	authInterceptor, err := NewAuthInterceptor(authCtx, authClient, cfg.AuthRefreshEvery)
	if err != nil {
		_ = authConn.Close()
		return nil, fmt.Errorf("authenticate with server: %w", err)
	}

	mainConn, err := grpc.NewClient(
		cfg.ServerAddr, transport, keepaliveOpts,
		grpc.WithUnaryInterceptor(authInterceptor.Unary()),
		grpc.WithStreamInterceptor(authInterceptor.Stream()),
	)
	if err != nil {
		_ = authConn.Close()
		return nil, fmt.Errorf("create main gRPC connection: %w", err)
	}

	return &AgentConn{
		AuthConn:        authConn,
		MainConn:        mainConn,
		AuthInterceptor: authInterceptor,
	}, nil
}
