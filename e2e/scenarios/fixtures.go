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
	"embed"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

//go:embed fixtures/*.yaml fixtures/*.json fixtures/*/*.yaml fixtures/*/*.json
var fixtureFS embed.FS

// Scenario is the single source of truth for one integration test case.
//
// Single-workflow scenarios use a flat fixture pair:
//
//	fixtures/NN_name.yaml   — the pipeline YAML served by the mock forge
//	fixtures/NN_name.json   — assertions (Scenario fields)
//
// Multi-workflow scenarios use a subdirectory:
//
//	fixtures/NN_name/workflow-a.yaml
//	fixtures/NN_name/workflow-b.yaml
//	fixtures/NN_name/scenario.json   — assertions; Workflows field is populated from the YAMLs
type Scenario struct {
	// Name is a human-readable label shown in test output.
	Name string `json:"name"`

	// Event is the webhook event that triggers the pipeline (default: push).
	Event model.WebhookEvent `json:"event"`

	// ExpectedStatus is the final pipeline status we assert on.
	ExpectedStatus model.StatusValue `json:"expected_status"`

	// ExpectedSteps lists per-step assertions (matched by step name).
	// Steps not listed here are not checked.
	ExpectedSteps []ExpectedStep `json:"expected_steps"`

	// ExpectedWorkflows lists per-workflow assertions (matched by workflow name).
	// Only checked when non-empty. For single-workflow pipelines, the workflow
	// name is derived from the YAML filename by the step builder.
	ExpectedWorkflows []ExpectedWorkflow `json:"expected_workflows"`

	// Files is the set of workflow YAML files served by the mock forge.
	// Single-workflow: one entry named ".woodpecker.yaml".
	// Multi-workflow:  one entry per file in the fixtures subdirectory,
	//                  with paths like ".woodpecker/workflow-a.yaml".
	// Populated by LoadScenarios — not present in the JSON.
	Files []*forge_types.FileMeta `json:"-"`
}

// ExpectedStep describes what we expect for one named step after the pipeline finishes.
type ExpectedStep struct {
	Name     string            `json:"name"`
	Status   model.StatusValue `json:"status"`
	ExitCode int               `json:"exit_code"`
}

// ExpectedWorkflow describes what we expect for one named workflow after the pipeline finishes.
type ExpectedWorkflow struct {
	Name   string            `json:"name"`
	Status model.StatusValue `json:"status"`
}

// LoadScenarios reads all fixture pairs and subdirectories from the embedded
// fixtures/ directory and returns them sorted by filesystem order.
//
// Flat pairs  (NN_name.yaml + NN_name.json)   → single-workflow scenario.
// Directories (NN_name/ with *.yaml + scenario.json) → multi-workflow scenario.
func LoadScenarios(t *testing.T) []Scenario {
	t.Helper()

	entries, err := fixtureFS.ReadDir("fixtures")
	require.NoError(t, err, "read fixtures dir")

	// Index flat YAML files by stem.
	yamlByStem := make(map[string][]byte)
	jsonByStem := make(map[string][]byte)

	var scenarios []Scenario

	for _, e := range entries {
		name := e.Name()

		if e.IsDir() {
			// Multi-workflow scenario: load scenario.json + all *.yaml files.
			s := loadMultiWorkflowScenario(t, name)
			scenarios = append(scenarios, s)
			continue
		}

		data, err := fixtureFS.ReadFile(filepath.Join("fixtures", name))
		require.NoError(t, err, "read fixture %s", name)

		stem := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".json")
		switch filepath.Ext(name) {
		case ".yaml":
			yamlByStem[stem] = data
		case ".json":
			jsonByStem[stem] = data
		}
	}

	// Pair flat YAML + JSON files.
	for stem, jsonData := range jsonByStem {
		var s Scenario
		require.NoError(t, json.Unmarshal(jsonData, &s), "parse %s.json", stem)

		yamlData, ok := yamlByStem[stem]
		require.True(t, ok, "missing %s.yaml for %s.json", stem, stem)

		// Single-workflow: serve as ".woodpecker.yaml" so the config service
		// calls File() and gets back the YAML directly.
		s.Files = []*forge_types.FileMeta{
			{Name: ".woodpecker.yaml", Data: yamlData},
		}

		if s.Event == "" {
			s.Event = model.EventPush
		}
		scenarios = append(scenarios, s)
	}

	require.NotEmpty(t, scenarios, "no scenarios loaded")
	return scenarios
}

// loadMultiWorkflowScenario reads a fixtures/dirName/ subdirectory.
// It expects a scenario.json and one or more *.yaml workflow files.
func loadMultiWorkflowScenario(t *testing.T, dirName string) Scenario {
	t.Helper()

	dir := filepath.Join("fixtures", dirName)
	entries, err := fixtureFS.ReadDir(dir)
	require.NoError(t, err, "read multi-workflow dir %s", dir)

	var s Scenario
	var files []*forge_types.FileMeta

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		data, err := fixtureFS.ReadFile(filepath.Join(dir, name))
		require.NoError(t, err, "read %s/%s", dirName, name)

		switch {
		case name == "scenario.json":
			require.NoError(t, json.Unmarshal(data, &s), "parse %s/scenario.json", dirName)
		case strings.HasSuffix(name, ".yaml"):
			// Serve under .woodpecker/<filename> so Dir() returns them.
			files = append(files, &forge_types.FileMeta{
				Name: ".woodpecker/" + name,
				Data: data,
			})
		}
	}

	require.NotEmpty(t, files, "no YAML files in multi-workflow dir %s", dirName)
	require.NotEmpty(t, s.Name, "scenario.json missing 'name' in %s", dirName)

	s.Files = forge_types.SortByName(files)
	if s.Event == "" {
		s.Event = model.EventPush
	}
	return s
}
