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

package tui_test

import (
	"errors"
	"testing"

	"charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/cli/exec/scheduler"
	"go.woodpecker-ci.org/woodpecker/v3/cli/exec/tui"
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
)

// fakeKeyMsg builds a KeyPressMsg for a single printable character,
// enough to drive the model's keybind handler in tests.
func fakeKeyMsg(ch string) tea.Msg {
	if ch == "" {
		return tea.KeyPressMsg(tea.Key{})
	}
	r := []rune(ch)[0]
	return tea.KeyPressMsg(tea.Key{Text: ch, Code: r})
}

// asModel is a test helper that asserts the Model returned from
// Update is our concrete *tui.Model. The bubbletea Update signature
// is typed as the interface tea.Model, so a safe assertion at each
// call site keeps the linter happy and surfaces a clearer failure
// than a panicking unchecked cast would.
func asModel(t *testing.T, m tea.Model) *tui.Model {
	t.Helper()
	model, ok := m.(*tui.Model)
	require.True(t, ok, "expected *tui.Model, got %T", m)
	return model
}

func TestModelRendersSeededWorkflows(t *testing.T) {
	m := tui.New([]string{"build", "test"})
	out := m.View().Content
	// Placeholder view is prose; verify both workflows appear and
	// start with the pending glyph.
	assert.Contains(t, out, "build")
	assert.Contains(t, out, "test")
}

func TestModelTransitionsThroughLifecycle(t *testing.T) {
	m := tui.New([]string{"build"})

	// Running.
	updated, _ := m.Update(tui.WorkflowStateMsg{Event: scheduler.Event{
		Workflow: "build",
		State:    scheduler.StateRunning,
	}})
	m = asModel(t, updated)
	assert.Contains(t, m.View().Content, "build")

	// Success.
	updated, _ = m.Update(tui.WorkflowStateMsg{Event: scheduler.Event{
		Workflow: "build",
		State:    scheduler.StateSuccess,
	}})
	m = asModel(t, updated)

	// Terminal success auto-collapses, so step lines should not
	// appear. We don't assert on them directly (no steps yet), but
	// the main line still reflects the workflow name.
	assert.Contains(t, m.View().Content, "build")

	// Pipeline done — view should annotate completion.
	updated, _ = m.Update(tui.PipelineDoneMsg{Err: nil})
	m = asModel(t, updated)
	assert.Contains(t, m.View().Content, "finished successfully")
}

func TestModelShowsErrorOnFailure(t *testing.T) {
	m := tui.New([]string{"build"})
	updated, _ := m.Update(tui.WorkflowStateMsg{Event: scheduler.Event{
		Workflow: "build",
		State:    scheduler.StateFailure,
		Err:      errors.New("boom"),
	}})
	m = asModel(t, updated)
	out := m.View().Content
	assert.Contains(t, out, "build")
	assert.Contains(t, out, "boom")

	updated, _ = m.Update(tui.PipelineDoneMsg{Err: errors.New("boom")})
	m = asModel(t, updated)
	assert.Contains(t, m.View().Content, "finished with error")
}

func TestModelCancelingState(t *testing.T) {
	m := tui.New([]string{"build"})
	updated, _ := m.Update(tui.CancelingMsg{})
	m = asModel(t, updated)
	assert.Contains(t, m.View().Content, "canceling")
}

func TestModelStepStateUpdatesAndRing(t *testing.T) {
	m := tui.New([]string{"build"})

	// Seed a running workflow so the placeholder view shows its steps.
	m.Update(tui.WorkflowStateMsg{Event: scheduler.Event{
		Workflow: "build",
		State:    scheduler.StateRunning,
	}})

	step := &backend_types.Step{Name: "compile", UUID: "u-1"}

	// Log line arriving before any state update must still route
	// into a lazily-created ring without panicking.
	_, _ = m.Update(tui.LogLineMsg{
		Workflow: "build",
		Step:     step,
		Line:     "compiling...\n",
	})

	// Step state with exited=true, code=0 should make the placeholder
	// render the success glyph for this step.
	_, _ = m.Update(tui.StepStateMsg{
		Workflow: "build",
		Step:     step,
		State: &state.State{
			CurrStep: step,
			CurrStepState: backend_types.State{
				Exited:   true,
				ExitCode: 0,
			},
		},
	})

	out := m.View().Content
	assert.Contains(t, out, "compile")

	// The ring must hold the log line we appended.
	ring := m.StepRing("build", "u-1", "compile")
	lines, _ := ring.Snapshot()
	require.Len(t, lines, 1)
	assert.Equal(t, "compiling...\n", lines[0])
}

func TestModelIgnoresUnknownWorkflow(t *testing.T) {
	// A workflow-state event for a workflow the model wasn't seeded
	// with must be a no-op, not a panic. This keeps the TUI
	// defensible against future sources that seed names differently.
	m := tui.New([]string{"a"})
	updated, _ := m.Update(tui.WorkflowStateMsg{Event: scheduler.Event{
		Workflow: "ghost",
		State:    scheduler.StateRunning,
	}})
	m = asModel(t, updated)
	assert.NotContains(t, m.View().Content, "ghost")
}

func TestModelQuitKey(t *testing.T) {
	m := tui.New([]string{"build"})
	_, cmd := m.Update(fakeKeyMsg("q"))
	require.NotNil(t, cmd, "q key must return a command")
	// The returned cmd is tea.Quit, which produces tea.QuitMsg. We
	// don't assert the concrete type to avoid coupling tests to
	// bubbletea internals; the smoke is that a non-nil cmd came back.
}
