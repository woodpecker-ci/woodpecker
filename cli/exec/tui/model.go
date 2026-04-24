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

// Package tui implements the interactive split-pane display for the
// cli exec command. It consumes workflow-level events from the
// scheduler, step-level events from the pipeline tracer, and
// per-line log output from the pipeline logger, then renders a tree
// of workflows + steps alongside a log viewport and a debug tab.
//
// The package exposes a Model implementing the bubbletea Model
// interface. Callers (cli/exec) construct a Model, wrap it in a
// tea.Program, then Send messages from the scheduler's event
// consumer and from the tracer/logger callbacks.
//
// This file contains the scaffolding only — model state, message
// dispatch, and placeholder View. Real rendering (lipgloss styles,
// tree layout, log viewport, debug pane, keybind handling) is built
// on top in subsequent chunks.
package tui

import (
	"time"

	"charm.land/bubbles/v2/viewport"
	"charm.land/bubbletea/v2"

	"go.woodpecker-ci.org/woodpecker/v3/cli/exec/scheduler"
)

// workflowNode is the model's per-workflow bookkeeping. It mirrors
// scheduler state with presentation fields added (expanded, cursor
// position within its steps, per-step state and log rings).
type workflowNode struct {
	name  string
	state scheduler.State
	// err is non-nil only in terminal error states.
	err error
	// expanded controls whether child steps are rendered in the tree.
	// Defaults to true for running workflows, false once terminal.
	expanded bool
	// steps is ordered by first-seen; step-level events populate it
	// as the pipeline runtime emits tracer updates.
	steps []*stepNode
}

// stepNode is the model's per-step bookkeeping inside a workflow.
type stepNode struct {
	name     string
	uuid     string
	exited   bool
	exitCode int
	skipped  bool
	oomKill  bool
	errText  string
	// log is the per-step line ring. Owned by the model, shared with
	// the budget controller.
	log *Ring
}

// Focus identifies which pane currently accepts keyboard input.
type Focus int

const (
	// FocusTree is the default: the workflow/step tree on the top left.
	FocusTree Focus = iota
	// FocusLog is the log viewport on the top right.
	FocusLog
	// FocusMessages is the bottom-strip pane that collects pre-run
	// output (lint warnings, metadata, anything printed before the
	// TUI took over stdout) plus zerolog diagnostics captured
	// during the run. It replaces the earlier "debug tab" concept:
	// one dedicated pane for everything that is neither the tree
	// nor a step's own log output.
	FocusMessages
)

// Model is the bubbletea Model for the cli exec TUI.
//
// Construct with New. Send scheduler and pipeline messages via
// tea.Program.Send during the run; Send a PipelineDoneMsg when the
// scheduler returns. The program exits when the user presses q/ctrl-c
// after a terminal state, matching bubbletea convention.
type Model struct {
	// workflows is insertion-ordered so the tree renders the same way
	// across runs (matching yaml file ordering).
	workflows []*workflowNode
	// byName indexes into workflows for O(1) event dispatch.
	byName map[string]*workflowNode

	// Ring for the messages pane: pre-run output (lint warnings,
	// metadata diagnostics, anything printed before the TUI took
	// over stdout) plus zerolog log output captured during the run
	// via a RingWriter installed as the zerolog destination.
	messages *Ring

	// budget is the shared cap across all step rings. The messages
	// ring is NOT registered here — it has its own separate cap so
	// diagnostic noise cannot crowd out step logs.
	budget *Budget

	// UI state.
	width, height int
	focus         Focus
	// cursor is the index into the flattened navigable-items list
	// produced by flatten(). It points at either a workflow or a
	// step; the setter clamps it to the list bounds so out-of-range
	// values from a collapse/terminate cannot desync the view.
	cursor int
	// logView is the top-right viewport for step logs. It is reused
	// across selections — SetContent is called when the selection
	// changes.
	logView viewport.Model
	// messagesView is the bottom-strip viewport for diagnostics.
	messagesView viewport.Model
	// viewReady gates rendering on the first WindowSizeMsg. Before
	// the first size message we don't know how wide the panes should
	// be, so we fall back to the placeholder view.
	viewReady bool

	// Terminal state flags.
	canceling bool
	done      bool
	doneErr   error
}

// New constructs a Model seeded with the given workflow names.
// Workflow order here determines rendering order. The caller should
// pass names in the same order as scheduler.Options.Items, which is
// the order from the yaml build output.
func New(workflowNames []string) *Model {
	m := &Model{
		byName:       make(map[string]*workflowNode, len(workflowNames)),
		messages:     NewRing(MessagesLogCapBytes),
		budget:       NewBudget(GlobalLogCapBytes),
		focus:        FocusTree,
		logView:      viewport.New(),
		messagesView: viewport.New(),
	}
	for _, name := range workflowNames {
		n := &workflowNode{
			name:     name,
			state:    scheduler.StatePending,
			expanded: true,
		}
		m.workflows = append(m.workflows, n)
		m.byName[name] = n
	}
	return m
}

// MessagesRing returns the Ring backing the bottom messages pane.
// Exposed so callers can wrap it in a RingWriter and install it as
// the zerolog destination before starting the program, and/or seed
// the ring with pre-run output (lint warnings, metadata) that was
// produced before the TUI took control.
func (m *Model) MessagesRing() *Ring {
	return m.messages
}

// fallbackStepRingCapBytes is the per-ring cap used only for the
// defensive "unknown workflow" path in StepRing. Real step rings
// rely on the shared global budget; this is a throwaway buffer size
// that should never be reached in practice.
const fallbackStepRingCapBytes = 1024 * 1024

// StepRing returns (or lazily creates) the per-step log ring for the
// given workflow/step pair. The ring is registered with the model's
// shared budget on creation so eviction policy applies from line one.
//
// Called by the pipeline logger callback (once per step, before the
// first log line). Thread-safe: the model's map is mutated only here
// and only from callers guarded by the caller's own serialization.
// Because the pipeline runtime creates one logger goroutine per step
// and Go's map access is not safe for concurrent writers, callers
// that may interleave must go through tea.Program.Send instead.
func (m *Model) StepRing(workflow, stepUUID, stepName string) *Ring {
	wf := m.byName[workflow]
	if wf == nil {
		// Defensive: step for an unknown workflow. Return a throwaway
		// ring so logging does not panic; the user will not see these
		// lines.
		return NewRing(fallbackStepRingCapBytes)
	}
	for _, s := range wf.steps {
		if s.uuid == stepUUID {
			return s.log
		}
	}
	// Per-step cap is generous; the global budget enforces the real
	// ceiling across all steps.
	r := NewRing(0)
	m.budget.Register(r)
	wf.steps = append(wf.steps, &stepNode{
		name: stepName,
		uuid: stepUUID,
		log:  r,
	})
	return r
}

// debugTickInterval is the rate at which the TUI refreshes the
// zerolog debug pane and enforces the memory budget. A slow tick is
// fine: zerolog writes are rare compared to step output, and budget
// eviction is cheap enough that a lazy schedule beats re-doing it
// on every log line.
const debugTickInterval = 250 * time.Millisecond

// Init implements tea.Model. Most inputs arrive as Send-ed messages
// from the caller, but the debug tick is internal — we schedule it
// at startup and the handler re-schedules itself to keep the loop
// alive until tea.Quit is issued.
func (m *Model) Init() tea.Cmd {
	return tickDebug()
}

// tickDebug returns a command that will fire a DebugTickMsg after
// the debugTickInterval. The model's Update handler should return
// another tickDebug() after processing the message so the loop
// continues.
func tickDebug() tea.Cmd {
	return tea.Tick(debugTickInterval, func(time.Time) tea.Msg {
		return DebugTickMsg{}
	})
}

// Update implements tea.Model. It dispatches each message to a
// dedicated handler and returns the (possibly updated) model plus
// any command to run next.
//
// The Update method is the single serialization point for model
// state; the caller is responsible for feeding all external events
// through tea.Program.Send so writes are naturally serialized.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeViewports()
		m.viewReady = true
		// Refresh both viewports on resize so reflow picks up new
		// width.
		m.refreshLogView()
		m.refreshMessagesView()
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case WorkflowStateMsg:
		m.handleWorkflowState(msg)
		return m, nil

	case StepStateMsg:
		m.handleStepState(msg)
		return m, nil

	case LogLineMsg:
		m.handleLogLine(msg)
		// If the line belongs to the step currently displayed in the
		// log viewport, refresh so the user sees it immediately. A
		// timer-driven debounce could batch these for very chatty
		// steps; chunk 7 can add that if it becomes an issue.
		if m.logLineBelongsToSelection(msg) {
			m.refreshLogView()
		}
		return m, nil

	case DebugTickMsg:
		// Debug ring is written by zerolog directly; this message is
		// just a redraw trigger. Enforcing the budget here debounces
		// eviction work to roughly the tick rate.
		m.budget.Enforce()
		m.refreshMessagesView()
		// Re-arm the ticker so the loop continues until tea.Quit.
		return m, tickDebug()

	case CancelingMsg:
		m.canceling = true
		return m, nil

	case PipelineDoneMsg:
		m.done = true
		m.doneErr = msg.Err
		return m, nil
	}

	return m, nil
}

// View implements tea.Model. Renders the split-pane layout once the
// first WindowSizeMsg has arrived; before that, the placeholder
// view keeps the program runnable.
func (m *Model) View() tea.View {
	v := renderViewTea(m)
	// AltScreen puts the TUI in the terminal's alternate buffer, so
	// the user's scrollback is preserved and is restored on exit.
	v.AltScreen = true
	return v
}

// handleWorkflowState applies a scheduler.Event to the model's
// workflow bookkeeping.
func (m *Model) handleWorkflowState(msg WorkflowStateMsg) {
	wf := m.byName[msg.Event.Workflow]
	if wf == nil {
		return
	}
	wf.state = msg.Event.State
	wf.err = msg.Event.Err
	// Auto-collapse finished workflows so the tree stays readable in
	// long runs. The user can re-expand with enter.
	if msg.Event.State.Terminal() && msg.Event.State != scheduler.StateFailure {
		wf.expanded = false
	}
}

// handleStepState applies a tracer-sourced step update.
func (m *Model) handleStepState(msg StepStateMsg) {
	wf := m.byName[msg.Workflow]
	if wf == nil || msg.Step == nil || msg.State == nil {
		return
	}
	// Find or create the step node. StepRing also does this lazily,
	// so in practice the step already exists by the time its first
	// state update arrives; the find path is expected.
	var sn *stepNode
	for _, s := range wf.steps {
		if s.uuid == msg.Step.UUID {
			sn = s
			break
		}
	}
	if sn == nil {
		sn = &stepNode{
			name: msg.Step.Name,
			uuid: msg.Step.UUID,
			log:  NewRing(0),
		}
		m.budget.Register(sn.log)
		wf.steps = append(wf.steps, sn)
	}
	st := msg.State.CurrStepState
	sn.exited = st.Exited
	sn.exitCode = st.ExitCode
	sn.skipped = st.Skipped
	sn.oomKill = st.OOMKilled
	if st.Error != nil {
		sn.errText = st.Error.Error()
	}
}

// handleLogLine routes a log line to the appropriate per-step ring.
func (m *Model) handleLogLine(msg LogLineMsg) {
	if msg.Step == nil {
		return
	}
	ring := m.StepRing(msg.Workflow, msg.Step.UUID, msg.Step.Name)
	ring.Append(msg.Line)
}

// handleKey dispatches key presses according to the focus. Tree
// navigation is shared across modes; pane-specific keys (viewport
// scrolling) only fire when that pane is focused.
//
// Two-stage ctrl-c will land in chunk 6 once the sigint plumbing is
// on the cli/exec side; here the key just quits the program.
func (m *Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys that fire regardless of focus.
	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "tab":
		m.cycleFocus()
		return m, nil
	case "L":
		m.focus = FocusMessages
		return m, nil
	}

	// Focus-specific handling.
	switch m.focus {
	case FocusTree:
		return m.handleKeyTree(msg)
	case FocusLog:
		return m.handleKeyViewport(msg, &m.logView)
	case FocusMessages:
		return m.handleKeyViewport(msg, &m.messagesView)
	}
	return m, nil
}

// handleKeyTree handles keys when the tree pane has focus.
func (m *Model) handleKeyTree(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.moveCursor(-1)
		return m, nil
	case "down", "j":
		m.moveCursor(1)
		return m, nil
	case "enter", " ":
		m.activateCursor()
		return m, nil
	case "g", "home":
		m.cursor = 0
		m.refreshLogView()
		return m, nil
	case "G", "end":
		items := m.flatten()
		if len(items) > 0 {
			m.cursor = len(items) - 1
		}
		m.refreshLogView()
		return m, nil
	}
	return m, nil
}

// handleKeyViewport forwards a key to a bubbles viewport and handles
// generic viewport-scope keys (g/G/etc.) consistently with the tree.
// The viewport's own KeyMap covers page-up/page-down; we just
// translate single-key navigation on top of that.
func (m *Model) handleKeyViewport(msg tea.KeyPressMsg, vp *viewport.Model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	// Extra keybinds that the viewport's default KeyMap does not
	// include.
	switch msg.String() {
	case "g", "home":
		vp.GotoTop()
		return m, nil
	case "G", "end":
		vp.GotoBottom()
		return m, nil
	}
	updated, cmd := vp.Update(msg)
	*vp = updated
	return m, cmd
}

// cycleFocus advances the focus ring: tree → log → debug → tree.
func (m *Model) cycleFocus() {
	switch m.focus {
	case FocusTree:
		m.focus = FocusLog
	case FocusLog:
		m.focus = FocusMessages
	case FocusMessages:
		m.focus = FocusTree
	}
}

// moveCursor applies a delta to the tree cursor, clamped to the
// bounds of the currently-flattened items list. Out-of-range deltas
// are silently saturated so holding a key down does not underflow.
func (m *Model) moveCursor(delta int) {
	items := m.flatten()
	if len(items) == 0 {
		m.cursor = 0
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(items) {
		m.cursor = len(items) - 1
	}
	m.refreshLogView()
}

// activateCursor implements the enter-key semantics on the tree. On
// a workflow row, it toggles expanded. On a step row, it focuses the
// log pane so the user can scroll that step's output.
func (m *Model) activateCursor() {
	items := m.flatten()
	if m.cursor < 0 || m.cursor >= len(items) {
		return
	}
	it := items[m.cursor]
	switch it.kind {
	case flatKindWorkflow:
		it.workflow.expanded = !it.workflow.expanded
		// Expanded/collapsed changes the list length; clamp cursor.
		m.moveCursor(0)
	case flatKindStep:
		m.focus = FocusLog
		m.refreshLogView()
	}
}

// logLineBelongsToSelection returns true when the incoming log line
// targets the step currently selected in the tree. Used to decide
// whether a refresh is worth doing; for non-selected steps the
// viewport will pick up the new lines on the next selection change.
func (m *Model) logLineBelongsToSelection(msg LogLineMsg) bool {
	if msg.Step == nil {
		return false
	}
	items := m.flatten()
	if m.cursor < 0 || m.cursor >= len(items) {
		return false
	}
	it := items[m.cursor]
	if it.kind != flatKindStep {
		return false
	}
	return it.workflow.name == msg.Workflow && it.step.uuid == msg.Step.UUID
}
