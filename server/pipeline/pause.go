// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pipeline

import (
	"context"
	"sync"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// PauseRegistry connects the agent waiting in a paused workflow (rpc Pause)
// with the user decision coming from the API (resume / decline).
//
// The decision is kept until the agent picks it up, so an agent that
// reconnects after the user decided still gets the decision.
// It is held in memory only: after a server restart the user has to decide again.
type PauseRegistry struct {
	mu      sync.Mutex
	entries map[int64]*pauseEntry
}

type pauseEntry struct {
	stepUUID string
	decided  bool
	resume   bool
	ch       chan bool // buffered(1), receives the decision
}

// NewPauseRegistry returns an empty PauseRegistry.
func NewPauseRegistry() *PauseRegistry {
	return &PauseRegistry{entries: map[int64]*pauseEntry{}}
}

// Pauses is the registry used by the server.
var Pauses = NewPauseRegistry()

// Wait blocks until a decision for the paused workflow is made or ctx is done.
// A decision made before (e.g. while the agent was reconnecting) is returned right away.
func (r *PauseRegistry) Wait(ctx context.Context, workflowID int64, stepUUID string) (resume bool, err error) {
	r.mu.Lock()
	entry, ok := r.entries[workflowID]
	switch {
	case ok && entry.decided && (entry.stepUUID == "" || entry.stepUUID == stepUUID):
		// decision already made for this pause.
		delete(r.entries, workflowID)
		r.mu.Unlock()
		return entry.resume, nil
	case !ok || entry.stepUUID != stepUUID:
		// first wait for this pause (an entry of an older pause is replaced).
		entry = &pauseEntry{stepUUID: stepUUID, ch: make(chan bool, 1)}
		r.entries[workflowID] = entry
	}
	ch := entry.ch
	r.mu.Unlock()

	select {
	case resume := <-ch:
		r.mu.Lock()
		if r.entries[workflowID] == entry {
			delete(r.entries, workflowID)
		}
		r.mu.Unlock()
		return resume, nil
	case <-ctx.Done():
		// keep the entry: the agent will call Wait again after reconnecting.
		return false, ctx.Err()
	}
}

// Take returns and removes a decision that was already made for this pause, if any.
func (r *PauseRegistry) Take(workflowID int64, stepUUID string) (resume, ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, exists := r.entries[workflowID]
	if !exists || !entry.decided || (entry.stepUUID != "" && entry.stepUUID != stepUUID) {
		return false, false
	}
	delete(r.entries, workflowID)
	return entry.resume, true
}

// Decide stores the decision for a paused workflow and wakes up the waiting agent, if any.
// It returns false if a decision was already made.
func (r *PauseRegistry) Decide(workflowID int64, resume bool) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.entries[workflowID]
	if !ok {
		// no agent is waiting right now (e.g. after a server restart): keep the
		// decision for the next Wait call of this workflow.
		entry = &pauseEntry{ch: make(chan bool, 1)}
		r.entries[workflowID] = entry
	}
	if entry.decided {
		return false
	}
	entry.decided = true
	entry.resume = resume
	entry.ch <- resume
	return true
}

// Forget drops the state of a workflow, e.g. when it finished.
func (r *PauseRegistry) Forget(workflowID int64) {
	r.mu.Lock()
	delete(r.entries, workflowID)
	r.mu.Unlock()
}

// ErrWorkflowNotPaused is returned when a workflow that is not paused gets resumed or declined.
var ErrWorkflowNotPaused = &ErrBadRequest{Msg: "workflow is not paused"}

// DecidePause resumes (resume=true) or declines (resume=false) a paused workflow.
// A declined workflow is stopped by its agent and reported as killed.
func DecidePause(_store store.Store, workflow *model.Workflow, resume bool) (*model.Workflow, error) {
	if workflow.State != model.StatusPaused {
		return nil, ErrWorkflowNotPaused
	}
	if !Pauses.Decide(workflow.ID, resume) {
		return nil, ErrWorkflowNotPaused
	}
	if resume {
		// a declined workflow keeps the paused state until the agent reports it as done.
		workflow.State = model.StatusRunning
		if err := _store.WorkflowUpdate(workflow); err != nil {
			return nil, err
		}
	}
	return workflow, nil
}
