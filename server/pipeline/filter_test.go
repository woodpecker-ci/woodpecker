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
			name:     "missing required dependency is rejected",
			selected: []string{"test"},
			wantErr:  `selected workflows depend on "lint", which is not in the selection`,
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
	// the deploy fixture depends on test optionally, and optional dependencies
	// never make a selection invalid, so they are not reported
	assert.Equal(t, WorkflowInfo{Name: "deploy"}, infos[2])
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
