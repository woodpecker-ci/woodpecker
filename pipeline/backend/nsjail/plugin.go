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

package nsjail

import (
	"context"
	"fmt"
	"os/exec"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// execPlugin executes a plugin step inside the nsjail sandbox.
// step.Image is the plugin binary path, mounted into the sandbox via bindmount.
func (e *nsjail) execPlugin(ctx context.Context, step *types.Step, state *workflowState, env []string) error {
	binary, err := exec.LookPath(step.Image)
	if err != nil {
		return fmt.Errorf("lookup plugin binary: %w", err)
	}

	// Mount plugin binary into sandbox and construct nsjail command
	// Inside sandbox, mount to /bin/plugin-bin
	pluginInJail := "/bin/plugin-bin"
	nsjailArgs := e.buildNsJailArgs(step, state, env)

	// Find "--" separator and truncate (keep "--")
	sepIdx := len(nsjailArgs) - 1
	for i := len(nsjailArgs) - 1; i >= 0; i-- {
		if nsjailArgs[i] == "--" {
			sepIdx = i
			break
		}
	}
	nsjailArgs = nsjailArgs[:sepIdx+1]

	// Insert bindmount before "--"
	extraArgs := []string{"--bindmount_ro", binary + ":" + pluginInJail}
	nsjailArgs = append(nsjailArgs[:sepIdx], append(extraArgs, nsjailArgs[sepIdx:]...)...)

	// Append plugin binary as the command to execute
	nsjailArgs = append(nsjailArgs, pluginInJail)

	cmd := newNsJailCmd(ctx, e.cfg.binPath, nsjailArgs...)
	cmd.Env = env
	cmd.Dir = state.workspaceDir

	reader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout

	state.stepState.Store(step.UUID, &stepState{cmd: cmd, output: reader})
	return cmd.Start()
}
