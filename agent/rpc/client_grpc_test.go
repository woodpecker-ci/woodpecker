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
	"testing"
	"testing/synctest"
	"time"

	"github.com/cenkalti/backoff/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/rpc/proto"
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
	classified := classifyRPCErr(t.Context(), err)

	assert.Equal(t, codes.Unauthenticated, status.Code(classified))
	assert.False(t, errors.Is(classified, backoff.ErrPermanent))
}

func TestRetryRPCUnauthenticatedHonorsFiniteTimeout(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
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

	ctx, cancel := context.WithCancelCause(t.Context())
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

type logRecorderClient struct {
	proto.WoodpeckerClient

	mu      sync.Mutex
	batches [][]*proto.LogEntry
}

func (l *logRecorderClient) Log(_ context.Context, in *proto.LogRequest, _ ...grpc.CallOption) (*proto.Empty, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.batches = append(l.batches, in.GetLogEntries())
	return &proto.Empty{}, nil
}

// sent returns the number of Log calls and the total number of entries sent.
func (l *logRecorderClient) sent() (batches, entries int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, b := range l.batches {
		entries += len(b)
	}
	return len(l.batches), entries
}

func TestProcessLogsFlushesContinuousStream(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(t.Context())
		defer cancel(nil)

		recorder := &logRecorderClient{}
		c := &client{
			client: recorder,
			logs:   make(chan *proto.LogEntry, 10),
		}
		go c.processLogs(ctx)

		// Emit a small log line more often than maxLogFlushPeriod, so the stream
		// never goes quiet long enough for an idle timeout to fire.
		const interval = maxLogFlushPeriod / 4
		const total = 20 // 5 * maxLogFlushPeriod of continuous output
		for i := range total {
			c.EnqueueLog(&rpc.LogEntry{StepUUID: "step", Line: i, Data: []byte("line")})
			time.Sleep(interval)
			synctest.Wait()
		}

		// While output is still streaming, logs must have been flushed
		// periodically instead of being held back until the stream goes quiet.
		batches, entries := recorder.sent()
		assert.GreaterOrEqual(t, batches, 4)
		assert.Greater(t, entries, total-4)

		// Once the stream is idle everything must be delivered.
		time.Sleep(maxLogFlushPeriod)
		synctest.Wait()
		_, entries = recorder.sent()
		assert.Equal(t, total, entries)

		cancel(nil)
		synctest.Wait()
	})
}
