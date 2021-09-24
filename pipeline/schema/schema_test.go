package schema_test

import (
	"testing"

	"github.com/woodpecker-ci/woodpecker/pipeline/schema"
)

func TestSchema(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name     string
		testFile string
	}{
		{
			name:     "Clone",
			testFile: "./test/test-clone.yml",
		},
		{
			name:     "Matrix",
			testFile: "./test/test-matrix.yml",
		},
		{
			name:     "Plugin",
			testFile: "./test/test-plugin.yml",
		},
		{
			name:     "Service",
			testFile: "./test/test-service.yml",
		},
		{
			name:     "Step",
			testFile: "./test/test-step.yml",
		},
		{
			name:     "When",
			testFile: "./test/test-when.yml",
		},
		{
			name:     "Workspace",
			testFile: "./test/test-workspace.yml",
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			err, configErrors := schema.LintSchema(tt.testFile)
			if err != nil {
				t.Error("Validation failed", err, configErrors)
			}
		})
	}
}
