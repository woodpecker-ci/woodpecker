// Copyright 2024 Woodpecker Authors
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

//go:build test
// +build test

package dummy

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	backend "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

type dummy struct {
	kv sync.Map
}

const (
	// Step names to control behavior of dummy backend.
	WorkflowSetupFailUUID = "WorkflowSetupShouldFail"
	EnvKeyStepSleep       = "SLEEP"
	EnvKeyStepType        = "EXPECT_TYPE"
	EnvKeyStepStartFail   = "STEP_START_FAIL"
	EnvKeyStepExitCode    = "STEP_EXIT_CODE"
	EnvKeyStepTailFail    = "STEP_TAIL_FAIL"
	EnvKeyStepOOMKilled   = "STEP_OOM_KILLED"

	// Internal const.
	stepStateStarted   = "started"
	stepStateDone      = "done"
	testServiceTimeout = 1 * time.Second
)

// New returns a dummy backend.
func New() backend.Backend {
	return &dummy{
		kv: sync.Map{},
	}
}

func (e *dummy) Name() string {
	return "dummy"
}

func (e *dummy) IsAvailable(_ context.Context) bool {
	return true
}

func (e *dummy) Flags() []cli.Flag {
	return nil
}

// Load new client for Docker Backend using environment variables.
func (e *dummy) Load(_ context.Context) (*backend.BackendInfo, error) {
	return &backend.BackendInfo{
		Platform: "dummy",
	}, nil
}

func (e *dummy) SetupWorkflow(_ context.Context, _ *backend.Config, taskUUID string) error {
	if taskUUID == WorkflowSetupFailUUID {
		return fmt.Errorf("expected fail to setup workflow")
	}
	log.Trace().Str("taskUUID", taskUUID).Msg("create workflow environment")
	e.kv.Store("task_"+taskUUID, "setup")
	return nil
}

func (e *dummy) StartStep(_ context.Context, step *backend.Step, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("start step %s", step.Name)

	// internal state checks
	_, exist := e.kv.Load("task_" + taskUUID)
	if !exist {
		return fmt.Errorf("expect env of workflow %s to exist but found none to destroy", taskUUID)
	}
	stepState, stepExist := e.kv.Load(fmt.Sprintf("task_%s_step_%s", taskUUID, step.UUID))
	if stepExist {
		// Detect issues like https://github.com/woodpecker-ci/woodpecker/issues/3494
		return fmt.Errorf("StartStep detected already started step '%s' (%s) in state: %s", step.Name, step.UUID, stepState)
	}

	if stepStartFail, _ := strconv.ParseBool(step.Environment[EnvKeyStepStartFail]); stepStartFail {
		return fmt.Errorf("expected fail to start step")
	}

	expectStepType, testStepType := step.Environment[EnvKeyStepType]
	if testStepType && string(step.Type) != expectStepType {
		return fmt.Errorf("expected step type '%s' but got '%s'", expectStepType, step.Type)
	}

	e.kv.Store(fmt.Sprintf("task_%s_step_%s", taskUUID, step.UUID), stepStateStarted)
	return nil
}

func (e *dummy) WaitStep(ctx context.Context, step *backend.Step, taskUUID string) (*backend.State, error) {
	log.Trace().Str("taskUUID", taskUUID).Msgf("wait for step %s", step.Name)

	_, exist := e.kv.Load("task_" + taskUUID)
	if !exist {
		err := fmt.Errorf("expect env of workflow %s to exist but found none to destroy", taskUUID)
		return &backend.State{Error: err}, err
	}

	// check state
	stepState, stepExist := e.kv.Load(fmt.Sprintf("task_%s_step_%s", taskUUID, step.UUID))
	if !stepExist {
		err := fmt.Errorf("WaitStep expect step '%s' (%s) to be created but found none", step.Name, step.UUID)
		return &backend.State{Error: err}, err
	}
	if stepState != stepStateStarted {
		err := fmt.Errorf("WaitStep expect step '%s' (%s) to be '%s' but it is: %s", step.Name, step.UUID, stepStateStarted, stepState)
		return &backend.State{Error: err}, err
	}

	// extend wait time logic
	if sleep, sleepExist := step.Environment[EnvKeyStepSleep]; sleepExist {
		toSleep, err := time.ParseDuration(sleep)
		if err != nil {
			err = fmt.Errorf("WaitStep fail to parse sleep duration: %w", err)
			return &backend.State{Error: err}, err
		}
		time.Sleep(toSleep)
	} else {
		if step.Type == backend.StepTypeService {
			select {
			case <-time.NewTimer(testServiceTimeout).C:
				err := fmt.Errorf("WaitStep fail due to timeout of service after 1 second")
				return &backend.State{Error: err}, err
			case <-ctx.Done():
				// context for service closed ... we can move forward
			}
		} else {
			time.Sleep(time.Nanosecond)
		}
	}

	e.kv.Store(fmt.Sprintf("task_%s_step_%s", taskUUID, step.UUID), stepStateDone)

	oomKilled, _ := strconv.ParseBool(step.Environment[EnvKeyStepOOMKilled])
	exitCode := 0

	if code, exist := step.Environment[EnvKeyStepExitCode]; exist {
		exitCode, _ = strconv.Atoi(strings.TrimSpace(code))
	}

	return &backend.State{
		ExitCode:  exitCode,
		Exited:    true,
		OOMKilled: oomKilled,
	}, nil
}

func (e *dummy) TailStep(_ context.Context, step *backend.Step, taskUUID string) (io.ReadCloser, error) {
	log.Trace().Str("taskUUID", taskUUID).Msgf("tail logs of step %s", step.Name)

	_, exist := e.kv.Load("task_" + taskUUID)
	if !exist {
		return nil, fmt.Errorf("expect env of workflow %s to exist but found none to destroy", taskUUID)
	}

	// check state
	stepState, stepExist := e.kv.Load(fmt.Sprintf("task_%s_step_%s", taskUUID, step.UUID))
	if !stepExist {
		return nil, fmt.Errorf("WaitStep expect step '%s' (%s) to be created but found none", step.Name, step.UUID)
	}
	if stepState != stepStateStarted {
		return nil, fmt.Errorf("WaitStep expect step '%s' (%s) to be '%s' but it is: %s", step.Name, step.UUID, stepStateStarted, stepState)
	}

	if tailShouldFail, _ := strconv.ParseBool(step.Environment[EnvKeyStepTailFail]); tailShouldFail {
		return nil, fmt.Errorf("expected fail to read stdout of step")
	}

	return io.NopCloser(strings.NewReader(dummyExecStepOutput(step))), nil
}

func (e *dummy) DestroyStep(_ context.Context, step *backend.Step, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("stop step %s", step.Name)

	_, exist := e.kv.Load("task_" + taskUUID)
	if !exist {
		return fmt.Errorf("expect env of workflow %s to exist but found none to destroy", taskUUID)
	}

	// check state
	stepState, stepExist := e.kv.Load(fmt.Sprintf("task_%s_step_%s", taskUUID, step.UUID))
	if !stepExist {
		return fmt.Errorf("WaitStep expect step '%s' (%s) to be created but found none", step.Name, step.UUID)
	}
	if stepState != stepStateDone {
		return fmt.Errorf("WaitStep expect step '%s' (%s) to be '%s' but it is: %s", step.Name, step.UUID, stepStateDone, stepState)
	}

	e.kv.Delete(fmt.Sprintf("task_%s_step_%s", taskUUID, step.UUID))
	return nil
}

func (e *dummy) DestroyWorkflow(_ context.Context, _ *backend.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("delete workflow environment")

	_, exist := e.kv.Load("task_" + taskUUID)
	if !exist {
		return fmt.Errorf("expect env of workflow %s to exist but found none to destroy", taskUUID)
	}
	e.kv.Delete("task_" + taskUUID)
	return nil
}

func dummyExecStepOutput(step *backend.Step) string {
	return fmt.Sprintf(`StepName: %s
StepType: %s
StepUUID: %s
StepCommands:
------------------
%s
------------------
`, step.Name, step.Type, step.UUID, strings.Join(step.Commands, "\n"))
}
