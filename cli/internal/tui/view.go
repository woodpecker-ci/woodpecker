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
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/status"
)

const (
	// What border and padding of a pane take of its width and height.
	paneFrameWidth  = 4
	paneFrameHeight = 2

	minStepsWidth     = 28
	maxStepsWidth     = 50
	maxMessagesHeight = 8
)

var (
	colorOk      = lipgloss.Green
	colorError   = lipgloss.Red
	colorInfo    = lipgloss.Blue
	colorWarn    = lipgloss.Yellow
	colorNeutral = lipgloss.BrightBlack

	styleSelected = lipgloss.NewStyle().Reverse(true)
	styleTitle    = lipgloss.NewStyle().Bold(true)
	styleMuted    = lipgloss.NewStyle().Foreground(colorNeutral)
)

func paneStyle(focused bool) lipgloss.Style {
	borderColor := colorNeutral
	if focused {
		borderColor = colorInfo
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)
}

// statusIcon returns the sign of a status, in the color the web UI uses for it.
func statusIcon(state string, colored bool) string {
	icon, color := "○", colorWarn
	switch status.Value(state) {
	case status.Success:
		icon, color = "✓", colorOk
	case status.Failure, status.Error, status.Declined:
		icon, color = "✗", colorError
	case status.Running:
		icon, color = "●", colorInfo
	case status.Skipped, status.Canceled, status.Killed, status.Blocked:
		icon, color = "⊘", colorNeutral
	}
	if !colored {
		return icon
	}
	return lipgloss.NewStyle().Foreground(color).Render(icon)
}

// duration returns how long a workflow or step took or is running.
func duration(started, finished int64) string {
	if started == 0 {
		return ""
	}
	if finished == 0 {
		finished = time.Now().Unix()
	}
	d := max(0, finished-started)

	const minute, hour = 60, 60 * 60
	if d >= hour {
		return fmt.Sprintf("%02d:%02d:%02d", d/hour, d%hour/minute, d%minute)
	}
	return fmt.Sprintf("%02d:%02d", d/minute, d%minute)
}

// layout returns the width of the step list and the height of the messages pane.
func (m *Model) layout() (stepsWidth, messagesHeight int) {
	stepsWidth = min(max(m.width/4, minStepsWidth), maxStepsWidth, m.width/2) //nolint:mnd
	if len(m.messages) > 0 {
		messagesHeight = min(maxMessagesHeight, m.height/3) //nolint:mnd
	}
	return stepsWidth, messagesHeight
}

// resize fits the log and messages to the window size.
func (m *Model) resize() {
	stepsWidth, messagesHeight := m.layout()
	bodyHeight := m.height - messagesHeight - 1 // footer

	// what showed its end has to show it in the new size too
	logAtBottom, messagesAtBottom := m.logView.AtBottom(), m.messagesView.AtBottom()

	m.logView.SetWidth(max(0, m.width-stepsWidth-paneFrameWidth))
	m.logView.SetHeight(max(0, bodyHeight-paneFrameHeight-2)) //nolint:mnd // title and step status
	m.messagesView.SetWidth(max(0, m.width-paneFrameWidth))
	m.messagesView.SetHeight(max(0, messagesHeight-paneFrameHeight))

	if logAtBottom {
		m.logView.GotoBottom()
	}
	if messagesAtBottom {
		m.messagesView.GotoBottom()
	}
}

func (m *Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m *Model) render() string {
	if m.width == 0 {
		return ""
	}

	stepsWidth, messagesHeight := m.layout()
	bodyHeight := m.height - messagesHeight - 1

	pane := func(p pane, width, height int, content string) string {
		return paneStyle(m.focus == p).
			Width(width).Height(height).MaxHeight(height).
			Render(content)
	}

	view := lipgloss.JoinHorizontal(lipgloss.Top,
		pane(paneSteps, stepsWidth, bodyHeight, m.renderSteps(stepsWidth-paneFrameWidth, bodyHeight-paneFrameHeight)),
		pane(paneLog, m.width-stepsWidth, bodyHeight, m.renderLog(m.width-stepsWidth-paneFrameWidth)),
	)
	if messagesHeight > 0 {
		view += "\n" + pane(paneMessages, m.width, messagesHeight, m.messagesView.View())
	}
	return view + "\n" + m.renderFooter()
}

// renderSteps renders the part of the step list around the cursor.
func (m *Model) renderSteps(width, height int) string {
	rows := m.rows()
	m.stepsOffset = max(min(m.stepsOffset, m.cursor), m.cursor-height+1, 0)

	var lines []string
	for i := m.stepsOffset; i < min(len(rows), m.stepsOffset+height); i++ {
		lines = append(lines, m.renderRow(rows[i], i == m.cursor, width))
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderRow(r row, selected bool, width int) string {
	var name, took string
	if r.step == nil {
		fold := "▾"
		if m.collapsed[r.workflow.ID] {
			fold = "▸"
		}
		name = fold + " " + statusIcon(r.workflow.State, !selected) + " " + r.workflow.Name
		// like the badges of a matrix workflow
		for _, key := range slices.Sorted(maps.Keys(r.workflow.Environ)) {
			name += " " + key + "=" + r.workflow.Environ[key]
		}
		took = duration(r.workflow.Started, r.workflow.Stopped)
	} else {
		name = "    " + statusIcon(r.step.State, !selected) + " " + r.step.Name
		took = duration(r.step.Started, r.step.Stopped)
	}

	name = ansi.Truncate(name, width-len(took)-1, "…")
	line := name + strings.Repeat(" ", max(0, width-ansi.StringWidth(name)-len(took))) + took
	if selected {
		return styleSelected.Render(line)
	}
	return line
}

// renderLog renders the log of the selected step, with its name above and
// how it ended below, like the web UI.
func (m *Model) renderLog(width int) string {
	workflow, step := m.step()
	if step == nil {
		return styleTitle.Render("Step Logs")
	}

	title := styleTitle.Render(ansi.Truncate("Step Logs: "+workflow.Name+" / "+step.Name, width, "…"))

	log := m.logView.View()
	if len(m.logs[step.ID]) == 0 {
		hint := "No logs"
		switch state := status.Value(step.State); {
		case state == status.Canceled:
			hint = "This step has been canceled."
		case state == status.Skipped:
			hint = "This step has been skipped."
		case step.Started == 0:
			hint = "This step hasn't started yet."
		case state.IsActive():
			hint = "Loading…"
		}
		log = lipgloss.Place(m.logView.Width(), m.logView.Height(), lipgloss.Center, lipgloss.Center, styleMuted.Render(hint))
	}

	end := ""
	if step.Stopped != 0 {
		end = fmt.Sprintf("Exit Code %d", step.ExitCode)
		if step.Error != "" {
			end = step.Error
		}
		end = statusIcon(step.State, true) + " " + styleTitle.Render(ansi.Truncate(end, width-2, "…")) //nolint:mnd
	}

	return title + "\n" + log + "\n" + end
}

// logGutter renders the line number and, each time it changes, the seconds
// since the step started in front of a log line.
func (m *Model) logGutter(ctx viewport.GutterContext) string {
	entries := m.logs[m.selected]
	if ctx.Soft || ctx.Index >= len(entries) {
		return strings.Repeat(" ", 12) //nolint:mnd
	}

	entry := entries[ctx.Index]
	took := ""
	if ctx.Index == 0 || entries[ctx.Index-1].Time != entry.Time {
		took = fmt.Sprintf("%ds", entry.Time)
	}
	return styleMuted.Render(fmt.Sprintf("%5d %5s ", entry.Line+1, took))
}

func (m *Model) renderFooter() string {
	quit := "quit"
	if m.OnCancel != nil && !m.canceled && m.active() {
		quit = "cancel"
	}

	pipelineStatus := m.pipelineStatus()
	state := string(pipelineStatus)
	if m.canceled && pipelineStatus.IsActive() {
		state = "canceling…"
	}

	footer := " " + statusIcon(string(pipelineStatus), true) + " " + state +
		styleMuted.Render("  ↑/↓ select · enter fold/open log · tab switch pane · q "+quit)
	return ansi.Truncate(footer, m.width, "…")
}

// Summary returns the messages and the status of all workflows and steps as
// text, to leave something in the terminal once the view is closed.
func (m *Model) Summary() string {
	var b strings.Builder
	for _, message := range m.messages {
		b.WriteString(message + "\n")
	}

	for _, workflow := range m.workflows {
		fmt.Fprintf(&b, "%s %s %s\n", statusIcon(workflow.State, true), workflow.Name, duration(workflow.Started, workflow.Stopped))
		if workflow.Error != "" {
			fmt.Fprintf(&b, "    %s\n", workflow.Error)
		}
		for _, step := range workflow.Children {
			fmt.Fprintf(&b, "    %s %s %s", statusIcon(step.State, true), step.Name, duration(step.Started, step.Stopped))
			switch {
			case step.Error != "":
				fmt.Fprintf(&b, " (%s)", step.Error)
			case step.ExitCode != 0:
				fmt.Fprintf(&b, " (exit code %d)", step.ExitCode)
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}
