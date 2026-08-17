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

package agent

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/compile"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/rpc/mocks"
)

// runStep drives one step's output through the runner's logger and returns
// every log entry the peer received.
func runStep(t *testing.T, phase rpc.WorkflowPhase, step *backend_types.Step, output string, secrets ...string) (*compileResults, []*rpc.LogEntry) {
	t.Helper()

	var entries []*rpc.LogEntry
	peer := mocks.NewMockPeer(t)
	peer.On("EnqueueLog", mock.Anything).Run(func(args mock.Arguments) {
		entry, ok := args.Get(0).(*rpc.LogEntry)
		require.True(t, ok)
		entries = append(entries, entry)
	})

	config := new(backend_types.Config)
	for _, secret := range secrets {
		config.Secrets = append(config.Secrets, &backend_types.Secret{Value: secret})
	}

	runner := Runner{client: peer}
	results := newCompileResults()
	logger := runner.createLogger(zerolog.Nop(), &rpc.Workflow{Phase: phase, Config: config}, results)

	require.NoError(t, logger(step, io.NopCloser(strings.NewReader(output))))

	return results, entries
}

func compileStep() *backend_types.Step {
	return &backend_types.Step{UUID: "step-1", Name: "generate", Type: backend_types.StepTypeCommands}
}

func TestLoggerCollectsFromACompileWorkflow(t *testing.T) {
	block, err := compile.EncodeResponse([]compile.Config{{Name: "build", Data: []byte("steps: {}\n")}})
	require.NoError(t, err)

	results, entries := runStep(t, rpc.WorkflowPhaseCompile, compileStep(), "generating\n"+block)

	result := results.report()
	require.NotNil(t, result)
	require.Len(t, result.Configs, 1)
	assert.Equal(t, "build", result.Configs[0].Name)

	// the block stays in the log, retyped, so it can be audited
	var types []int
	for _, entry := range entries {
		types = append(types, entry.Type)
	}
	assert.Equal(t, rpc.LogEntryStdout, types[0])
	assert.Contains(t, types, rpc.LogEntryCompileConfig)
	assert.Equal(t, []byte("generating"), entries[0].Data)

	// line numbering must stay continuous across the two writers
	for i, entry := range entries {
		assert.Equal(t, i, entry.Line)
	}
}

func TestLoggerIgnoresAnOrdinaryWorkflow(t *testing.T) {
	block, err := compile.EncodeResponse([]compile.Config{{Name: "build", Data: []byte("steps: {}\n")}})
	require.NoError(t, err)

	// A run workflow printing something that looks like a response must not be
	// able to inject configuration.
	results, entries := runStep(t, rpc.WorkflowPhaseRun, compileStep(), block)

	assert.Nil(t, results.report())
	for _, entry := range entries {
		assert.Equal(t, rpc.LogEntryStdout, entry.Type)
	}
}

func TestLoggerIgnoresTheCloneStep(t *testing.T) {
	block, err := compile.EncodeResponse([]compile.Config{{Name: "build", Data: []byte("steps: {}\n")}})
	require.NoError(t, err)

	// Only steps the user declared under compile: may emit a response, so a
	// malicious clone plugin cannot rewrite the pipeline.
	step := &backend_types.Step{UUID: "clone", Name: "clone", Type: backend_types.StepTypeClone}
	results, _ := runStep(t, rpc.WorkflowPhaseCompile, step, block)

	assert.Nil(t, results.report())
}

func TestLoggerReportsAMissingResponse(t *testing.T) {
	results, _ := runStep(t, rpc.WorkflowPhaseCompile, compileStep(), "I forgot to emit anything\n")

	result := results.report()
	require.NotNil(t, result)
	assert.Contains(t, result.Error, compile.ErrNoResponse.Error())
	assert.Empty(t, result.Configs)
}

func TestLoggerScansUpstreamOfSecretMasking(t *testing.T) {
	block, err := compile.EncodeResponse([]compile.Config{{Name: "build", Data: []byte("steps:\n  test:\n    image: golang\n")}})
	require.NoError(t, err)

	// Pick a short secret that occurs inside the encoded payload by chance.
	// That is not contrived: the masker replaces every secret longer than three
	// characters on every line, and base64's alphabet makes such a collision
	// likely in a large payload. Masking before the scan would corrupt it.
	payload := strings.Split(block, "\n")[1]
	secret := payload[8:12]

	results, entries := runStep(t, rpc.WorkflowPhaseCompile, compileStep(), block, secret)

	result := results.report()
	require.NotNil(t, result)
	require.Len(t, result.Configs, 1, "the authoritative copy must be unmasked")
	assert.Equal(t, "steps:\n  test:\n    image: golang\n", string(result.Configs[0].Data))

	// what reaches the log is still masked
	var logged strings.Builder
	for _, entry := range entries {
		logged.Write(entry.Data)
	}
	assert.NotContains(t, logged.String(), secret)
}

func TestCompileResultsReport(t *testing.T) {
	t.Run("no step scanned means not a compile workflow", func(t *testing.T) {
		assert.Nil(t, newCompileResults().report())
	})

	t.Run("steps fold in uuid order", func(t *testing.T) {
		results := newCompileResults()
		results.set("step-2", []compile.Config{{Name: "b"}}, nil)
		results.set("step-1", []compile.Config{{Name: "a"}}, nil)

		result := results.report()
		require.NotNil(t, result)
		assert.Equal(t, []string{"a", "b"}, []string{result.Configs[0].Name, result.Configs[1].Name},
			"the merge must not depend on which step finished first")
	})

	t.Run("a failing step outweighs a succeeding one", func(t *testing.T) {
		results := newCompileResults()
		results.set("step-1", []compile.Config{{Name: "a"}}, nil)
		results.set("step-2", nil, errors.New("boom"))

		result := results.report()
		require.NotNil(t, result)
		assert.Equal(t, "boom", result.Error)
		assert.Empty(t, result.Configs, "acting on half a response is worse than not acting")
	})
}
