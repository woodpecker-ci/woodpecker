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

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

type workflowState struct {
	stepCMDs        sync.Map // map of *exec.Cmd
	stepOutputs     sync.Map // map of io.ReadCloser
	baseDir         string
	homeDir         string
	workspaceDir    string
	pluginGitBinary string
}

type local struct {
	tempDir         string
	workflows       sync.Map
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

func (e *local) SetupWorkflow(_ context.Context, _ *types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msg("create workflow environment")

	baseDir, err := os.MkdirTemp(e.tempDir, "woodpecker-local-*")
	if err != nil {
		return err
	}

	state := &workflowState{
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

	e.workflows.Store(taskUUID, state)

	return nil
}

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

func (e *local) WaitStep(_ context.Context, step *types.Step, taskUUID string) (*types.State, error) {
	log.Trace().Str("taskUUID", taskUUID).Msgf("wait for step %s", step.Name)

	state, err := e.getState(taskUUID)
	if err != nil {
		return nil, err
	}

	cmd, ok := state.stepCMDs.Load(step.UUID)
	if !ok {
		return nil, fmt.Errorf("step cmd for %s not found", step.UUID)
	}

	err = cmd.(*exec.Cmd).Wait()
	ExitCode := 0

	var execExitError *exec.ExitError
	if errors.As(err, &execExitError) {
		ExitCode = execExitError.ExitCode()
		// Non-zero exit code is a step failure, but not an agent error.
		err = nil
	}

	return &types.State{
		Exited:   true,
		ExitCode: ExitCode,
	}, err
}

func (e *local) TailStep(_ context.Context, step *types.Step, taskUUID string) (io.ReadCloser, error) {
	state, err := e.getState(taskUUID)
	if err != nil {
		return nil, err
	}
	reader, found := state.stepOutputs.Load(step.UUID)
	if !found || reader == nil {
		return nil, ErrStepReaderNotFound
	}

	log.Trace().Str("taskUUID", taskUUID).Msgf("tail logs of step %s", step.Name)
	return reader.(io.ReadCloser), nil
}

func (e *local) DestroyStep(_ context.Context, step *types.Step, taskUUID string) error {
	state, err := e.getState(taskUUID)
	if err != nil {
		return err
	}

	// WaitStep uses cmd.Wait() witch ensures the process already finished and
	// the io pipe is closed on process end, so there is nothing to do here.
	// we just remove the state values
	state.stepOutputs.Delete(step.UUID)
	state.stepCMDs.Delete(step.UUID)

	return nil
}

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

	// hint for the gc to clean stuff
	state.stepCMDs.Clear()
	state.stepOutputs.Clear()
	e.workflows.Delete(taskUUID)

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
