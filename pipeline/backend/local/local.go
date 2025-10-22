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
	stepState       sync.Map // map of *stepState
	baseDir         string
	homeDir         string
	workspaceDir    string
	pluginGitBinary string
}

type stepState struct {
	cmd    *exec.Cmd
	output io.ReadCloser
}

type local struct {
	tempDir         string
	workflows       sync.Map
	pluginGitBinary string
	os, arch        string
}

var CLIWorkaroundExecAtDir string // To handle edge case for running local backend via cli exec

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
		baseDir: baseDir,
		homeDir: filepath.Join(baseDir, "home"),
	}
	e.workflows.Store(taskUUID, state)

	if err := os.Mkdir(state.homeDir, 0o700); err != nil {
		return err
	}

	// normal workspace setup case
	if CLIWorkaroundExecAtDir == "" {
		state.workspaceDir = filepath.Join(baseDir, "workspace")
		if err := os.Mkdir(state.workspaceDir, 0o700); err != nil {
			return err
		}
	} else
	// setup workspace via internal flag signaled from cli exec to a specific dir
	{
		state.workspaceDir = CLIWorkaroundExecAtDir
		if stat, err := os.Stat(CLIWorkaroundExecAtDir); os.IsNotExist(err) {
			log.Debug().Msgf("create workspace directory '%s' set by internal flag", CLIWorkaroundExecAtDir)
			if err := os.Mkdir(state.workspaceDir, 0o700); err != nil {
				return err
			}
		} else if !stat.IsDir() {
			//nolint:forbidigo
			log.Fatal().Msg("This should never happen! internalExecDir was set to an non directory path!")
		}
	}

	e.workflows.Store(taskUUID, state)

	return nil
}

func (e *local) StartStep(ctx context.Context, step *types.Step, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("start step %s", step.Name)

	state, err := e.getWorkflowState(taskUUID)
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

	state, err := e.getStepState(taskUUID, step.UUID)
	if err != nil {
		return nil, err
	}

	// normally we use cmd.Wait() to wait for *exec.Cmd, but cmd.StdoutPipe() tells us not
	// as Wait() would close the io pipe even if not all logs where read and send back
	// so we have to do use the underlying functions
	if state.cmd.Process == nil {
		return nil, errors.New("exec: not started")
	}
	if state.cmd.ProcessState == nil {
		cmdState, err := state.cmd.Process.Wait()
		if err != nil {
			return nil, err
		}
		state.cmd.ProcessState = cmdState
	}

	return &types.State{
		Exited:   true,
		ExitCode: state.cmd.ProcessState.ExitCode(),
	}, err
}

func (e *local) TailStep(_ context.Context, step *types.Step, taskUUID string) (io.ReadCloser, error) {
	state, err := e.getStepState(taskUUID, step.UUID)
	if err != nil {
		return nil, err
	} else if state.output == nil {
		return nil, ErrStepReaderNotFound
	}
	return state.output, nil
}

func (e *local) DestroyStep(_ context.Context, step *types.Step, taskUUID string) error {
	state, err := e.getStepState(taskUUID, step.UUID)
	if err != nil {
		return err
	}

	// As WaitStep can not use cmd.Wait() witch ensures the process already finished and
	// the io pipe is closed on process end, we make sure it is done.
	_ = state.output.Close()
	state.output = nil
	_ = state.cmd.Cancel()
	state.cmd = nil
	workflowState, _ := e.getWorkflowState(taskUUID)
	workflowState.stepState.Delete(step.UUID)

	return nil
}

func (e *local) DestroyWorkflow(_ context.Context, _ *types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msg("delete workflow environment")

	state, err := e.getWorkflowState(taskUUID)
	if err != nil {
		return err
	}

	// clean up steps not cleaned up because of context cancel or detached function
	state.stepState.Range(func(_, value any) bool {
		state, _ := value.(*stepState)
		_ = state.output.Close()
		state.output = nil
		_ = state.cmd.Cancel()
		state.cmd = nil
		return true
	})

	err = os.RemoveAll(state.baseDir)
	if err != nil {
		return err
	}

	// hint for the gc to clean stuff
	state.stepState.Clear()
	e.workflows.Delete(taskUUID)

	return err
}

func (e *local) getWorkflowState(taskUUID string) (*workflowState, error) {
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

func (e *local) getStepState(taskUUID, stepUUID string) (*stepState, error) {
	wState, err := e.getWorkflowState(taskUUID)
	if err != nil {
		return nil, err
	}

	state, ok := wState.stepState.Load(stepUUID)
	if !ok {
		return nil, ErrStepStateNotFound
	}

	s, ok := state.(*stepState)
	if !ok {
		return nil, fmt.Errorf("could not parse state: %v", state)
	}

	return s, nil
}
