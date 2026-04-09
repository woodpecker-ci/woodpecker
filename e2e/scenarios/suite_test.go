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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/e2e/setup"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
)

// TestScenarios is the table-driven runner for all fixture-based scenarios.
// Each subtest gets its own isolated server+agent environment so they cannot
// interfere with each other.
//
// Subtests do NOT run in parallel because StartServer writes to the
// server.Config package-level global — running concurrently would race.
func TestScenarios(t *testing.T) {
	for _, sc := range LoadScenarios(t) {
		t.Run(sc.Name, func(t *testing.T) {
			runScenario(t, sc)
		})
	}
}

// runScenario starts a fresh server+agent, triggers one pipeline described by
// sc, waits for it to finish, then asserts the expected DB state.
func runScenario(t *testing.T, sc Scenario) {
	t.Helper()

	env := setup.StartServer(t.Context(), t, sc.Files)
	agent := setup.StartAgent(t.Context(), t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, &model.Pipeline{
		Event:  sc.Event,
		Branch: "main",
		Commit: "deadbeef",
		Ref:    "refs/heads/main",
		Author: env.Fixtures.Owner.Login,
		Sender: env.Fixtures.Owner.Login,
	})
	require.NoError(t, err, "create pipeline")
	require.NotNil(t, created)

	finished := setup.WaitForPipeline(t, env.Store, created.ID)
	assert.Equal(t, sc.ExpectedStatus, finished.Status, "pipeline final status")

	if len(sc.ExpectedSteps) == 0 {
		return
	}

	steps, err := env.Store.StepList(finished)
	require.NoError(t, err, "list steps for pipeline %d", finished.ID)

	require.ElementsMatch(t, expStepsToName(sc.ExpectedSteps), modelStepsToName(steps), "we got different steps reported back as we expected")

	// Index steps by name for O(1) lookup.
	byName := make(map[string]*model.Step, len(steps))
	for _, s := range steps {
		byName[s.Name] = s
	}

	for _, want := range sc.ExpectedSteps {
		step, ok := byName[want.Name]
		if !assert.Truef(t, ok, "step %q not found in pipeline %d", want.Name, finished.ID) {
			continue
		}
		assert.Equalf(t, want.Status, step.State, "step %q status", want.Name)
		assert.Equalf(t, want.ExitCode, step.ExitCode, "step %q exit code", want.Name)
	}

	if len(sc.ExpectedWorkflows) == 0 {
		return
	}

	workflows, err := env.Store.WorkflowGetTree(finished)
	require.NoError(t, err, "list workflows for pipeline %d", finished.ID)

	require.ElementsMatch(t, expWorkflowsToName(sc.ExpectedWorkflows), modelWorkflowsToName(workflows), "we got different workflows reported back as we expected")

	byWorkflowName := make(map[string]*model.Workflow, len(workflows))
	for _, w := range workflows {
		byWorkflowName[w.Name] = w
	}

	for _, want := range sc.ExpectedWorkflows {
		wf, ok := byWorkflowName[want.Name]
		if !assert.Truef(t, ok, "workflow %q not found in pipeline %d", want.Name, finished.ID) {
			continue
		}
		assert.Equalf(t, want.Status, wf.State, "workflow %q status", want.Name)
	}
}

func expStepsToName(in []ExpectedStep) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		out = append(out, s.Name)
	}
	return out
}

func modelStepsToName(in []*model.Step) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		out = append(out, s.Name)
	}
	return out
}

func expWorkflowsToName(in []ExpectedWorkflow) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		out = append(out, s.Name)
	}
	return out
}

func modelWorkflowsToName(in []*model.Workflow) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		out = append(out, s.Name)
	}
	return out
}
