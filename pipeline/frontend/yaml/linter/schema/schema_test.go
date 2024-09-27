// Copyright 2023 Woodpecker Authors
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

package schema_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/linter/schema"
)

func TestSchema(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name     string
		testFile string
		fail     bool
	}{
		{
			name:     "Clone",
			testFile: ".woodpecker/test-clone.yaml",
		},
		{
			name:     "Clone skip",
			testFile: ".woodpecker/test-clone-skip.yaml",
		},
		{
			name:     "Matrix",
			testFile: ".woodpecker/test-matrix.yaml",
		},
		{
			name:     "Multi Pipeline",
			testFile: ".woodpecker/test-multi.yaml",
		},
		{
			name:     "Plugin",
			testFile: ".woodpecker/test-plugin.yaml",
		},
		{
			name:     "Run on",
			testFile: ".woodpecker/test-run-on.yaml",
		},
		{
			name:     "Service",
			testFile: ".woodpecker/test-service.yaml",
		},
		{
			name:     "Step",
			testFile: ".woodpecker/test-step.yaml",
		},
		{
			name:     "When",
			testFile: ".woodpecker/test-when.yaml",
		},
		{
			name:     "Workspace",
			testFile: ".woodpecker/test-workspace.yaml",
		},
		{
			name:     "Labels",
			testFile: ".woodpecker/test-labels.yaml",
		},
		{
			name:     "Map and Sequence Merge", // https://woodpecker-ci.org/docs/next/usage/advanced-yaml-syntax
			testFile: ".woodpecker/test-merge-map-and-sequence.yaml",
		},
		{
			name:     "Broken Config",
			testFile: ".woodpecker/test-broken.yaml",
			fail:     true,
		},
		{
			name:     "Array syntax",
			testFile: ".woodpecker/test-array-syntax.yaml",
			fail:     false,
		},
		{
			name:     "Step DAG syntax",
			testFile: ".woodpecker/test-dag.yaml",
			fail:     false,
		},
		{
			name:     "Custom backend",
			testFile: ".woodpecker/test-custom-backend.yaml",
			fail:     false,
		},
		{
			name:     "Broken Plugin by environment",
			testFile: ".woodpecker/test-broken-plugin.yaml",
			fail:     true,
		},
		{
			name:     "Broken Plugin by commands",
			testFile: ".woodpecker/test-broken-plugin2.yaml",
			fail:     true,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			fi, err := os.Open(tt.testFile)
			assert.NoError(t, err, "could not open test file")
			defer fi.Close()
			configErrors, err := schema.Lint(fi)
			if tt.fail {
				if len(configErrors) == 0 {
					assert.Error(t, err, "Expected config errors but got none")
				}
			} else {
				assert.NoError(t, err, fmt.Sprintf("Validation failed: %v", configErrors))
			}
		})
	}
}
