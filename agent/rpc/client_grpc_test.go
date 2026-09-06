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
	"testing"
	"time"

	"github.com/cenkalti/backoff/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSetConnectionRetryTimeout(t *testing.T) {
	tc := []struct {
		name    string
		timeout time.Duration
	}{
		{"finite", 5 * time.Minute},
		{"zero means infinite", 0},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			cl := &client{}
			SetConnectionRetryTimeout(c.timeout)(cl)
			assert.Equal(t, c.timeout, cl.connectionRetryTimeout)
		})
	}
}

func TestIsConnected(t *testing.T) {
	cl := &client{conn: newTestConn(t)}
	defer cl.conn.Close()

	t.Run("idle connection reports connected", func(t *testing.T) {
		assert.True(t, cl.IsConnected())
	})

	t.Run("closed connection reports not connected", func(t *testing.T) {
		assert.NoError(t, cl.conn.Close())
		assert.False(t, cl.IsConnected())
	})
}

func TestClassifyRPCErrUnauthenticatedIsRetryable(t *testing.T) {
	t.Parallel()

	err := status.Error(codes.Unauthenticated, "expired token")
	classified := classifyRPCErr(context.Background(), err)

	assert.Equal(t, codes.Unauthenticated, status.Code(classified))
	assert.False(t, errors.Is(classified, backoff.ErrPermanent))
}

func TestRetryRPCUnauthenticatedHonorsFiniteTimeout(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	c := &client{connectionRetryTimeout: time.Nanosecond}
	var attempts int

	_, err := retryRPC(ctx, c, "test", func() (struct{}, error) {
		attempts++
		return struct{}{}, classifyRPCErr(ctx, status.Error(codes.Unauthenticated, "expired token"))
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, backoff.ErrMaxElapsedTime)
	assert.Equal(t, 1, attempts)
}

func TestRetryRPCUnauthenticatedRetriesUntilContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	c := &client{connectionRetryTimeout: 0}
	var attempts int

	_, err := retryRPC(ctx, c, "test", func() (struct{}, error) {
		attempts++
		if attempts == 3 {
			cancel(nil)
		}
		return struct{}{}, classifyRPCErr(ctx, status.Error(codes.Unauthenticated, "expired token"))
	})

	require.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

// fakeClock drives failureWindowBackOff without sleeping.
type fakeClock struct{ t time.Time }

func (f *fakeClock) now() time.Time          { return f.t }
func (f *fakeClock) advance(d time.Duration) { f.t = f.t.Add(d) }
func newFakeWindow(window time.Duration) (*failureWindowBackOff, *fakeClock) {
	clock := &fakeClock{t: time.Unix(1_700_000_000, 0)}
	b := &failureWindowBackOff{BackOff: backoff.NewConstantBackOff(time.Millisecond), window: window, now: clock.now}
	return b, clock
}

func TestFailureWindowBackOff(t *testing.T) {
	t.Parallel()

	t.Run("delegates until the window elapses since the first failure", func(t *testing.T) {
		t.Parallel()

		b, clock := newFakeWindow(time.Minute)
		b.attemptFailed(clock.now())
		assert.Equal(t, time.Millisecond, b.NextBackOff())
		clock.advance(59 * time.Second)
		assert.Equal(t, time.Millisecond, b.NextBackOff())
		clock.advance(2 * time.Second)
		assert.Equal(t, backoff.Stop, b.NextBackOff())
	})

	t.Run("reset starts over", func(t *testing.T) {
		t.Parallel()

		b, clock := newFakeWindow(time.Minute)
		b.attemptFailed(clock.now())
		clock.advance(2 * time.Minute)
		assert.Equal(t, backoff.Stop, b.NextBackOff())
		b.Reset()
		assert.Equal(t, time.Millisecond, b.NextBackOff())
	})

	t.Run("an attempt that outlived the window starts a new streak", func(t *testing.T) {
		t.Parallel()

		b, clock := newFakeWindow(time.Minute)
		b.attemptFailed(clock.now()) // quick failure: streak begins
		clock.advance(3 * time.Hour) // the next attempt reconnected and long-polled
		b.attemptFailed(clock.now().Add(-3 * time.Hour))
		assert.Equal(t, time.Millisecond, b.NextBackOff(), "failure of a long-lived attempt must be retried")

		clock.advance(30 * time.Second)
		b.attemptFailed(clock.now().Add(-time.Second)) // quick failure inside the new streak
		assert.Equal(t, time.Millisecond, b.NextBackOff())
		clock.advance(31 * time.Second)
		b.attemptFailed(clock.now().Add(-time.Second))
		assert.Equal(t, backoff.Stop, b.NextBackOff(), "new streak is bounded by the window too")
	})

	t.Run("zero window never stops", func(t *testing.T) {
		t.Parallel()

		b, clock := newFakeWindow(0)
		b.attemptFailed(clock.now())
		clock.advance(24 * time.Hour)
		assert.Equal(t, time.Millisecond, b.NextBackOff())
	})
}

func TestRetryRPCWindowStartsAtFirstFailure(t *testing.T) {
	t.Parallel()

	// A long poll (Next/Wait) blocks longer than the whole retry window before
	// it fails for the first time. The failure must still be retried.
	ctx := context.Background()
	c := &client{connectionRetryTimeout: 200 * time.Millisecond}
	var attempts int

	_, err := retryRPC(ctx, c, "test", func() (struct{}, error) {
		attempts++
		if attempts == 1 {
			time.Sleep(2 * c.connectionRetryTimeout)
			return struct{}{}, classifyRPCErr(ctx, status.Error(codes.Unavailable, "server restarting"))
		}
		return struct{}{}, nil
	})

	require.NoError(t, err)
	assert.Equal(t, 2, attempts)
}

func TestRetryRPCLongAttemptStartsNewFailureStreak(t *testing.T) {
	t.Parallel()

	// Within one retry loop: a quick failure starts a streak, the next attempt
	// reconnects and blocks for longer than the whole window (a long poll)
	// before it fails again. That failure must start a new streak and be
	// retried instead of exhausting the old one.
	ctx := context.Background()
	c := &client{connectionRetryTimeout: 200 * time.Millisecond}
	var attempts int

	_, err := retryRPC(ctx, c, "test", func() (struct{}, error) {
		attempts++
		switch attempts {
		case 1:
			return struct{}{}, classifyRPCErr(ctx, status.Error(codes.Unavailable, "server restarting"))
		case 2:
			time.Sleep(2 * c.connectionRetryTimeout)
			return struct{}{}, classifyRPCErr(ctx, status.Error(codes.Unavailable, "server restarting again"))
		default:
			return struct{}{}, nil
		}
	})

	require.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestRetryRPCGivesUpAfterWindow(t *testing.T) {
	t.Parallel()

	tc := []struct {
		name               string
		err                error
		classify           bool
		wantConnectionLost bool
	}{
		// the guard error is returned by the operations as-is
		{"not connected guard", errNotConnected, false, true},
		{"transport unavailable", status.Error(codes.Unavailable, "connection refused"), true, false},
		{"unauthenticated", status.Error(codes.Unauthenticated, "expired token"), true, false},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			c := &client{connectionRetryTimeout: 200 * time.Millisecond}
			var attempts int
			started := time.Now()

			_, err := retryRPC(ctx, c, "test", func() (struct{}, error) {
				attempts++
				if tt.classify {
					return struct{}{}, classifyRPCErr(ctx, tt.err)
				}
				return struct{}{}, tt.err
			})

			require.Error(t, err)
			assert.Greater(t, attempts, 1, "must retry before giving up")
			assert.Less(t, time.Since(started), 3*c.connectionRetryTimeout, "must give up close to the window")
			if tt.wantConnectionLost {
				assert.ErrorIs(t, err, ErrConnectionLost)
			} else {
				assert.ErrorIs(t, err, backoff.ErrMaxElapsedTime)
				assert.NotErrorIs(t, err, ErrConnectionLost)
			}
		})
	}
}
