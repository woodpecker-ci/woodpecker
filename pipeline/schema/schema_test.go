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

	"github.com/woodpecker-ci/woodpecker/pipeline/schema"
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
			testFile: ".woodpecker/test-clone.yml",
		},
		{
			name:     "Clone skip",
			testFile: ".woodpecker/test-clone-skip.yml",
		},
		{
			name:     "Matrix",
			testFile: ".woodpecker/test-matrix.yml",
		},
		{
			name:     "Multi Pipeline",
			testFile: ".woodpecker/test-multi.yml",
		},
		{
			name:     "Plugin",
			testFile: ".woodpecker/test-plugin.yml",
		},
		{
			name:     "Run on",
			testFile: ".woodpecker/test-run-on.yml",
		},
		{
			name:     "Service",
			testFile: ".woodpecker/test-service.yml",
		},
		{
			name:     "Step",
			testFile: ".woodpecker/test-step.yml",
		},
		{
			name:     "When",
			testFile: ".woodpecker/test-when.yml",
		},
		{
			name:     "Workspace",
			testFile: ".woodpecker/test-workspace.yml",
		},
		{
			name:     "Labels",
			testFile: ".woodpecker/test-labels.yml",
		},
		{
			name:     "Map and Sequence Merge", // https://woodpecker-ci.org/docs/next/usage/advanced-yaml-syntax
			testFile: ".woodpecker/test-merge-map-and-sequence.yml",
		},
		{
			name:     "Broken Config",
			testFile: ".woodpecker/test-broken.yml",
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
