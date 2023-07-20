// Copyright 2023 Woodpecker Authors
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
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/shared/constant"
)

func checkGitCloneCap() error {
	_, err := exec.LookPath("git")
	return err
}

func (e *local) loadClone() {
	binary, err := exec.LookPath("plugin-git")
	if err != nil || binary == "" {
		// could not found global git plugin, just ignore it
		return
	}
	e.pluginGitBinary = binary
}

func (e *local) setupClone(state *workflowState) error {
	if e.pluginGitBinary != "" {
		state.pluginGitBinary = e.pluginGitBinary
		return nil
	}

	log.Info().Msg("no global 'plugin-git' installed, try to download for current workflow")
	// TODO: download plugin-git binary to homeDir and set PATH
	return fmt.Errorf("download not implemented")
}

func (e *local) execClone(ctx context.Context, step *types.Step, state *workflowState, env []string) error {
	if err := e.setupClone(state); err != nil {
		return fmt.Errorf("setup clone step failed: %w", err)
	}

	if err := checkGitCloneCap(); err != nil {
		return fmt.Errorf("check for git clone capabilities failed: %w", err)
	}

	if step.Image != constant.DefaultCloneImage {
		// TODO: write message into log
		log.Warn().Msgf("clone step image '%s' does not match default git clone image. We ignore it assume git.", step.Image)
	}

	rmCmd, err := writeNetRC(step, state)
	if err != nil {
		return err
	}

	env = append(env, "CI_WORKSPACE="+state.workspaceDir)

	// Prepare command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		pwsh, err := exec.LookPath("powershell.exe")
		if err != nil {
			return err
		}
		cmd = exec.CommandContext(ctx, pwsh, "-Command", fmt.Sprintf("%s ; $code=$? ; %s ; if (!$code) {[Environment]::Exit(1)}", state.pluginGitBinary, rmCmd))
	} else {
		cmd = exec.CommandContext(ctx, "/bin/sh", "-c", fmt.Sprintf("%s ; $code=$? ; %s ; exit $code", state.pluginGitBinary, rmCmd))
	}
	cmd.Env = env
	cmd.Dir = state.workspaceDir

	// Get output and redirect Stderr to Stdout
	e.output, _ = cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	state.stepCMDs[step.Name] = cmd

	return cmd.Start()
}

func writeNetRC(step *types.Step, state *workflowState) (string, error) {
	if step.Environment["CI_NETRC_MACHINE"] == "" {
		return "", nil
	}

	file := filepath.Join(state.homeDir, ".netrc")
	rmCmd := fmt.Sprintf("rm \"%s\"", file)
	if runtime.GOOS == "windows" {
		file = filepath.Join(state.homeDir, "_netrc")
		rmCmd = fmt.Sprintf("del \"%s\"", file)
	}

	return rmCmd, os.WriteFile(file, []byte(fmt.Sprintf(
		netrcFile,
		step.Environment["CI_NETRC_MACHINE"],
		step.Environment["CI_NETRC_USERNAME"],
		step.Environment["CI_NETRC_PASSWORD"],
	)), 0o600)
}
