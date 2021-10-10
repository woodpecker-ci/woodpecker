// Copyright 2019 mhmxs.
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

package shared

import (
	"testing"
	"time"

	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

type mockUpdateProcStore struct {
}

func (m *mockUpdateProcStore) ProcUpdate(build *model.Proc) error {
	return nil
}

func TestUpdateProcStatusNotExited(t *testing.T) {
	t.Parallel()

	state := rpc.State{
		Started: int64(42),
		Exited:  false,
		// Dummy data
		Finished: int64(1),
		ExitCode: 137,
		Error:    "not an error",
	}
	proc, _ := UpdateProcStatus(&mockUpdateProcStore{}, model.Proc{}, state, int64(1))

	if proc.State != model.StatusRunning {
		t.Errorf("Proc status not equals '%s' != '%s'", model.StatusRunning, proc.State)
	} else if proc.Started != int64(42) {
		t.Errorf("Proc started not equals 42 != %d", proc.Started)
	} else if proc.Stopped != int64(0) {
		t.Errorf("Proc stopped not equals 0 != %d", proc.Stopped)
	} else if proc.ExitCode != 0 {
		t.Errorf("Proc exit code not equals 0 != %d", proc.ExitCode)
	} else if proc.Error != "" {
		t.Errorf("Proc error not equals '' != '%s'", proc.Error)
	}
}

func TestUpdateProcStatusNotExitedButStopped(t *testing.T) {
	t.Parallel()

	proc := &model.Proc{Stopped: int64(64)}

	state := rpc.State{
		Exited: false,
		// Dummy data
		Finished: int64(1),
		ExitCode: 137,
		Error:    "not an error",
	}
	proc, _ = UpdateProcStatus(&mockUpdateProcStore{}, *proc, state, int64(42))

	if proc.State != model.StatusRunning {
		t.Errorf("Proc status not equals '%s' != '%s'", model.StatusRunning, proc.State)
	} else if proc.Started != int64(42) {
		t.Errorf("Proc started not equals 42 != %d", proc.Started)
	} else if proc.Stopped != int64(64) {
		t.Errorf("Proc stopped not equals 64 != %d", proc.Stopped)
	} else if proc.ExitCode != 0 {
		t.Errorf("Proc exit code not equals 0 != %d", proc.ExitCode)
	} else if proc.Error != "" {
		t.Errorf("Proc error not equals '' != '%s'", proc.Error)
	}
}

func TestUpdateProcStatusExited(t *testing.T) {
	t.Parallel()

	state := rpc.State{
		Started:  int64(42),
		Exited:   true,
		Finished: int64(34),
		ExitCode: 137,
		Error:    "an error",
	}
	proc, _ := UpdateProcStatus(&mockUpdateProcStore{}, model.Proc{}, state, int64(42))

	if proc.State != model.StatusKilled {
		t.Errorf("Proc status not equals '%s' != '%s'", model.StatusKilled, proc.State)
	} else if proc.Started != int64(42) {
		t.Errorf("Proc started not equals 42 != %d", proc.Started)
	} else if proc.Stopped != int64(34) {
		t.Errorf("Proc stopped not equals 34 != %d", proc.Stopped)
	} else if proc.ExitCode != 137 {
		t.Errorf("Proc exit code not equals 137 != %d", proc.ExitCode)
	} else if proc.Error != "an error" {
		t.Errorf("Proc error not equals 'an error' != '%s'", proc.Error)
	}
}

func TestUpdateProcStatusExitedButNot137(t *testing.T) {
	t.Parallel()

	state := rpc.State{
		Started:  int64(42),
		Exited:   true,
		Finished: int64(34),
		Error:    "an error",
	}
	proc, _ := UpdateProcStatus(&mockUpdateProcStore{}, model.Proc{}, state, int64(42))

	if proc.State != model.StatusFailure {
		t.Errorf("Proc status not equals '%s' != '%s'", model.StatusFailure, proc.State)
	} else if proc.Started != int64(42) {
		t.Errorf("Proc started not equals 42 != %d", proc.Started)
	} else if proc.Stopped != int64(34) {
		t.Errorf("Proc stopped not equals 34 != %d", proc.Stopped)
	} else if proc.ExitCode != 0 {
		t.Errorf("Proc exit code not equals 0 != %d", proc.ExitCode)
	} else if proc.Error != "an error" {
		t.Errorf("Proc error not equals 'an error' != '%s'", proc.Error)
	}
}

func TestUpdateProcStatusExitedWithCode(t *testing.T) {
	t.Parallel()

	state := rpc.State{
		Started:  int64(42),
		Exited:   true,
		Finished: int64(34),
		ExitCode: 1,
		Error:    "an error",
	}
	proc, _ := UpdateProcStatus(&mockUpdateProcStore{}, model.Proc{}, state, int64(42))

	if proc.State != model.StatusFailure {
		t.Errorf("Proc status not equals '%s' != '%s'", model.StatusFailure, proc.State)
	} else if proc.ExitCode != 1 {
		t.Errorf("Proc exit code not equals 1 != %d", proc.ExitCode)
	}
}

func TestUpdateProcToStatusStarted(t *testing.T) {
	t.Parallel()

	state := rpc.State{Started: int64(42)}
	proc, _ := UpdateProcToStatusStarted(&mockUpdateProcStore{}, model.Proc{}, state)

	if proc.State != model.StatusRunning {
		t.Errorf("Proc status not equals '%s' != '%s'", model.StatusRunning, proc.State)
	} else if proc.Started != int64(42) {
		t.Errorf("Proc started not equals 42 != %d", proc.Started)
	}
}

func TestUpdateProcToStatusSkipped(t *testing.T) {
	t.Parallel()

	proc, _ := UpdateProcToStatusSkipped(&mockUpdateProcStore{}, model.Proc{}, int64(1))

	if proc.State != model.StatusSkipped {
		t.Errorf("Proc status not equals '%s' != '%s'", model.StatusSkipped, proc.State)
	} else if proc.Stopped != int64(0) {
		t.Errorf("Proc stopped not equals 0 != %d", proc.Stopped)
	}
}

func TestUpdateProcToStatusSkippedButStarted(t *testing.T) {
	t.Parallel()

	proc := &model.Proc{
		Started: int64(42),
	}

	proc, _ = UpdateProcToStatusSkipped(&mockUpdateProcStore{}, *proc, int64(1))

	if proc.State != model.StatusSuccess {
		t.Errorf("Proc status not equals '%s' != '%s'", model.StatusSuccess, proc.State)
	} else if proc.Stopped != int64(1) {
		t.Errorf("Proc stopped not equals 1 != %d", proc.Stopped)
	}
}

func TestUpdateProcStatusToDoneSkipped(t *testing.T) {
	t.Parallel()

	state := rpc.State{
		Finished: int64(34),
	}

	proc, _ := UpdateProcStatusToDone(&mockUpdateProcStore{}, model.Proc{}, state)

	if proc.State != model.StatusSkipped {
		t.Errorf("Proc status not equals '%s' != '%s'", model.StatusSkipped, proc.State)
	} else if proc.Stopped != int64(34) {
		t.Errorf("Proc stopped not equals 34 != %d", proc.Stopped)
	} else if proc.Error != "" {
		t.Errorf("Proc error not equals '' != '%s'", proc.Error)
	} else if proc.ExitCode != 0 {
		t.Errorf("Proc exit code not equals 0 != %d", proc.ExitCode)
	}
}

func TestUpdateProcStatusToDoneSuccess(t *testing.T) {
	t.Parallel()

	state := rpc.State{
		Started:  int64(42),
		Finished: int64(34),
	}

	proc, _ := UpdateProcStatusToDone(&mockUpdateProcStore{}, model.Proc{}, state)

	if proc.State != model.StatusSuccess {
		t.Errorf("Proc status not equals '%s' != '%s'", model.StatusSuccess, proc.State)
	} else if proc.Stopped != int64(34) {
		t.Errorf("Proc stopped not equals 34 != %d", proc.Stopped)
	} else if proc.Error != "" {
		t.Errorf("Proc error not equals '' != '%s'", proc.Error)
	} else if proc.ExitCode != 0 {
		t.Errorf("Proc exit code not equals 0 != %d", proc.ExitCode)
	}
}

func TestUpdateProcStatusToDoneFailureWithError(t *testing.T) {
	t.Parallel()

	state := rpc.State{Error: "an error"}

	proc, _ := UpdateProcStatusToDone(&mockUpdateProcStore{}, model.Proc{}, state)

	if proc.State != model.StatusFailure {
		t.Errorf("Proc status not equals '%s' != '%s'", model.StatusFailure, proc.State)
	}
}

func TestUpdateProcStatusToDoneFailureWithExitCode(t *testing.T) {
	t.Parallel()

	state := rpc.State{ExitCode: 43}

	proc, _ := UpdateProcStatusToDone(&mockUpdateProcStore{}, model.Proc{}, state)

	if proc.State != model.StatusFailure {
		t.Errorf("Proc status not equals '%s' != '%s'", model.StatusFailure, proc.State)
	}
}

func TestUpdateProcToStatusKilledStarted(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	proc, _ := UpdateProcToStatusKilled(&mockUpdateProcStore{}, model.Proc{})

	if proc.State != model.StatusKilled {
		t.Errorf("Proc status not equals '%s' != '%s'", model.StatusKilled, proc.State)
	} else if proc.Stopped < now {
		t.Errorf("Proc stopped not equals %d < %d", now, proc.Stopped)
	} else if proc.Started != proc.Stopped {
		t.Errorf("Proc started not equals %d != %d", proc.Stopped, proc.Started)
	} else if proc.ExitCode != 137 {
		t.Errorf("Proc exit code not equals 137 != %d", proc.ExitCode)
	}
}

func TestUpdateProcToStatusKilledNotStarted(t *testing.T) {
	t.Parallel()

	proc, _ := UpdateProcToStatusKilled(&mockUpdateProcStore{}, model.Proc{Started: int64(1)})

	if proc.Started != int64(1) {
		t.Errorf("Proc started not equals 1 != %d", proc.Started)
	}
}
