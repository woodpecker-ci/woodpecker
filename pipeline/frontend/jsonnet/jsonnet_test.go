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

package jsonnet

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsJsonnetFile(t *testing.T) {
	assert.True(t, IsJsonnetFile(".woodpecker.jsonnet"))
	assert.True(t, IsJsonnetFile(".woodpecker/pipelines.jsonnet"))
	assert.False(t, IsJsonnetFile(".woodpecker.yaml"))
	assert.False(t, IsJsonnetFile(".woodpecker.yml"))
	assert.False(t, IsJsonnetFile("jsonnet"))
}

func TestCompileSingleObject(t *testing.T) {
	files, err := Compile(".woodpecker.jsonnet", []byte(`{
		steps: [{ name: 'test', image: 'alpine', commands: ['echo hi'] }],
	}`))
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "woodpecker", files[0].Name)
	assert.JSONEq(t, `{"steps":[{"name":"test","image":"alpine","commands":["echo hi"]}]}`, string(files[0].Data))
}

func TestCompileSingleObjectWithName(t *testing.T) {
	files, err := Compile(".woodpecker.jsonnet", []byte(`{
		name: 'build',
		steps: [{ name: 'test', image: 'alpine' }],
	}`))
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "build", files[0].Name)

	var workflow map[string]any
	require.NoError(t, json.Unmarshal(files[0].Data, &workflow))
	assert.NotContains(t, workflow, "name")
}

func TestCompileArray(t *testing.T) {
	files, err := Compile(".woodpecker.jsonnet", []byte(`
		local pipeline(name) = {
			name: name,
			steps: [{ name: 'test', image: 'alpine', commands: ['echo ' + name] }],
		};
		std.map(pipeline, ['alpha', 'beta', 'gamma'])
	`))
	require.NoError(t, err)
	require.Len(t, files, 3)
	assert.Equal(t, "alpha", files[0].Name)
	assert.Equal(t, "beta", files[1].Name)
	assert.Equal(t, "gamma", files[2].Name)
	for _, file := range files {
		var workflow map[string]any
		require.NoError(t, json.Unmarshal(file.Data, &workflow))
		assert.NotContains(t, workflow, "name")
		assert.Contains(t, workflow, "steps")
	}
}

func TestCompileTextBlocks(t *testing.T) {
	files, err := Compile(".woodpecker.jsonnet", []byte(`{
		steps: [{
			name: 'test',
			image: 'alpine',
			commands: [|||
				echo one
				echo two
			|||],
		}],
	}`))
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Contains(t, string(files[0].Data), "echo one\\necho two")
}

func TestCompileErrors(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{
			name:    "syntax error",
			config:  `{ steps: [ }`,
			wantErr: "failed to compile",
		},
		{
			name:    "top-level scalar",
			config:  `"not a workflow"`,
			wantErr: "must evaluate to an object or an array",
		},
		{
			name:    "array element not an object",
			config:  `[{ name: 'ok', steps: [] }, 42]`,
			wantErr: "element 1 is not an object",
		},
		{
			name:    "array element without name",
			config:  `[{ steps: [] }]`,
			wantErr: "has no name",
		},
		{
			name:    "duplicate names",
			config:  `[{ name: 'a', steps: [] }, { name: 'a', steps: [] }]`,
			wantErr: "duplicate workflow name",
		},
		{
			name:    "non-string name",
			config:  `{ name: 42, steps: [] }`,
			wantErr: "non-string name",
		},
		{
			name:    "invalid name",
			config:  `{ name: 'foo/bar', steps: [] }`,
			wantErr: "invalid workflow name",
		},
		{
			name:    "import",
			config:  `import 'other.jsonnet'`,
			wantErr: "imports are not supported",
		},
		{
			name:    "importstr",
			config:  `{ data: importstr 'file.txt' }`,
			wantErr: "imports are not supported",
		},
		{
			name:    "external variable",
			config:  `{ name: std.extVar('x'), steps: [] }`,
			wantErr: "failed to compile",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Compile(".woodpecker.jsonnet", []byte(tt.config))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestCompileOutputSizeLimit(t *testing.T) {
	_, err := Compile(".woodpecker.jsonnet", []byte(`{
		steps: [{ name: 'test', image: 'alpine', commands: [std.repeat('x', 2 * 1024 * 1024)] }],
	}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeding the limit")
}
