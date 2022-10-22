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
	"encoding/base64"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/shared/constant"
)

type local struct {
	cmd        *exec.Cmd
	output     io.ReadCloser
	workingdir string
}

// make sure local implements Engine
var _ types.Engine = &local{}

// New returns a new local Engine.
func New() types.Engine {
	return &local{}
}

func (e *local) Name() string {
	return "local"
}

func (e *local) IsAvailable() bool {
	return true
}

func (e *local) Load() error {
	dir, err := os.MkdirTemp("", "woodpecker-local-*")
	e.workingdir = dir
	return err
}

// Setup the pipeline environment.
func (e *local) Setup(ctx context.Context, proc *types.Config) error {
	return nil
}

// Exec the pipeline step.
func (e *local) Exec(ctx context.Context, proc *types.Step) error {
	// Get environment variables
	Env := os.Environ()
	for a, b := range proc.Environment {
		if a != "HOME" && a != "SHELL" { // Don't override $HOME and $SHELL
			Env = append(Env, a+"="+b)
		}
	}

	Command := []string{}
	if proc.Image == constant.DefaultCloneImage {
		// Default clone step
		Env = append(Env, "CI_WORKSPACE="+e.workingdir+"/"+proc.Environment["CI_REPO"])
		Command = append(Command, "plugin-git")
	} else {
		// Use "image name" as run command
		Command = append(Command, proc.Image)
		Command = append(Command, "-c")

		// Decode script and delete initial lines
		// Deleting the initial lines removes netrc support but adds compatibility for more shells like fish
		Script, _ := base64.RawStdEncoding.DecodeString(proc.Environment["CI_SCRIPT"])
		Command = append(Command, string(Script)[strings.Index(string(Script), "\n\n")+2:])
	}

	// Prepare command
	e.cmd = exec.CommandContext(ctx, Command[0], Command[1:]...)
	e.cmd.Env = Env

	// Prepare working directory
	if proc.Image == constant.DefaultCloneImage {
		e.cmd.Dir = e.workingdir + "/" + proc.Environment["CI_REPO_OWNER"]
	} else {
		e.cmd.Dir = e.workingdir + "/" + proc.Environment["CI_REPO"]
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
	var (
		exitCode  = 0
		exitError = &exec.ExitError{}
	)

	if errors.As(err, &exitError) {
		exitCode = exitError.ExitCode()
		// Non-zero exit code is a pipeline failure, but not an agent error.
		err = nil
	}
	return &types.State{
		Exited:   true,
		ExitCode: exitCode,
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
