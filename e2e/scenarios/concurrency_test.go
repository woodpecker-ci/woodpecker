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

package scenarios

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/e2e/setup"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
)

// Two workflows in one pipeline:
//   - workflow A opts into concurrency limit 1 (fully serialized across runs)
//   - workflow B opts into concurrency limit 2 (at most two in flight at once)
//
// The concurrency group defaults to the workflow name and is scoped per repo,
// so every A instance across all pipeline runs shares one group, and likewise
// for B. Each step sleeps so the workflows overlap in wall-clock time, making
// the limit observable through the recorded start/finish timestamps.
var (
	concurrencyWorkflowA = []byte(`
skip_clone: true
concurrency: 1
steps:
  - name: work
    image: dummy
    commands:
      - echo workflow-a
    environment:
      SLEEP: "2s"
`)

	concurrencyWorkflowB = []byte(`
skip_clone: true
concurrency:
  limit: 2
steps:
  - name: work
    image: dummy
    commands:
      - echo workflow-b
    environment:
      SLEEP: "2s"
`)
)

// TestWorkflowConcurrencyLimit runs three rounds of a two-workflow pipeline on
// an agent with six free slots — enough capacity that the concurrency limits,
// not the agent, are the only thing constraining parallelism. It then inspects
// the recorded workflow timings and asserts:
//   - every pipeline succeeds,
//   - workflow A (limit 1) never overlaps itself,
//   - workflow B (limit 2) never exceeds two concurrent instances and does in
//     fact reach two at some point (otherwise the test would not distinguish a
//     working limit-2 from an accidental limit-1).
func TestWorkflowConcurrencyLimit(t *testing.T) {
	const rounds = 3

	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker/a.yaml", Data: concurrencyWorkflowA},
		{Name: ".woodpecker/b.yaml", Data: concurrencyWorkflowB},
	})

	// Six slots: 3 rounds × 2 workflows = 6 workflows could all run at once
	// if nothing limited them, so any serialization we observe is the
	// concurrency limit at work, not slot starvation.
	agent := setup.StartAgent(t, env.GRPCAddr, setup.WithCapacity(6))
	setup.WaitForAgentRegistered(t, env.Store, agent)

	// Trigger all rounds up front so they compete for the concurrency groups.
	created := make([]*model.Pipeline, 0, rounds)
	for i := range rounds {
		pipeDraft := env.DummyPipeline(model.EventPush)
		pipeDraft.Commit = fmt.Sprintf("deadbeef%d", i)

		p, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, pipeDraft)
		require.NoErrorf(t, err, "create pipeline round %d", i)
		require.NotNil(t, p)
		created = append(created, p)
	}

	// Wait for every round to finish, collecting workflow timings by name.
	byWorkflow := map[string][]wfInterval{}

	for i, p := range created {
		finished := setup.WaitForPipeline(t, env.Store, p.ID)
		assert.Equalf(t, model.StatusSuccess, finished.Status, "round %d pipeline status", i)

		workflows, err := env.Store.WorkflowGetTree(finished)
		require.NoErrorf(t, err, "workflow tree round %d", i)

		for _, wf := range workflows {
			require.NotZerof(t, wf.Started, "round %d workflow %q has no start time", i, wf.Name)
			require.NotZerof(t, wf.Finished, "round %d workflow %q has no finish time", i, wf.Name)
			byWorkflow[wf.Name] = append(byWorkflow[wf.Name], wfInterval{wf.Started, wf.Finished})
		}
	}

	require.Len(t, byWorkflow["a"], rounds, "expected one A workflow per round")
	require.Len(t, byWorkflow["b"], rounds, "expected one B workflow per round")

	maxConcurrentA := maxConcurrent(byWorkflow["a"])
	maxConcurrentB := maxConcurrent(byWorkflow["b"])

	assert.LessOrEqualf(t, maxConcurrentA, 1,
		"workflow A has concurrency limit 1 but %d ran at once", maxConcurrentA)
	assert.LessOrEqualf(t, maxConcurrentB, 2,
		"workflow B has concurrency limit 2 but %d ran at once", maxConcurrentB)
	assert.GreaterOrEqualf(t, maxConcurrentB, 2,
		"workflow B has concurrency limit 2 but never ran more than %d at once — limit not exercised", maxConcurrentB)
}

// wfInterval is a workflow's [start, finish) wall-clock window in unix seconds.
type wfInterval struct{ start, finish int64 }

// maxConcurrent returns the largest number of intervals that overlap at any
// instant. Intervals are treated as half-open [start, finish): a workflow that
// finishes at the same second another starts is not counted as overlapping, so
// back-to-back serialized runs report a max concurrency of 1.
func maxConcurrent(intervals []wfInterval) int {
	type event struct {
		t     int64
		delta int
	}
	events := make([]event, 0, len(intervals)*2)
	for _, iv := range intervals {
		events = append(events, event{iv.start, +1}, event{iv.finish, -1})
	}
	// Insertion sort by time; at an equal timestamp process finishes (-1)
	// before starts (+1) so a touching boundary does not count as overlap.
	for i := 1; i < len(events); i++ {
		for j := i; j > 0; j-- {
			a, b := events[j-1], events[j]
			if a.t < b.t || (a.t == b.t && a.delta <= b.delta) {
				break
			}
			events[j-1], events[j] = b, a
		}
	}

	cur, best := 0, 0
	for _, e := range events {
		cur += e.delta
		if cur > best {
			best = cur
		}
	}
	return best
}
