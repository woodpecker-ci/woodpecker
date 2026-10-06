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

package exec

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/dummy"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

// The scenarios the e2e tests run with a server and an agent.
// A local run has to end with the very same status.
const scenariosDir = "../../e2e/scenarios/fixtures"

type scenario struct {
	Name           string `json:"name"`
	Event          string `json:"event"`
	ExpectedStatus string `json:"expected_status"`
	ExpectedSteps  []struct {
		Name     string `json:"name"`
		Status   string `json:"status"`
		ExitCode int    `json:"exit_code"`
	} `json:"expected_steps"`
	ExpectedWorkflows []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
		Error  string `json:"error"`
	} `json:"expected_workflows"`

	yamls []*builder.YamlFile
}

// loadScenarios reads the scenarios the way the e2e tests do: a single
// workflow is a yaml file with a json file of the same name next to it,
// multiple workflows are a directory with their files and a scenario.json.
func loadScenarios(t *testing.T) []*scenario {
	t.Helper()

	read := func(path ...string) []byte {
		data, err := os.ReadFile(filepath.Join(append([]string{scenariosDir}, path...)...))
		require.NoError(t, err)
		return data
	}

	entries, err := os.ReadDir(scenariosDir)
	require.NoError(t, err)

	var scenarios []*scenario
	for _, entry := range entries {
		name := entry.Name()
		sc := new(scenario)

		switch {
		case entry.IsDir():
			require.NoError(t, json.Unmarshal(read(name, "scenario.json"), sc))
			files, err := os.ReadDir(filepath.Join(scenariosDir, name))
			require.NoError(t, err)
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".yaml") {
					sc.yamls = append(sc.yamls, &builder.YamlFile{Name: ".woodpecker/" + file.Name(), Data: read(name, file.Name())})
				}
			}
		case strings.HasSuffix(name, ".json"):
			require.NoError(t, json.Unmarshal(read(name), sc))
			sc.yamls = []*builder.YamlFile{{Name: ".woodpecker.yaml", Data: read(strings.TrimSuffix(name, ".json") + ".yaml")}}
		default:
			continue
		}
		scenarios = append(scenarios, sc)
	}

	require.NotEmpty(t, scenarios)
	return scenarios
}

// recordOutput keeps the last reported state of each workflow.
type recordOutput struct {
	sync.Mutex
	workflows map[int64]*woodpecker.Workflow
}

func (o *recordOutput) Workflow(workflow *woodpecker.Workflow) {
	o.Lock()
	defer o.Unlock()
	o.workflows[workflow.ID] = workflow
}

func (*recordOutput) Log(*woodpecker.LogEntry) {}

func (*recordOutput) Message(string) {}

func TestRunMatchesServerScenarios(t *testing.T) {
	for _, sc := range loadScenarios(t) {
		t.Run(sc.Name, func(t *testing.T) {
			if sc.Name == "service runs alongside steps" {
				// The dummy backend fails a service that starts to wait after
				// its workflow is gone. A server is slow enough to never see
				// that, a local run is not.
				t.Skip("depends on the service being slower than the reports of the other steps")
			}

			b := builder.PipelineBuilder{
				Yamls:       sc.yamls,
				RepoTrusted: &metadata.TrustedConfiguration{},
				GetWorkflowMetadata: func(w *builder.Workflow) metadata.Metadata {
					return metadata.Metadata{
						Curr:     metadata.Pipeline{Event: metadata.Event(sc.Event)},
						Workflow: metadata.Workflow{Name: w.Name, Number: w.PID, Matrix: w.Environ},
					}
				},
			}
			items, err := b.Build()
			if sc.ExpectedStatus == "error" {
				// the pipeline never starts, the server reports that as status
				require.Error(t, err)
				assert.Empty(t, items)
				return
			}

			ctx, cancel := context.WithCancelCause(t.Context())
			defer cancel(nil)
			out := &recordOutput{workflows: make(map[int64]*woodpecker.Workflow)}
			run := newPipelineRun(items, []execBackend{{Backend: dummy.New()}}, time.Minute, cancel)
			run.out = out

			assert.EqualValues(t, sc.ExpectedStatus, run.execute(ctx), "pipeline status")

			steps := make(map[string]*woodpecker.Step)
			workflows := make(map[string]*woodpecker.Workflow)
			for _, workflow := range out.workflows {
				workflows[workflow.Name] = workflow
				for _, step := range workflow.Children {
					steps[step.Name] = step
				}
			}

			if len(sc.ExpectedSteps) == 0 {
				return
			}
			assert.Len(t, steps, len(sc.ExpectedSteps), "steps")
			for _, want := range sc.ExpectedSteps {
				if step := steps[want.Name]; assert.NotNil(t, step, "step %q", want.Name) {
					assert.Equal(t, want.Status, step.State, "step %q status", want.Name)
					assert.Equal(t, want.ExitCode, step.ExitCode, "step %q exit code", want.Name)
				}
			}

			if len(sc.ExpectedWorkflows) == 0 {
				return
			}
			assert.Len(t, workflows, len(sc.ExpectedWorkflows), "workflows")
			for _, want := range sc.ExpectedWorkflows {
				if workflow := workflows[want.Name]; assert.NotNil(t, workflow, "workflow %q", want.Name) {
					assert.Equal(t, want.Status, workflow.State, "workflow %q status", want.Name)
					assert.Equal(t, want.Error, workflow.Error, "workflow %q error", want.Name)
				}
			}
		})
	}
}
