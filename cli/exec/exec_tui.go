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

package exec

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"charm.land/bubbletea/v2"

	"go.woodpecker-ci.org/woodpecker/v3/cli/exec/scheduler"
	"go.woodpecker-ci.org/woodpecker/v3/cli/exec/tui"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/logging"
	pipeline_runtime "go.woodpecker-ci.org/woodpecker/v3/pipeline/runtime"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/tracing"
	pipeline_utils "go.woodpecker-ci.org/woodpecker/v3/pipeline/utils"
	"go.woodpecker-ci.org/woodpecker/v3/shared/logger"
)

// sigintExitCode is the conventional exit code for a ctrl-c
// interrupted process (128 + SIGINT's value).
const sigintExitCode = 130

// sigCh is buffered for two pending signals (first cancels, second
// exits). If both arrive before the goroutine drains the first, the
// second one remains queued and triggers os.Exit on the next read.
const sigChanBuffer = 2

// runTUIMode drives the scheduler with an interactive split-pane
// display built on bubbletea + lipgloss. Per-step logs go into
// in-memory rings rendered in the right pane; zerolog output is
// routed into a separate debug ring so diagnostic noise cannot tear
// the alt-screen buffer.
//
// Lifecycle:
//
//  1. Construct the tui.Model seeded with the workflow names from
//     items (so the tree is complete before any workflow actually
//     starts).
//  2. Install a RingWriter as the zerolog destination; defer restore.
//  3. Build a tea.Program with AltScreen enabled (set on View.AltScreen).
//  4. Install a two-stage sigint handler: first signal cancels the
//     pipeline context and flips the model to canceling; second
//     signal exits immediately with code 130.
//  5. Start a goroutine draining scheduler events into p.Send as
//     tui.WorkflowStateMsg; start scheduler.Run in another goroutine,
//     Send a PipelineDoneMsg when it returns.
//  6. p.Run blocks until the user quits or the pipeline completes.
//  7. On exit, flush the debug ring back to the original stderr so
//     nothing diagnostic is lost, restore the zerolog output, and
//     return the aggregated scheduler error.
func runTUIMode(pipelineCtx context.Context, items []*builder.Item, backendEngine backend_types.Backend, preRunMessages string) (retErr error) {
	// The TUI owns the alt-screen buffer. sigint cancels via the
	// pipeline context, not via os.Exit, so the program can flush and
	// restore on shutdown.
	runCtx, cancel := context.WithCancel(pipelineCtx) //nolint:forbidigo // needed for two-stage sigint
	defer cancel()

	// Seed the model with each workflow's full step list from the
	// compiled backend config so every step appears in the tree with
	// a 'pending' glyph before the scheduler starts. The tracer
	// events during execution will then visibly flip each step
	// pending → running → success/failure/skipped.
	seeds := make([]tui.WorkflowSeed, len(items))
	for i, it := range items {
		seed := tui.WorkflowSeed{Name: it.Workflow.Name}
		if it.Config != nil {
			for _, stage := range it.Config.Stages {
				for _, step := range stage.Steps {
					seed.Steps = append(seed.Steps, tui.StepSeed{
						Name: step.Name,
						UUID: step.UUID,
					})
				}
			}
		}
		seeds[i] = seed
	}
	model := tui.NewFromSeeds(seeds)

	// Seed the messages pane with pre-run output (lint warnings,
	// validator output, anything printed before the TUI took over).
	// Each line goes in as its own ring entry so the viewport
	// wraps/scrolls correctly.
	if preRunMessages != "" {
		msgRing := model.MessagesRing()
		for _, line := range strings.SplitAfter(preRunMessages, "\n") {
			if line == "" {
				continue
			}
			msgRing.Append(line)
		}
	}

	// Route zerolog into the messages ring so stderr writes don't
	// tear the alt-screen view. Non-pretty + no-color: the TUI will
	// style what it displays; raw json lines in the ring keep
	// rendering flexibility.
	ringWriter := tui.NewRingWriter(model.MessagesRing())
	restoreLog := logger.SetOutput(ringWriter, false, true)
	defer func() {
		// Order is critical: restore the logger first so any log
		// calls emitted during the flush itself (unlikely but
		// possible) go to real stderr rather than back into the
		// ring we're draining. Then flush the ring content to
		// stderr so diagnostics survive the alt-screen tear-down.
		// Finally drain any carried-over fragment from the writer.
		//
		// This runs on any return path from runTUIMode — success,
		// error, or panic — because it is deferred. That is the
		// whole point: if prog.Run panics, we still want the user
		// to see what zerolog captured on the way down.
		restoreLog()
		ringWriter.Flush()
		flushMessagesRingToStderr(model.MessagesRing())
	}()

	prog := tea.NewProgram(model,
		tea.WithContext(runCtx),
	)

	// Two-stage sigint handler. The first signal cancels the pipeline
	// context — which the scheduler picks up and propagates to every
	// running workflow — and flips the model to 'canceling'. The
	// second signal exits immediately; cleanup is best-effort but the
	// user has chosen speed over neatness by pressing ctrl-c twice.
	sigCh := make(chan os.Signal, sigChanBuffer)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	go func() {
		count := 0
		for range sigCh {
			count++
			switch count {
			case 1:
				cancel()
				prog.Send(tui.CancelingMsg{})
			default:
				os.Exit(sigintExitCode)
			}
		}
	}()

	// Scheduler events goroutine: forward each event to the tea
	// program as a WorkflowStateMsg. The scheduler closes the events
	// channel when it returns, which terminates this loop.
	events := make(chan scheduler.Event, schedulerEventBuffer)
	eventsDone := make(chan struct{})
	go func() {
		defer close(eventsDone)
		for ev := range events {
			prog.Send(tui.WorkflowStateMsg{Event: ev})
		}
	}()

	runFunc := tuiRunFunc(prog, backendEngine)

	sched := scheduler.New(scheduler.Options{
		Items:  items,
		Run:    runFunc,
		Events: events,
	})

	// Scheduler in its own goroutine so p.Run can block on the tea
	// event loop in the main goroutine. When scheduler.Run returns,
	// send PipelineDoneMsg so the model can transition to its final
	// state; the user then chooses when to quit.
	schedDone := make(chan error, 1)
	go func() {
		err := sched.Run(runCtx)
		schedDone <- err
		prog.Send(tui.PipelineDoneMsg{Err: err})
	}()

	if _, err := prog.Run(); err != nil {
		retErr = fmt.Errorf("tui program: %w", err)
	}

	// Make sure all derived goroutines have wound down before we
	// return and the deferred restore/flush runs. cancel() propagates
	// through runCtx so scheduler workflows tear down cleanly.
	cancel()
	<-eventsDone

	var execErr error
	select {
	case execErr = <-schedDone:
	default:
		// User quit before the scheduler finished. Wait for it.
		execErr = <-schedDone
	}

	if retErr != nil {
		return retErr
	}
	return execErr
}

// tuiRunFunc returns a scheduler.RunFunc that executes a workflow
// with tracer + logger hooks forwarding step state and log lines to
// the tea program as messages.
//
// Constructed once per runTUIMode call and captured by closure; the
// returned func is safe to invoke from multiple goroutines because
// each call builds its own per-workflow tracer/logger.
func tuiRunFunc(prog *tea.Program, backendEngine backend_types.Backend) scheduler.RunFunc {
	return func(runCtx context.Context, item *builder.Item) error {
		workflow := item.Workflow.Name

		// Per-workflow tracer: invoke DefaultTracer first so env vars
		// still get populated for the running step, then forward the
		// state update to the tea program so the tree can reflect
		// step-level transitions (exited / skipped / etc).
		tracer := tracing.TraceFunc(func(s *state.State) error {
			if err := tracing.DefaultTracer.Trace(s); err != nil {
				return err
			}
			prog.Send(tui.StepStateMsg{
				Workflow: workflow,
				Step:     s.CurrStep,
				State:    s,
			})
			return nil
		})

		// Per-workflow logger: one goroutine per step reads from rc
		// and forwards each complete line as a LogLineMsg. The tea
		// program serializes appends via the model's Update.
		logger := logging.Logger(func(step *backend_types.Step, rc io.ReadCloser) error {
			lw := &tuiStepWriter{prog: prog, workflow: workflow, step: step}
			return pipeline_utils.CopyLineByLine(lw, rc, pipeline.MaxLogLineLength)
		})

		return pipeline_runtime.New(item.Config, backendEngine,
			pipeline_runtime.WithContext(runCtx),
			pipeline_runtime.WithTracer(tracer),
			pipeline_runtime.WithLogger(logger),
			pipeline_runtime.WithDescription(map[string]string{
				"CLI": "exec",
			}),
		).Run(runCtx)
	}
}

// tuiStepWriter is the io.Writer that CopyLineByLine feeds. Each
// Write corresponds to one logical log line, which we forward to the
// tea program as a LogLineMsg.
//
// Unlike LineWriter, this writer does not emit to stderr — the TUI
// owns that channel. The line is stored in the model's ring and
// rendered by the log viewport.
type tuiStepWriter struct {
	prog     *tea.Program
	workflow string
	step     *backend_types.Step
}

// Write implements io.Writer. Returns len(p) per the io.Writer
// contract so upstream CopyLineByLine accounting stays correct.
func (w *tuiStepWriter) Write(p []byte) (int, error) {
	w.prog.Send(tui.LogLineMsg{
		Workflow: w.workflow,
		Step:     w.step,
		Line:     string(p),
	})
	return len(p), nil
}

// Close implements io.Closer. No-op: the Writer doesn't own any
// resources that need releasing.
func (w *tuiStepWriter) Close() error { return nil }

// flushMessagesRingToStderr writes the accumulated debug ring contents
// to os.Stderr after the TUI has exited. This preserves any zerolog
// output the user might want to see (errors, warnings) that was
// collected while the alt-screen was active.
func flushMessagesRingToStderr(ring *tui.Ring) {
	lines, truncated := ring.Snapshot()
	if truncated == 0 && len(lines) == 0 {
		return
	}
	if truncated > 0 {
		fmt.Fprintf(os.Stderr, "[… %d diagnostic line(s) truncated]\n", truncated)
	}
	for _, ln := range lines {
		// The ring retains trailing newlines, so Write, not Writeln.
		_, _ = os.Stderr.WriteString(ln)
	}
}
