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

package scheduler

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"

	"go.uber.org/multierr"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
)

// RunFunc is the caller-provided function that executes a single
// workflow. It is called concurrently from the scheduler's worker
// goroutines, so it must be safe to call from any goroutine. The
// context passed to it is a child of the scheduler's context and will
// be canceled when the scheduler's context is canceled or when the
// scheduler is asked to abort.
//
// A non-nil return value marks the workflow as failed. Failing
// workflows cause their dependents to be marked blocked but do not
// stop other independent workflows from running.
type RunFunc func(ctx context.Context, item *builder.Item) error

// Options configures a Scheduler.
type Options struct {
	// Items is the set of workflows to schedule. It is consumed
	// without modification. Items must have unique Workflow.Name
	// values; duplicates cause Run to return an error immediately.
	Items []*builder.Item

	// Run executes one workflow. See RunFunc for details. Required.
	Run RunFunc

	// Events, if non-nil, receives a stream of workflow lifecycle
	// events. The scheduler sends on the channel synchronously from
	// its control goroutine, so a slow consumer will back-pressure the
	// scheduler. Callers that cannot afford that should use a
	// sufficiently buffered channel. The scheduler closes the channel
	// when Run returns.
	Events chan<- Event

	// Parallel is the maximum number of workflows that may be running
	// concurrently. Zero means runtime.NumCPU(). A negative value
	// means unbounded. This is mirrored after the plan's default; a
	// dedicated --parallel flag may be added later.
	Parallel int
}

// Scheduler is the cli-local DAG runner. Construct with New, then call
// Run. A Scheduler is intended for a single run and is not reusable.
type Scheduler struct {
	opts Options
}

// New constructs a Scheduler. It does not start any goroutines and
// does not validate the DAG; validation happens at Run time so that
// the caller can set up channels and sinks before any work is
// attempted.
func New(opts Options) *Scheduler {
	return &Scheduler{opts: opts}
}

// workflowState is the scheduler's private per-item bookkeeping.
type workflowState struct {
	item     *builder.Item
	state    State
	err      error
	depNames []string
}

// Run executes the DAG.
//
// It returns a multierr aggregating the errors of all workflows whose
// run function returned non-nil. Blocked and canceled workflows do
// not contribute to the return value — they are observable only via
// the Events channel — because historically the CLI's exec command
// treated a single non-successful workflow as a single error, and
// decorating the caller with BlockedError/context.Canceled for
// workflows that never actually ran would be noisy without adding
// information.
//
// Run blocks until every workflow is in a terminal state, including
// when ctx is canceled. On ctx cancel, currently-running workflows
// receive a canceled context (through RunFunc) and whatever error
// they return is aggregated; still-pending workflows transition to
// StateCanceled without ever calling RunFunc.
func (s *Scheduler) Run(ctx context.Context) error {
	if s.opts.Run == nil {
		return errors.New("scheduler: Options.Run is required")
	}

	states, err := s.buildStateMap()
	if err != nil {
		// We must still close Events to honor the documented contract.
		if s.opts.Events != nil {
			close(s.opts.Events)
		}
		return err
	}

	parallel := s.opts.Parallel
	if parallel == 0 {
		parallel = runtime.NumCPU()
	}

	// sem is the worker cap. A negative Parallel means unbounded, which
	// we model as a nil semaphore for simplicity.
	var sem chan struct{}
	if parallel > 0 {
		sem = make(chan struct{}, parallel)
	}

	// done carries results from worker goroutines back to the
	// controller. Buffered so workers never block on the send when the
	// controller is busy emitting events.
	done := make(chan workflowDone, len(states))

	var wg sync.WaitGroup
	var aggErr error

	// Initial emission of pending state so the UI has a baseline for
	// every workflow before anything starts.
	for _, name := range s.orderedNames(states) {
		s.emit(Event{Workflow: name, State: StatePending})
	}

	for {
		// 1. Compute ready set. A workflow is ready when every dep is
		//    in StateSuccess. Deps in a failed/blocked/canceled state
		//    cause the workflow itself to become blocked, unless any
		//    dep is still non-terminal in which case we wait.
		for _, name := range s.orderedNames(states) {
			ws := states[name]
			if ws.state != StatePending {
				continue
			}
			ready, blockedBy, wait := s.depCheck(ws, states)
			switch {
			case blockedBy != "":
				ws.state = StateBlocked
				ws.err = &BlockedError{Dependency: blockedBy}
				s.emit(Event{Workflow: name, State: StateBlocked, Err: ws.err})
			case wait:
				// leave as pending
			case ready:
				ws.state = StateReady
				s.emit(Event{Workflow: name, State: StateReady})
			}
		}

		// 2. Launch ready items respecting the worker cap. We do the
		//    acquire BEFORE launching so a burst of ready items does
		//    not create a burst of goroutines that then block on the
		//    semaphore — that would make ctx cancel slower because
		//    every blocked goroutine would need to wake up.
		for _, name := range s.orderedNames(states) {
			ws := states[name]
			if ws.state != StateReady {
				continue
			}
			if ctx.Err() != nil {
				// Don't launch anything new after cancellation.
				break
			}
			if sem != nil {
				select {
				case sem <- struct{}{}:
				case <-ctx.Done():
					// Canceled while waiting for a worker slot; bail
					// out of the launch loop, the main loop below will
					// handle cancellation.
				}
				if ctx.Err() != nil {
					break
				}
			}
			ws.state = StateRunning
			s.emit(Event{Workflow: name, State: StateRunning})

			wg.Add(1)
			go func(item *builder.Item) {
				defer wg.Done()
				defer func() {
					if sem != nil {
						<-sem
					}
				}()
				runErr := s.opts.Run(ctx, item)
				done <- workflowDone{name: item.Workflow.Name, err: runErr}
			}(ws.item)
		}

		// 3. Decide what to do next. If nothing is running and nothing
		//    is pending/ready, we're finished. Otherwise we wait for
		//    either a completion or ctx cancellation.
		pending, running := s.countActive(states)
		if pending == 0 && running == 0 {
			break
		}

		select {
		case d := <-done:
			ws := states[d.name]
			if d.err != nil {
				ws.state = StateFailure
				ws.err = d.err
				aggErr = multierr.Append(aggErr, d.err)
				s.emit(Event{Workflow: d.name, State: StateFailure, Err: d.err})
			} else {
				ws.state = StateSuccess
				s.emit(Event{Workflow: d.name, State: StateSuccess})
			}
		case <-ctx.Done():
			// Cancellation. Mark everything that has not yet started as
			// canceled and let running workflows drain via done.
			for _, name := range s.orderedNames(states) {
				ws := states[name]
				switch ws.state {
				case StatePending, StateReady:
					ws.state = StateCanceled
					s.emit(Event{Workflow: name, State: StateCanceled, Err: ctx.Err()})
				}
			}
			// Drain: collect remaining done entries from running
			// goroutines. We still emit their terminal events.
			wg.Wait()
		drain:
			for {
				select {
				case d := <-done:
					ws := states[d.name]
					if d.err != nil {
						ws.state = StateFailure
						ws.err = d.err
						aggErr = multierr.Append(aggErr, d.err)
						s.emit(Event{Workflow: d.name, State: StateFailure, Err: d.err})
					} else {
						ws.state = StateSuccess
						s.emit(Event{Workflow: d.name, State: StateSuccess})
					}
				default:
					break drain
				}
			}
			if s.opts.Events != nil {
				close(s.opts.Events)
			}
			return aggErr
		}
	}

	wg.Wait()
	if s.opts.Events != nil {
		close(s.opts.Events)
	}
	return aggErr
}

// buildStateMap validates input and produces the initial state map
// keyed by workflow name. It detects duplicate names and unknown
// dependency references.
//
// Note: the builder package drops items with missing dependencies
// before the scheduler sees them (see builder.PipelineBuilder.Build
// and its use of utils.dependsOnExists), so an unknown-dep error here
// is a programming error by the caller rather than a user-facing bug.
func (s *Scheduler) buildStateMap() (map[string]*workflowState, error) {
	states := make(map[string]*workflowState, len(s.opts.Items))
	for _, it := range s.opts.Items {
		name := it.Workflow.Name
		if _, dup := states[name]; dup {
			return nil, fmt.Errorf("scheduler: duplicate workflow name %q", name)
		}
		states[name] = &workflowState{
			item:     it,
			state:    StatePending,
			depNames: append([]string(nil), it.DependsOn...),
		}
	}
	for _, ws := range states {
		for _, d := range ws.depNames {
			if _, ok := states[d]; !ok {
				return nil, fmt.Errorf(
					"scheduler: workflow %q depends on unknown workflow %q",
					ws.item.Workflow.Name, d,
				)
			}
		}
	}
	return states, nil
}

// orderedNames returns the workflow names in a deterministic order,
// derived from the original slice position. The scheduler relies on
// this both for event emission stability and for reproducible test
// behavior.
func (s *Scheduler) orderedNames(states map[string]*workflowState) []string {
	// Rebuilding from s.opts.Items each call is O(n) and the caller
	// invokes this a bounded number of times per DAG tick, so a
	// cached slice would be micro-optimization.
	out := make([]string, 0, len(states))
	for _, it := range s.opts.Items {
		if _, ok := states[it.Workflow.Name]; ok {
			out = append(out, it.Workflow.Name)
		}
	}
	return out
}

// depCheck inspects the deps of ws and returns:
//   - ready=true if all deps are in StateSuccess
//   - blockedBy=<name> if a dep reached a non-success terminal state
//   - wait=true if at least one dep is still non-terminal
//
// At most one of ready/wait/blockedBy indicates "yes". The function
// prioritizes blockedBy over wait so that a workflow whose failed
// dependency is already known can be marked blocked without waiting
// for unrelated deps to finish.
func (s *Scheduler) depCheck(ws *workflowState, states map[string]*workflowState) (ready bool, blockedBy string, wait bool) {
	allSuccess := true
	for _, d := range ws.depNames {
		dep := states[d]
		switch dep.state {
		case StateSuccess:
			// ok
		case StateFailure, StateBlocked, StateCanceled:
			return false, d, false
		default:
			// pending/ready/running — not terminal yet.
			allSuccess = false
			wait = true
		}
	}
	if wait {
		return false, "", true
	}
	return allSuccess, "", false
}

// countActive returns the number of workflows that still have work
// ahead of them (pending/ready → not yet started, running → started
// but not done).
func (s *Scheduler) countActive(states map[string]*workflowState) (pending, running int) {
	for _, ws := range states {
		switch ws.state {
		case StatePending, StateReady:
			pending++
		case StateRunning:
			running++
		}
	}
	return pending, running
}

// emit sends an event if a sink is configured. Sends are synchronous;
// see the docstring on Options.Events.
func (s *Scheduler) emit(ev Event) {
	if s.opts.Events != nil {
		s.opts.Events <- ev
	}
}

// workflowDone is the internal message workers send back to the
// controller goroutine.
type workflowDone struct {
	name string
	err  error
}
