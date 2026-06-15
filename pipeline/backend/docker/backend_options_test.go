// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func Test_parseBackendOptions(t *testing.T) {
	tests := []struct {
		name    string
		step    *backend_types.Step
		want    BackendOptions
		wantErr bool
	}{
		{
			name: "nil options",
			step: &backend_types.Step{BackendOptions: nil},
			want: BackendOptions{},
		},
		{
			name: "empty options",
			step: &backend_types.Step{BackendOptions: map[string]any{}},
			want: BackendOptions{},
		},
		{
			name: "with user option",
			step: &backend_types.Step{BackendOptions: map[string]any{
				"docker": map[string]any{
					"user": "1000:1000",
				},
			}},
			want: BackendOptions{User: "1000:1000"},
		},
		{
			name:    "invalid backend options",
			step:    &backend_types.Step{BackendOptions: map[string]any{"docker": "invalid"}},
			want:    BackendOptions{},
			wantErr: true,
		},
		{
			name: "oom_score_adj zero is valid",
			step: &backend_types.Step{BackendOptions: map[string]any{
				"docker": map[string]any{
					"oom_score_adj": 0,
				},
			}},
			want: BackendOptions{OomScoreAdj: 0},
		},
		{
			name: "oom_score_adj 500 is valid",
			step: &backend_types.Step{BackendOptions: map[string]any{
				"docker": map[string]any{
					"oom_score_adj": 500,
				},
			}},
			want: BackendOptions{OomScoreAdj: 500},
		},
		{
			name: "oom_score_adj 1000 is valid",
			step: &backend_types.Step{BackendOptions: map[string]any{
				"docker": map[string]any{
					"oom_score_adj": 1000,
				},
			}},
			want: BackendOptions{OomScoreAdj: 1000},
		},
		{
			name: "oom_score_adj -1 is invalid",
			step: &backend_types.Step{BackendOptions: map[string]any{
				"docker": map[string]any{
					"oom_score_adj": -1,
				},
			}},
			want:    BackendOptions{},
			wantErr: true,
		},
		{
			name: "oom_score_adj 1001 is invalid",
			step: &backend_types.Step{BackendOptions: map[string]any{
				"docker": map[string]any{
					"oom_score_adj": 1001,
				},
			}},
			want:    BackendOptions{},
			wantErr: true,
		},
		{
			name: "valid memory 512m",
			step: &backend_types.Step{BackendOptions: map[string]any{
				"docker": map[string]any{
					"resources": map[string]any{
						"limits": map[string]any{
							"memory": "512m",
						},
					},
				},
			}},
			want: BackendOptions{Resources: Resources{Limits: ResourceList{Memory: "512m"}}},
		},
		{
			name: "valid memory 1g",
			step: &backend_types.Step{BackendOptions: map[string]any{
				"docker": map[string]any{
					"resources": map[string]any{
						"limits": map[string]any{
							"memory": "1g",
						},
					},
				},
			}},
			want: BackendOptions{Resources: Resources{Limits: ResourceList{Memory: "1g"}}},
		},
		{
			name: "valid memory bare bytes",
			step: &backend_types.Step{BackendOptions: map[string]any{
				"docker": map[string]any{
					"resources": map[string]any{
						"limits": map[string]any{
							"memory": "536870912",
						},
					},
				},
			}},
			want: BackendOptions{Resources: Resources{Limits: ResourceList{Memory: "536870912"}}},
		},
		{
			name: "valid memory 1K",
			step: &backend_types.Step{BackendOptions: map[string]any{
				"docker": map[string]any{
					"resources": map[string]any{
						"limits": map[string]any{
							"memory": "1K",
						},
					},
				},
			}},
			want: BackendOptions{Resources: Resources{Limits: ResourceList{Memory: "1K"}}},
		},
		{
			name: "invalid memory string",
			step: &backend_types.Step{BackendOptions: map[string]any{
				"docker": map[string]any{
					"resources": map[string]any{
						"limits": map[string]any{
							"memory": "abc",
						},
					},
				},
			}},
			want:    BackendOptions{},
			wantErr: true,
		},
		{
			name: "empty memory is not set",
			step: &backend_types.Step{BackendOptions: map[string]any{
				"docker": map[string]any{},
			}},
			want: BackendOptions{},
		},
		{
			name: "all three fields set together",
			step: &backend_types.Step{BackendOptions: map[string]any{
				"docker": map[string]any{
					"user":          "1000:1000",
					"oom_score_adj": 500,
					"resources": map[string]any{
						"limits": map[string]any{
							"memory": "512m",
							"cpus":   1.5,
						},
					},
				},
			}},
			want: BackendOptions{
				User:        "1000:1000",
				OomScoreAdj: 500,
				Resources:   Resources{Limits: ResourceList{Memory: "512m", CPUs: 1.5}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBackendOptions(tt.step)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseMemory(t *testing.T) {
	tests := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"", 0, false},
		{"512m", 512 * 1024 * 1024, false},
		{"512M", 512 * 1024 * 1024, false},
		{"1g", 1 * 1024 * 1024 * 1024, false},
		{"1G", 1 * 1024 * 1024 * 1024, false},
		{"1k", 1024, false},
		{"1K", 1024, false},
		{"536870912", 536870912, false},
		{"abc", 0, true},
		{"1x", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseMemory(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
