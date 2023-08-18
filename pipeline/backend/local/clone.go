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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/shared/constant"
)

// checkGitCloneCap check if we have the git binary on hand
func checkGitCloneCap() error {
	_, err := exec.LookPath("git")
	return err
}

// loadClone on backend start determine if there is a global plugin-git binary
func (e *local) loadClone() {
	binary, err := exec.LookPath("plugin-git")
	if err != nil || binary == "" {
		// could not found global git plugin, just ignore it
		return
	}
	e.pluginGitBinary = binary
}

// setupClone prepare the clone environment before exec
func (e *local) setupClone(state *workflowState) error {
	if e.pluginGitBinary != "" {
		state.pluginGitBinary = e.pluginGitBinary
		return nil
	}

	log.Info().Msg("no global 'plugin-git' installed, try to download for current workflow")
	state.pluginGitBinary = filepath.Join(state.homeDir, "plugin-git")
	if runtime.GOOS == "windows" {
		state.pluginGitBinary += ".exe"
	}
	return downloadLatestGitPluginBinary(state.pluginGitBinary)
}

// execClone executes a clone-step locally
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
	if rmCmd != "" {
		// if we have a netrc injected we have to make sure it's deleted in any case after clone was attempted
		if runtime.GOOS == "windows" {
			pwsh, err := exec.LookPath("powershell.exe")
			if err != nil {
				return err
			}
			cmd = exec.CommandContext(ctx, pwsh, "-Command", fmt.Sprintf("%s ; $code=$? ; %s ; if (!$code) {[Environment]::Exit(1)}", state.pluginGitBinary, rmCmd))
		} else {
			cmd = exec.CommandContext(ctx, "/bin/sh", "-c", fmt.Sprintf("%s ; export code=$? ; %s ; exit $code", state.pluginGitBinary, rmCmd))
		}
	} else {
		// if we have NO netrc, we can just exec the clone directly
		cmd = exec.CommandContext(ctx, state.pluginGitBinary)
	}
	cmd.Env = env
	cmd.Dir = state.workspaceDir

	// Get output and redirect Stderr to Stdout
	e.output, _ = cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	state.stepCMDs[step.Name] = cmd

	return cmd.Start()
}

// writeNetRC write a netrc file into the home dir of a given workflow state
func writeNetRC(step *types.Step, state *workflowState) (string, error) {
	if step.Environment["CI_NETRC_MACHINE"] == "" {
		log.Trace().Msg("no netrc to write")
		return "", nil
	}

	file := filepath.Join(state.homeDir, ".netrc")
	rmCmd := fmt.Sprintf("rm \"%s\"", file)
	if runtime.GOOS == "windows" {
		file = filepath.Join(state.homeDir, "_netrc")
		rmCmd = fmt.Sprintf("del \"%s\"", file)
	}

	log.Trace().Msgf("try to write netrc to '%s'", file)
	return rmCmd, os.WriteFile(file, []byte(fmt.Sprintf(
		netrcFile,
		step.Environment["CI_NETRC_MACHINE"],
		step.Environment["CI_NETRC_USERNAME"],
		step.Environment["CI_NETRC_PASSWORD"],
	)), 0o600)
}

// downloadLatestGitPluginBinary download the latest plugin-git binary based on runtime OS and Arch
// and saves it to dest
func downloadLatestGitPluginBinary(dest string) error {
	type asset struct {
		Name               string
		BrowserDownloadURL string `json:"browser_download_url"`
	}

	type release struct {
		Assets []asset
	}

	// get latest release
	req, _ := http.NewRequest(http.MethodGet, "https://api.github.com/repos/woodpecker-ci/plugin-git/releases/latest", nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not get latest release: %w", err)
	}
	raw, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	var rel release
	if err := json.Unmarshal(raw, &rel); err != nil {
		return fmt.Errorf("could not unmarshal github response: %w", err)
	}

	for _, at := range rel.Assets {
		if strings.Contains(at.Name, runtime.GOOS) && strings.Contains(at.Name, runtime.GOARCH) {
			resp2, err := http.Get(at.BrowserDownloadURL)
			if err != nil {
				return fmt.Errorf("could not download plugin-git: %w", err)
			}
			defer resp2.Body.Close()

			file, err := os.Create(dest)
			if err != nil {
				return fmt.Errorf("could not create plugin-git: %w", err)
			}
			defer file.Close()

			if _, err := io.Copy(file, resp2.Body); err != nil {
				return fmt.Errorf("could not download plugin-git: %w", err)
			}
			if err := os.Chmod(dest, 0o755); err != nil {
				return err
			}

			// download successful
			log.Trace().Msgf("download of 'plugin-git' to '%s' successful", dest)
			return nil
		}
	}

	return fmt.Errorf("could not download plugin-git, binary for this os/arch not found")
}
