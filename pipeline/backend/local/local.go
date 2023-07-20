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
	"path/filepath"
	"strings"
	"sync"

	"github.com/alessio/shellescape"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

type workflowState struct {
	stepCMDs        map[string]*exec.Cmd
	baseDir         string
	homeDir         string
	workspaceDir    string
	pluginGitBinary string
}

type local struct {
	workflows       sync.Map
	output          io.ReadCloser
	pluginGitBinary string
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
	e.loadClone()

	return nil
}

// SetupWorkflow the pipeline environment.
func (e *local) SetupWorkflow(_ context.Context, _ *types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msg("create workflow environment")

	baseDir, err := os.MkdirTemp("", "woodpecker-local-*")
	if err != nil {
		return err
	}

	state := &workflowState{
		stepCMDs:     make(map[string]*exec.Cmd),
		baseDir:      baseDir,
		workspaceDir: filepath.Join(baseDir, "workspace"),
		homeDir:      filepath.Join(baseDir, "home"),
	}

	if err := os.Mkdir(state.homeDir, 0o700); err != nil {
		return err
	}

	if err := os.Mkdir(state.workspaceDir, 0o700); err != nil {
		return err
	}

	e.saveState(taskUUID, state)

	return nil
}

// StartStep the pipeline step.
func (e *local) StartStep(ctx context.Context, step *types.Step, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("start step %s", step.Name)

	state, err := e.getState(taskUUID)
	if err != nil {
		return err
	}

	// Get environment variables
	env := os.Environ()
	for a, b := range step.Environment {
		// append allowed env vars to command env
		if !slices.Contains(notAllowedEnvVarOverwrites, a) {
			env = append(env, a+"="+b)
		}
	}

	// Set HOME
	env = append(env, "HOME="+state.homeDir)

	switch step.Type {
	case types.StepTypeClone:
		return e.execClone(ctx, step, state, env)
	case types.StepTypeCommands:
		return e.execCommands(ctx, step, state, env)
	default:
		return ErrUnsupportedStepType
	}
}

func (e *local) execCommands(ctx context.Context, step *types.Step, state *workflowState, env []string) error {
	// TODO: use commands directly
	script := ""
	for _, cmd := range step.Commands {
		script += fmt.Sprintf("echo + %s\n%s", strings.TrimSpace(shellescape.Quote(cmd)), cmd)
	}
	script = strings.TrimSpace(script)

	// Prepare command
	// Use "image name" as run command (indicate shell)
	cmd := exec.CommandContext(ctx, step.Image, "-c", script)
	cmd.Env = env
	cmd.Dir = state.workspaceDir

	// Get output and redirect Stderr to Stdout
	e.output, _ = cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	state.stepCMDs[step.Name] = cmd

	return cmd.Start()
}

// WaitStep for the pipeline step to complete and returns
// the completion results.
func (e *local) WaitStep(_ context.Context, step *types.Step, taskUUID string) (*types.State, error) {
	log.Trace().Str("taskUUID", taskUUID).Msgf("wait for step %s", step.Name)

	state, err := e.getState(taskUUID)
	if err != nil {
		return nil, err
	}

	cmd, ok := state.stepCMDs[step.Name]
	if !ok {
		return nil, fmt.Errorf("step cmd %s not found", step.Name)
	}

	err = cmd.Wait()
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

// TailStep the pipeline step logs.
func (e *local) TailStep(_ context.Context, step *types.Step, taskUUID string) (io.ReadCloser, error) {
	log.Trace().Str("taskUUID", taskUUID).Msgf("tail logs of step %s", step.Name)
	return e.output, nil
}

// DestroyWorkflow the pipeline environment.
func (e *local) DestroyWorkflow(_ context.Context, _ *types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("delete workflow environment")

	state, err := e.getState(taskUUID)
	if err != nil {
		return err
	}

	err = os.RemoveAll(state.baseDir)
	if err != nil {
		return err
	}

	e.deleteState(taskUUID)

	return err
}

func (e *local) getState(taskUUID string) (*workflowState, error) {
	state, ok := e.workflows.Load(taskUUID)
	if !ok {
		return nil, ErrWorkflowStateNotFound
	}
	return state.(*workflowState), nil
}

func (e *local) saveState(taskUUID string, state *workflowState) {
	e.workflows.Store(taskUUID, state)
}

func (e *local) deleteState(taskUUID string) {
	e.workflows.Delete(taskUUID)
}
