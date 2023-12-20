// Copyright 2022 Woodpecker Authors
// Copyright 2019 mhmxs
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

package pipeline

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type mockUpdateStepStore struct{}

func (m *mockUpdateStepStore) StepUpdate(_ *model.Step) error {
	return nil
}

func TestUpdateStepStatusNotExited(t *testing.T) {
	t.Parallel()

	state := rpc.State{
		Started: int64(42),
		Exited:  false,
		// Dummy data
		Finished: int64(1),
		ExitCode: pipeline.ExitCodeKilled,
		Error:    "not an error",
	}
	step := &model.Step{}
	err := UpdateStepStatus(&mockUpdateStepStore{}, step, state, int64(1))
	assert.NoError(t, err)

	switch {
	case step.State != model.StatusRunning:
		t.Errorf("Step status not equals '%s' != '%s'", model.StatusRunning, step.State)
	case step.Started != int64(42):
		t.Errorf("Step started not equals 42 != %d", step.Started)
	case step.Stopped != int64(0):
		t.Errorf("Step stopped not equals 0 != %d", step.Stopped)
	case step.ExitCode != 0:
		t.Errorf("Step exit code not equals 0 != %d", step.ExitCode)
	case step.Error != "":
		t.Errorf("Step error not equals '' != '%s'", step.Error)
	}
}

func TestUpdateStepStatusNotExitedButStopped(t *testing.T) {
	t.Parallel()

	step := &model.Step{Stopped: int64(64)}

	state := rpc.State{
		Exited: false,
		// Dummy data
		Finished: int64(1),
		ExitCode: pipeline.ExitCodeKilled,
		Error:    "not an error",
	}
	err := UpdateStepStatus(&mockUpdateStepStore{}, step, state, int64(42))
	assert.NoError(t, err)

	switch {
	case step.State != model.StatusRunning:
		t.Errorf("Step status not equals '%s' != '%s'", model.StatusRunning, step.State)
	case step.Started != int64(42):
		t.Errorf("Step started not equals 42 != %d", step.Started)
	case step.Stopped != int64(64):
		t.Errorf("Step stopped not equals 64 != %d", step.Stopped)
	case step.ExitCode != 0:
		t.Errorf("Step exit code not equals 0 != %d", step.ExitCode)
	case step.Error != "":
		t.Errorf("Step error not equals '' != '%s'", step.Error)
	}
}

func TestUpdateStepStatusExited(t *testing.T) {
	t.Parallel()

	state := rpc.State{
		Started:  int64(42),
		Exited:   true,
		Finished: int64(34),
		ExitCode: pipeline.ExitCodeKilled,
		Error:    "an error",
	}

	step := &model.Step{}
	err := UpdateStepStatus(&mockUpdateStepStore{}, step, state, int64(42))
	assert.NoError(t, err)

	switch {
	case step.State != model.StatusKilled:
		t.Errorf("Step status not equals '%s' != '%s'", model.StatusKilled, step.State)
	case step.Started != int64(42):
		t.Errorf("Step started not equals 42 != %d", step.Started)
	case step.Stopped != int64(34):
		t.Errorf("Step stopped not equals 34 != %d", step.Stopped)
	case step.ExitCode != pipeline.ExitCodeKilled:
		t.Errorf("Step exit code not equals %d != %d", pipeline.ExitCodeKilled, step.ExitCode)
	case step.Error != "an error":
		t.Errorf("Step error not equals 'an error' != '%s'", step.Error)
	}
}

func TestUpdateStepStatusExitedButNot137(t *testing.T) {
	t.Parallel()

	state := rpc.State{
		Started:  int64(42),
		Exited:   true,
		Finished: int64(34),
		Error:    "an error",
	}
	step := &model.Step{}
	err := UpdateStepStatus(&mockUpdateStepStore{}, step, state, int64(42))
	assert.NoError(t, err)

	switch {
	case step.State != model.StatusFailure:
		t.Errorf("Step status not equals '%s' != '%s'", model.StatusFailure, step.State)
	case step.Started != int64(42):
		t.Errorf("Step started not equals 42 != %d", step.Started)
	case step.Stopped != int64(34):
		t.Errorf("Step stopped not equals 34 != %d", step.Stopped)
	case step.ExitCode != 0:
		t.Errorf("Step exit code not equals 0 != %d", step.ExitCode)
	case step.Error != "an error":
		t.Errorf("Step error not equals 'an error' != '%s'", step.Error)
	}
}

func TestUpdateStepStatusExitedWithCode(t *testing.T) {
	t.Parallel()

	state := rpc.State{
		Started:  int64(42),
		Exited:   true,
		Finished: int64(34),
		ExitCode: 1,
		Error:    "an error",
	}
	step := &model.Step{}
	err := UpdateStepStatus(&mockUpdateStepStore{}, step, state, int64(42))
	assert.NoError(t, err)

	if step.State != model.StatusFailure {
		t.Errorf("Step status not equals '%s' != '%s'", model.StatusFailure, step.State)
	} else if step.ExitCode != 1 {
		t.Errorf("Step exit code not equals 1 != %d", step.ExitCode)
	}
}

func TestUpdateStepPToStatusStarted(t *testing.T) {
	t.Parallel()

	state := rpc.State{Started: int64(42)}
	step, _ := UpdateStepToStatusStarted(&mockUpdateStepStore{}, model.Step{}, state)

	if step.State != model.StatusRunning {
		t.Errorf("Step status not equals '%s' != '%s'", model.StatusRunning, step.State)
	} else if step.Started != int64(42) {
		t.Errorf("Step started not equals 42 != %d", step.Started)
	}
}

func TestUpdateStepToStatusSkipped(t *testing.T) {
	t.Parallel()

	step, _ := UpdateStepToStatusSkipped(&mockUpdateStepStore{}, model.Step{}, int64(1))

	if step.State != model.StatusSkipped {
		t.Errorf("Step status not equals '%s' != '%s'", model.StatusSkipped, step.State)
	} else if step.Stopped != int64(0) {
		t.Errorf("Step stopped not equals 0 != %d", step.Stopped)
	}
}

func TestUpdateStepToStatusSkippedButStarted(t *testing.T) {
	t.Parallel()

	step := &model.Step{
		Started: int64(42),
	}

	step, _ = UpdateStepToStatusSkipped(&mockUpdateStepStore{}, *step, int64(1))

	if step.State != model.StatusSuccess {
		t.Errorf("Step status not equals '%s' != '%s'", model.StatusSuccess, step.State)
	} else if step.Stopped != int64(1) {
		t.Errorf("Step stopped not equals 1 != %d", step.Stopped)
	}
}

func TestUpdateStepStatusToDoneSkipped(t *testing.T) {
	t.Parallel()

	state := rpc.State{
		Finished: int64(34),
	}

	step, _ := UpdateStepStatusToDone(&mockUpdateStepStore{}, model.Step{}, state)

	switch {
	case step.State != model.StatusSkipped:
		t.Errorf("Step status not equals '%s' != '%s'", model.StatusSkipped, step.State)
	case step.Stopped != int64(34):
		t.Errorf("Step stopped not equals 34 != %d", step.Stopped)
	case step.Error != "":
		t.Errorf("Step error not equals '' != '%s'", step.Error)
	case step.ExitCode != 0:
		t.Errorf("Step exit code not equals 0 != %d", step.ExitCode)
	}
}

func TestUpdateStepStatusToDoneSuccess(t *testing.T) {
	t.Parallel()

	state := rpc.State{
		Started:  int64(42),
		Finished: int64(34),
	}

	step, _ := UpdateStepStatusToDone(&mockUpdateStepStore{}, model.Step{}, state)

	switch {
	case step.State != model.StatusSuccess:
		t.Errorf("Step status not equals '%s' != '%s'", model.StatusSuccess, step.State)
	case step.Stopped != int64(34):
		t.Errorf("Step stopped not equals 34 != %d", step.Stopped)
	case step.Error != "":
		t.Errorf("Step error not equals '' != '%s'", step.Error)
	case step.ExitCode != 0:
		t.Errorf("Step exit code not equals 0 != %d", step.ExitCode)
	}
}

func TestUpdateStepStatusToDoneFailureWithError(t *testing.T) {
	t.Parallel()

	state := rpc.State{Error: "an error"}

	step, _ := UpdateStepStatusToDone(&mockUpdateStepStore{}, model.Step{}, state)

	if step.State != model.StatusFailure {
		t.Errorf("Step status not equals '%s' != '%s'", model.StatusFailure, step.State)
	}
}

func TestUpdateStepStatusToDoneFailureWithExitCode(t *testing.T) {
	t.Parallel()

	state := rpc.State{ExitCode: 43}

	step, _ := UpdateStepStatusToDone(&mockUpdateStepStore{}, model.Step{}, state)

	if step.State != model.StatusFailure {
		t.Errorf("Step status not equals '%s' != '%s'", model.StatusFailure, step.State)
	}
}

func TestUpdateStepToStatusKilledStarted(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	step, _ := UpdateStepToStatusKilled(&mockUpdateStepStore{}, model.Step{})

	switch {
	case step.State != model.StatusKilled:
		t.Errorf("Step status not equals '%s' != '%s'", model.StatusKilled, step.State)
	case step.Stopped < now:
		t.Errorf("Step stopped not equals %d < %d", now, step.Stopped)
	case step.Started != step.Stopped:
		t.Errorf("Step started not equals %d != %d", step.Stopped, step.Started)
	case step.ExitCode != pipeline.ExitCodeKilled:
		t.Errorf("Step exit code not equals %d != %d", pipeline.ExitCodeKilled, step.ExitCode)
	}
}

func TestUpdateStepToStatusKilledNotStarted(t *testing.T) {
	t.Parallel()

	step, _ := UpdateStepToStatusKilled(&mockUpdateStepStore{}, model.Step{Started: int64(1)})

	if step.Started != int64(1) {
		t.Errorf("Step started not equals 1 != %d", step.Started)
	}
}
