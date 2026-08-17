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

package compile

import (
	"encoding/base64"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scanned is what a scanner produced for one step's output.
type scanned struct {
	logs      string
	blockLogs string
	configs   []Config
	err       error
}

// scan feeds output through a scanner in one write.
func scan(t *testing.T, output string, opts ...ScanOption) scanned {
	t.Helper()

	var log, block strings.Builder
	src, result := ScanWriter(&log, &block, opts...)

	// a rejected block fails the write, but the result still has to explain why
	_, _ = io.WriteString(src, output)

	out := scanned{logs: log.String(), blockLogs: block.String()}
	out.configs, out.err = result()

	return out
}

func TestScanWriterCollectsAResponse(t *testing.T) {
	block, err := EncodeResponse([]Config{
		{Name: "build", Data: []byte("steps:\n  test:\n    image: golang\n")},
		{Name: "deploy"},
	})
	require.NoError(t, err)

	out := scan(t, "generating\n"+block+"done\n")
	require.NoError(t, out.err)

	require.Len(t, out.configs, 2)
	assert.Equal(t, "build", out.configs[0].Name)
	assert.Equal(t, "steps:\n  test:\n    image: golang\n", string(out.configs[0].Data))
	assert.Equal(t, "deploy", out.configs[1].Name)
	assert.Empty(t, out.configs[1].Data, "an entry without data is a removal")

	assert.Equal(t, "generating\ndone\n", out.logs, "ordinary lines stay in the normal log")
	assert.Equal(t, block, out.blockLogs, "the block is kept verbatim, not swallowed")
}

func TestScanWriterCollectsSeveralBlocks(t *testing.T) {
	first, err := EncodeResponse([]Config{{Name: "a", Data: []byte("steps: {}\n")}})
	require.NoError(t, err)
	second, err := EncodeResponse([]Config{{Name: "b", Data: []byte("steps: {}\n")}})
	require.NoError(t, err)

	out := scan(t, first+"between\n"+second)
	require.NoError(t, out.err)

	require.Len(t, out.configs, 2)
	assert.Equal(t, "a", out.configs[0].Name)
	assert.Equal(t, "b", out.configs[1].Name)
}

func TestScanWriterEmptyConfigListIsNotNoResponse(t *testing.T) {
	block, err := EncodeResponse(nil)
	require.NoError(t, err)

	out := scan(t, block)
	require.NoError(t, out.err, `"proceed unchanged" is a valid answer`)
	assert.Empty(t, out.configs)
}

func TestScanWriterWithoutABlock(t *testing.T) {
	assert.ErrorIs(t, scan(t, "I forgot to emit anything\n").err, ErrNoResponse)
}

func TestScanWriterHandlesSplitWritesAndMissingFinalNewline(t *testing.T) {
	block, err := EncodeResponse([]Config{{Name: "build", Data: []byte("steps: {}\n")}})
	require.NoError(t, err)

	output := "noise\n" + strings.TrimSuffix(block, "\n")

	var log, blockLog strings.Builder
	src, result := ScanWriter(&log, &blockLog)

	// one byte at a time: markers must be recognized across write boundaries
	for _, b := range []byte(output) {
		_, err := src.Write([]byte{b})
		require.NoError(t, err)
	}

	configs, err := result()
	require.NoError(t, err, "a block may end without a trailing newline")
	require.Len(t, configs, 1)
	assert.Equal(t, "build", configs[0].Name)
}

func TestScanWriterAcceptsCarriageReturns(t *testing.T) {
	block, err := EncodeResponse([]Config{{Name: "build", Data: []byte("steps: {}\n")}})
	require.NoError(t, err)

	out := scan(t, strings.ReplaceAll(block, "\n", "\r\n"))
	require.NoError(t, out.err)
	assert.Len(t, out.configs, 1)
}

func TestScanWriterRejectsMalformedBlocks(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{"nested begin", BlockBegin + "\n" + BlockBegin + "\n" + BlockEnd + "\n"},
		{"end without begin", BlockEnd + "\n"},
		{"unterminated", BlockBegin + "\nZm9v\n"},
		{"not base64", BlockBegin + "\nnot valid base64 !!\n" + BlockEnd + "\n"},
		{"not json", BlockBegin + "\n" + base64.StdEncoding.EncodeToString([]byte("nope")) + "\n" + BlockEnd + "\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.ErrorIs(t, scan(t, test.output).err, ErrMalformedBlock)
		})
	}
}

func TestScanWriterRejectsAnUnknownVersion(t *testing.T) {
	payload := base64.StdEncoding.EncodeToString([]byte(`{"version":2,"configs":[]}`))

	assert.ErrorIs(t, scan(t, BlockBegin+"\n"+payload+"\n"+BlockEnd+"\n").err, ErrUnsupportedVersion)
}

func TestScanWriterCapsWhileReading(t *testing.T) {
	// The cap has to bite before the block is terminated, otherwise a step that
	// never stops printing exhausts the agent's memory before anything is
	// measured.
	var log strings.Builder
	src, result := ScanWriter(&log, nil, WithMaxPayloadSize(64))

	_, err := io.WriteString(src, BlockBegin+"\n")
	require.NoError(t, err)

	_, err = io.WriteString(src, strings.Repeat("A", 1024)+"\n")
	assert.ErrorIs(t, err, ErrMalformedBlock, "the oversized write itself must fail")

	_, err = result()
	assert.ErrorIs(t, err, ErrMalformedBlock, "and the recorded result must agree")
}

func TestScanWriterCapsTheDecodedPayload(t *testing.T) {
	block, err := EncodeResponse([]Config{{Name: "build", Data: []byte(strings.Repeat("x", 512))}})
	require.NoError(t, err)

	assert.ErrorIs(t, scan(t, block, WithMaxPayloadSize(256)).err, ErrMalformedBlock)
}

func TestScanWriterDiscardsWithoutWriters(t *testing.T) {
	block, err := EncodeResponse([]Config{{Name: "build", Data: []byte("steps: {}\n")}})
	require.NoError(t, err)

	src, result := ScanWriter(nil, nil)
	_, err = io.WriteString(src, "noise\n"+block)
	require.NoError(t, err)

	configs, err := result()
	require.NoError(t, err)
	assert.Len(t, configs, 1)
}

func TestEncodeResponseWraps(t *testing.T) {
	block, err := EncodeResponse([]Config{{Name: "build", Data: []byte(strings.Repeat("x", 1024))}})
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSuffix(block, "\n"), "\n")
	require.Greater(t, len(lines), 3, "a payload this size must wrap over several lines")
	assert.Equal(t, BlockBegin, lines[0])
	assert.Equal(t, BlockEnd, lines[len(lines)-1])

	for _, line := range lines[1 : len(lines)-1] {
		assert.LessOrEqual(t, len(line), BlockLineWidth)
	}
}
