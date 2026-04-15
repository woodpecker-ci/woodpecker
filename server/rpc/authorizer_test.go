// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func newAuthorizer(t *testing.T) *Authorizer {
	t.Helper()
	return NewAuthorizer(NewJWTManager("auth-test-secret"))
}

// validTokenForAgent generates a JWT that the authorizer will accept.
func validTokenForAgent(t *testing.T, agentID int64) string {
	t.Helper()
	token, err := NewJWTManager("auth-test-secret").Generate(agentID)
	require.NoError(t, err)
	return token
}

// ctxWithToken builds an incoming gRPC context carrying metadata["token"].
func ctxWithToken(ctx context.Context, token string) context.Context {
	return metadata.NewIncomingContext(ctx, metadata.Pairs("token", token))
}

func TestAuthorize(t *testing.T) {
	t.Parallel()

	t.Run("Auth endpoint bypasses JWT validation", func(t *testing.T) {
		t.Parallel()

		a := newAuthorizer(t)
		// Plain context with no metadata – would normally fail, but Auth is exempt.
		ctx, err := a.authorize(t.Context(), "/proto.WoodpeckerAuth/Auth")

		require.NoError(t, err)
		assert.NotNil(t, ctx)
	})

	t.Run("missing metadata returns Unauthenticated", func(t *testing.T) {
		t.Parallel()

		a := newAuthorizer(t)
		// A plain context has no gRPC incoming metadata.
		_, err := a.authorize(t.Context(), "/proto.WoodpeckerServer/Next")

		require.Error(t, err)
		s, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, s.Code())
		assert.Contains(t, s.Message(), "metadata is not provided")
	})

	t.Run("metadata present but token key absent returns Unauthenticated", func(t *testing.T) {
		t.Parallel()

		a := newAuthorizer(t)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("other-key", "value"))

		_, err := a.authorize(ctx, "/proto.WoodpeckerServer/Next")

		require.Error(t, err)
		s, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, s.Code())
		assert.Contains(t, s.Message(), "token is not provided")
	})

	t.Run("invalid (garbage) token returns Unauthenticated", func(t *testing.T) {
		t.Parallel()

		a := newAuthorizer(t)
		ctx := ctxWithToken(t.Context(), "this-is-not-a-jwt")

		_, err := a.authorize(ctx, "/proto.WoodpeckerServer/Next")

		require.Error(t, err)
		s, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, s.Code())
		assert.Contains(t, s.Message(), "access token is invalid")
	})

	t.Run("token signed with wrong secret returns Unauthenticated", func(t *testing.T) {
		t.Parallel()

		wrongManager := NewJWTManager("DIFFERENT-secret")
		token, err := wrongManager.Generate(55)
		require.NoError(t, err)

		a := newAuthorizer(t) // uses "auth-test-secret"
		ctx := ctxWithToken(t.Context(), token)

		_, err = a.authorize(ctx, "/proto.WoodpeckerServer/Next")

		require.Error(t, err)
		s, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, s.Code())
	})

	t.Run("valid token enriches context with agent_id metadata", func(t *testing.T) {
		t.Parallel()

		a := newAuthorizer(t)
		token := validTokenForAgent(t, 77)
		ctx := ctxWithToken(t.Context(), token)

		newCtx, err := a.authorize(ctx, "/proto.WoodpeckerServer/Next")

		require.NoError(t, err)

		md, ok := metadata.FromIncomingContext(newCtx)
		require.True(t, ok)
		agentIDs := md["agent_id"]
		require.Len(t, agentIDs, 1)
		assert.Equal(t, "77", agentIDs[0])
	})

	t.Run("valid token preserves existing metadata keys", func(t *testing.T) {
		t.Parallel()

		a := newAuthorizer(t)
		token := validTokenForAgent(t, 10)
		ctx := metadata.NewIncomingContext(t.Context(),
			metadata.Pairs("token", token, "hostname", "worker-1"),
		)

		newCtx, err := a.authorize(ctx, "/proto.WoodpeckerServer/Init")

		require.NoError(t, err)
		md, _ := metadata.FromIncomingContext(newCtx)
		assert.Equal(t, []string{"worker-1"}, md["hostname"])
		assert.Equal(t, []string{"10"}, md["agent_id"])
	})

	t.Run("empty token value in metadata slice returns Unauthenticated", func(t *testing.T) {
		t.Parallel()

		a := newAuthorizer(t)
		// Passing an empty string as the token value.
		ctx := ctxWithToken(t.Context(), "")

		_, err := a.authorize(ctx, "/proto.WoodpeckerServer/Next")

		require.Error(t, err)
		s, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, s.Code())
	})
}

func TestUnaryInterceptor(t *testing.T) {
	t.Parallel()

	t.Run("valid token calls handler with enriched context", func(t *testing.T) {
		t.Parallel()

		a := newAuthorizer(t)
		token := validTokenForAgent(t, 21)
		ctx := ctxWithToken(t.Context(), token)

		var capturedCtx context.Context
		handler := func(ctx context.Context, _ any) (any, error) {
			capturedCtx = ctx
			return "ok", nil
		}

		resp, err := a.UnaryInterceptor(ctx, nil, &grpc.UnaryServerInfo{
			FullMethod: "/proto.WoodpeckerServer/Next",
		}, handler)

		require.NoError(t, err)
		assert.Equal(t, "ok", resp)

		md, ok := metadata.FromIncomingContext(capturedCtx)
		require.True(t, ok)
		assert.Equal(t, []string{"21"}, md["agent_id"])
	})

	t.Run("invalid token does not call handler", func(t *testing.T) {
		t.Parallel()

		a := newAuthorizer(t)
		ctx := ctxWithToken(t.Context(), "bad-token")

		handlerCalled := false
		handler := func(_ context.Context, _ any) (any, error) {
			handlerCalled = true
			return nil, nil
		}

		_, err := a.UnaryInterceptor(ctx, nil, &grpc.UnaryServerInfo{
			FullMethod: "/proto.WoodpeckerServer/Next",
		}, handler)

		require.Error(t, err)
		assert.False(t, handlerCalled)
	})

	t.Run("Auth endpoint bypasses token check and calls handler", func(t *testing.T) {
		t.Parallel()

		a := newAuthorizer(t)
		// No token in context – fine because Auth is exempt.
		ctx := metadata.NewIncomingContext(t.Context(), metadata.MD{})

		handlerCalled := false
		handler := func(_ context.Context, _ any) (any, error) {
			handlerCalled = true
			return nil, nil
		}

		_, err := a.UnaryInterceptor(ctx, nil, &grpc.UnaryServerInfo{
			FullMethod: "/proto.WoodpeckerAuth/Auth",
		}, handler)

		require.NoError(t, err)
		assert.True(t, handlerCalled)
	})

	t.Run("handler error is propagated", func(t *testing.T) {
		t.Parallel()

		a := newAuthorizer(t)
		token := validTokenForAgent(t, 1)
		ctx := ctxWithToken(t.Context(), token)

		handler := func(_ context.Context, _ any) (any, error) {
			return nil, errors.New("handler boom")
		}

		_, err := a.UnaryInterceptor(ctx, nil, &grpc.UnaryServerInfo{
			FullMethod: "/proto.WoodpeckerServer/Next",
		}, handler)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "handler boom")
	})
}

// mockServerStream is a minimal grpc.ServerStream for testing.
type mockServerStream struct {
	ctx context.Context
}

func (m *mockServerStream) SetHeader(metadata.MD) error  { return nil }
func (m *mockServerStream) SendHeader(metadata.MD) error { return nil }
func (m *mockServerStream) SetTrailer(metadata.MD)       {}
func (m *mockServerStream) Context() context.Context     { return m.ctx }
func (m *mockServerStream) SendMsg(any) error            { return nil }
func (m *mockServerStream) RecvMsg(any) error            { return nil }

func TestStreamInterceptor(t *testing.T) {
	t.Parallel()

	t.Run("valid token calls handler with enriched stream context", func(t *testing.T) {
		t.Parallel()

		a := newAuthorizer(t)
		token := validTokenForAgent(t, 33)
		ctx := ctxWithToken(t.Context(), token)
		stream := &mockServerStream{ctx: ctx}

		var capturedStream grpc.ServerStream
		handler := func(_ any, s grpc.ServerStream) error {
			capturedStream = s
			return nil
		}

		err := a.StreamInterceptor(nil, stream, &grpc.StreamServerInfo{
			FullMethod: "/proto.WoodpeckerServer/Next",
		}, handler)

		require.NoError(t, err)

		md, ok := metadata.FromIncomingContext(capturedStream.Context())
		require.True(t, ok)
		assert.Equal(t, []string{"33"}, md["agent_id"])
	})

	t.Run("invalid token does not call handler", func(t *testing.T) {
		t.Parallel()

		a := newAuthorizer(t)
		ctx := ctxWithToken(t.Context(), "garbage")
		stream := &mockServerStream{ctx: ctx}

		handlerCalled := false
		handler := func(_ any, _ grpc.ServerStream) error {
			handlerCalled = true
			return nil
		}

		err := a.StreamInterceptor(nil, stream, &grpc.StreamServerInfo{
			FullMethod: "/proto.WoodpeckerServer/Next",
		}, handler)

		require.Error(t, err)
		assert.False(t, handlerCalled)
		s, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, s.Code())
	})

	t.Run("stream context wrapper SetContext and Context round-trip", func(t *testing.T) {
		t.Parallel()

		stream := &mockServerStream{ctx: t.Context()}
		wrapper := newStreamContextWrapper(stream)

		newCtx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("foo", "bar"))
		wrapper.SetContext(newCtx)

		md, ok := metadata.FromIncomingContext(wrapper.Context())
		require.True(t, ok)
		assert.Equal(t, []string{"bar"}, md["foo"])
	})
}
