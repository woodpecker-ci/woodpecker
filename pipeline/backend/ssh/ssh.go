// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ssh

import (
	"context"
	"io"
	"strings"

	"github.com/melbahja/goph"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/common"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/shared/constant"
)

type ssh struct {
	cmd        *goph.Cmd
	output     io.ReadCloser
	client     *goph.Client
	workingdir string
}

type readCloser struct {
	io.Reader
}

func (c readCloser) Close() error {
	return nil
}

// New returns a new ssh Engine.
func New() types.Engine {
	return &ssh{}
}

func (e *ssh) Name() string {
	return "ssh"
}

func (e *ssh) IsAvailable(ctx context.Context) bool {
	c, ok := ctx.Value(types.CliContext).(*cli.Context)
	return ok && c.String("backend-ssh-address") != "" && c.String("backend-ssh-user") != "" && (c.String("backend-ssh-key") != "" || c.String("backend-ssh-password") != "")
}

func (e *ssh) Load(ctx context.Context) error {
	cmd, err := e.client.Command("/bin/env", "mktemp", "-d", "-p", "/tmp", "woodpecker-ssh-XXXXXXXXXX")
	if err != nil {
		return err
	}

	dir, err := cmd.Output()
	if err != nil {
		return err
	}

	e.workingdir = string(dir)
	c, ok := ctx.Value(types.CliContext).(*cli.Context)
	if !ok {
		return types.ErrNoCliContextFound
	}
	address := c.String("backend-ssh-address")
	user := c.String("backend-ssh-user")
	var auth goph.Auth
	if file := c.String("backend-ssh-key"); file != "" {
		keyAuth, err := goph.Key(file, c.String("backend-ssh-key-password"))
		if err != nil {
			return err
		}
		auth = append(auth, keyAuth...)
	}
	if password := c.String("backend-ssh-password"); password != "" {
		auth = append(auth, goph.Password(password)...)
	}
	client, err := goph.New(user, address, auth)
	if err != nil {
		return err
	}
	e.client = client
	return nil
}

// SetupWorkflow create the workflow environment.
func (e *ssh) SetupWorkflow(context.Context, *types.Config, string) error {
	return nil
}

// StartStep start the step.
func (e *ssh) StartStep(ctx context.Context, step *types.Step, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("Start step %s", step.Name)

	// Get environment variables
	var command []string
	for a, b := range step.Environment {
		if a != "HOME" && a != "SHELL" { // Don't override $HOME and $SHELL
			command = append(command, a+"="+b)
		}
	}

	if step.Image == constant.DefaultCloneImage {
		// Default clone step
		command = append(command, "CI_WORKSPACE="+e.workingdir+"/"+step.Environment["CI_REPO"])
		command = append(command, "plugin-git")
	} else {
		// Use "image name" as run command
		command = append(command, step.Image)
		command = append(command, "-c")

		// TODO: use commands directly
		script := common.GenerateScript(step.Commands)
		// Deleting the initial lines removes netrc support but adds compatibility for more shells like fish
		command = append(command, "cd "+e.workingdir+"/"+step.Environment["CI_REPO"]+" && "+script[strings.Index(script, "\n\n")+2:])
	}

	// Prepare command
	var err error
	e.cmd, err = e.client.CommandContext(ctx, "/bin/env", command...)
	if err != nil {
		return err
	}

	// Get output and redirect Stderr to Stdout
	std, _ := e.cmd.StdoutPipe()
	e.output = readCloser{std}
	e.cmd.Stderr = e.cmd.Stdout

	return e.cmd.Start()
}

// WaitStep for the pipeline step to complete and returns
// the completion results.
func (e *ssh) WaitStep(context.Context, *types.Step, string) (*types.State, error) {
	return &types.State{
		Exited: true,
	}, e.cmd.Wait()
}

// TailStep the pipeline step logs.
func (e *ssh) TailStep(context.Context, *types.Step, string) (io.ReadCloser, error) {
	return e.output, nil
}

// DestroyWorkflow delete the workflow environment.
func (e *ssh) DestroyWorkflow(context.Context, *types.Config, string) error {
	e.client.Close()
	sftp, err := e.client.NewSftp()
	if err != nil {
		return err
	}

	return sftp.RemoveDirectory(e.workingdir)
}
