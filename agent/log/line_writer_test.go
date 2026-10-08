// Copyright 2019 Woodpecker Authors
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

package log_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/agent/log"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/shared"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	rpc_mocks "go.woodpecker-ci.org/woodpecker/v3/rpc/mocks"
)

func TestLineWriter(t *testing.T) {
	peer := rpc_mocks.NewMockPeer(t)
	peer.On("EnqueueLog", mock.Anything)

	lw := log.NewLineWriter(peer, "e9ea76a5-44a1-4059-9c4a-6956c478b26d")

	_, err := lw.Write([]byte("hello world\n"))
	assert.NoError(t, err)
	_, err = lw.Write([]byte("the previous line had no newline at the end"))
	assert.NoError(t, err)

	peer.AssertCalled(t, "EnqueueLog", &rpc.LogEntry{
		StepUUID: "e9ea76a5-44a1-4059-9c4a-6956c478b26d",
		Time:     0,
		Type:     rpc.LogEntryStdout,
		Line:     0,
		Data:     []byte("hello world"),
	})

	peer.AssertCalled(t, "EnqueueLog", &rpc.LogEntry{
		StepUUID: "e9ea76a5-44a1-4059-9c4a-6956c478b26d",
		Time:     0,
		Type:     rpc.LogEntryStdout,
		Line:     1,
		Data:     []byte("the previous line had no newline at the end"),
	})

	peer.AssertExpectations(t)
}

// TestLineWriterWithSecretsReplaceWriter guards the agent contract: wrapping the
// line writer in shared.NewSecretsReplaceWriter masks secret values before they
// are enqueued, matching the previous built-in masking behavior.
func TestLineWriterWithSecretsReplaceWriter(t *testing.T) {
	peer := rpc_mocks.NewMockPeer(t)
	peer.On("EnqueueLog", mock.Anything)

	lw := shared.NewSecretsReplaceWriter(
		log.NewLineWriter(peer, "e9ea76a5-44a1-4059-9c4a-6956c478b26d"),
		[]string{"world"},
	)

	_, err := lw.Write([]byte("hello world\n"))
	assert.NoError(t, err)

	peer.AssertCalled(t, "EnqueueLog", &rpc.LogEntry{
		StepUUID: "e9ea76a5-44a1-4059-9c4a-6956c478b26d",
		Time:     0,
		Type:     rpc.LogEntryStdout,
		Line:     0,
		Data:     []byte("hello ********"),
	})

	peer.AssertExpectations(t)
}

// TestLineWriterDoesNotRetainInput guards the io.Writer contract: the caller
// may reuse its buffer after Write returns, while the entry stays queued.
func TestLineWriterDoesNotRetainInput(t *testing.T) {
	var got *rpc.LogEntry
	peer := rpc_mocks.NewMockPeer(t)
	peer.On("EnqueueLog", mock.Anything).Run(func(args mock.Arguments) {
		got, _ = args.Get(0).(*rpc.LogEntry)
	})

	lw := log.NewLineWriter(peer, "e9ea76a5-44a1-4059-9c4a-6956c478b26d")

	buf := []byte("hello world\n")
	_, err := lw.Write(buf)
	assert.NoError(t, err)
	copy(buf, "XXXXXXXXXXX")

	if assert.NotNil(t, got) {
		assert.Equal(t, "hello world", string(got.Data))
	}
}
