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

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

//go:embed fixtures/*.yaml fixtures/*.json
var fixtureFS embed.FS

// Scenario is the single source of truth for one integration test case.
// The pipeline YAML drives the forge mock (what config it returns).
// The expected fields describe what the DB must contain after the pipeline finishes.
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

	// PipelineYAML is loaded from the matching .yaml file — not in JSON.
	PipelineYAML []byte `json:"-"`
}

// ExpectedStep describes what we expect for one named step after the pipeline finishes.
type ExpectedStep struct {
	Name     string            `json:"name"`
	Status   model.StatusValue `json:"status"`
	ExitCode int               `json:"exit_code"`
}

// LoadScenarios reads all fixture pairs (NN_name.yaml + NN_name.json) from
// the embedded fixtures/ directory and returns them sorted by filename.
func LoadScenarios(t *testing.T) []Scenario {
	t.Helper()

	entries, err := fixtureFS.ReadDir("fixtures")
	require.NoError(t, err, "read fixtures dir")

	// Index YAML files by stem so we can pair them with JSON.
	yamlByStem := make(map[string][]byte)
	jsonByStem := make(map[string][]byte)

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
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

	var scenarios []Scenario
	for stem, jsonData := range jsonByStem {
		var s Scenario
		require.NoError(t, json.Unmarshal(jsonData, &s), "parse %s.json", stem)

		yamlData, ok := yamlByStem[stem]
		require.True(t, ok, "missing %s.yaml for %s.json", stem, stem)
		s.PipelineYAML = yamlData

		if s.Event == "" {
			s.Event = model.EventPush
		}
		scenarios = append(scenarios, s)
	}

	require.NotEmpty(t, scenarios, "no scenarios loaded")
	return scenarios
}
