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
	"go.woodpecker-ci.org/woodpecker/v3/cli/exec/scheduler"
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
)

// Message types sent into the TUI via tea.Program.Send. Producers
// (the scheduler's event consumer, the pipeline tracer, the pipeline
// logger) construct these; the model's Update method handles them.
//
// Keeping the messages as data-only structs means we can unit-test
// the model by feeding synthetic messages — no need to spin up a
// real pipeline.

// WorkflowStateMsg announces a workflow-level state transition. It
// is a direct translation of scheduler.Event for ingestion into the
// tea program's event loop.
type WorkflowStateMsg struct {
	Event scheduler.Event
}

// StepStateMsg announces a step-level state transition, sourced from
// the pipeline runtime's tracer. Workflow is attached by the
// producer because the tracer itself does not know which workflow it
// is tracing — the TUI needs it to route the update to the correct
// tree node.
type StepStateMsg struct {
	Workflow string
	Step     *backend_types.Step
	State    *state.State
}

// LogLineMsg carries one line of step output. One message per
// logical line; the model appends to the appropriate per-step ring.
type LogLineMsg struct {
	Workflow string
	Step     *backend_types.Step
	Line     string
}

// DebugTickMsg is emitted on a timer to tell the model to refresh
// its view of the zerolog debug ring. The ring itself is written to
// directly by zerolog; this message exists only so the model can
// batch redraws rather than re-rendering on every zerolog line.
type DebugTickMsg struct{}

// PipelineDoneMsg is emitted when the scheduler has returned. It
// carries the final aggregate error so the model can transition to
// its final display state (summary, footer text, quit key hint).
type PipelineDoneMsg struct {
	Err error
}

// CancellingMsg is emitted by the signal handler on the first
// ctrl-c, so the model can flip its status to "canceling…" while
// the actual pipeline context cancellation propagates through the
// runtimes.
type CancellingMsg struct{}
