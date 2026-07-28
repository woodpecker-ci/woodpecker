// Copyright 2022 Woodpecker Authors
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
	"fmt"
	"io"
	"os"
	"strings"
)

// maxPrefixWidth caps how wide the [prefix] column can grow when
// multiple workflows or steps are running concurrently. 24 characters
// is enough for typical step names ("test-integration-1") without
// pushing the log body too far right on an 80-column terminal.
const maxPrefixWidth = 24

// LineWriter writes pipeline step log lines to stderr, one logical
// line per Write. Each line is prefixed with the step (and, when the
// run contains more than one workflow, the workflow) so that output
// from parallel workflows remains attributable when interleaved.
//
// The format is deliberately grep-friendly: no ANSI escape sequences,
// no dynamic counters, no timestamps. Tools that consume the output
// get one predictable line per log line. Terminal users who want
// richer output should use the TUI (the default on a tty).
type LineWriter struct {
	// stepName is the name of the pipeline step whose output this
	// writer consumes.
	stepName string
	// stepUUID is retained for forward-compat with future log routing
	// that needs a stable key, but is not rendered.
	stepUUID string
	// workflowName, when non-empty, is emitted before the step name
	// separated by a slash. It is left empty in single-workflow runs
	// to keep the prefix terse.
	workflowName string
	// prefix is the precomputed "[wf/step] " or "[step] " form.
	prefix string
	// out is the destination. In production this is os.Stderr; tests
	// can swap it.
	out io.Writer
}

// NewLineWriter returns a writer for the given step in a
// single-workflow run. The workflow prefix is omitted.
func NewLineWriter(stepName, stepUUID string) io.WriteCloser {
	return newLineWriter("", stepName, stepUUID, os.Stderr)
}

// NewWorkflowLineWriter returns a writer for a step inside a specific
// workflow. The workflow name is rendered before the step name as
// "[workflow/step]". Intended for multi-workflow runs where output
// from parallel workflows will interleave on stderr.
func NewWorkflowLineWriter(workflowName, stepName, stepUUID string) io.WriteCloser {
	return newLineWriter(workflowName, stepName, stepUUID, os.Stderr)
}

func newLineWriter(workflowName, stepName, stepUUID string, out io.Writer) *LineWriter {
	return &LineWriter{
		stepName:     stepName,
		stepUUID:     stepUUID,
		workflowName: workflowName,
		prefix:       buildPrefix(workflowName, stepName),
		out:          out,
	}
}

// buildPrefix constructs the "[workflow/step]" or "[step]" label,
// truncating with an ellipsis if the combined length exceeds
// maxPrefixWidth. The result always ends with a trailing space so
// callers can concatenate the log body directly.
func buildPrefix(workflowName, stepName string) string {
	var body string
	if workflowName != "" {
		body = workflowName + "/" + stepName
	} else {
		body = stepName
	}
	if len(body) > maxPrefixWidth {
		// Truncate with an ellipsis character. We reserve one rune for
		// the ellipsis, hence maxPrefixWidth-1.
		body = body[:maxPrefixWidth-1] + "…"
	}
	return "[" + body + "] "
}

// Write implements io.Writer. Each call corresponds to one line
// emitted by the pipeline's line-by-line copier
// (pipeline/utils.CopyLineByLine), so we can prepend the prefix once
// per call without splitting p. The returned n is len(p) per the
// io.Writer contract — we do not want partial writes to cascade into
// duplicate lines upstream.
func (w *LineWriter) Write(p []byte) (n int, err error) {
	// Defensive: if the upstream writer somehow passes us content
	// without a trailing newline, the prefix of the next line would
	// land on the same visible line. Append one so the output stays
	// aligned. CopyLineByLine always includes the newline today, but
	// future callers might not.
	needsNL := len(p) == 0 || p[len(p)-1] != '\n'

	if _, werr := fmt.Fprint(w.out, w.prefix); werr != nil {
		return 0, werr
	}
	if _, werr := w.out.Write(p); werr != nil {
		return 0, werr
	}
	if needsNL {
		if _, werr := fmt.Fprintln(w.out); werr != nil {
			return 0, werr
		}
	}
	return len(p), nil
}

// Close implements io.Closer. The underlying stderr is not owned by
// this writer, so Close is a no-op.
func (w *LineWriter) Close() error {
	return nil
}

// WorkflowHeader prints a human-readable banner announcing the start
// of a workflow. In multi-workflow runs the scheduler may emit
// multiple banners as workflows become ready; the caller decides
// whether to print one at all.
//
// The format is stable and matches the pre-refactor output so
// downstream tools that grep for "# <n>" keep working.
func WorkflowHeader(out io.Writer, name string) {
	// Keep the legacy "# name" format — it's short, unambiguous, and
	// already consumed by user workflows in the wild.
	fmt.Fprintln(out, "#", strings.TrimSpace(name))
}
