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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// checkGitCloneCap checks if the git command is available on the host.
func checkGitCloneCap() error {
	_, err := exec.LookPath("git")
	return err
}

// loadClone on backend start determines if there is a global plugin-git binary.
func (e *nsjail) loadClone() {
	binary, err := exec.LookPath("plugin-git")
	if err != nil || binary == "" {
		return
	}
	e.pluginGitBinary = binary
}

// setupClone prepares the clone environment before exec.
func (e *nsjail) setupClone(ctx context.Context, state *workflowState) error {
	if e.pluginGitBinary != "" {
		state.pluginGitBinary = e.pluginGitBinary
		return nil
	}

	log.Info().Msg("no global 'plugin-git' installed, download for current workflow")
	state.pluginGitBinary = filepath.Join(state.homeDir, "plugin-git")
	if e.os == "windows" {
		state.pluginGitBinary += ".exe"
	}
	return e.downloadLatestGitPluginBinary(ctx, state.pluginGitBinary)
}

// execClone executes a clone-step inside the nsjail sandbox.
func (e *nsjail) execClone(ctx context.Context, step *types.Step, state *workflowState, env []string) error {
	if err := checkGitCloneCap(); err != nil {
		return fmt.Errorf("check for git clone capabilities failed: %w", err)
	}

	if err := e.setupClone(ctx, state); err != nil {
		return fmt.Errorf("setup clone step failed: %w", err)
	}

	if !strings.Contains(step.Image, "plugin-git") {
		log.Warn().Msgf("clone step image '%s' does not match default git clone image", step.Image)
	}

	rmCmd, netrcPath, err := e.writeNetRC(step, state)
	if err != nil {
		return err
	}

	// Build nsjail args (without the command after "--")
	nsjailArgs := e.buildNsJailArgsForClone(step, state, env, state.pluginGitBinary, netrcPath)

	// Append actual command (after "--")
	var cmd *exec.Cmd
	if rmCmd != "" {
		// Execute plugin-git in sandbox, clean up netrc after
		nsjailArgs = append(nsjailArgs,
			"/bin/sh", "-c",
			fmt.Sprintf("/bin/plugin-git ; export code=$? ; rm -f '%s/.netrc' ; exit $code",
				state.homeDir),
		)
		cmd = newNsJailCmd(ctx, e.cfg.binPath, nsjailArgs...)
	} else {
		nsjailArgs = append(nsjailArgs, "/bin/plugin-git")
		cmd = newNsJailCmd(ctx, e.cfg.binPath, nsjailArgs...)
	}
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

// writeNetRC writes a netrc file into the home dir of a given workflow state.
// Returns (rmCmd, netrcFilePath, error) to support netrc file injection into the nsjail sandbox.
func (e *nsjail) writeNetRC(step *types.Step, state *workflowState) (string, string, error) {
	if step.Environment["CI_NETRC_MACHINE"] == "" {
		log.Trace().Msg("no netrc to write")
		return "", "", nil
	}

	if !e.cfg.isolatedHome {
		log.Trace().Msg("writing .netrc skipped due to disabled isolated home")
		return "", "", nil
	}

	file := filepath.Join(state.homeDir, ".netrc")
	rmCmd := fmt.Sprintf("rm \"%s\"", file)
	if e.os == "windows" {
		file = filepath.Join(state.homeDir, "_netrc")
		rmCmd = fmt.Sprintf("del \"%s\"", file)
	}

	log.Trace().Msgf("try to write netrc to '%s'", file)
	return rmCmd, file, os.WriteFile(file, []byte(genNetRC(step.Environment)), 0o600)
}

// downloadLatestGitPluginBinary downloads the latest plugin-git binary based on runtime OS and Arch
// and saves it to dest.
func (e *nsjail) downloadLatestGitPluginBinary(ctx context.Context, dest string) error {
	type asset struct {
		Name               string
		BrowserDownloadURL string `json:"browser_download_url"`
	}

	type release struct {
		Assets []asset
	}

	// get latest release
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/woodpecker-ci/plugin-git/releases/latest", nil)
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
		if strings.Contains(at.Name, e.os) && strings.Contains(at.Name, e.arch) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, at.BrowserDownloadURL, nil)
			if err != nil {
				return err
			}
			assetResp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("could not download plugin-git: %w", err)
			}
			defer assetResp.Body.Close()

			file, err := os.Create(dest)
			if err != nil {
				return fmt.Errorf("could not create plugin-git: %w", err)
			}
			defer file.Close()

			if _, err := io.Copy(file, assetResp.Body); err != nil {
				return fmt.Errorf("could not download plugin-git: %w", err)
			}
			if err := os.Chmod(dest, 0o755); err != nil {
				return err
			}

			log.Trace().Msgf("download of 'plugin-git' to '%s' successful", dest)
			return nil
		}
	}

	return fmt.Errorf("could not download plugin-git, binary for this os/arch not found")
}
