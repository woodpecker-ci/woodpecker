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
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/agent/log"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/rpc/mocks"
)

func TestLineWriter(t *testing.T) {
	peer := mocks.NewMockPeer(t)
	peer.On("EnqueueLog", mock.Anything)

	secrets := []string{"world"}
	lw := log.NewLineWriter(peer, "e9ea76a5-44a1-4059-9c4a-6956c478b26d", secrets...)

	_, err := lw.Write([]byte("hello world\n"))
	assert.NoError(t, err)
	_, err = lw.Write([]byte("the previous line had no newline at the end"))
	assert.NoError(t, err)

	peer.AssertCalled(t, "EnqueueLog", &rpc.LogEntry{
		StepUUID: "e9ea76a5-44a1-4059-9c4a-6956c478b26d",
		Time:     0,
		Type:     rpc.LogEntryStdout,
		Line:     0,
		Data:     []byte("hello ********"),
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

func TestLineWriterWithType(t *testing.T) {
	peer := mocks.NewMockPeer(t)
	var entries []*rpc.LogEntry
	peer.On("EnqueueLog", mock.Anything).Run(func(args mock.Arguments) {
		entry, ok := args.Get(0).(*rpc.LogEntry)
		require.True(t, ok)
		entries = append(entries, entry)
	})

	lw := log.NewLineWriter(peer, "step", "world")
	typed := lw.WithType(rpc.LogEntryCompileConfig)

	_, err := lw.Write([]byte("hello world\n"))
	assert.NoError(t, err)
	_, err = typed.Write([]byte("tagged world\n"))
	assert.NoError(t, err)
	_, err = lw.Write([]byte("plain again\n"))
	assert.NoError(t, err)

	// numbering is shared, or a log view interleaving both types would show
	// duplicate line numbers
	assert.Equal(t, []int{0, 1, 2}, []int{entries[0].Line, entries[1].Line, entries[2].Line})
	assert.Equal(t, []int{rpc.LogEntryStdout, rpc.LogEntryCompileConfig, rpc.LogEntryStdout},
		[]int{entries[0].Type, entries[1].Type, entries[2].Type})

	// the typed writer masks like its parent
	assert.Equal(t, []byte("tagged ********"), entries[1].Data)
}
