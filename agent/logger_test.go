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

package agent

import (
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/rpc/mocks"
)

// A step that runs several times (restart signal) calls the logger once per
// attempt: the line numbers must continue instead of starting at 0 again.
func TestLoggerContinuesLineNumbersAcrossCalls(t *testing.T) {
	peer := mocks.NewMockPeer(t)
	var (
		mu      sync.Mutex
		entries []*rpc.LogEntry
	)
	peer.On("EnqueueLog", mock.Anything).Run(func(args mock.Arguments) {
		entry, ok := args.Get(0).(*rpc.LogEntry)
		assert.True(t, ok)
		mu.Lock()
		entries = append(entries, entry)
		mu.Unlock()
	}).Return()

	r := &Runner{client: peer}
	logger := r.createLogger(zerolog.Nop(), &rpc.Workflow{Config: &backend_types.Config{}})

	step := &backend_types.Step{UUID: "step-1"}
	other := &backend_types.Step{UUID: "step-2"}
	require.NoError(t, logger(step, io.NopCloser(strings.NewReader("a\nb\n"))))
	require.NoError(t, logger(other, io.NopCloser(strings.NewReader("x\n"))))
	require.NoError(t, logger(step, io.NopCloser(strings.NewReader("c\n"))))

	var stepLines, otherLines []int
	for _, e := range entries {
		if e.StepUUID == "step-1" {
			stepLines = append(stepLines, e.Line)
		} else {
			otherLines = append(otherLines, e.Line)
		}
	}
	assert.Equal(t, []int{0, 1, 2}, stepLines, "second call for the same step continues numbering")
	assert.Equal(t, []int{0}, otherLines, "other steps have their own numbering")
}
