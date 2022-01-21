package local

import (
	"context"
	"encoding/base64"
	"io"
	"os"
	"os/exec"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

type local struct {
	cmd    *exec.Cmd
	output io.ReadCloser
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
	return nil
}

// Setup the pipeline environment.
func (e *local) Setup(ctx context.Context, proc *types.Config) error {
	return nil
}

// Exec the pipeline step.
func (e *local) Exec(ctx context.Context, proc *types.Step) error {
	// Get environment variables
	Command := []string{}
	for a, b := range proc.Environment {
		if a != "HOME" && a != "SHELL" {
			Command = append(Command, a+"="+b)
		}
	}

	// Use "image name" as run command
	Command = append(Command, proc.Image[18:len(proc.Image)-7])
	Command = append(Command, "-c")

	// Decode script and remove initial lines
	Script, _ := base64.RawStdEncoding.DecodeString(proc.Environment["CI_SCRIPT"])
	Command = append(Command, string(Script))

	// Prepare command and working directory
	e.cmd = exec.CommandContext(ctx, "/bin/env", Command...)
	e.cmd.Dir = "/tmp/" + proc.Environment["CI_REPO"]
	_ = os.MkdirAll(e.cmd.Dir, 0o700)

	// Get output and redirect Stderr to Stdout
	e.output, _ = e.cmd.StdoutPipe()
	e.cmd.Stderr = e.cmd.Stdout

	return e.cmd.Start()
}

// Wait for the pipeline step to complete and returns
// the completion results.
func (e *local) Wait(context.Context, *types.Step) (*types.State, error) {
	return &types.State{
		Exited: true,
	}, e.cmd.Wait()
}

// Tail the pipeline step logs.
func (e *local) Tail(context.Context, *types.Step) (io.ReadCloser, error) {
	return e.output, nil
}

// Destroy the pipeline environment.
func (e *local) Destroy(context.Context, *types.Config) error {
	os.RemoveAll(e.cmd.Dir)
	return nil
}
