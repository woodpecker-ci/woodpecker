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

package tui

import (
	"fmt"
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/cli/exec/scheduler"
)

// Status glyphs rendered next to each workflow and step. Unicode
// round-trips fine in every modern terminal; ASCII fallbacks can be
// added later if real users hit issues.
const (
	glyphSuccess  = "✓"
	glyphFailure  = "✗"
	glyphSkipped  = "⊘"
	glyphBlocked  = "⏸"
	glyphCanceled = "⊗"
	glyphRunning  = "●"
	glyphPending  = "·"
)

// stateGlyph returns a single-character status marker for a workflow
// state. Used by the tree renderer; also handy for the placeholder
// view so operators can eyeball skeleton output even before the full
// rendering lands.
func stateGlyph(s scheduler.State) string {
	switch s {
	case scheduler.StateSuccess:
		return glyphSuccess
	case scheduler.StateFailure:
		return glyphFailure
	case scheduler.StateBlocked:
		return glyphBlocked
	case scheduler.StateCanceled:
		return glyphCanceled
	case scheduler.StateRunning:
		return glyphRunning
	}
	return glyphPending
}

// placeholderHeaderWidth is the width of the horizontal rule in the
// skeleton placeholder view. Replaced by lipgloss-aware sizing in
// the full layout (chunk 5).
const placeholderHeaderWidth = 40

// placeholderView is the bare-bones view used until the full tree +
// log + debug layout lands. It renders one line per workflow with
// state glyph, name, and (if running or finished) a short summary.
// Enough to verify the wiring end-to-end without committing to a
// visual design yet.
func placeholderView(m *Model) string {
	var b strings.Builder

	fmt.Fprintln(&b, "Woodpecker exec")
	fmt.Fprintln(&b, strings.Repeat("─", placeholderHeaderWidth))

	for _, wf := range m.workflows {
		fmt.Fprintf(&b, "  %s %s", stateGlyph(wf.state), wf.name)
		if wf.err != nil {
			fmt.Fprintf(&b, "  (%s)", wf.err.Error())
		}
		fmt.Fprintln(&b)

		if wf.expanded {
			for _, s := range wf.steps {
				glyph := glyphPending
				switch {
				case s.skipped:
					glyph = glyphSkipped
				case s.exited && s.exitCode == 0:
					glyph = glyphSuccess
				case s.exited:
					glyph = glyphFailure
				case s.errText != "":
					glyph = glyphFailure
				}
				fmt.Fprintf(&b, "      %s %s\n", glyph, s.name)
			}
		}
	}

	fmt.Fprintln(&b)
	if m.canceling {
		fmt.Fprintln(&b, "canceling…")
	}
	if m.done {
		if m.doneErr != nil {
			fmt.Fprintf(&b, "finished with error: %s\n", m.doneErr.Error())
		} else {
			fmt.Fprintln(&b, "finished successfully")
		}
	}
	fmt.Fprintln(&b, "q / ctrl-c: quit")

	return b.String()
}
