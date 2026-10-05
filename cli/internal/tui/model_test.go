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
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

func testWorkflows() WorkflowsMsg {
	return WorkflowsMsg{
		{ID: 1, Name: "build", State: "running", Started: 1, Children: []*woodpecker.Step{
			{ID: 11, Name: "clone", State: "success", Started: 1, Stopped: 2},
			{ID: 12, Name: "compile", State: "running", Started: 2},
			{ID: 13, Name: "package", State: "pending"},
		}},
		{ID: 2, Name: "deploy", State: "pending", Children: []*woodpecker.Step{
			{ID: 21, Name: "upload", State: "pending"},
		}},
	}
}

func key(k string) tea.KeyPressMsg {
	switch k {
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	}
	return tea.KeyPressMsg{Code: rune(k[0]), Text: k}
}

// send updates the model and runs the returned command, if any.
func send(m *Model, msg tea.Msg) tea.Msg {
	_, cmd := m.Update(msg)
	if cmd == nil {
		return nil
	}
	return cmd()
}

func TestWorkflowsMsgReplacesWorkflowsByID(t *testing.T) {
	m := New()
	send(m, testWorkflows())

	send(m, WorkflowsMsg{{ID: 2, Name: "deploy", State: "skipped"}})
	send(m, WorkflowsMsg{{ID: 3, Name: "notify", State: "pending"}})

	require.Len(t, m.workflows, 3)
	assert.Equal(t, "running", m.workflows[0].State)
	assert.Equal(t, "skipped", m.workflows[1].State)
	assert.Equal(t, "notify", m.workflows[2].Name)
}

func TestSelectionFollowsRunningStepUntilUserMoves(t *testing.T) {
	var selected []string
	m := New()
	m.OnSelect = func(step *woodpecker.Step) { selected = append(selected, step.Name) }

	send(m, testWorkflows())
	assert.EqualValues(t, 12, m.selected, "the running step is selected")

	workflows := testWorkflows()
	workflows[0].Children[1].State = "success"
	workflows[0].Children[2].State = "running"
	send(m, workflows)
	assert.EqualValues(t, 13, m.selected, "the selection follows the running step")

	send(m, key("up"))
	workflows = testWorkflows()
	workflows[1].Children[0].State = "running"
	workflows[0].Children[1].State = "success"
	send(m, workflows)
	assert.EqualValues(t, 12, m.selected, "the selection stays where the user moved it")

	assert.Equal(t, []string{"compile", "package", "compile"}, selected)
}

func TestWorkflowRowFoldsItsSteps(t *testing.T) {
	m := New()
	send(m, testWorkflows())
	require.Len(t, m.rows(), 6)
	assert.EqualValues(t, 12, m.selected)

	send(m, key("g"))
	send(m, key("enter"))
	assert.Len(t, m.rows(), 3, "the steps of the first workflow are hidden")
	assert.EqualValues(t, 12, m.selected, "the log of the last selected step stays")

	send(m, key("enter"))
	assert.Len(t, m.rows(), 6)
}

func TestEnterOnStepFocusesItsLog(t *testing.T) {
	m := New()
	send(m, testWorkflows())

	send(m, key("enter"))
	assert.Equal(t, paneLog, m.focus)

	send(m, key("tab"))
	assert.Equal(t, paneSteps, m.focus, "the empty messages pane is skipped")
}

func TestQuitCancelsRunningPipelineFirst(t *testing.T) {
	canceled := 0
	m := New()
	m.OnCancel = func() { canceled++ }
	send(m, testWorkflows())

	assert.Nil(t, send(m, key("q")))
	assert.Equal(t, 1, canceled)

	assert.Equal(t, tea.QuitMsg{}, send(m, key("q")), "the second request quits")
	assert.Equal(t, 1, canceled)
}

func TestQuitWithoutRunningPipeline(t *testing.T) {
	m := New()
	m.OnCancel = func() { t.Error("nothing to cancel") }
	send(m, WorkflowsMsg{{ID: 1, Name: "build", State: "success"}})

	assert.Equal(t, tea.QuitMsg{}, send(m, key("q")))
}

func TestLogKeepsOnlyTheLastLines(t *testing.T) {
	m := New()
	for i := range maxLogLines + 10 {
		send(m, LogMsg{{StepID: 1, Line: i}})
	}

	require.Len(t, m.logs[1], maxLogLines)
	assert.Equal(t, 10, m.logs[1][0].Line)
}

func TestLogIgnoresLinesItAlreadyHas(t *testing.T) {
	m := New()
	send(m, LogMsg{{StepID: 1, Line: 0}, {StepID: 1, Line: 1}})
	send(m, LogMsg{{StepID: 1, Line: 0}, {StepID: 1, Line: 1}, {StepID: 1, Line: 2}, {StepID: 2, Line: 0}})

	assert.Len(t, m.logs[1], 3)
	assert.Len(t, m.logs[2], 1)
}

func TestStepLogMsgReplacesTheLogOfAStep(t *testing.T) {
	m := New()
	send(m, testWorkflows())
	// line 1 got lost
	send(m, LogMsg{{StepID: 12, Line: 0}, {StepID: 12, Line: 2}, {StepID: 13, Line: 0}})
	send(m, tickMsg{})

	send(m, StepLogMsg{StepID: 12, Entries: []*woodpecker.LogEntry{{Line: 0}, {Line: 1}, {Line: 2}}})

	require.Len(t, m.logs[12], 3)
	assert.Equal(t, 1, m.logs[12][1].Line)
	assert.Len(t, m.logs[13], 1, "other steps keep their log")
	assert.True(t, m.logChanged, "the shown log gets redrawn")
}

func TestCleanLine(t *testing.T) {
	tests := map[string]string{
		"plain text":                             "plain text",
		"trailing newline\r\n":                   "trailing newline",
		"\x1b[1;32mcolored\x1b[0m":               "\x1b[1;32mcolored\x1b[0m",
		"clear\x1b[2J screen\x1b[H":              "clear screen",
		"\x1b]0;window title\x07text":            "text",
		"bell\a and backspace\b":                 "bell and backspace",
		"progress 10%\rprogress 50%\rdone":       "done",
		"a\tb":                                   "a    b",
		"wide ✓ 日本":                              "wide ✓ 日本",
		"\x1b[?25lhidden cursor\x1b[?25h":        "hidden cursor",
		"link \x1b]8;;http://x\x1b\\x\x1b]8;;\a": "link x",
	}
	for in, want := range tests {
		assert.Equal(t, want, cleanLine([]byte(in)), "%q", in)
	}
}

func TestRenderFitsWindow(t *testing.T) {
	m := New()
	send(m, testWorkflows())
	for i := range 100 {
		send(m, LogMsg{{StepID: 12, Line: i, Time: int64(i), Data: []byte(strings.Repeat("long log line ", 20))}})
	}
	send(m, tickMsg{})

	check := func(width, height int) {
		send(m, tea.WindowSizeMsg{Width: width, Height: height})
		lines := strings.Split(m.render(), "\n")
		assert.Len(t, lines, height, "%dx%d", width, height)
		for _, line := range lines {
			assert.LessOrEqual(t, ansi.StringWidth(line), width, "%dx%d: %q", width, height, line)
		}
	}

	for _, size := range [][2]int{{200, 60}, {100, 24}, {80, 24}, {60, 15}, {40, 10}} {
		check(size[0], size[1])
	}

	send(m, MessageMsg("a warning\nand another one"))
	for _, size := range [][2]int{{200, 60}, {100, 24}, {80, 24}, {60, 15}, {40, 10}} {
		check(size[0], size[1])
	}
}

func TestRenderShowsStepsAndLog(t *testing.T) {
	m := New()
	send(m, tea.WindowSizeMsg{Width: 100, Height: 20})
	workflows := testWorkflows()
	workflows[0].Children[1] = &woodpecker.Step{ID: 12, Name: "compile", State: "failure", Started: 2, Stopped: 3, ExitCode: 2}
	send(m, workflows)
	send(m, key("down"))
	send(m, LogMsg{{StepID: 12, Line: 0, Time: 3, Data: []byte("gcc: error")}})
	send(m, tickMsg{})

	view := ansi.Strip(m.render())
	for _, want := range []string{
		"▾ ● build", "✓ clone", "✗ compile", "00:01", "○ package", "▾ ○ deploy",
		"Step Logs: build / compile", "    1    3s gcc: error", "✗ Exit Code 2",
	} {
		assert.Contains(t, view, want)
	}
}

func TestRenderShowsWhatArrivedBeforeTheWindowSize(t *testing.T) {
	m := New()
	send(m, testWorkflows())
	send(m, MessageMsg("a warning"))
	for i := range 100 {
		send(m, LogMsg{{StepID: 12, Line: i, Data: fmt.Appendf(nil, "log line %d", i)}})
	}
	send(m, tickMsg{})
	send(m, tea.WindowSizeMsg{Width: 100, Height: 20})

	view := ansi.Strip(m.render())
	assert.Contains(t, view, "a warning")
	assert.Contains(t, view, "log line 99", "the end of the log is shown")
}

func TestRenderShowsPipelineStatus(t *testing.T) {
	m := New()
	send(m, tea.WindowSizeMsg{Width: 100, Height: 20})
	assert.Contains(t, ansi.Strip(m.render()), "○ pending", "nothing known yet")

	send(m, WorkflowsMsg{{ID: 1, State: "success"}, {ID: 2, State: "failure"}, {ID: 3, State: "skipped"}})
	assert.Contains(t, ansi.Strip(m.render()), "✗ failure", "derived from the workflows")

	send(m, StatusMsg("error"))
	assert.Contains(t, ansi.Strip(m.render()), "✗ error", "the status set wins")
}

func TestRenderLogHints(t *testing.T) {
	tests := map[string]*woodpecker.Step{
		"This step hasn't started yet.": {ID: 1, State: "pending"},
		"This step has been skipped.":   {ID: 1, State: "skipped"},
		"This step has been canceled.":  {ID: 1, State: "canceled"},
		"Loading…":                      {ID: 1, State: "running", Started: 1},
		"No logs":                       {ID: 1, State: "success", Started: 1, Stopped: 2},
	}
	for want, step := range tests {
		m := New()
		send(m, tea.WindowSizeMsg{Width: 100, Height: 20})
		send(m, WorkflowsMsg{{ID: 9, Name: "w", State: "running", Children: []*woodpecker.Step{step}}})
		assert.Contains(t, m.render(), want, fmt.Sprintf("%+v", step))
	}
}

func TestSummary(t *testing.T) {
	m := New()
	send(m, MessageMsg("a warning"))
	send(m, WorkflowsMsg{{ID: 1, Name: "build", State: "failure", Error: "no space left", Children: []*woodpecker.Step{
		{ID: 11, Name: "clone", State: "success"},
		{ID: 12, Name: "compile", State: "failure", ExitCode: 2},
		{ID: 13, Name: "upload", State: "failure", Error: "image not found"},
	}}})

	summary := ansi.Strip(m.Summary())
	for _, want := range []string{
		"a warning\n", "✗ build", "no space left", "✓ clone", "✗ compile  (exit code 2)", "✗ upload  (image not found)",
	} {
		assert.Contains(t, summary, want)
	}
}
