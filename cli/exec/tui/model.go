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

	// Log ring for the zerolog debug tab. Populated by a RingWriter
	// that cli/exec installs as the zerolog destination before the
	// tea program starts.
	debug *Ring

	// budget is the shared cap across all step rings. The debug ring
	// is NOT registered here — it has its own separate cap.
	budget *Budget

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
		byName: make(map[string]*workflowNode, len(workflowNames)),
		debug:  NewRing(DebugLogCapBytes),
		budget: NewBudget(GlobalLogCapBytes),
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

// DebugRing returns the Ring backing the zerolog debug tab. Exposed
// so callers can wrap it in a RingWriter and install as the zerolog
// destination before starting the program.
func (m *Model) DebugRing() *Ring {
	return m.debug
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

// Init implements tea.Model. The TUI does not start any commands on
// init — all inputs arrive as Send-ed messages from the caller.
func (m *Model) Init() tea.Cmd {
	return nil
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
		return m, nil

	case DebugTickMsg:
		// Debug ring is written by zerolog directly; this message is
		// just a redraw trigger. Enforcing the budget here debounces
		// eviction work to roughly the tick rate.
		m.budget.Enforce()
		return m, nil

	case CancellingMsg:
		m.canceling = true
		return m, nil

	case PipelineDoneMsg:
		m.done = true
		m.doneErr = msg.Err
		return m, nil
	}

	return m, nil
}

// View implements tea.Model. Chunk 4 ships a placeholder so the
// program is runnable; real rendering lands in the next chunk.
func (m *Model) View() tea.View {
	// Placeholder. The real view will join a tree panel with a log +
	// debug panel and render a footer.
	return tea.NewView(placeholderView(m))
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

// handleKey is the keybind dispatcher. Chunk 4 only wires the quit
// key so the program is exitable; fuller keybinds land in chunk 5.
func (m *Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		// chunk 6 will turn this into two-stage sigint; for now a
		// single press quits.
		return m, tea.Quit
	}
	return m, nil
}
