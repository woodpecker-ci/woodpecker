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

//go:build test

package exec

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecSkipsFilteredWorkflowWithoutBackendSetup(t *testing.T) {
	repoDir := t.TempDir()
	workflowPath := filepath.Join(repoDir, "workflow.yaml")
	require.NoError(t, os.WriteFile(workflowPath, []byte(`
when:
  - event: manual

steps:
  - name: build
    image: alpine
    commands:
      - echo hello
`), 0o600))

	err := Command.Run(t.Context(), []string{
		"woodpecker-cli",
		"--backend-engine", "dummy",
		"--repo-path", repoDir,
		workflowPath,
	})

	require.NoError(t, err)
}
