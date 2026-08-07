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
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

func TestPipelineCreateOptions(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected *woodpecker.PipelineOptions
		wantErr  string
	}{
		{
			name:     "branch",
			args:     []string{"create", "--branch", "main"},
			expected: &woodpecker.PipelineOptions{Branch: "main"},
		},
		{
			name:     "tag",
			args:     []string{"create", "--tag", "v1.0.0"},
			expected: &woodpecker.PipelineOptions{Tag: "v1.0.0"},
		},
		{
			name:     "sha",
			args:     []string{"create", "--sha", "abc123"},
			expected: &woodpecker.PipelineOptions{SHA: "abc123"},
		},
		{
			name:    "missing ref",
			args:    []string{"create"},
			wantErr: "exactly one of --branch, --tag, or --sha must be set",
		},
		{
			name:    "multiple refs",
			args:    []string{"create", "--branch", "main", "--tag", "v1.0.0"},
			wantErr: "exactly one of --branch, --tag, or --sha must be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := &cli.Command{
				Name:   "create",
				Writer: io.Discard,
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "branch"},
					&cli.StringFlag{Name: "tag"},
					&cli.StringFlag{Name: "sha"},
				},
				Action: func(_ context.Context, c *cli.Command) error {
					options, err := pipelineCreateOptions(c)
					if tt.wantErr != "" {
						assert.EqualError(t, err, tt.wantErr)
						return nil
					}

					assert.NoError(t, err)
					assert.Equal(t, tt.expected, options)
					return nil
				},
			}

			assert.NoError(t, command.Run(t.Context(), tt.args))
		})
	}
}
