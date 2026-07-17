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
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go.woodpecker-ci.org/woodpecker/v3/rpc/proto"
)

type authClientFunc func(context.Context, *proto.AuthRequest, ...grpc.CallOption) (*proto.AuthResponse, error)

func (f authClientFunc) Auth(ctx context.Context, req *proto.AuthRequest, opts ...grpc.CallOption) (*proto.AuthResponse, error) {
	return f(ctx, req, opts...)
}

func testAuthInterceptor(token string, auth authClientFunc) *AuthInterceptor {
	return &AuthInterceptor{
		authClient:  &AuthClient{client: auth},
		accessToken: token,
	}
}

func TestAuthInterceptorAttachToken(t *testing.T) {
	t.Parallel()

	interceptor := &AuthInterceptor{accessToken: "token"}
	base := metadata.AppendToOutgoingContext(context.Background(), "extra", "value")
	ctx, token := interceptor.attachToken(base)

	md, ok := metadata.FromOutgoingContext(ctx)
	require.True(t, ok)
	assert.Equal(t, []string{"token"}, md.Get("token"))
	assert.Equal(t, []string{"value"}, md.Get("extra"))
	assert.Equal(t, "token", token)
}

func TestAuthInterceptorRefreshesRejectedTokenForCallerRetry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	wantDeadline, _ := ctx.Deadline()

	var authCalls int
	interceptor := testAuthInterceptor("old-token", func(ctx context.Context, _ *proto.AuthRequest, _ ...grpc.CallOption) (*proto.AuthResponse, error) {
		authCalls++
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		assert.Equal(t, wantDeadline, deadline)
		return &proto.AuthResponse{AccessToken: "new-token"}, nil
	})
	unauthenticatedErr := status.Error(codes.Unauthenticated, "expired token")

	var tokens []string
	invoker := func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		require.True(t, ok)
		tokens = append(tokens, md.Get("token")[0])
		if len(tokens) == 1 {
			return unauthenticatedErr
		}
		return nil
	}

	_, err := retryRPC(ctx, &client{connectionRetryTimeout: time.Second}, "test", func() (struct{}, error) {
		err := interceptor.Unary()(ctx, "/proto.Woodpecker/Next", nil, nil, nil, invoker)
		return struct{}{}, classifyRPCErr(ctx, err)
	})

	require.NoError(t, err)
	assert.Equal(t, 1, authCalls)
	assert.Equal(t, []string{"old-token", "new-token"}, tokens)
}

func TestAuthInterceptorPreservesErrors(t *testing.T) {
	t.Run("non-authentication error", func(t *testing.T) {
		var authCalls int
		interceptor := testAuthInterceptor("token", func(context.Context, *proto.AuthRequest, ...grpc.CallOption) (*proto.AuthResponse, error) {
			authCalls++
			return &proto.AuthResponse{AccessToken: "new-token"}, nil
		})
		permissionErr := status.Error(codes.PermissionDenied, "denied")

		err := interceptor.Unary()(
			context.Background(), "/proto.Woodpecker/Next", nil, nil, nil,
			func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
				return permissionErr
			},
		)

		assert.Equal(t, permissionErr, err)
		assert.Zero(t, authCalls)
	})

	t.Run("failed reauthentication", func(t *testing.T) {
		authErr := errors.New("authentication unavailable")
		interceptor := testAuthInterceptor("old-token", func(context.Context, *proto.AuthRequest, ...grpc.CallOption) (*proto.AuthResponse, error) {
			return nil, authErr
		})
		unauthenticatedErr := status.Error(codes.Unauthenticated, "expired token")

		err := interceptor.Unary()(
			context.Background(), "/proto.Woodpecker/Next", nil, nil, nil,
			func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
				return unauthenticatedErr
			},
		)

		assert.Equal(t, unauthenticatedErr, err)
	})
}

func TestAuthInterceptorCoalescesConcurrentRefresh(t *testing.T) {
	var authCalls atomic.Int32
	interceptor := testAuthInterceptor("old-token", func(context.Context, *proto.AuthRequest, ...grpc.CallOption) (*proto.AuthResponse, error) {
		authCalls.Add(1)
		return &proto.AuthResponse{AccessToken: "new-token"}, nil
	})

	const callers = 2
	start := make(chan struct{})
	errs := make(chan error, callers)
	var wg sync.WaitGroup
	for range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- interceptor.refreshTokenAfterUnauthenticated(context.Background(), "old-token")
		}()
	}
	close(start)
	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
	_, token := interceptor.attachToken(context.Background())
	assert.Equal(t, "new-token", token)
	assert.Equal(t, int32(1), authCalls.Load())
}
