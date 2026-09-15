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
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/kubernetes"
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// The defaults of the exec command's workspace flags.
const (
	testWorkspaceBase = "/woodpecker"
	testWorkspacePath = "src"
)

// volumeRecordingBackend records the workflow volume a backend is asked to
// create and the volumes its steps reference.
type volumeRecordingBackend struct {
	name           string
	workflowVolume string
	stepVolumes    [][]string
}

func (b *volumeRecordingBackend) Name() string                     { return b.name }
func (b *volumeRecordingBackend) IsAvailable(context.Context) bool { return true }
func (b *volumeRecordingBackend) Flags() []cli.Flag                { return nil }

func (b *volumeRecordingBackend) Load(context.Context) (*backend_types.BackendInfo, error) {
	return &backend_types.BackendInfo{}, nil
}

func (b *volumeRecordingBackend) SetupWorkflow(_ context.Context, conf *backend_types.Config, _ string) error {
	b.workflowVolume = conf.Volume
	return nil
}

func (b *volumeRecordingBackend) StartStep(_ context.Context, step *backend_types.Step, _ string) error {
	b.stepVolumes = append(b.stepVolumes, step.Volumes)
	return nil
}

func (b *volumeRecordingBackend) TailStep(context.Context, *backend_types.Step, string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}

func (b *volumeRecordingBackend) WaitStep(context.Context, *backend_types.Step, string) (*backend_types.State, error) {
	return &backend_types.State{Exited: true}, nil
}

func (b *volumeRecordingBackend) DestroyStep(context.Context, *backend_types.Step, string) error {
	return nil
}

func (b *volumeRecordingBackend) DestroyWorkflow(context.Context, *backend_types.Config, string) error {
	return nil
}

func volumeSource(volume string) string {
	source, _, _ := strings.Cut(volume, ":")
	return source
}

// TestExecStepVolumesReferenceWorkflowVolume ensures that the volume mounted
// into the steps is the workflow volume the backend creates, and that the
// Kubernetes backend never references a local host path as a PVC.
func TestExecStepVolumesReferenceWorkflowVolume(t *testing.T) {
	repoDir := t.TempDir()
	workflowPath := filepath.Join(repoDir, "workflow.yaml")
	require.NoError(t, os.WriteFile(workflowPath, []byte(`when:
  - event: manual

steps:
  - name: build
    image: alpine
    commands:
      - echo hello
`), 0o600))

	for _, engine := range []string{kubernetes.EngineName, "docker"} {
		t.Run(engine, func(t *testing.T) {
			recorder := &volumeRecordingBackend{name: engine}
			originalBackends := backends
			backends = []backend_types.Backend{recorder}
			t.Cleanup(func() { backends = originalBackends })

			// This is important, else the metadata of the running test leaks
			// into the exec command.
			clearEnv(t)

			require.NoError(t, Command.Run(t.Context(), []string{
				"woodpecker-cli",
				"--backend-engine", engine,
				"--repo-path", repoDir,
				workflowPath,
			}))

			require.NotEmpty(t, recorder.workflowVolume, "backend should create a workflow volume")
			require.NotEmpty(t, recorder.stepVolumes, "backend should start at least one step")

			workspaceMount := recorder.workflowVolume + ":" + testWorkspaceBase
			for _, volumes := range recorder.stepVolumes {
				assert.Contains(t, volumes, workspaceMount, "steps must mount the workflow volume")

				if engine == kubernetes.EngineName {
					for _, volume := range volumes {
						assert.Equal(t, recorder.workflowVolume, volumeSource(volume),
							"kubernetes maps every step volume to a PVC, so steps must only reference the workflow volume")
					}
				}
			}

			if engine != kubernetes.EngineName {
				assert.Contains(t, recorder.stepVolumes[0],
					repoDir+":"+testWorkspaceBase+"/"+testWorkspacePath,
					"backends running on the host must still mount the local repository")
			}
		})
	}
}
