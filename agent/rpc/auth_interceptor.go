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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor is a client interceptor for authentication.
type AuthInterceptor struct {
	authClient  *AuthClient
	mu          sync.RWMutex
	accessToken string
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

// Unary returns a client interceptor to authenticate unary RPC. If the server
// rejects the attached token, it refreshes the token before returning the
// error so the caller's retry policy can repeat the call.
func (interceptor *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		callCtx, rejectedToken := interceptor.attachToken(ctx)
		err := invoker(callCtx, method, req, reply, cc, opts...)
		if status.Code(err) == codes.Unauthenticated {
			refreshErr := interceptor.refreshTokenAfterUnauthenticated(ctx, rejectedToken)
			if refreshErr != nil {
				log.Warn().Err(refreshErr).Msg("could not reauthenticate after the server rejected the gRPC token")
			}
		}
		return err
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
		callCtx, _ := interceptor.attachToken(ctx)
		return streamer(callCtx, desc, cc, method, opts...)
	}
}

func (interceptor *AuthInterceptor) attachToken(ctx context.Context) (context.Context, string) {
	interceptor.mu.RLock()
	accessToken := interceptor.accessToken
	interceptor.mu.RUnlock()

	return metadata.AppendToOutgoingContext(ctx, "token", accessToken), accessToken
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
				err := interceptor.refreshToken(ctx)
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

func (interceptor *AuthInterceptor) refreshToken(ctx context.Context) error {
	interceptor.mu.Lock()
	defer interceptor.mu.Unlock()

	return interceptor.refreshTokenLocked(ctx)
}

func (interceptor *AuthInterceptor) refreshTokenAfterUnauthenticated(ctx context.Context, rejectedToken string) error {
	interceptor.mu.Lock()
	defer interceptor.mu.Unlock()

	// Another rejected RPC or the refresh timer may already have replaced the
	// token while this call was in flight. Reuse that refresh when it did.
	if interceptor.accessToken != rejectedToken {
		return nil
	}

	return interceptor.refreshTokenLocked(ctx)
}

func (interceptor *AuthInterceptor) refreshTokenLocked(ctx context.Context) error {
	accessToken, _, err := interceptor.authClient.Auth(ctx)
	if err != nil {
		return err
	}

	interceptor.accessToken = accessToken
	log.Trace().Msg("token refreshed")

	return nil
}
