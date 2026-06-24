// Copyright 2023 Woodpecker Authors
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
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthInterceptor is a client interceptor for authentication.
type AuthInterceptor struct {
	authClient *AuthClient

	// mu guards accessToken, which is read by every outgoing RPC (attachToken)
	// and written by both the background refresh goroutine and on-demand
	// RefreshToken calls.
	mu          sync.RWMutex
	accessToken string

	// refreshMu serializes re-authentication so that a burst of RPCs failing
	// with an expired token does not trigger a stampede of Auth calls.
	refreshMu sync.Mutex
}

// NewAuthInterceptor returns a new auth interceptor.
func NewAuthInterceptor(ctx context.Context, authClient *AuthClient, refreshDuration time.Duration) (*AuthInterceptor, error) {
	interceptor := &AuthInterceptor{
		authClient: authClient,
	}

	err := interceptor.scheduleRefreshToken(ctx, refreshDuration)
	if err != nil {
		return nil, err
	}

	return interceptor, nil
}

// Unary returns a client interceptor to authenticate unary RPC.
func (interceptor *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		return invoker(interceptor.attachToken(ctx), method, req, reply, cc, opts...)
	}
}

// Stream returns a client interceptor to authenticate stream RPC.
func (interceptor *AuthInterceptor) Stream() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		return streamer(interceptor.attachToken(ctx), desc, cc, method, opts...)
	}
}

func (interceptor *AuthInterceptor) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "token", interceptor.Token())
}

// Token returns the access token currently in use. It is safe for concurrent
// use.
func (interceptor *AuthInterceptor) Token() string {
	interceptor.mu.RLock()
	defer interceptor.mu.RUnlock()
	return interceptor.accessToken
}

func (interceptor *AuthInterceptor) setToken(token string) {
	interceptor.mu.Lock()
	defer interceptor.mu.Unlock()
	interceptor.accessToken = token
}

// RefreshToken forces a re-authentication with the server and atomically
// replaces the stored access token. It is used by the RPC client to recover
// from an expired access token without waiting for the background refresh
// timer.
//
// It is safe to call concurrently. Callers pass the token they last used as
// staleToken; if another goroutine already refreshed the token in the
// meantime, the redundant re-authentication is skipped. No token material is
// logged.
func (interceptor *AuthInterceptor) RefreshToken(ctx context.Context, staleToken string) error {
	interceptor.refreshMu.Lock()
	defer interceptor.refreshMu.Unlock()

	// Another goroutine already obtained a fresh token while we were waiting
	// for the lock; nothing to do.
	if staleToken != "" && interceptor.Token() != staleToken {
		return nil
	}

	return interceptor.refreshToken(ctx)
}

func (interceptor *AuthInterceptor) scheduleRefreshToken(ctx context.Context, refreshInterval time.Duration) error {
	err := interceptor.refreshToken(ctx)
	if err != nil {
		return err
	}

	go func() {
		wait := refreshInterval

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(wait):
				// Serialize with on-demand RefreshToken calls so the two
				// never race to authenticate at the same time.
				interceptor.refreshMu.Lock()
				err := interceptor.refreshToken(ctx)
				interceptor.refreshMu.Unlock()
				if err != nil {
					wait = time.Second
				} else {
					wait = refreshInterval
				}
			}
		}
	}()

	return nil
}

// refreshToken authenticates with the server and stores the new access token.
// Callers must hold refreshMu (or be the constructor, before the refresh
// goroutine is started).
func (interceptor *AuthInterceptor) refreshToken(ctx context.Context) error {
	accessToken, _, err := interceptor.authClient.Auth(ctx)
	if err != nil {
		return err
	}

	interceptor.setToken(accessToken)
	log.Trace().Msg("token refreshed")

	return nil
}
