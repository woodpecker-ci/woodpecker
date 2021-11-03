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
	"time"

	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

type UpdateProcStore interface {
	ProcUpdate(*model.Proc) error
}

func UpdateProcStatus(store UpdateProcStore, proc model.Proc, state rpc.State, started int64) (*model.Proc, error) {
	if state.Exited {
		proc.Stopped = state.Finished
		proc.ExitCode = state.ExitCode
		proc.Error = state.Error
		proc.State = model.StatusSuccess
		if state.ExitCode != 0 || state.Error != "" {
			proc.State = model.StatusFailure
		}
		if state.ExitCode == 137 {
			proc.State = model.StatusKilled
		}
	} else {
		proc.Started = state.Started
		proc.State = model.StatusRunning
	}

	if proc.Started == 0 && proc.Stopped != 0 {
		proc.Started = started
	}
	return &proc, store.ProcUpdate(&proc)
}

func UpdateProcToStatusStarted(store UpdateProcStore, proc model.Proc, state rpc.State) (*model.Proc, error) {
	proc.Started = state.Started
	proc.State = model.StatusRunning
	return &proc, store.ProcUpdate(&proc)
}

func UpdateProcToStatusSkipped(store UpdateProcStore, proc model.Proc, stopped int64) (*model.Proc, error) {
	proc.State = model.StatusSkipped
	if proc.Started != 0 {
		proc.State = model.StatusSuccess // for daemons that are killed
		proc.Stopped = stopped
	}
	return &proc, store.ProcUpdate(&proc)
}

func UpdateProcStatusToDone(store UpdateProcStore, proc model.Proc, state rpc.State) (*model.Proc, error) {
	proc.Stopped = state.Finished
	proc.Error = state.Error
	proc.ExitCode = state.ExitCode
	if state.Started == 0 {
		proc.State = model.StatusSkipped
	} else {
		proc.State = model.StatusSuccess
	}
	if proc.ExitCode != 0 || proc.Error != "" {
		proc.State = model.StatusFailure
	}
	return &proc, store.ProcUpdate(&proc)
}

func UpdateProcToStatusKilled(store UpdateProcStore, proc model.Proc) (*model.Proc, error) {
	proc.State = model.StatusKilled
	proc.Stopped = time.Now().Unix()
	if proc.Started == 0 {
		proc.Started = proc.Stopped
	}
	proc.ExitCode = 137
	return &proc, store.ProcUpdate(&proc)
}
