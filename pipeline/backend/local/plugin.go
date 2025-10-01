// Copyright 2025 Woodpecker Authors
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

package local

import (
	"context"
	"fmt"
	"os/exec"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// execPlugin use step.Image as exec binary.
func (e *local) execPlugin(ctx context.Context, step *types.Step, state *workflowState, env []string) error {
	binary, err := exec.LookPath(step.Image)
	if err != nil {
		return fmt.Errorf("lookup plugin binary: %w", err)
	}

	cmd := exec.CommandContext(ctx, binary)
	cmd.Env = env
	cmd.Dir = state.workspaceDir

	reader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// Get output and redirect Stderr to Stdout
	cmd.Stderr = cmd.Stdout

	// Save state
	state.stepCMDs.Store(step.UUID, cmd)
	state.stepOutputs.Store(step.UUID, reader)

	return cmd.Start()
}
