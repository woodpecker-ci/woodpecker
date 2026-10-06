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

package tui

import (
	"bytes"
	"fmt"
	"math"
	"slices"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/common"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/status"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

const (
	// The block of the lines before the first command, others are named by
	// the line of their command.
	blockInit = -1
	// Width of what logGutter renders.
	logGutterWidth = 14
)

// logRow is a line of the log pane. Like in the web UI the output of each
// command forms a block that can be folded down to the line of the command.
type logRow struct {
	// entry is nil for the row that heads the lines before the first command
	entry *woodpecker.LogEntry
	text  string
	block int
	// head is set for the row a block is folded down to
	head bool
	// folded is set for the head of a folded block
	folded bool
}

// isCommand reports if a log line is the echo of a command about to run.
func isCommand(entry *woodpecker.LogEntry) bool {
	return bytes.HasPrefix(bytes.TrimLeft(entry.Data, " \t"), []byte(common.CommandMarker))
}

// logRows returns the rows to show of a log: all lines, in blocks if the log
// has commands, without the lines of the blocks that are folded. A block is
// folded unless set otherwise, except for the last one of a running step,
// which is the one that still gets output.
func logRows(entries []*woodpecker.LogEntry, folded map[int]bool, running bool) []logRow {
	lastBlock := blockInit
	for _, entry := range entries {
		if isCommand(entry) {
			lastBlock = entry.Line
		}
	}
	if lastBlock == blockInit {
		rows := make([]logRow, 0, len(entries))
		for _, entry := range entries {
			rows = append(rows, logRow{entry: entry, text: cleanLine(entry.Data), block: blockInit})
		}
		return rows
	}

	isFolded := func(block int) bool {
		if set, ok := folded[block]; ok {
			return set
		}
		return !running || block != lastBlock
	}

	var rows []logRow
	block := blockInit
	for i, entry := range entries {
		switch {
		case isCommand(entry):
			block = entry.Line
			rows = append(rows, logRow{entry: entry, text: cleanLine(entry.Data), block: block, head: true, folded: isFolded(block)})
			continue
		case i == 0:
			rows = append(rows, logRow{text: "Initialization", block: blockInit, head: true, folded: isFolded(blockInit)})
		}
		if !isFolded(block) {
			rows = append(rows, logRow{entry: entry, text: cleanLine(entry.Data), block: block})
		}
	}
	return rows
}

// showLog puts the log of the selected step into the log pane. The cursor
// stays on the last row if it was there, to follow a growing log. Else it
// stays on its line, or on the head of its block if the line got folded away.
func (m *Model) showLog() {
	m.logChanged = false

	follow := m.logCursor >= len(m.logRows)-1
	var cursor logRow
	if !follow {
		cursor = m.logRows[m.logCursor]
	}

	_, step := m.step()
	running := step != nil && status.Value(step.State).IsActive()
	m.logRows = logRows(m.logs[m.selected], m.folded[m.selected], running)

	sameLine := func(row logRow) bool {
		return row.block == cursor.block && row.entry != nil && cursor.entry != nil && row.entry.Line == cursor.entry.Line
	}
	head := func(row logRow) bool { return row.block == cursor.block && row.head }
	switch {
	case follow:
		m.logCursor = len(m.logRows) - 1
	case slices.ContainsFunc(m.logRows, sameLine):
		m.logCursor = slices.IndexFunc(m.logRows, sameLine)
	case slices.ContainsFunc(m.logRows, head):
		m.logCursor = slices.IndexFunc(m.logRows, head)
	}

	lines := make([]string, 0, len(m.logRows))
	for _, row := range m.logRows {
		lines = append(lines, row.text)
	}
	m.logView.SetContentLines(lines)
	m.moveLogCursor(0)
}

// moveLogCursor moves the cursor of the log pane by the given number of rows
// and scrolls the log to have it in view.
func (m *Model) moveLogCursor(by int) {
	m.logCursor = max(0, min(m.logCursor+by, len(m.logRows)-1))
	if len(m.logRows) == 0 {
		return
	}

	// long lines are wrapped, so rows differ in height
	width := float64(max(1, m.logView.Width()-logGutterWidth))
	height := func(row logRow) int {
		return max(1, int(math.Ceil(float64(ansi.StringWidth(row.text))/width)))
	}
	top := 0
	for _, row := range m.logRows[:m.logCursor] {
		top += height(row)
	}
	bottom := top + height(m.logRows[m.logCursor])

	switch {
	case top < m.logView.YOffset():
		m.logView.SetYOffset(top)
	case bottom > m.logView.YOffset()+m.logView.Height():
		m.logView.SetYOffset(bottom - m.logView.Height())
	}
}

// foldLog folds the blocks of the shown log the way fold tells, which gets
// each block and if it is folded now.
func (m *Model) foldLog(fold func(block int, folded bool) bool) {
	folded := m.folded[m.selected]
	if folded == nil {
		folded = make(map[int]bool)
		m.folded[m.selected] = folded
	}
	for _, row := range m.logRows {
		if row.head {
			folded[row.block] = fold(row.block, row.folded)
		}
	}
	m.showLog()
}

func (m *Model) handleLogKey(msg tea.KeyPressMsg) {
	switch msg.String() {
	case "up", "k":
		m.moveLogCursor(-1)
	case "down", "j":
		m.moveLogCursor(1)
	// cspell:words pgup pgdown
	case "pgup":
		m.moveLogCursor(-m.logView.Height())
	case "pgdown":
		m.moveLogCursor(m.logView.Height())
	case "g", "home":
		m.moveLogCursor(-len(m.logRows))
	case "G", "end":
		m.moveLogCursor(len(m.logRows))
	case "enter", "space":
		// only the block the cursor is in changes
		m.foldLog(func(block int, folded bool) bool {
			return folded != (block == m.logRows[m.logCursor].block)
		})
	case "c":
		m.foldLog(func(int, bool) bool { return true })
	case "e":
		m.foldLog(func(int, bool) bool { return false })
	}
}

// logGutter renders what stands in front of a log line: its number, the
// seconds since the step started each time they change, and if the block
// headed by the line is folded.
func (m *Model) logGutter(ctx viewport.GutterContext) string {
	if ctx.Soft || ctx.Index >= len(m.logRows) {
		return strings.Repeat(" ", logGutterWidth)
	}
	row := m.logRows[ctx.Index]

	number, took := "", ""
	if row.entry != nil {
		number = fmt.Sprint(row.entry.Line + 1)
		if ctx.Index == 0 || m.logRows[ctx.Index-1].entry == nil || m.logRows[ctx.Index-1].entry.Time != row.entry.Time {
			took = fmt.Sprintf("%ds", row.entry.Time)
		}
	}
	fold := " "
	switch {
	case row.folded:
		fold = "▸"
	case row.head:
		fold = "▾"
	}

	style := styleMuted
	if m.focus == paneLog && ctx.Index == m.logCursor {
		style = styleSelected
	}
	return style.Render(fmt.Sprintf("%5s %5s %s", number, took, fold)) + " "
}
