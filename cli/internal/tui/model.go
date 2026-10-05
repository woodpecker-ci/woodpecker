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

// Package tui shows the workflows of a pipeline with their steps and logs in
// the terminal, like the pipeline page of the web UI does in the browser.
//
// It does not know where the pipeline runs. Whoever does feeds the model via
// tea.Program.Send with the messages defined here.
package tui

import (
	"bytes"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/status"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

const (
	// Log lines kept per step, older ones are dropped.
	// It is the default limit of the web UI.
	maxLogLines = 5000
	// Lines kept in the messages pane.
	maxMessages = 1000

	// How often durations and a changed log are redrawn.
	tickInterval = 250 * time.Millisecond
)

type (
	// WorkflowsMsg adds the workflows or replaces the ones with the same id.
	WorkflowsMsg []*woodpecker.Workflow
	// LogMsg adds log lines to the steps they belong to. Lines of a step have
	// to come in order, the ones it already has are ignored.
	LogMsg []*woodpecker.LogEntry
	// StepLogMsg replaces the log of a step.
	StepLogMsg struct {
		StepID  int64
		Entries []*woodpecker.LogEntry
	}
	// MessageMsg adds text that belongs to no step to the messages pane.
	MessageMsg string
	// StatusMsg sets the status of the pipeline. Without it the status is
	// derived from the workflows.
	StatusMsg status.Value

	tickMsg struct{}
)

type pane int

const (
	paneSteps pane = iota
	paneLog
	paneMessages
)

// row is a line of the step list: a workflow or, if step is set, one of its steps.
type row struct {
	workflow *woodpecker.Workflow
	step     *woodpecker.Step
}

// Model is the bubbletea model of the pipeline view.
type Model struct {
	// OnSelect, if set, is called each time another step gets selected,
	// e.g. to load its log.
	OnSelect func(step *woodpecker.Step)
	// OnCancel, if set, is called instead of quitting the first time the user
	// asks to quit while the pipeline is not finished yet.
	OnCancel func()

	status    status.Value
	workflows []*woodpecker.Workflow
	collapsed map[int64]bool                   // by workflow id
	logs      map[int64][]*woodpecker.LogEntry // by step id
	messages  []string

	// cursor is the selected row of the step list
	cursor int
	// follow moves the cursor to the running step until the user moves it
	follow bool
	// selected is the id of the step the log is shown of
	selected int64
	// logChanged is set if the shown log has lines not drawn yet
	logChanged bool
	canceled   bool

	focus         pane
	width, height int
	stepsOffset   int
	logView       viewport.Model
	messagesView  viewport.Model
}

// New returns the model of an empty pipeline view.
func New() *Model {
	m := &Model{
		collapsed:    make(map[int64]bool),
		logs:         make(map[int64][]*woodpecker.LogEntry),
		follow:       true,
		logView:      viewport.New(),
		messagesView: viewport.New(),
	}
	m.logView.SoftWrap = true
	m.logView.LeftGutterFunc = m.logGutter
	m.messagesView.SoftWrap = true
	return m
}

// pipelineStatus returns the status of the pipeline: the one set, or else
// the one its workflows have together.
func (m *Model) pipelineStatus() status.Value {
	if m.status != "" {
		return m.status
	}
	if len(m.workflows) == 0 {
		return status.Pending
	}

	pipelineStatus := status.Success
	for _, workflow := range m.workflows {
		pipelineStatus = status.Merge(pipelineStatus, status.Value(workflow.State))
	}
	return pipelineStatus
}

// active reports if a workflow is still to run or running.
func (m *Model) active() bool {
	for _, workflow := range m.workflows {
		if status.Value(workflow.State).IsActive() {
			return true
		}
	}
	return false
}

func (m *Model) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resize()

	case tea.KeyPressMsg:
		return m, m.handleKey(msg)

	case WorkflowsMsg:
		m.setWorkflows(msg)
		return m, m.selectStep()

	case LogMsg:
		for _, entry := range msg {
			m.addLog(entry)
		}

	case StepLogMsg:
		m.logs[msg.StepID] = msg.Entries[max(0, len(msg.Entries)-maxLogLines):]
		m.logChanged = m.logChanged || msg.StepID == m.selected

	case MessageMsg:
		m.addMessage(string(msg))

	case StatusMsg:
		m.status = status.Value(msg)

	case tickMsg:
		if m.logChanged {
			m.showLog()
		}
		return m, tick()
	}

	return m, nil
}

// setWorkflows adds the workflows or replaces the ones with the same id.
func (m *Model) setWorkflows(workflows []*woodpecker.Workflow) {
	for _, workflow := range workflows {
		i := 0
		for i < len(m.workflows) && m.workflows[i].ID != workflow.ID {
			i++
		}
		if i == len(m.workflows) {
			m.workflows = append(m.workflows, workflow)
		} else {
			m.workflows[i] = workflow
		}
	}

	rows := m.rows()
	if m.follow {
		// the running step, or to start with the first one
		i := slices.IndexFunc(rows, func(r row) bool { return r.step != nil && status.Value(r.step.State) == status.Running })
		if i < 0 && m.selected == 0 {
			i = slices.IndexFunc(rows, func(r row) bool { return r.step != nil })
		}
		if i >= 0 {
			m.cursor = i
		}
	}
	m.cursor = max(0, min(m.cursor, len(rows)-1))
}

// addLog adds a log line to its step, unless it is already there, so a log
// can be sent again from its start to complete it.
func (m *Model) addLog(entry *woodpecker.LogEntry) {
	entries := m.logs[entry.StepID]
	if len(entries) > 0 && entry.Line <= entries[len(entries)-1].Line {
		return
	}
	m.logs[entry.StepID] = append(entries, entry)
	if len(entries) >= maxLogLines {
		m.logs[entry.StepID] = m.logs[entry.StepID][len(entries)+1-maxLogLines:]
	}

	if entry.StepID == m.selected {
		m.logChanged = true
	}
}

func (m *Model) addMessage(text string) {
	for line := range strings.SplitSeq(strings.TrimRight(text, "\n"), "\n") {
		m.messages = append(m.messages, cleanLine([]byte(line)))
	}
	if len(m.messages) > maxMessages {
		m.messages = m.messages[len(m.messages)-maxMessages:]
	}

	// the pane appears with the first message
	m.resize()
	atBottom := m.messagesView.AtBottom()
	m.messagesView.SetContentLines(m.messages)
	if atBottom {
		m.messagesView.GotoBottom()
	}
}

// rows returns the visible lines of the step list.
func (m *Model) rows() []row {
	var rows []row
	for _, workflow := range m.workflows {
		rows = append(rows, row{workflow: workflow})
		if m.collapsed[workflow.ID] {
			continue
		}
		for _, step := range workflow.Children {
			rows = append(rows, row{workflow: workflow, step: step})
		}
	}
	return rows
}

// current returns the row the cursor is at.
func (m *Model) current() row {
	if rows := m.rows(); m.cursor < len(rows) {
		return rows[m.cursor]
	}
	return row{}
}

// step returns the step the log is shown of.
func (m *Model) step() (*woodpecker.Workflow, *woodpecker.Step) {
	for _, workflow := range m.workflows {
		for _, step := range workflow.Children {
			if step.ID == m.selected {
				return workflow, step
			}
		}
	}
	return nil, nil
}

// selectStep shows the log of the step the cursor is at.
// On a workflow the log of the last selected step stays.
func (m *Model) selectStep() tea.Cmd {
	step := m.current().step
	if step == nil || step.ID == m.selected {
		return nil
	}

	m.selected = step.ID
	m.showLog()
	m.logView.GotoBottom()

	if m.OnSelect == nil {
		return nil
	}
	selected := *step
	return func() tea.Msg {
		m.OnSelect(&selected)
		return nil
	}
}

// showLog puts the log of the selected step into the log pane and keeps
// following it, if its end was shown.
func (m *Model) showLog() {
	m.logChanged = false

	entries := m.logs[m.selected]
	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		lines = append(lines, cleanLine(entry.Data))
	}

	atBottom := m.logView.AtBottom()
	m.logView.SetContentLines(lines)
	if atBottom {
		m.logView.GotoBottom()
	}
}

func (m *Model) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "q", "ctrl+c":
		if m.OnCancel != nil && !m.canceled && m.active() {
			m.canceled = true
			return func() tea.Msg {
				m.OnCancel()
				return nil
			}
		}
		return tea.Quit
	case "tab":
		m.focus = (m.focus + 1) % (paneMessages + 1)
		if m.focus == paneMessages && len(m.messages) == 0 {
			m.focus = paneSteps
		}
		return nil
	case "esc":
		m.focus = paneSteps
		return nil
	}

	var view *viewport.Model
	switch m.focus {
	case paneSteps:
		return m.handleStepsKey(msg)
	case paneLog:
		view = &m.logView
	case paneMessages:
		view = &m.messagesView
	}

	switch msg.String() {
	case "g", "home":
		view.GotoTop()
	case "G", "end":
		view.GotoBottom()
	default:
		var cmd tea.Cmd
		*view, cmd = view.Update(msg)
		return cmd
	}
	return nil
}

func (m *Model) handleStepsKey(msg tea.KeyPressMsg) tea.Cmd {
	rows := m.rows()
	if len(rows) == 0 {
		return nil
	}

	cursor := m.cursor
	switch msg.String() {
	case "up", "k":
		cursor--
	case "down", "j":
		cursor++
	case "g", "home":
		cursor = 0
	case "G", "end":
		cursor = len(rows) - 1
	case "enter", "space":
		if r := rows[m.cursor]; r.step == nil {
			m.collapsed[r.workflow.ID] = !m.collapsed[r.workflow.ID]
		} else {
			m.focus = paneLog
		}
		return nil
	default:
		return nil
	}

	m.follow = false
	m.cursor = max(0, min(cursor, len(rows)-1))
	return m.selectStep()
}

// cleanLine returns the text of a log line that is safe to print inside the
// view: colors and styles are kept, everything else that controls the
// terminal is removed and a line rewritten with carriage returns is reduced
// to its last version.
func cleanLine(data []byte) string {
	data = bytes.TrimRight(data, "\r\n")
	if i := bytes.LastIndexByte(data, '\r'); i >= 0 {
		data = data[i+1:]
	}

	parser := ansi.GetParser()
	defer ansi.PutParser(parser)

	var line strings.Builder
	var state byte
	for len(data) > 0 {
		seq, width, n, newState := ansi.DecodeSequence(data, state, parser)
		switch {
		case width > 0:
			line.Write(seq)
		case string(seq) == "\t":
			line.WriteString("    ")
		case ansi.HasCsiPrefix(seq) && ansi.Cmd(parser.Command()).Final() == 'm':
			line.Write(seq)
		}
		state = newState
		data = data[n:]
	}
	return line.String()
}
