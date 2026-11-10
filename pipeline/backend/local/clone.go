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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// checkGitCloneCap check if we have the git binary on hand.
func checkGitCloneCap() error {
	_, err := exec.LookPath("git")
	return err
}

// loadClone on backend start determine if there is a global plugin-git binary.
func (e *local) loadClone() {
	binary, err := exec.LookPath("plugin-git")
	if err != nil || binary == "" {
		// could not found global git plugin, just ignore it
		return
	}
	e.pluginGitBinary = binary
}

// setupClone prepare the clone environment before exec.
func (e *local) setupClone(ctx context.Context, state *workflowState) error {
	if e.pluginGitBinary != "" {
		state.pluginGitBinary = e.pluginGitBinary
		return nil
	}

	log.Info().Msg("no global 'plugin-git' installed, try to download for current workflow")
	state.pluginGitBinary = filepath.Join(state.homeDir, "plugin-git")
	if e.os == "windows" {
		state.pluginGitBinary += ".exe"
	}
	return e.downloadGitPluginBinary(ctx, state.pluginGitBinary)
}

// execClone executes a clone-step locally.
func (e *local) execClone(ctx context.Context, step *types.Step, state *workflowState, env []string) error {
	if err := checkGitCloneCap(); err != nil {
		return fmt.Errorf("check for git clone capabilities failed: %w", err)
	}

	if err := e.setupClone(ctx, state); err != nil {
		return fmt.Errorf("setup clone step failed: %w", err)
	}

	if !strings.Contains(step.Image, "plugin-git") {
		log.Warn().Msgf("clone step image '%s' does not match default git clone image. We ignore it and use our plugin-git anyway.", step.Image)
	}

	rmCmd, err := e.writeNetRC(step, state)
	if err != nil {
		return err
	}

	// Prepare command
	var cmd *exec.Cmd
	if rmCmd != "" {
		// if we have a netrc injected we have to make sure it's deleted in any case after clone was attempted
		if e.os == "windows" {
			pwsh, err := exec.LookPath("powershell.exe")
			if err != nil {
				return err
			}
			cmd = newCmd(ctx, pwsh, "-Command", fmt.Sprintf("%s ; $code=$? ; %s ; if (!$code) {[Environment]::Exit(1)}", state.pluginGitBinary, rmCmd))
		} else {
			cmd = newCmd(ctx, "/bin/sh", "-c", fmt.Sprintf("%s ; export code=$? ; %s ; exit $code", state.pluginGitBinary, rmCmd))
		}
	} else {
		// if we have NO netrc, we can just exec the clone directly
		cmd = newCmd(ctx, state.pluginGitBinary)
	}
	cmd.Env = env
	cmd.Dir = state.workspaceDir

	reader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// Save state
	state.stepState.Store(step.UUID, &stepState{
		cmd:    cmd,
		output: reader,
	})

	// Get output and redirect Stderr to Stdout
	cmd.Stderr = cmd.Stdout

	return cmd.Start()
}

// writeNetRC write a netrc file into the home dir of a given workflow state.
func (e *local) writeNetRC(step *types.Step, state *workflowState) (string, error) {
	if step.Environment["CI_NETRC_MACHINE"] == "" {
		log.Trace().Msg("no netrc to write")
		return "", nil
	}

	if !e.isolatedHome {
		log.Trace().Msg("writing .netrc skipped due to disabled isolated home")
		return "", nil
	}

	file := filepath.Join(state.homeDir, ".netrc")
	rmCmd := fmt.Sprintf("rm \"%s\"", file)
	if e.os == "windows" {
		file = filepath.Join(state.homeDir, "_netrc")
		rmCmd = fmt.Sprintf("del \"%s\"", file)
	}

	log.Trace().Msgf("try to write netrc to '%s'", file)
	return rmCmd, os.WriteFile(file, []byte(genNetRC(step.Environment)), 0o600)
}

// pluginGitReleaseURL is the base URL of the plugin-git releases (overridden in tests).
var pluginGitReleaseURL = "https://github.com/woodpecker-ci/plugin-git/releases"

// downloadGitPluginBinary downloads the plugin-git binary of the configured version for the
// runtime OS and Arch, verifies it against the release's checksums.txt and saves it to dest.
func (e *local) downloadGitPluginBinary(ctx context.Context, dest string) error {
	asset := fmt.Sprintf("%s-%s_plugin-git", e.os, e.arch)
	if e.os == "windows" {
		asset += ".exe"
	}

	baseURL := pluginGitReleaseURL + "/latest/download"
	if e.pluginGitVersion != "" && e.pluginGitVersion != "latest" {
		baseURL = pluginGitReleaseURL + "/download/" + url.PathEscape(e.pluginGitVersion)
	}

	checksums, err := httpGet(ctx, baseURL+"/checksums.txt")
	if err != nil {
		return fmt.Errorf("could not download plugin-git checksums: %w", err)
	}
	var wantSum string
	for _, line := range strings.Split(string(checksums), "\n") {
		if fields := strings.Fields(line); len(fields) == 2 && fields[1] == asset {
			wantSum = fields[0]
			break
		}
	}
	if wantSum == "" {
		return fmt.Errorf("could not download plugin-git: no release asset '%s' for this os/arch", asset)
	}

	binary, err := httpGet(ctx, baseURL+"/"+asset)
	if err != nil {
		return fmt.Errorf("could not download plugin-git: %w", err)
	}
	if sum := sha256.Sum256(binary); hex.EncodeToString(sum[:]) != wantSum {
		return fmt.Errorf("could not download plugin-git: checksum mismatch for '%s'", asset)
	}

	if err := os.WriteFile(dest, binary, 0o755); err != nil {
		return fmt.Errorf("could not write plugin-git: %w", err)
	}

	log.Trace().Msgf("download of 'plugin-git' to '%s' successful", dest)
	return nil
}

// httpGet returns the body of a GET request and fails on non-200 responses.
func httpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: unexpected status %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}
