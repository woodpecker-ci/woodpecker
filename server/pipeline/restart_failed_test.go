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

package pipeline

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/constraint"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func ran(pipelineNumber int64, workflows ...*model.Workflow) []ranWorkflow {
	out := make([]ranWorkflow, 0, len(workflows))
	for _, wf := range workflows {
		out = append(out, ranWorkflow{Workflow: wf, PipelineNumber: pipelineNumber})
	}
	return out
}

func TestSelectFailedWorkflows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    model.StatusValue
		workflows []*model.Workflow
		// rerun lists the workflows expected to be rerun as name or name/axis.
		rerun []string
	}{
		{
			name:   "nothing failed",
			status: model.StatusSuccess,
			workflows: []*model.Workflow{
				{Name: "build", State: model.StatusSuccess},
				{Name: "notify", State: model.StatusSkipped, DependsOn: []string{"build"}},
			},
		},
		{
			name:   "failure, error and killed count as failed, skipped did not pass either",
			status: model.StatusFailure,
			workflows: []*model.Workflow{
				{Name: "a", State: model.StatusFailure},
				{Name: "b", State: model.StatusError},
				{Name: "c", State: model.StatusKilled},
				{Name: "d", State: model.StatusSuccess},
				{Name: "e", State: model.StatusSkipped},
			},
			rerun: []string{"a", "b", "c", "e"},
		},
		{
			// Canceling skips the workflows that did not start yet.
			name:   "canceled while running",
			status: model.StatusKilled,
			workflows: []*model.Workflow{
				{Name: "build", State: model.StatusKilled},
				{Name: "lint", State: model.StatusSkipped},
				{Name: "docs", State: model.StatusSuccess},
			},
			rerun: []string{"build", "lint"},
		},
		{
			// The cancel came in as the last workflow finished.
			name:   "canceled with nothing left to restart",
			status: model.StatusKilled,
			workflows: []*model.Workflow{
				{Name: "build", State: model.StatusSuccess},
			},
		},
		{
			name:   "canceled before anything started",
			status: model.StatusCanceled,
			workflows: []*model.Workflow{
				{Name: "build", State: model.StatusSkipped},
				{Name: "lint", State: model.StatusSkipped},
			},
			rerun: []string{"build", "lint"},
		},
		{
			name: "dependents are rerun transitively",
			workflows: []*model.Workflow{
				{Name: "build", State: model.StatusSuccess},
				{Name: "test", State: model.StatusFailure, DependsOn: []string{"build"}},
				{Name: "package", State: model.StatusSkipped, DependsOn: []string{"test"}},
				{Name: "deploy", State: model.StatusSkipped, DependsOn: []string{"package"}},
				{Name: "lint", State: model.StatusSuccess},
			},
			rerun: []string{"test", "package", "deploy"},
		},
		{
			name: "dependents declared before their dependency",
			workflows: []*model.Workflow{
				{Name: "deploy", State: model.StatusSkipped, DependsOn: []string{"package"}},
				{Name: "package", State: model.StatusSkipped, DependsOn: []string{"test"}},
				{Name: "test", State: model.StatusFailure},
			},
			rerun: []string{"deploy", "package", "test"},
		},
		{
			name: "diamond",
			workflows: []*model.Workflow{
				{Name: "base", State: model.StatusFailure},
				{Name: "left", State: model.StatusSkipped, DependsOn: []string{"base"}},
				{Name: "right", State: model.StatusSkipped, DependsOn: []string{"base"}},
				{Name: "join", State: model.StatusSkipped, DependsOn: []string{"left", "right"}},
			},
			rerun: []string{"base", "left", "right", "join"},
		},
		{
			name: "only the failed matrix variant",
			workflows: []*model.Workflow{
				{Name: "test", State: model.StatusSuccess, Environ: map[string]string{"GO": "1.25"}},
				{Name: "test", State: model.StatusFailure, Environ: map[string]string{"GO": "1.26"}},
			},
			rerun: []string{"test/1.26"},
		},
		{
			name: "all variants of a dependent matrix workflow",
			workflows: []*model.Workflow{
				{Name: "build", State: model.StatusFailure},
				{Name: "test", State: model.StatusSkipped, DependsOn: []string{"build"}, Environ: map[string]string{"GO": "1.25"}},
				{Name: "test", State: model.StatusSkipped, DependsOn: []string{"build"}, Environ: map[string]string{"GO": "1.26"}},
			},
			rerun: []string{"build", "test/1.25", "test/1.26"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status := tt.status
			if status == "" {
				status = model.StatusFailure
			}
			s := selectFailedWorkflows(ran(1, tt.workflows...), status)
			if len(tt.rerun) == 0 {
				assert.Nil(t, s)
				return
			}
			require.NotNil(t, s)

			var rerun []string
			for _, wf := range tt.workflows {
				if !s.rerun(wf.Name, wf.Environ) {
					continue
				}
				if v, ok := wf.Environ["GO"]; ok {
					rerun = append(rerun, wf.Name+"/"+v)
				} else {
					rerun = append(rerun, wf.Name)
				}
			}
			assert.ElementsMatch(t, tt.rerun, rerun)
		})
	}
}

func TestWorkflowSelectionFilterItems(t *testing.T) {
	t.Parallel()

	old := ran(1,
		&model.Workflow{ID: 10, Name: "build", State: model.StatusSuccess},
		&model.Workflow{ID: 11, Name: "lint", State: model.StatusSuccess},
		&model.Workflow{ID: 12, Name: "test", State: model.StatusSuccess, Environ: map[string]string{"GO": "1.25"}},
		&model.Workflow{ID: 13, Name: "test", State: model.StatusFailure, Environ: map[string]string{"GO": "1.26"}},
		&model.Workflow{ID: 14, Name: "deploy", State: model.StatusSkipped, DependsOn: []string{"build", "lint", "test"}},
		// The config the pipeline ran with had no dependency of report on test.
		&model.Workflow{ID: 15, Name: "report", State: model.StatusSuccess},
	)
	s := selectFailedWorkflows(old, model.StatusFailure)
	require.NotNil(t, s)

	item := func(name string, environ map[string]string, deps ...string) *builder.Item {
		var dependsOn constraint.DependsOn
		for _, dep := range deps {
			dependsOn = append(dependsOn, constraint.Dependency{Name: dep})
		}
		return &builder.Item{
			Workflow:  &builder.Workflow{Name: name, Environ: environ},
			DependsOn: dependsOn,
		}
	}
	items := []*builder.Item{
		item("build", nil),
		item("lint", nil),
		item("test", map[string]string{"GO": "1.25"}, "build"),
		item("test", map[string]string{"GO": "1.26"}, "build"),
		item("deploy", nil, "build", "lint", "test"),
		item("docs", nil, "build"),
		// A config extension served a config where report waits for test.
		item("report", nil, "test"),
	}

	kept := s.filterItems(items)

	var names []string
	for _, item := range kept {
		names = append(names, item.Workflow.Name+fmt.Sprint(item.Workflow.Environ))
	}
	assert.Equal(t, []string{"testmap[GO:1.26]", "deploymap[]", "docsmap[]", "reportmap[]"}, names,
		"the failed variant, its dependents, a workflow added to the config since and a dependent in the new config run")
	assert.Equal(t, []string{"build", "lint", "test"}, kept[1].DependsOn.Names(), "the dependencies are kept as configured")

	assert.Equal(t, map[string]map[string]model.StatusValue{
		"test": {"10": model.StatusSuccess},
		"deploy": {
			"10": model.StatusSuccess,
			"11": model.StatusSuccess,
			"12": model.StatusSuccess,
		},
		"docs":   {"10": model.StatusSuccess},
		"report": {"12": model.StatusSuccess},
	}, s.depStatus, "the dependencies that are not rerun keep their old state, including the passed matrix variant")
}

func TestPipelineTasksDepStatus(t *testing.T) {
	t.Parallel()

	repo := &model.Repo{ID: 1, FullName: "owner/repo"}
	item := &builder.Item{
		Workflow:  &builder.Workflow{ID: 2, Name: "deploy"},
		DependsOn: constraint.DependsOn{{Name: "test"}},
		Config:    nil,
	}
	tasks, err := pipelineTasks(repo, &model.Pipeline{ID: 3}, []*builder.Item{item},
		map[string]map[string]model.StatusValue{"deploy": {"10": model.StatusSkipped}})
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, map[string]model.StatusValue{"10": model.StatusSkipped}, tasks[0].DepStatus)
	assert.False(t, tasks[0].ShouldRun(), "a skipped dependency outside the pipeline still gates the task")
}

func TestWorkflowSelectionReused(t *testing.T) {
	t.Parallel()

	old := append(ran(5,
		&model.Workflow{Name: "test", State: model.StatusFailure},
		&model.Workflow{Name: "matrix", State: model.StatusSuccess, Environ: map[string]string{"GO": "1.26"}, AxisID: 2},
	), ran(3,
		&model.Workflow{Name: "build", State: model.StatusSuccess},
		&model.Workflow{Name: "deploy", State: model.StatusSkipped, DependsOn: []string{"test"}},
		&model.Workflow{Name: "matrix", State: model.StatusSuccess, Environ: map[string]string{"GO": "1.25"}, AxisID: 1},
		&model.Workflow{Name: "lint", State: model.StatusSuccess, Environ: map[string]string{"OS": "linux"}, AxisID: 1},
		&model.Workflow{Name: "lint", State: model.StatusFailure, Environ: map[string]string{"OS": "darwin"}, AxisID: 2},
	)...)
	s := selectFailedWorkflows(old, model.StatusFailure)
	require.NotNil(t, s)

	assert.Equal(t, map[string]int64{
		"build":    3,
		"matrix/1": 3,
		"matrix/2": 5,
		"lint/1":   3,
	}, s.reused(), "rerun workflows are left out, matrix variants are listed one by one")
}

func TestEffectiveWorkflows(t *testing.T) {
	t.Parallel()

	repo := &model.Repo{ID: 1}
	full := &model.Pipeline{ID: 3, Number: 3}
	first := &model.Pipeline{ID: 4, Number: 4, Parent: 3, PartialRestart: true}
	second := &model.Pipeline{ID: 5, Number: 5, Parent: 4, PartialRestart: true}

	s := store_mocks.NewMockStore(t)
	s.On("WorkflowGetTree", full).Return([]*model.Workflow{
		{ID: 1, Name: "build", State: model.StatusSuccess},
		{ID: 2, Name: "test", State: model.StatusFailure},
		{ID: 3, Name: "lint", State: model.StatusFailure},
	}, nil)
	s.On("WorkflowGetTree", first).Return([]*model.Workflow{
		{ID: 10, Name: "test", State: model.StatusFailure},
		{ID: 11, Name: "lint", State: model.StatusSuccess},
	}, nil)
	s.On("WorkflowGetTree", second).Return([]*model.Workflow{
		{ID: 20, Name: "test", State: model.StatusFailure},
	}, nil)
	s.On("GetPipelineNumber", repo, int64(4)).Return(first, nil)
	s.On("GetPipelineNumber", repo, int64(3)).Return(full, nil)

	effective, err := effectiveWorkflows(s, repo, second)
	require.NoError(t, err)

	got := make(map[int64]int64)
	for _, wf := range effective {
		got[wf.ID] = wf.PipelineNumber
	}
	assert.Equal(t, map[int64]int64{20: 5, 11: 4, 1: 3}, got,
		"each workflow is taken from the latest pipeline it ran in, up to the first full run")

	// A regular pipeline is not followed up to its parent.
	effective, err = effectiveWorkflows(s, repo, full)
	require.NoError(t, err)
	assert.Len(t, effective, 3)
}

func TestEffectiveWorkflowsDeletedParent(t *testing.T) {
	t.Parallel()

	repo := &model.Repo{ID: 1, FullName: "owner/repo"}
	partial := &model.Pipeline{ID: 4, Number: 4, Parent: 3, PartialRestart: true}

	s := store_mocks.NewMockStore(t)
	s.On("WorkflowGetTree", partial).Return([]*model.Workflow{
		{ID: 10, Name: "test", State: model.StatusFailure},
	}, nil)
	s.On("GetPipelineNumber", repo, int64(3)).Return(nil, types.ErrRecordNotExist)

	effective, err := effectiveWorkflows(s, repo, partial)
	require.NoError(t, err, "a deleted parent ends the chain")
	require.Len(t, effective, 1)
	assert.Equal(t, int64(10), effective[0].ID)
}
