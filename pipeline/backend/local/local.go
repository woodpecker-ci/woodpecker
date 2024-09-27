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
	"runtime"
	"slices"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

type workflowState struct {
	stepCMDs        map[string]*exec.Cmd
	baseDir         string
	homeDir         string
	workspaceDir    string
	pluginGitBinary string
}

type local struct {
	tempDir         string
	workflows       sync.Map
	output          io.ReadCloser
	pluginGitBinary string
	os, arch        string
}

// New returns a new local Backend.
func New() types.Backend {
	return &local{
		os:   runtime.GOOS,
		arch: runtime.GOARCH,
	}
}

func (e *local) Name() string {
	return "local"
}

func (e *local) IsAvailable(ctx context.Context) bool {
	if c, ok := ctx.Value(types.CliCommand).(*cli.Command); ok {
		if c.String("backend-engine") == e.Name() {
			return true
		}
	}
	_, inContainer := os.LookupEnv("WOODPECKER_IN_CONTAINER")
	return !inContainer
}

func (e *local) Flags() []cli.Flag {
	return Flags
}

func (e *local) Load(ctx context.Context) (*types.BackendInfo, error) {
	c, ok := ctx.Value(types.CliCommand).(*cli.Command)
	if ok {
		e.tempDir = c.String("backend-local-temp-dir")
	}

	e.loadClone()

	return &types.BackendInfo{
		Platform: e.os + "/" + e.arch,
	}, nil
}

// SetupWorkflow the pipeline environment.
func (e *local) SetupWorkflow(_ context.Context, _ *types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msg("create workflow environment")

	baseDir, err := os.MkdirTemp(e.tempDir, "woodpecker-local-*")
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

	// Set HOME and CI_WORKSPACE
	env = append(env, "HOME="+state.homeDir)
	env = append(env, "USERPROFILE="+state.homeDir)
	env = append(env, "CI_WORKSPACE="+state.workspaceDir)

	switch step.Type {
	case types.StepTypeClone:
		return e.execClone(ctx, step, state, env)
	case types.StepTypeCommands:
		return e.execCommands(ctx, step, state, env)
	case types.StepTypePlugin:
		return e.execPlugin(ctx, step, state, env)
	default:
		return ErrUnsupportedStepType
	}
}

// execCommands use step.Image as shell and run the commands in it.
func (e *local) execCommands(ctx context.Context, step *types.Step, state *workflowState, env []string) error {
	// Prepare commands
	// TODO: support `entrypoint` from pipeline config
	args, err := e.genCmdByShell(step.Image, step.Commands)
	if err != nil {
		return fmt.Errorf("could not convert commands into args: %w", err)
	}

	// Use "image name" as run command (indicate shell)
	cmd := exec.CommandContext(ctx, step.Image, args...)
	cmd.Env = env
	cmd.Dir = state.workspaceDir

	// Get output and redirect Stderr to Stdout
	e.output, _ = cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	if e.os == "windows" {
		// we get non utf8 output from windows so just sanitize it
		// TODO: remove hack
		e.output = io.NopCloser(transform.NewReader(e.output, unicode.UTF8.NewDecoder().Transformer))
	}

	state.stepCMDs[step.UUID] = cmd

	return cmd.Start()
}

// execPlugin use step.Image as exec binary.
func (e *local) execPlugin(ctx context.Context, step *types.Step, state *workflowState, env []string) error {
	binary, err := exec.LookPath(step.Image)
	if err != nil {
		return fmt.Errorf("lookup plugin binary: %w", err)
	}

	cmd := exec.CommandContext(ctx, binary)
	cmd.Env = env
	cmd.Dir = state.workspaceDir

	// Get output and redirect Stderr to Stdout
	e.output, _ = cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	state.stepCMDs[step.UUID] = cmd

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

	cmd, ok := state.stepCMDs[step.UUID]
	if !ok {
		return nil, fmt.Errorf("step cmd for %s not found", step.UUID)
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

func (e *local) DestroyStep(_ context.Context, _ *types.Step, _ string) error {
	// WaitStep already waits for the command to finish, so there is nothing to do here.
	return nil
}

// DestroyWorkflow the pipeline environment.
func (e *local) DestroyWorkflow(_ context.Context, _ *types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msg("delete workflow environment")

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

	s, ok := state.(*workflowState)
	if !ok {
		return nil, fmt.Errorf("could not parse state: %v", state)
	}

	return s, nil
}

func (e *local) saveState(taskUUID string, state *workflowState) {
	e.workflows.Store(taskUUID, state)
}

func (e *local) deleteState(taskUUID string) {
	e.workflows.Delete(taskUUID)
}
