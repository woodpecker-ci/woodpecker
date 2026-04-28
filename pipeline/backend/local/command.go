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

// cSpell:ignore ERRORLEVEL

package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"al.essio.dev/pkg/shellescape"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// execCommands use step.Image as shell and run the commands in it.
func (e *local) execCommands(ctx context.Context, step *types.Step, state *workflowState, env []string) error {
	if err := checkShellExistence(step.Image); err != nil {
		return err
	}

	// Prepare commands
	// TODO: support `entrypoint` from pipeline config
	args, err := e.genCmdByShell(step.Image, step.Commands, state.baseDir)
	if err != nil {
		return fmt.Errorf("could not convert commands into args: %w", err)
	}

	// Use "image name" as run command (indicate shell)
	cmd := exec.CommandContext(ctx, step.Image, args...)
	cmd.Env = env
	cmd.Dir = state.workspaceDir

	reader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if e.os == "windows" {
		// we get non utf8 output from windows so just sanitize it
		// TODO: remove hack
		reader = io.NopCloser(transform.NewReader(reader, unicode.UTF8.NewDecoder().Transformer))
	}

	// Get output and redirect Stderr to Stdout
	cmd.Stderr = cmd.Stdout

	// Save state
	state.stepState.Store(step.UUID, &stepState{
		cmd:    cmd,
		output: reader,
	})

	return cmd.Start()
}

func checkShellExistence(shell string) error {
	_, err := exec.LookPath(shell)
	return err
}

func (e *local) genCmdByShell(shell string, cmdList []string, baseDir string) (args []string, err error) {
	if len(cmdList) == 0 {
		return nil, ErrNoCmdSet
	}

	script := ""
	for _, cmd := range cmdList {
		script += fmt.Sprintf("echo %s\n%s\n", strings.TrimSpace(shellescape.Quote("+ "+cmd)), cmd)
	}
	script = strings.TrimSpace(script)

	shell = strings.TrimSuffix(strings.ToLower(shell), ".exe")
	switch shell {
	default:
		// assume posix shell
		if err := probeShellIsPosix(shell); err != nil {
			return nil, err
		}
		fallthrough
		// normal posix shells
	case "sh", "bash", "zsh":
		return []string{"-e", "-c", script}, nil
	case "":
		return nil, ErrNoShellSet
	case "cmd":
		script := "@SET PROMPT=$\n"
		for _, cmd := range cmdList {
			quotedCmd := strings.TrimSpace(shellescape.Quote(cmd))
			// As cmd echo does not allow strings with newlines we need to replace them ...
			quotedCmd = strings.ReplaceAll(quotedCmd, "\n", "\\n")
			// Also the shellescape.Quote fail with any | or & char and wrapping them in quotes again can be bypassed
			// by just leaving an string halve quoted we just replace them with symbolic representations
			quotedCmd = strings.ReplaceAll(quotedCmd, "&", "\\AND")
			quotedCmd = strings.ReplaceAll(quotedCmd, "|", "\\OR")

			script += fmt.Sprintf("@echo + %s\n", quotedCmd)
			script += fmt.Sprintf("@%s\n", cmd)
			script += "@IF NOT %ERRORLEVEL% == 0 exit %ERRORLEVEL%\n"
		}
		cmd, err := os.CreateTemp(baseDir, "*.cmd")
		if err != nil {
			return nil, err
		}
		defer cmd.Close()
		if _, err := cmd.WriteString(script); err != nil {
			return nil, err
		}
		return []string{"/c", cmd.Name()}, nil
	case "fish":
		script := ""
		for _, cmd := range cmdList {
			script += fmt.Sprintf("echo %s\n%s || exit $status\n", strings.TrimSpace(shellescape.Quote("+ "+cmd)), cmd)
		}
		return []string{"-c", script}, nil
	case "nu":
		return []string{"--commands", script}, nil
	case "powershell", "pwsh":
		// cspell:disable-next-line
		return []string{"-noprofile", "-noninteractive", "-c", "$ErrorActionPreference = \"Stop\"; " + script}, nil
	}
}

// before we generate a generic posix shell we test.
func probeShellIsPosix(shell string) error {
	script := `x=1 && [ "$x" = "1" ] && command -v test >/dev/null && printf ok`

	cmd := exec.Command(shell, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ErrNoPosixShell{Shell: shell, Err: err}
	}

	if strings.TrimSpace(string(output)) != "ok" {
		return &ErrNoPosixShell{Shell: shell, Err: fmt.Errorf("unexpected output returned: %q", string(output))}
	}

	return nil
}
