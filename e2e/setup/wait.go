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

//go:build test

package setup

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

const (
	defaultTimeout  = 30 * time.Second
	defaultRetry    = 3
	shortTimeout    = 10 * time.Second
	defaultInterval = 100 * time.Millisecond
)

// isTerminal returns true if the status is a final (non-running) state.
func isTerminal(s model.StatusValue) bool {
	switch s {
	case model.StatusSuccess, model.StatusFailure, model.StatusKilled,
		model.StatusError, model.StatusDeclined, model.StatusCanceled:
		return true
	}
	return false
}

// WaitForPipeline polls the store until the pipeline with the given ID reaches
// a terminal status, then returns it. Fails the test if timeout is exceeded.
func WaitForPipeline(t *testing.T, s store.Store, pipelineID int64) *model.Pipeline {
	t.Helper()
	return WaitForPipelineStatus(t, s, pipelineID, "", defaultTimeout)
}

// WaitForPipelineStatus polls until the pipeline reaches wantStatus (or any
// terminal status if wantStatus is empty). Fails the test on timeout.
func WaitForPipelineStatus(t *testing.T, s store.Store, pipelineID int64, wantStatus model.StatusValue, timeout time.Duration) *model.Pipeline {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		p, err := s.GetPipeline(pipelineID)
		require.NoError(t, err, "get pipeline %d", pipelineID)

		if wantStatus != "" {
			if p.Status == wantStatus {
				return p
			}
		} else if isTerminal(p.Status) {
			return p
		}

		time.Sleep(defaultInterval)
	}

	p, _ := s.GetPipeline(pipelineID)
	t.Fatalf("timeout waiting for pipeline %d: last status=%q (want %q)", pipelineID, p.Status, wantStatus)
	return nil
}

// WaitForAgentRegistered polls until all provided agents appear in the store
// (by AgentID), then applies any deferred DB patches (e.g. OrgID).
// Pass every *AgentEnv returned by StartAgent before triggering pipelines.
func WaitForAgentRegistered(t *testing.T, s store.Store, agents ...*AgentEnv) {
	t.Helper()

	deadline := time.Now().Add(shortTimeout)
	for time.Now().Before(deadline) {
		allFound := true
		for _, env := range agents {
			if env.AgentID == 0 {
				allFound = false
				break
			}
			if _, err := s.AgentFind(env.AgentID); err != nil {
				allFound = false
				break
			}
		}
		if allFound {
			// Apply any deferred OrgID patches.
			for _, env := range agents {
				if env.requestOrgID == model.IDNotSet {
					continue
				}
				agent, err := s.AgentFind(env.AgentID)
				require.NoError(t, err, "find agent %d to patch OrgID", env.AgentID)
				agent.OrgID = env.requestOrgID
				require.NoError(t, s.AgentUpdate(agent),
					"patch OrgID on agent %d", env.AgentID)
			}
			return
		}
		time.Sleep(defaultInterval)
	}

	t.Fatal("timeout: not all agents registered with the server")
}

// WaitForStep polls the store until a named step in the given pipeline reaches
// a terminal status. It returns the final step state. Fails the test on timeout.
func WaitForStep(t *testing.T, s store.Store, pipeline *model.Pipeline, stepName string) *model.Step {
	t.Helper()
	return WaitForStepStatus(t, s, pipeline, stepName, "", defaultTimeout)
}

// WaitForStepStatus polls until a named step reaches wantState (or any terminal
// state when wantState is empty). This is useful after a pipeline.Cancel() call
// where the agent sends its final step status asynchronously via gRPC Done(),
// independently of the pipeline itself reaching a terminal status.
func WaitForStepStatus(t *testing.T, s store.Store, pipeline *model.Pipeline, stepName string, wantState model.StatusValue, timeout time.Duration) *model.Step {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		steps, err := s.StepList(pipeline.ID)
		require.NoError(t, err, "list steps for pipeline %d", pipeline.ID)

		for _, step := range steps {
			if step.Name != stepName {
				continue
			}
			if wantState != "" {
				if step.State == wantState {
					return step
				}
			} else if isTerminal(step.State) {
				return step
			}
		}
		time.Sleep(defaultInterval)
	}

	steps, _ := s.StepList(pipeline.ID)
	var lastState model.StatusValue
	for _, step := range steps {
		if step.Name == stepName {
			lastState = step.State
			break
		}
	}
	if wantState != "" {
		t.Fatalf("timeout waiting for step %q in pipeline %d to reach state %q: last state=%q",
			stepName, pipeline.ID, wantState, lastState)
	} else {
		t.Fatalf("timeout waiting for step %q in pipeline %d to reach terminal state: last state=%q",
			stepName, pipeline.ID, lastState)
	}
	return nil
}

// AssertWorkflowRanOnAgent asserts that the named workflow in the finished
// pipeline was executed by the given agent. Use this to verify label-based
// routing and org-agent preference.
func AssertWorkflowRanOnAgent(t *testing.T, s store.Store, pipeline *model.Pipeline, workflowName string, agent *AgentEnv) {
	t.Helper()

	workflows, err := s.WorkflowGetTree(pipeline)
	require.NoError(t, err, "get workflow tree for pipeline %d", pipeline.ID)

	for _, wf := range workflows {
		if wf.Name == workflowName {
			assert.Equalf(t, agent.AgentID, wf.AgentID,
				"workflow %q should have run on agent %d (%s) but ran on agent %d",
				workflowName, agent.AgentID, agent.name, wf.AgentID)
			return
		}
	}
	t.Errorf("workflow %q not found in pipeline %d", workflowName, pipeline.ID)
}

// WaitForWorkersReady polls the queue until at least minWorkers worker slots
// are active (i.e. agents have connected and are blocking on Poll). Call this
// after WaitForAgentRegistered and before pipeline.Create in tests that rely
// on specific routing: the org-id label is read from the DB at Poll time, so
// the org-agent must have started its poll loop *after* its OrgID has been
// patched — otherwise the global agent can win the race and steal the task
// before the org-agent advertises its exact org-id label.
func WaitForWorkersReady(t *testing.T, q queue.Queue, minWorkers int) {
	t.Helper()

	deadline := time.Now().Add(shortTimeout)
	for time.Now().Before(deadline) {
		info := q.Info(context.Background())
		if info.Stats.Workers >= minWorkers {
			return
		}
		time.Sleep(defaultInterval)
	}

	info := q.Info(context.Background())
	t.Fatalf("timeout waiting for %d workers to be ready in queue: got %d", minWorkers, info.Stats.Workers)
}

// WaitForStepRunning polls the store until a named step in the pipeline with
// the given ID reaches StatusRunning. This is used before triggering a cancel
// so we know the dummy backend's sleepWithContext is genuinely blocking — if
// we cancel before the step is running, the step may finish with StatusSuccess
// before the cancel context propagates to WaitStep.
func WaitForStepRunning(t *testing.T, s store.Store, pipelineID int64, stepName string) {
	t.Helper()

	deadline := time.Now().Add(shortTimeout)
	for time.Now().Before(deadline) {
		p, err := s.GetPipeline(pipelineID)
		require.NoError(t, err, "get pipeline %d", pipelineID)

		steps, err := s.StepList(p.ID)
		require.NoError(t, err, "list steps for pipeline %d", pipelineID)

		for _, step := range steps {
			if step.Name == stepName && step.State == model.StatusRunning {
				return
			}
		}
		time.Sleep(defaultInterval)
	}

	t.Fatalf("timeout waiting for step %q in pipeline %d to reach StatusRunning", stepName, pipelineID)
}
