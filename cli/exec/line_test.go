// Copyright 2024 Woodpecker Authors
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

package exec

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildPrefixSingleWorkflow(t *testing.T) {
	assert.Equal(t, "[build] ", buildPrefix("", "build"))
}

func TestBuildPrefixMultiWorkflow(t *testing.T) {
	assert.Equal(t, "[test/unit] ", buildPrefix("test", "unit"))
}

func TestBuildPrefixTruncatesLongBody(t *testing.T) {
	// The combined body is 30 chars, beyond the 24-char cap. We expect
	// exactly maxPrefixWidth chars of body content (23 + ellipsis) then
	// the bracket + trailing space.
	got := buildPrefix("long-workflow-name", "very-long-step")
	body := strings.TrimSuffix(strings.TrimPrefix(got, "["), "] ")
	assert.LessOrEqual(t, len(body), maxPrefixWidth+len("…")-1,
		"truncated body must not exceed the configured cap")
	assert.Contains(t, body, "…", "truncated prefix must carry an ellipsis marker")
}

func TestLineWriterPrefixesEachWrite(t *testing.T) {
	var buf bytes.Buffer
	w := newLineWriter("", "build", "uuid-1", &buf)

	n, err := w.Write([]byte("hello world\n"))
	assert.NoError(t, err)
	assert.Equal(t, len("hello world\n"), n,
		"Write must report the original byte count, not the post-prefix count; "+
			"io.Writer consumers rely on this invariant")

	_, err = w.Write([]byte("second line\n"))
	assert.NoError(t, err)

	assert.Equal(t,
		"[build] hello world\n[build] second line\n",
		buf.String())
}

func TestLineWriterAppendsMissingNewline(t *testing.T) {
	// CopyLineByLine always includes the trailing newline today, but
	// the defensive fix-up keeps the prefix aligned for any future
	// upstream that forgets. Test that behavior explicitly so it does
	// not silently regress.
	var buf bytes.Buffer
	w := newLineWriter("", "build", "uuid-1", &buf)

	_, err := w.Write([]byte("no newline here"))
	assert.NoError(t, err)
	_, err = w.Write([]byte("next line\n"))
	assert.NoError(t, err)

	assert.Equal(t,
		"[build] no newline here\n[build] next line\n",
		buf.String())
}

func TestLineWriterMultiWorkflowPrefix(t *testing.T) {
	var buf bytes.Buffer
	w := newLineWriter("test", "unit", "uuid-2", &buf)

	_, err := w.Write([]byte("ok\n"))
	assert.NoError(t, err)

	assert.Equal(t, "[test/unit] ok\n", buf.String())
}

func TestLineWriterCloseIsNoop(t *testing.T) {
	// Close must not touch stderr or any underlying stream — other
	// writers may be pointing at it. A return value of nil documents
	// the no-op contract.
	w := NewLineWriter("build", "uuid-1")
	assert.NoError(t, w.Close())
}

func TestWorkflowHeaderFormat(t *testing.T) {
	// The "# name" banner is a user-visible contract; downstream tools
	// may grep for it. Guard the exact format.
	var buf bytes.Buffer
	WorkflowHeader(&buf, "build")
	assert.Equal(t, "# build\n", buf.String())
}

func TestWorkflowHeaderTrimsWhitespace(t *testing.T) {
	var buf bytes.Buffer
	WorkflowHeader(&buf, "  padded  ")
	assert.Equal(t, "# padded\n", buf.String())
}
