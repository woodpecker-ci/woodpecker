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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/e2e/setup"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
)

var statusWorkflows = []*forge_types.FileMeta{
	{Name: ".woodpecker/a.yaml", Data: []byte("skip_clone: true\nsteps:\n  - name: a\n    image: dummy\n    commands: [echo a]\n")},
	{Name: ".woodpecker/b.yaml", Data: []byte("skip_clone: true\nsteps:\n  - name: b\n    image: dummy\n    commands: [echo b]\n")},
}

// TestStatusesReachTheForge: the forge ends up with the final status of
// every workflow.
func TestStatusesReachTheForge(t *testing.T) {
	env := setup.StartServer(t.Context(), t, statusWorkflows)
	runAndCheckStatuses(t, env)
}

// TestStatusesOutlastAForgeOutage: statuses that fail while the forge is
// down are sent again, and the final one still lands last.
func TestStatusesOutlastAForgeOutage(t *testing.T) {
	env := setup.StartServer(t.Context(), t, statusWorkflows)
	env.Statuses.FailNext(5)
	runAndCheckStatuses(t, env)
	assert.Zero(t, env.Statuses.FailuresLeft(), "the forge was never down")
}

func runAndCheckStatuses(t *testing.T, env *setup.ServerEnv) {
	t.Helper()
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, env.DummyPipeline(model.EventPush))
	require.NoError(t, err, "create pipeline")
	finished := setup.WaitForPipelineStatus(t, env.Store, created.ID, model.StatusSuccess, 30*time.Second)

	workflows, err := env.Store.WorkflowGetTree(finished)
	require.NoError(t, err)
	require.Len(t, workflows, len(statusWorkflows))
	for _, w := range workflows {
		env.Statuses.WaitForLast(t, w.ID, model.StatusSuccess, 10*time.Second)
		taken := env.Statuses.Taken(w.ID)
		assert.Equal(t, 1, countOf(taken, model.StatusSuccess), "workflow %s: the forge took %v", w.Name, taken)
	}
}

func countOf(states []model.StatusValue, state model.StatusValue) int {
	n := 0
	for _, s := range states {
		if s == state {
			n++
		}
	}
	return n
}
