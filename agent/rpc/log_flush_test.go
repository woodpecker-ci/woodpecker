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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/rpc/proto"
)

// logRecorder records what the batcher sent, so a test can assert that a flush
// really transmitted rather than merely returned.
type logRecorder struct {
	proto.WoodpeckerClient

	mu      sync.Mutex
	entries []*proto.LogEntry
}

func (r *logRecorder) Log(_ context.Context, in *proto.LogRequest, _ ...grpc.CallOption) (*proto.Empty, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, in.GetLogEntries()...)
	return new(proto.Empty), nil
}

func (r *logRecorder) sent() []*proto.LogEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]*proto.LogEntry(nil), r.entries...)
}

func newRecordingClient(t *testing.T) (*client, *logRecorder) {
	t.Helper()

	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })

	recorder := new(logRecorder)
	c := &client{
		client:             recorder,
		logEntryBufferSize: 100,
		logs:               make(chan *proto.LogEntry, 100),
		flush:              make(chan chan struct{}),
	}
	go c.processLogs(ctx)

	return c, recorder
}

func TestFlushLogsSendsQueuedEntries(t *testing.T) {
	c, recorder := newRecordingClient(t)

	// A batch this small is far below maxLogBatchSize and would otherwise sit
	// in the buffer until maxLogFlushPeriod elapsed, which is exactly the
	// window in which Done used to overtake it.
	for i := range 3 {
		c.EnqueueLog(&rpc.LogEntry{StepUUID: "step", Line: i, Data: []byte("line")})
	}

	require.NoError(t, c.FlushLogs(t.Context()))
	assert.Len(t, recorder.sent(), 3, "flush must transmit before it returns")
}

func TestFlushLogsWithoutQueuedEntries(t *testing.T) {
	c, recorder := newRecordingClient(t)

	require.NoError(t, c.FlushLogs(t.Context()))
	assert.Empty(t, recorder.sent(), "an empty flush must not send an empty batch")
}

func TestFlushLogsHonorsCanceledContext(t *testing.T) {
	recorder := new(logRecorder)
	// No processLogs goroutine: nothing will ever accept the flush request, so
	// the context is the only way out.
	c := &client{
		client: recorder,
		logs:   make(chan *proto.LogEntry, 1),
		flush:  make(chan chan struct{}),
	}

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()

	assert.ErrorIs(t, c.FlushLogs(ctx), context.DeadlineExceeded)
}
