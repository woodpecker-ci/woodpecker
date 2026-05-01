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

// Package scenarios contains end-to-end integration tests that run a real
// in-process Woodpecker server (with MockForge) and a real in-process agent
// (with the dummy backend). Tests trigger pipelines via server/pipeline.Create
// and assert on final DB state.
package scenarios

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/e2e/setup"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
)

// TestMain sets global log level to warn so test output isn't buried in JSON.
// Override by setting WOODPECKER_LOG_LEVEL=trace before running tests.
func TestMain(m *testing.M) {
	level := zerolog.WarnLevel
	if lvl := os.Getenv("WOODPECKER_LOG_LEVEL"); lvl != "" {
		if l, err := zerolog.ParseLevel(lvl); err == nil {
			level = l
		}
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true})
	os.Exit(m.Run())
}

// simpleSuccessYAML is the minimal pipeline config for the smoke test.
// "image: dummy" is handled by the dummy backend (requires -tags test).
var simpleSuccessYAML = []byte(`
steps:
  - name: step-one
    image: dummy
    commands:
      - echo hello

  - name: step-two
    image: dummy
    commands:
      - echo world
`)

// TestInfraSmoke verifies the full server+agent stack can start, accept a
// pipeline, run it through the dummy backend, and reach StatusSuccess.
// This is the "does the plumbing work at all" gate — it runs first.
func TestInfraSmoke(t *testing.T) {
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker.yaml", Data: simpleSuccessYAML},
	})
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	draftPipeline := &model.Pipeline{
		Event:  model.EventPush,
		Branch: "main",
		Commit: "deadbeef",
		Ref:    "refs/heads/main",
		Author: env.Fixtures.Owner.Login,
		Sender: env.Fixtures.Owner.Login,
	}
	createdPipeline, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, draftPipeline)
	require.NoError(t, err, "create pipeline")
	require.NotNil(t, createdPipeline)
	t.Logf("pipeline %d created with status=%s", createdPipeline.ID, createdPipeline.Status)

	finished := setup.WaitForPipeline(t, env.Store, createdPipeline.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status, "pipeline should succeed")
}
