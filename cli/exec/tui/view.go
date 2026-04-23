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

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"go.woodpecker-ci.org/woodpecker/v3/cli/exec/scheduler"
)

// Layout tunables. These are constants rather than configurable at
// construction time because the split-pane layout has no meaningful
// alternatives to offer; users who need a different layout can use
// --no-tui.

const (
	// TreePaneNumerator over TreePaneDenominator is the fraction of
	// terminal width dedicated to the tree on the left. 3/8 leaves
	// a comfortable log pane on the right without squeezing long
	// step names in the tree.
	treePaneNumerator   = 3
	treePaneDenominator = 8

	// MinTreeWidth is the narrowest the tree pane will ever get. On
	// very narrow terminals we still prefer a legible tree over a
	// proportional split.
	minTreeWidth = 22

	// FooterHeight is the number of terminal rows reserved at the
	// bottom for the keybind hint line.
	footerHeight = 1

	// PaneBorderWidth accounts for the two vertical border columns
	// lipgloss draws around each pane.
	paneBorderWidth = 2

	// RightPaneTabsHeight accounts for the tab header row plus the
	// top/bottom border rows of the right pane.
	rightPaneTabsHeight = 3

	// MinTotalWidthMultiple keeps the combined tree+log width above
	// twice the minimum tree width so both panes stay legible when
	// the terminal is narrower than ideal.
	minTotalWidthMultiple = 2

	// RowInnerPadding is the horizontal padding lipgloss adds to a
	// pane when Padding(0, 1) is set. We subtract it from a row's
	// width cap to avoid line-wrapping inside the pane.
	rowInnerPadding = 2
)

// flatKind tags a flatItem as pointing at a workflow row or a step
// row in the flattened tree list.
type flatKind int

const (
	flatKindWorkflow flatKind = iota
	flatKindStep
)

// flatItem is one row in the navigable tree list. Used by cursor
// movement and by the renderer so both agree on what the user sees.
type flatItem struct {
	kind     flatKind
	workflow *workflowNode
	step     *stepNode // nil when kind is flatKindWorkflow
}

// flatten returns the currently visible tree rows in render order.
// Workflows are always visible; steps appear only for expanded
// workflows. The returned slice reflects the model's current state
// and is safe to iterate alongside rendering.
func (m *Model) flatten() []flatItem {
	out := make([]flatItem, 0, len(m.workflows))
	for _, wf := range m.workflows {
		out = append(out, flatItem{kind: flatKindWorkflow, workflow: wf})
		if !wf.expanded {
			continue
		}
		for _, st := range wf.steps {
			out = append(out, flatItem{kind: flatKindStep, workflow: wf, step: st})
		}
	}
	return out
}

// selectedStep returns the step currently under the cursor, or nil
// if the cursor is on a workflow row or out of range. Used by the
// log viewport to decide which per-step ring to show.
func (m *Model) selectedStep() (wf *workflowNode, st *stepNode) {
	items := m.flatten()
	if m.cursor < 0 || m.cursor >= len(items) {
		return nil, nil
	}
	it := items[m.cursor]
	if it.kind != flatKindStep {
		return it.workflow, nil
	}
	return it.workflow, it.step
}

// layout computes the two pane widths plus the body height (rows
// available for content after reserving the footer). Called from
// resizeViewports and View so both agree on sizes.
func (m *Model) layout() (treeWidth, logWidth, bodyHeight int) {
	totalWidth := m.width
	if totalWidth < minTreeWidth*minTotalWidthMultiple {
		totalWidth = minTreeWidth * minTotalWidthMultiple
	}
	treeWidth = totalWidth * treePaneNumerator / treePaneDenominator
	if treeWidth < minTreeWidth {
		treeWidth = minTreeWidth
	}
	logWidth = totalWidth - treeWidth
	if logWidth < minTreeWidth {
		logWidth = minTreeWidth
	}
	bodyHeight = m.height - footerHeight
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	return treeWidth, logWidth, bodyHeight
}

// resizeViewports propagates the current terminal size into the two
// bubbles viewports. Called from the WindowSizeMsg handler.
func (m *Model) resizeViewports() {
	_, logWidth, bodyHeight := m.layout()
	// Leave one column for the pane border on the right side.
	innerWidth := logWidth - paneBorderWidth
	if innerWidth < 1 {
		innerWidth = 1
	}
	innerHeight := bodyHeight - rightPaneTabsHeight
	if innerHeight < 1 {
		innerHeight = 1
	}
	m.logView.SetWidth(innerWidth)
	m.logView.SetHeight(innerHeight)
	m.debugView.SetWidth(innerWidth)
	m.debugView.SetHeight(innerHeight)
}

// refreshLogView rebuilds the log viewport contents from the ring
// backing the currently-selected step. If no step is selected (or a
// workflow row is selected), the viewport shows a hint instead.
func (m *Model) refreshLogView() {
	_, st := m.selectedStep()
	if st == nil {
		m.logView.SetContent("select a step to view its log")
		return
	}
	lines, truncated := st.log.Snapshot()
	var b strings.Builder
	if truncated > 0 {
		fmt.Fprintf(&b, "[… %d line(s) truncated]\n", truncated)
	}
	for _, ln := range lines {
		b.WriteString(ln)
	}
	m.logView.SetContent(b.String())
	// Most users want to see the latest output; auto-scroll to the
	// bottom on refresh unless they've manually navigated elsewhere.
	// The viewport's AtBottom check keeps us from stealing the
	// scroll position when the user is reading history.
	if m.logView.AtBottom() {
		m.logView.GotoBottom()
	}
}

// refreshDebugView rebuilds the debug viewport contents.
func (m *Model) refreshDebugView() {
	lines, truncated := m.debug.Snapshot()
	var b strings.Builder
	if truncated > 0 {
		fmt.Fprintf(&b, "[… %d line(s) truncated]\n", truncated)
	}
	for _, ln := range lines {
		b.WriteString(ln)
	}
	m.debugView.SetContent(b.String())
	if m.debugView.AtBottom() {
		m.debugView.GotoBottom()
	}
}

// renderView composes the full TUI frame from the current model
// state. Split out of Model.View so the tea.View wrapper stays thin.
func renderView(m *Model) string {
	if !m.viewReady {
		return placeholderView(m)
	}
	treeWidth, logWidth, bodyHeight := m.layout()

	tree := renderTree(m, treeWidth, bodyHeight)
	right := renderRightPane(m, logWidth, bodyHeight)

	body := lipgloss.JoinHorizontal(lipgloss.Top, tree, right)
	footer := renderFooter(m, m.width)
	return lipgloss.JoinVertical(lipgloss.Left, body, footer)
}

// renderTree draws the left-hand workflow/step tree.
func renderTree(m *Model, width, height int) string {
	focused := m.focus == FocusTree
	style := paneStyle(focused).Width(width).Height(height)

	items := m.flatten()
	var b strings.Builder
	// The body is limited by the pane height; show as many rows as
	// fit, centered loosely around the cursor so it stays visible.
	//
	// We render every row and rely on truncation inside the pane
	// style for overflow — dynamic scrolling for a tree this short
	// is overkill for v1.
	for i, it := range items {
		selected := focused && i == m.cursor
		b.WriteString(renderTreeRow(it, selected, width))
		b.WriteByte('\n')
	}
	return style.Render(strings.TrimRight(b.String(), "\n"))
}

// renderTreeRow draws one row of the tree.
func renderTreeRow(it flatItem, selected bool, width int) string {
	var glyph, label string
	var indent string
	switch it.kind {
	case flatKindWorkflow:
		glyph = stateGlyph(it.workflow.state)
		label = it.workflow.name
		if it.workflow.expanded {
			indent = "▾ "
		} else {
			indent = "▸ "
		}
	case flatKindStep:
		glyph = stepGlyph(it.step)
		label = it.step.name
		indent = "    "
	}

	// Build the row; add an arrow prefix for selected lines so the
	// focus cue survives themes that can't do reverse video.
	prefix := "  "
	if selected {
		prefix = "› "
	}
	body := prefix + indent + glyph + " " + label
	// Manual truncation keeps the row within the pane width even if
	// lipgloss's internal width handling decides to wrap. Reserve
	// rowInnerPadding for the style's horizontal padding.
	maxBody := width - rowInnerPadding
	if maxBody > 0 && lipgloss.Width(body) > maxBody {
		body = ansiTruncate(body, maxBody)
	}
	if selected {
		return selectedRowStyle.Render(body)
	}
	return body
}

// renderRightPane renders the tab header plus the active viewport.
func renderRightPane(m *Model, width, height int) string {
	focused := m.focus == FocusLog || m.focus == FocusDebug

	tabs := renderTabs(m, width)
	var body string
	switch m.focus {
	case FocusDebug:
		body = m.debugView.View()
	default:
		// Log is the default right-pane view even when focus is on
		// the tree. This matches the "inspect what you clicked last"
		// mental model users bring from IDE tree views.
		body = m.logView.View()
	}

	panel := lipgloss.JoinVertical(lipgloss.Left, tabs, body)
	return paneStyle(focused).Width(width).Height(height).Render(panel)
}

// renderTabs renders the "logs | debug" header at the top of the
// right pane.
func renderTabs(m *Model, width int) string {
	logs := " logs "
	dbg := " debug "
	if m.focus == FocusDebug {
		dbg = tabActiveStyle.Render(dbg)
		logs = tabInactiveStyle.Render(logs)
	} else {
		logs = tabActiveStyle.Render(logs)
		dbg = tabInactiveStyle.Render(dbg)
	}
	header := logs + " " + dbg
	_ = width // retained for future alignment
	return header
}

// renderFooter renders the keybind hint strip at the bottom.
func renderFooter(m *Model, width int) string {
	focusName := "tree"
	switch m.focus {
	case FocusLog:
		focusName = "log"
	case FocusDebug:
		focusName = "debug"
	}
	done, total := m.progressCounts()
	status := fmt.Sprintf("%d/%d", done, total)
	switch {
	case m.canceling:
		status = "canceling…"
	case m.done && m.doneErr != nil:
		status = "failed"
	case m.done:
		status = "done"
	}
	hint := fmt.Sprintf(
		"[%s] %s  j/k: move  enter: expand  tab: focus  L: debug  q: quit",
		focusName, status,
	)
	_ = width
	return footerStyle.Render(hint)
}

// progressCounts returns (finished, total) step counts across the
// whole DAG. Skipped and blocked workflows contribute their own step
// counts as "finished" so the number reflects visible progress, not
// only executed work.
func (m *Model) progressCounts() (done, total int) {
	for _, wf := range m.workflows {
		total += len(wf.steps)
		for _, st := range wf.steps {
			if st.exited || st.skipped {
				done++
			}
		}
		if wf.state.Terminal() && wf.state != scheduler.StateSuccess &&
			wf.state != scheduler.StateFailure {
			// Blocked / canceled workflows with no steps still count
			// visually: treat each such workflow as one unit.
			if len(wf.steps) == 0 {
				total++
				done++
			}
		}
	}
	return done, total
}

// stepGlyph returns the status glyph for a step node.
func stepGlyph(s *stepNode) string {
	switch {
	case s.skipped:
		return glyphSkipped
	case s.exited && s.exitCode == 0:
		return glyphSuccess
	case s.exited:
		return glyphFailure
	case s.errText != "":
		return glyphFailure
	case s.oomKill:
		return glyphFailure
	}
	return glyphRunning
}

// renderViewTea wraps renderView in a tea.View so Model.View has a
// one-liner.
func renderViewTea(m *Model) tea.View {
	return tea.NewView(renderView(m))
}

// ansiTruncate trims body to visible width fit. Lipgloss's Width
// counts printable cells; strings.Split/rune slicing would over-
// truncate styled content. For simplicity, this chunk assumes no
// styled content reaches the tree rows (they're plain strings), so
// we just rune-slice. If/when styled content arrives, swap this for
// lipgloss's built-in truncate.
func ansiTruncate(s string, maxCells int) string {
	if maxCells <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= maxCells {
		return s
	}
	if maxCells == 1 {
		return "…"
	}
	return string(r[:maxCells-1]) + "…"
}
