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
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/cenkalti/backoff/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fakeRefresher is a test double for the auth interceptor. Each RefreshToken
// call hands out a new, unique token value.
type fakeRefresher struct {
	mu         sync.Mutex
	token      string
	gen        int
	refreshN   int
	refreshErr error
}

func (f *fakeRefresher) Token() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.token
}

func (f *fakeRefresher) RefreshToken(_ context.Context, staleToken string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.refreshN++
	if f.refreshErr != nil {
		return f.refreshErr
	}
	// Mimic the real dedup: only rotate when the caller's token is still current.
	if staleToken == "" || staleToken == f.token {
		f.gen++
		f.token = fmt.Sprintf("token-%d", f.gen)
	}
	return nil
}

func (f *fakeRefresher) refreshCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.refreshN
}

func isPermanent(err error) bool {
	var perm *backoff.PermanentError
	return errors.As(err, &perm)
}

func TestWithAuthRefresh(t *testing.T) {
	unauth := status.Error(codes.Unauthenticated, "token is expired")

	t.Run("recovers after one refresh", func(t *testing.T) {
		fr := &fakeRefresher{token: "token-0"}
		c := &client{auth: fr}

		var calls int
		op := func() (int, error) {
			calls++
			if calls == 1 {
				return 0, unauth // first attempt: expired token
			}
			return 7, nil // after refresh: success
		}

		wrapped := withAuthRefresh(context.Background(), c, "test", op)

		// First invocation hits the expired token, triggers a refresh, and
		// returns a retryable error so backoff will call us again.
		_, err := wrapped()
		require.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
		assert.False(t, isPermanent(err))
		assert.Equal(t, 1, fr.refreshCalls())

		// Second invocation succeeds with the refreshed token.
		res, err := wrapped()
		require.NoError(t, err)
		assert.Equal(t, 7, res)
		assert.Equal(t, 1, fr.refreshCalls(), "should not refresh again on success")
	})

	t.Run("gives up after maxAuthRefreshes", func(t *testing.T) {
		fr := &fakeRefresher{token: "token-0"}
		c := &client{auth: fr}

		op := func() (int, error) { return 0, unauth } // never accepted

		wrapped := withAuthRefresh(context.Background(), c, "test", op)

		// The first maxAuthRefreshes failures stay retryable and each refreshes.
		for i := 0; i < maxAuthRefreshes; i++ {
			_, err := wrapped()
			require.Error(t, err)
			assert.False(t, isPermanent(err), "attempt %d should still be retryable", i)
		}
		assert.Equal(t, maxAuthRefreshes, fr.refreshCalls())

		// The next failure exhausts the budget and becomes permanent so the
		// RPC returns instead of looping forever on a dead token.
		_, err := wrapped()
		require.Error(t, err)
		assert.True(t, isPermanent(err), "should be permanent once refreshes are exhausted")
		assert.Equal(t, maxAuthRefreshes, fr.refreshCalls(), "no further refresh after giving up")
	})

	t.Run("keeps retrying when refresh itself fails", func(t *testing.T) {
		fr := &fakeRefresher{token: "token-0", refreshErr: errors.New("server down")}
		c := &client{auth: fr}

		op := func() (int, error) { return 0, unauth }
		wrapped := withAuthRefresh(context.Background(), c, "test", op)

		// A failing refresh must not make the error permanent before the budget
		// is exhausted — backoff should wait and try again.
		_, err := wrapped()
		require.Error(t, err)
		assert.False(t, isPermanent(err))
		assert.Equal(t, 1, fr.refreshCalls())
	})

	t.Run("passes non-auth errors through untouched", func(t *testing.T) {
		fr := &fakeRefresher{token: "token-0"}
		c := &client{auth: fr}

		other := status.Error(codes.Unavailable, "try later")
		op := func() (int, error) { return 0, other }
		wrapped := withAuthRefresh(context.Background(), c, "test", op)

		_, err := wrapped()
		assert.Equal(t, other, err)
		assert.Equal(t, 0, fr.refreshCalls(), "non-auth errors must not trigger a refresh")
	})

	t.Run("nil refresher makes unauthenticated permanent", func(t *testing.T) {
		c := &client{} // no auth refresher configured

		op := func() (int, error) { return 0, unauth }
		wrapped := withAuthRefresh(context.Background(), c, "test", op)

		_, err := wrapped()
		require.Error(t, err)
		assert.True(t, isPermanent(err))
	})
}

// TestAuthInterceptorTokenConcurrency exercises Token()/setToken under the race
// detector to prove access to the stored token is concurrency-safe.
func TestAuthInterceptorTokenConcurrency(t *testing.T) {
	interceptor := &AuthInterceptor{accessToken: "initial"}

	var wg sync.WaitGroup
	var writes int64

	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 1000 {
				interceptor.setToken(fmt.Sprintf("token-%d", i))
				atomic.AddInt64(&writes, 1)
			}
		}()
	}
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 1000 {
				_ = interceptor.Token()
				_ = interceptor.attachToken(context.Background())
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(8000), writes)
	assert.NotEmpty(t, interceptor.Token())
}
