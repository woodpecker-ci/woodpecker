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

	"github.com/alessio/shellescape"
	"github.com/rs/zerolog/log"
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

type workflowState struct {
	stepCMDs     map[string]*exec.Cmd
	baseDir      string
	homeDir      string
	workspaceDir string
}

type local struct {
	workflows map[string]*workflowState
	output    io.ReadCloser
}

// New returns a new local Engine.
func New() types.Engine {
	return &local{
		workflows: make(map[string]*workflowState),
	}
}

func (e *local) Name() string {
	return "local"
}

func (e *local) IsAvailable(context.Context) bool {
	return true
}

func (e *local) Load(context.Context) error {
	// TODO: download plugin-git binary if not exist

	return nil
}

// SetupWorkflow the pipeline environment.
func (e *local) SetupWorkflow(_ context.Context, conf *types.Config, taskUUID string) error {
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

	// TODO: copy plugin-git binary to homeDir and set PATH

	workflowID, err := e.getWorkflowIDFromConfig(conf)
	if err != nil {
		return err
	}

	e.workflows[workflowID] = state

	return nil
}

// StartStep the pipeline step.
func (e *local) StartStep(ctx context.Context, step *types.Step, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("start step %s", step.Name)

	state, err := e.getWorkflowStateFromStep(step)
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

	var command []string
	if step.Image == constant.DefaultCloneImage {
		// Default clone step
		// TODO: use tmp HOME and insert netrc and delete it after clone
		env = append(env, "CI_WORKSPACE="+state.workspaceDir)
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
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
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

	state, err := e.getWorkflowStateFromStep(step)
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
func (e *local) DestroyWorkflow(_ context.Context, conf *types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("delete workflow environment")

	state, err := e.getWorkflowStateFromConfig(conf)
	if err != nil {
		return err
	}

	err = os.RemoveAll(state.baseDir)
	if err != nil {
		return err
	}

	workflowID, err := e.getWorkflowIDFromConfig(conf)
	if err != nil {
		return err
	}

	delete(e.workflows, workflowID)

	return err
}

func (e *local) getWorkflowIDFromStep(step *types.Step) (string, error) {
	sep := "_step_"
	if strings.Contains(step.Name, sep) {
		prefix := strings.Split(step.Name, sep)
		if len(prefix) == 2 {
			return prefix[0], nil
		}
	}

	sep = "_clone"
	if strings.Contains(step.Name, sep) {
		prefix := strings.Split(step.Name, sep)
		if len(prefix) == 2 {
			return prefix[0], nil
		}
	}

	return "", fmt.Errorf("invalid step name (%s) %s", sep, step.Name)
}

func (e *local) getWorkflowIDFromConfig(c *types.Config) (string, error) {
	if len(c.Volumes) < 1 {
		return "", fmt.Errorf("no volumes found in config")
	}

	prefix := strings.Replace(c.Volumes[0].Name, "_default", "", 1)
	return prefix, nil
}

func (e *local) getWorkflowStateFromConfig(c *types.Config) (*workflowState, error) {
	workflowID, err := e.getWorkflowIDFromConfig(c)
	if err != nil {
		return nil, err
	}

	state, ok := e.workflows[workflowID]
	if !ok {
		return nil, fmt.Errorf("workflow %s not found", workflowID)
	}

	return state, nil
}

func (e *local) getWorkflowStateFromStep(step *types.Step) (*workflowState, error) {
	workflowID, err := e.getWorkflowIDFromStep(step)
	if err != nil {
		return nil, err
	}

	state, ok := e.workflows[workflowID]
	if !ok {
		return nil, fmt.Errorf("workflow %s not found", workflowID)
	}

	return state, nil
}
