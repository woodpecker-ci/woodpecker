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

package pipeline

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/cli/internal/tui"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker/mocks"
)

// testViewer returns a viewer of pipeline 2 of repo 1 and the messages it sent.
func testViewer(client woodpecker.Client) (*pipelineViewer, *[]tea.Msg) {
	var mu sync.Mutex
	var sent []tea.Msg
	return &pipelineViewer{
		client: client,
		repoID: 1,
		number: 2,
		send: func(msg tea.Msg) {
			mu.Lock()
			defer mu.Unlock()
			sent = append(sent, msg)
		},
		loaded: make(map[int64]bool),
	}, &sent
}

// streamLog lets the mocked log stream of step 3 send the entries and end with err.
func streamLog(client *mocks.MockClient, err error, entries ...*woodpecker.LogEntry) {
	client.On("StepLogStream", mock.Anything, int64(1), int64(2), int64(3), mock.Anything).
		Run(func(args mock.Arguments) {
			handle, _ := args.Get(4).(func(*woodpecker.LogEntry))
			for _, entry := range entries {
				handle(entry)
			}
		}).
		Return(err).Once()
}

func TestPipelineViewerLoadLog(t *testing.T) {
	line0 := &woodpecker.LogEntry{StepID: 3, Line: 0}
	line1 := &woodpecker.LogEntry{StepID: 3, Line: 1}
	stored := func(entries ...*woodpecker.LogEntry) tea.Msg {
		return tui.StepLogMsg{StepID: 3, Entries: entries}
	}

	t.Run("follows the log of a running step and completes it from the stored log", func(t *testing.T) {
		client := mocks.NewMockClient(t)
		// the stream ends fine, but a line got lost on the way
		streamLog(client, nil, line0)
		client.On("StepLogEntries", int64(1), int64(2), int64(3)).Return([]*woodpecker.LogEntry{line0, line1}, nil).Once()
		viewer, sent := testViewer(client)

		viewer.loadLog(t.Context(), &woodpecker.Step{ID: 3, State: "running"})

		assert.Equal(t, []tea.Msg{tui.LogMsg{line0}, stored(line0, line1)}, *sent)
	})

	t.Run("fetches the stored log if the step finished meanwhile", func(t *testing.T) {
		client := mocks.NewMockClient(t)
		streamLog(client, errors.New("step not running (anymore)"))
		client.On("StepLogEntries", int64(1), int64(2), int64(3)).Return([]*woodpecker.LogEntry{line0, line1}, nil).Once()
		viewer, sent := testViewer(client)

		viewer.loadLog(t.Context(), &woodpecker.Step{ID: 3, State: "pending"})

		assert.Equal(t, []tea.Msg{stored(line0, line1)}, *sent)
	})

	t.Run("fetches the stored log of a finished step only once", func(t *testing.T) {
		client := mocks.NewMockClient(t)
		client.On("StepLogEntries", int64(1), int64(2), int64(3)).Return([]*woodpecker.LogEntry{line0}, nil).Once()
		viewer, sent := testViewer(client)

		viewer.loadLog(t.Context(), &woodpecker.Step{ID: 3, State: "failure"})
		viewer.loadLog(t.Context(), &woodpecker.Step{ID: 3, State: "failure"})

		assert.Equal(t, []tea.Msg{stored(line0)}, *sent)
	})

	t.Run("asks for no log of a step that never ran", func(t *testing.T) {
		viewer, sent := testViewer(mocks.NewMockClient(t))

		viewer.loadLog(t.Context(), &woodpecker.Step{ID: 3, State: "skipped"})
		viewer.loadLog(t.Context(), &woodpecker.Step{ID: 4, State: "canceled"})

		assert.Empty(t, *sent)
	})

	t.Run("reports a failure and tries again on the next selection", func(t *testing.T) {
		client := mocks.NewMockClient(t)
		client.On("StepLogEntries", int64(1), int64(2), int64(3)).Return(nil, errors.New("server down")).Once()
		client.On("StepLogEntries", int64(1), int64(2), int64(3)).Return([]*woodpecker.LogEntry{line0}, nil).Once()
		viewer, sent := testViewer(client)

		step := &woodpecker.Step{ID: 3, Name: "build", State: "success"}
		viewer.loadLog(t.Context(), step)
		viewer.loadLog(t.Context(), step)

		assert.Equal(t, []tea.Msg{
			tui.MessageMsg("could not load log of step build: server down"),
			stored(line0),
		}, *sent)
	})

	t.Run("stays quiet if the view got closed", func(t *testing.T) {
		client := mocks.NewMockClient(t)
		streamLog(client, context.Canceled)
		viewer, sent := testViewer(client)

		ctx, cancel := context.WithCancelCause(t.Context())
		cancel(nil)
		viewer.loadLog(ctx, &woodpecker.Step{ID: 3, State: "running"})

		assert.Empty(t, *sent)
	})
}

func TestPipelineViewerWatch(t *testing.T) {
	oldInterval := viewRefreshInterval
	viewRefreshInterval = time.Millisecond
	t.Cleanup(func() { viewRefreshInterval = oldInterval })

	running := []*woodpecker.Workflow{{ID: 1, State: "running"}}
	done := []*woodpecker.Workflow{{ID: 1, State: "success"}}

	client := mocks.NewMockClient(t)
	client.On("Pipeline", int64(1), int64(2)).Return(&woodpecker.Pipeline{
		Status:    "running",
		Workflows: running,
		Errors:    []*woodpecker.PipelineError{{Type: "deprecation", Message: "old syntax"}},
	}, nil).Twice()
	client.On("Pipeline", int64(1), int64(2)).Return(nil, errors.New("server down")).Once()
	client.On("Pipeline", int64(1), int64(2)).Return(&woodpecker.Pipeline{Status: "success", Workflows: done}, nil).Once()
	viewer, sent := testViewer(client)

	// returns by itself as the pipeline finishes
	viewer.watch(t.Context())

	assert.Equal(t, []tea.Msg{
		tui.MessageMsg("deprecation: old syntax"),
		tui.StatusMsg("running"),
		tui.WorkflowsMsg(running),
		tui.StatusMsg("running"),
		tui.WorkflowsMsg(running),
		tui.MessageMsg("could not load pipeline: server down"),
		tui.StatusMsg("success"),
		tui.WorkflowsMsg(done),
	}, *sent)
}
