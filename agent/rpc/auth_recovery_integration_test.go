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
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"go.woodpecker-ci.org/woodpecker/v3/rpc/proto"
)

// tokenRegistry records which access tokens the fake server currently accepts.
// The fake authorizer rejects everything else with codes.Unauthenticated,
// mirroring how the real server rejects an expired token (see
// server/rpc/authorizer_test.go for the JWT-level coverage of that rejection).
type tokenRegistry struct {
	mu    sync.Mutex
	valid map[string]bool
}

func newTokenRegistry() *tokenRegistry {
	return &tokenRegistry{valid: make(map[string]bool)}
}

func (r *tokenRegistry) add(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.valid[token] = true
}

func (r *tokenRegistry) isValid(token string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.valid[token]
}

// authorize mimics server/rpc.Authorizer: it lets the Auth endpoint through
// unauthenticated and rejects any other call whose token is not (or no longer)
// valid.
func (r *tokenRegistry) authorize(ctx context.Context, fullMethod string) error {
	if fullMethod == proto.WoodpeckerAuth_Auth_FullMethodName {
		return nil
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "metadata is not provided")
	}
	tokens := md.Get("token")
	if len(tokens) == 0 || !r.isValid(tokens[0]) {
		return status.Error(codes.Unauthenticated, "access token is invalid: token is expired")
	}
	return nil
}

func (r *tokenRegistry) unaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if err := r.authorize(ctx, info.FullMethod); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

// fakeAuthServer hands out access tokens. The first token it issues is never
// registered as valid, simulating the situation in issue #4144 where the agent
// ends up holding a dead token; every subsequent token is valid. This lets us
// assert that the agent re-authenticates and recovers on its own.
type fakeAuthServer struct {
	proto.UnimplementedWoodpeckerAuthServer

	registry *tokenRegistry
	agentID  int64

	mu    sync.Mutex
	calls int
}

func (s *fakeAuthServer) Auth(_ context.Context, _ *proto.AuthRequest) (*proto.AuthResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++

	token := fmt.Sprintf("token-%d", s.calls)
	if s.calls > 1 {
		// Only tokens issued after the first one are accepted by the server.
		s.registry.add(token)
	}

	return &proto.AuthResponse{AccessToken: token, AgentId: s.agentID}, nil
}

func (s *fakeAuthServer) authCalls() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

// fakeWoodpeckerServer counts how many ReportHealth calls reached it after
// passing the authorizer (i.e. with a valid token).
type fakeWoodpeckerServer struct {
	proto.UnimplementedWoodpeckerServer

	mu                 sync.Mutex
	authedHealthChecks int
}

func (s *fakeWoodpeckerServer) ReportHealth(_ context.Context, _ *proto.ReportHealthRequest) (*proto.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.authedHealthChecks++
	return &proto.Empty{}, nil
}

func (s *fakeWoodpeckerServer) healthChecks() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.authedHealthChecks
}

// testServer is an in-process gRPC server (over bufconn) that wires a fake
// authorizer in front of a fake Woodpecker service, plus a fake auth service.
type testServer struct {
	auth   *fakeAuthServer
	wp     *fakeWoodpeckerServer
	dialer func(context.Context, string) (net.Conn, error)
}

func startTestServer(t *testing.T) *testServer {
	t.Helper()

	registry := newTokenRegistry()
	ts := &testServer{
		auth: &fakeAuthServer{registry: registry, agentID: 42},
		wp:   &fakeWoodpeckerServer{},
	}

	lis := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(registry.unaryInterceptor))
	proto.RegisterWoodpeckerAuthServer(grpcServer, ts.auth)
	proto.RegisterWoodpeckerServer(grpcServer, ts.wp)

	go func() {
		_ = grpcServer.Serve(lis)
	}()
	t.Cleanup(grpcServer.Stop)

	ts.dialer = func(ctx context.Context, _ string) (net.Conn, error) {
		return lis.DialContext(ctx)
	}
	return ts
}

func (ts *testServer) dial(t *testing.T, opts ...grpc.DialOption) *grpc.ClientConn {
	t.Helper()
	base := []grpc.DialOption{
		grpc.WithContextDialer(ts.dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient("passthrough:///bufnet", append(base, opts...)...)
	require.NoError(t, err)
	return conn
}

// TestExpiredTokenRecovery is the regression test for issue #4144: an agent that
// starts out holding a rejected (expired) access token must re-authenticate and
// retry the failed RPC instead of staying stuck.
func TestExpiredTokenRecovery(t *testing.T) {
	ts := startTestServer(t)
	ctx := t.Context()

	authConn := ts.dial(t)
	defer authConn.Close()

	authClient := NewAuthGrpcClient(authConn, "agent-secret", 42)
	// Long refresh interval so the background timer never fires during the
	// test; recovery must come purely from the on-demand refresh path.
	interceptor, err := NewAuthInterceptor(ctx, authClient, time.Hour)
	require.NoError(t, err)

	// The interceptor authenticated once during construction and is now holding
	// the token the server will reject.
	require.Equal(t, 1, ts.auth.authCalls())
	expired := interceptor.Token()
	require.NotEmpty(t, expired)

	mainConn := ts.dial(
		t,
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	defer mainConn.Close()

	client := NewGrpcClient(
		ctx, mainConn,
		SetConnectionRetryTimeout(30*time.Second),
		SetAuthRefresher(interceptor),
	)

	// This RPC initially carries the rejected token. The client must
	// transparently re-authenticate and retry.
	err = client.ReportHealth(ctx)
	require.NoError(t, err, "ReportHealth should succeed after re-authentication")

	// The agent re-authenticated (a second Auth call happened) and replaced the
	// dead token, and the retried RPC reached the server.
	assert.GreaterOrEqual(t, ts.auth.authCalls(), 2, "expected a re-authentication")
	assert.NotEqual(t, expired, interceptor.Token(), "expired token should have been replaced")
	assert.GreaterOrEqual(t, ts.wp.healthChecks(), 1, "an authenticated ReportHealth should reach the server")
}

// TestRefreshTokenDedup proves that concurrent RefreshToken calls sharing the
// same stale token collapse into a single re-authentication, so a burst of
// RPCs failing at once does not stampede the auth server.
func TestRefreshTokenDedup(t *testing.T) {
	ts := startTestServer(t)
	ctx := t.Context()

	authConn := ts.dial(t)
	defer authConn.Close()

	authClient := NewAuthGrpcClient(authConn, "agent-secret", 42)
	interceptor, err := NewAuthInterceptor(ctx, authClient, time.Hour)
	require.NoError(t, err)

	stale := interceptor.Token()
	require.Equal(t, 1, ts.auth.authCalls())

	// Fire many concurrent refreshes that all observed the same stale token.
	// Exactly one of them should reach the server; the rest see the token has
	// already changed and skip.
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NoError(t, interceptor.RefreshToken(ctx, stale))
		}()
	}
	wg.Wait()

	assert.Equal(t, 2, ts.auth.authCalls(), "concurrent stale refreshes should collapse into one re-auth")
	assert.NotEqual(t, stale, interceptor.Token())
}
