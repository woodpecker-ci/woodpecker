// Copyright 2022 Woodpecker Authors
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
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/alessio/shellescape"
	"golang.org/x/exp/slices"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/shared/constant"
)

// notAllowedEnvVarOverwrites are all env vars that can not be overwritten by step config
var notAllowedEnvVarOverwrites = []string{
	"CI_NETRC_MACHINE",
	"CI_NETRC_USERNAME",
	"CI_NETRC_PASSWORD",
	"CI_SCRIPT",
	"HOME",
	"SHELL",
}

type local struct {
	// TODO: make cmd a cmd list to iterate over, the hard part is to have a common ReadCloser
	cmd        *exec.Cmd
	output     io.ReadCloser
	workingdir string
}

// New returns a new local Engine.
func New() types.Engine {
	return &local{}
}

func (e *local) Name() string {
	return "local"
}

func (e *local) IsAvailable(context.Context) bool {
	return true
}

func (e *local) Load(context.Context) error {
	dir, err := os.MkdirTemp("", "woodpecker-local-*")
	e.workingdir = dir
	return err
}

// Setup the pipeline environment.
func (e *local) Setup(_ context.Context, _ *types.Config) error {
	return nil
}

// Exec the pipeline step.
func (e *local) Exec(ctx context.Context, step *types.Step) error {
	// Get environment variables
	env := os.Environ()
	for a, b := range step.Environment {
		// append allowed env vars to command env
		if !slices.Contains(notAllowedEnvVarOverwrites, a) {
			env = append(env, a+"="+b)
		}
	}

	var command []string
	if step.Image == constant.DefaultCloneImage {
		// Default clone step
		// TODO: creat tmp HOME and insert netrc
		// TODO: download plugin-git binary if not exist
		env = append(env, "CI_WORKSPACE="+e.workingdir+"/"+step.Environment["CI_REPO"])
		command = append(command, "plugin-git")
	} else {
		// Use "image name" as run command
		command = append(command, step.Image)
		command = append(command, "-c")

		// TODO: use commands directly
		script := ""
		for _, cmd := range step.Commands {
			script += fmt.Sprintf("echo + %s\n%s\n\n", shellescape.Quote(cmd), cmd)
		}
		script = strings.TrimSpace(script)

		// Deleting the initial lines removes netrc support but adds compatibility for more shells like fish
		command = append(command, script)
	}

	// Prepare command
	e.cmd = exec.CommandContext(ctx, command[0], command[1:]...)
	e.cmd.Env = env

	// Prepare working directory
	if step.Image == constant.DefaultCloneImage {
		e.cmd.Dir = e.workingdir + "/" + step.Environment["CI_REPO_OWNER"]
	} else {
		e.cmd.Dir = e.workingdir + "/" + step.Environment["CI_REPO"]
	}
	err := os.MkdirAll(e.cmd.Dir, 0o700)
	if err != nil {
		return err
	}
	// Get output and redirect Stderr to Stdout
	e.output, _ = e.cmd.StdoutPipe()
	e.cmd.Stderr = e.cmd.Stdout

	return e.cmd.Start()
}

// Wait for the pipeline step to complete and returns
// the completion results.
func (e *local) Wait(context.Context, *types.Step) (*types.State, error) {
	err := e.cmd.Wait()
	ExitCode := 0

	var execExitError *exec.ExitError
	if errors.As(err, &execExitError) {
		ExitCode = execExitError.ExitCode()
		// Non-zero exit code is a pipeline failure, but not an agent error.
		err = nil
	}

	return &types.State{
		Exited:   true,
		ExitCode: ExitCode,
	}, err
}

// Tail the pipeline step logs.
func (e *local) Tail(context.Context, *types.Step) (io.ReadCloser, error) {
	return e.output, nil
}

// Destroy the pipeline environment.
func (e *local) Destroy(context.Context, *types.Config) error {
	return os.RemoveAll(e.cmd.Dir)
}
