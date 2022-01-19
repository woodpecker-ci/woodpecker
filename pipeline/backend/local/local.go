package local

import (
	"context"
	"io"
	"os/exec"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

type local struct {
	cmd *exec.Cmd
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
	Command := ""
	for _, command := range proc.Command {
		Command += command + "\n"
	}
	
	e.cmd = exec.CommandContext(ctx, "bash", "-c", Command)

	e.output, _ = e.cmd.StdoutPipe()

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
	return nil
}
