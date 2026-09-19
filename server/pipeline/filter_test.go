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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
)

func testConfigs() []*forge_types.FileMeta {
	return []*forge_types.FileMeta{
		{Name: ".woodpecker/lint.yaml", Data: []byte("steps:\n  - name: lint\n    image: alpine\n")},
		{Name: ".woodpecker/test.yaml", Data: []byte("depends_on:\n  - lint\nsteps:\n  - name: test\n    image: alpine\n")},
		{Name: ".woodpecker/deploy.yaml", Data: []byte("depends_on:\n  - name: test\n    optional: true\nsteps:\n  - name: deploy\n    image: alpine\n")},
	}
}

func TestFilterConfigsByWorkflows(t *testing.T) {
	tests := []struct {
		name     string
		selected []string
		want     []string
		wantErr  string
	}{
		{
			name:     "no selection runs everything",
			selected: nil,
			want:     []string{".woodpecker/lint.yaml", ".woodpecker/test.yaml", ".woodpecker/deploy.yaml"},
		},
		{
			name:     "empty selection runs everything",
			selected: []string{},
			want:     []string{".woodpecker/lint.yaml", ".woodpecker/test.yaml", ".woodpecker/deploy.yaml"},
		},
		{
			name:     "select by sanitized workflow name",
			selected: []string{"lint"},
			want:     []string{".woodpecker/lint.yaml"},
		},
		{
			name:     "select by full forge path",
			selected: []string{".woodpecker/lint.yaml"},
			want:     []string{".woodpecker/lint.yaml"},
		},
		{
			name:     "select several, order follows the selection",
			selected: []string{"lint", "test"},
			want:     []string{".woodpecker/lint.yaml", ".woodpecker/test.yaml"},
		},
		{
			name:     "duplicate selection is tolerated",
			selected: []string{"lint", ".woodpecker/lint.yaml", "lint"},
			want:     []string{".woodpecker/lint.yaml"},
		},
		{
			name:     "blank entries are ignored",
			selected: []string{"lint", "  "},
			want:     []string{".woodpecker/lint.yaml"},
		},
		{
			name:     "unknown workflow is rejected",
			selected: []string{"nope"},
			wantErr:  `unknown workflow(s) "nope", this repo defines "lint", "test", "deploy"`,
		},
		{
			name:     "unknown workflow is rejected even alongside a valid one",
			selected: []string{"lint", "nope"},
			wantErr:  `unknown workflow(s) "nope"`,
		},
		{
			name:     "only blank entries selects nothing",
			selected: []string{" "},
			wantErr:  "no workflow selected",
		},
		{
			name:     "missing required dependency is left to the builder, not rejected here",
			selected: []string{"test"},
			want:     []string{".woodpecker/test.yaml"},
		},
		{
			name:     "required dependency included is fine",
			selected: []string{"test", "lint"},
			want:     []string{".woodpecker/test.yaml", ".woodpecker/lint.yaml"},
		},
		{
			name:     "missing optional dependency is fine",
			selected: []string{"deploy"},
			want:     []string{".woodpecker/deploy.yaml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filterConfigsByWorkflows(testConfigs(), tt.selected)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.ErrorIs(t, err, &ErrBadRequest{})
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			names := make([]string, 0, len(got))
			for _, config := range got {
				names = append(names, config.Name)
			}
			assert.Equal(t, tt.want, names)
		})
	}
}

// An unparsable config must not be reported as a dependency problem. The
// builder reports the real parse error a moment later with a better message.
func TestFilterConfigsByWorkflowsIgnoresUnparsableConfig(t *testing.T) {
	configs := []*forge_types.FileMeta{
		{Name: ".woodpecker/broken.yaml", Data: []byte("\tthis: is: not: yaml\n")},
	}

	got, err := filterConfigsByWorkflows(configs, []string{"broken"})

	require.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestWorkflowInfos(t *testing.T) {
	infos := WorkflowInfos(testConfigs())

	require.Len(t, infos, 3)
	assert.Equal(t, WorkflowInfo{Name: "lint"}, infos[0])
	assert.Equal(t, WorkflowInfo{Name: "test", DependsOn: []string{"lint"}}, infos[1])
	// required and optional dependencies are both informational now, so the
	// deploy fixture's optional dependency on test is reported too
	assert.Equal(t, WorkflowInfo{Name: "deploy", DependsOn: []string{"test"}}, infos[2])
}

func TestWorkflowInfosIgnoresUnparsableConfig(t *testing.T) {
	configs := []*forge_types.FileMeta{
		{Name: ".woodpecker/broken.yaml", Data: []byte("\tthis: is: not: yaml\n")},
	}

	assert.Equal(t, []WorkflowInfo{{Name: "broken"}}, WorkflowInfos(configs))
}

func TestWorkflowNames(t *testing.T) {
	assert.Equal(t, []string{"lint", "test", "deploy"}, WorkflowNames(testConfigs()))
	assert.Empty(t, WorkflowNames(nil))
}

func TestNamesExcludedBySelection(t *testing.T) {
	// "test" (selected) depends on "lint", which was left out of the
	// selection, so it belongs in the ignore set for the builder.
	only := []*forge_types.FileMeta{testConfigs()[1]} // .woodpecker/test.yaml

	assert.Equal(t, map[string]bool{"lint": true}, namesExcludedBySelection([]string{"test"}, only))
}

func TestNamesExcludedBySelectionEmptyWhenNoSelection(t *testing.T) {
	// no selection means every workflow runs, so there is nothing a `when`
	// filter can't already explain: the builder's default behavior applies.
	assert.Nil(t, namesExcludedBySelection(nil, testConfigs()))
	assert.Nil(t, namesExcludedBySelection([]string{}, testConfigs()))
}

func TestNamesExcludedBySelectionEmptyWhenDependencySatisfied(t *testing.T) {
	assert.Empty(t, namesExcludedBySelection([]string{"lint", "test"}, testConfigs()[:2]))
}

func TestNamesExcludedBySelectionIgnoresUnparsableConfig(t *testing.T) {
	configs := []*forge_types.FileMeta{
		{Name: ".woodpecker/broken.yaml", Data: []byte("\tthis: is: not: yaml\n")},
	}

	assert.Empty(t, namesExcludedBySelection([]string{"broken"}, configs))
}

// Two configs that sanitize to the same workflow name cannot be told apart by
// that name, so selecting it must not silently pick whichever the forge listed
// first.
func collidingConfigs() []*forge_types.FileMeta {
	return []*forge_types.FileMeta{
		{Name: ".woodpecker/a.yml", Data: []byte("steps:\n  - name: one\n    image: alpine\n")},
		{Name: ".woodpecker/a.yaml", Data: []byte("steps:\n  - name: two\n    image: alpine\n")},
		{Name: ".woodpecker/b.yaml", Data: []byte("steps:\n  - name: b\n    image: alpine\n")},
	}
}

func TestFilterConfigsByWorkflowsRejectsAmbiguousName(t *testing.T) {
	got, err := filterConfigsByWorkflows(collidingConfigs(), []string{"a"})

	require.Error(t, err)
	assert.ErrorIs(t, err, &ErrBadRequest{})
	assert.Contains(t, err.Error(), `ambiguous workflow "a" matches ".woodpecker/a.yml", ".woodpecker/a.yaml", select it by path`)
	assert.Nil(t, got)
}

func TestFilterConfigsByWorkflowsSelectsCollidingConfigByPath(t *testing.T) {
	got, err := filterConfigsByWorkflows(collidingConfigs(), []string{".woodpecker/a.yaml", "b"})

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, ".woodpecker/a.yaml", got[0].Name)
	assert.Equal(t, ".woodpecker/b.yaml", got[1].Name)
}

func TestWorkflowNamesUsePathWhenSanitizedNameCollides(t *testing.T) {
	assert.Equal(t, []string{".woodpecker/a.yml", ".woodpecker/a.yaml", "b"}, WorkflowNames(collidingConfigs()))
}

func TestWorkflowInfosUsePathWhenSanitizedNameCollides(t *testing.T) {
	infos := WorkflowInfos(collidingConfigs())

	require.Len(t, infos, 3)
	assert.Equal(t, ".woodpecker/a.yml", infos[0].Name)
	assert.Equal(t, ".woodpecker/a.yaml", infos[1].Name)
	assert.Equal(t, "b", infos[2].Name)
}

// A selected workflow whose required depends_on names a workflow that does not
// exist anywhere in the repo is a config bug, not a selection choice. It must
// be rejected here, otherwise namesExcludedBySelection cannot tell it apart
// from a dependency the trigger left out on purpose and the workflow would run
// standalone, hiding the typo.
func TestFilterConfigsByWorkflowsRejectsUnknownRequiredDependency(t *testing.T) {
	configs := []*forge_types.FileMeta{
		{Name: ".woodpecker/lint.yaml", Data: []byte("steps:\n  - name: lint\n    image: alpine\n")},
		{Name: ".woodpecker/test.yaml", Data: []byte("depends_on:\n  - lnit\nsteps:\n  - name: test\n    image: alpine\n")},
	}

	got, err := filterConfigsByWorkflows(configs, []string{"test"})

	require.Error(t, err)
	assert.ErrorIs(t, err, &ErrBadRequest{})
	assert.Contains(t, err.Error(), `workflow "test" depends on unknown workflow "lnit", this repo defines "lint", "test"`)
	assert.Nil(t, got)
}

func TestFilterConfigsByWorkflowsToleratesUnknownOptionalDependency(t *testing.T) {
	configs := []*forge_types.FileMeta{
		{Name: ".woodpecker/deploy.yaml", Data: []byte("depends_on:\n  - name: nope\n    optional: true\nsteps:\n  - name: deploy\n    image: alpine\n")},
	}

	got, err := filterConfigsByWorkflows(configs, []string{"deploy"})

	require.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestFilterConfigsByWorkflowsLeavesUnknownDependencyToBuilderWithoutSelection(t *testing.T) {
	configs := []*forge_types.FileMeta{
		{Name: ".woodpecker/test.yaml", Data: []byte("depends_on:\n  - lnit\nsteps:\n  - name: test\n    image: alpine\n")},
	}

	got, err := filterConfigsByWorkflows(configs, nil)

	require.NoError(t, err)
	assert.Len(t, got, 1)
}
