package schema_test

import (
	"fmt"
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
			name:     "Branches",
			testFile: ".woodpecker/test-branches.yml",
		},
		{
			name:     "Branches Array",
			testFile: ".woodpecker/test-branches-array.yml",
		},
		{
			name:     "Branches exclude & include",
			testFile: ".woodpecker/test-branches-exclude-include.yml",
		},
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
			name:     "Broken Config",
			testFile: ".woodpecker/test-broken.yml",
			fail:     true,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			err, configErrors := schema.Lint(tt.testFile)
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
