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
	"time"

	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func UpdateStepStatus(store model.UpdateStepStore, step *model.Step, state rpc.State, started int64) error {
	if state.Exited {
		step.Stopped = state.Finished
		step.ExitCode = state.ExitCode
		step.Error = state.Error
		step.State = model.StatusSuccess
		if state.ExitCode != 0 || state.Error != "" {
			step.State = model.StatusFailure
		}
		if state.ExitCode == 137 {
			step.State = model.StatusKilled
		}
	} else {
		step.Started = state.Started
		step.State = model.StatusRunning
	}

	if step.Started == 0 && step.Stopped != 0 {
		step.Started = started
	}
	return store.StepUpdate(step)
}

func UpdateStepToStatusStarted(store model.UpdateStepStore, step model.Step, state rpc.State) (*model.Step, error) {
	step.Started = state.Started
	step.State = model.StatusRunning
	return &step, store.StepUpdate(&step)
}

func UpdateStepToStatusSkipped(store model.UpdateStepStore, step model.Step, stopped int64) (*model.Step, error) {
	step.State = model.StatusSkipped
	if step.Started != 0 {
		step.State = model.StatusSuccess // for daemons that are killed
		step.Stopped = stopped
	}
	return &step, store.StepUpdate(&step)
}

func UpdateStepStatusToDone(store model.UpdateStepStore, step model.Step, state rpc.State) (*model.Step, error) {
	step.Stopped = state.Finished
	step.Error = state.Error
	step.ExitCode = state.ExitCode
	if state.Started == 0 {
		step.State = model.StatusSkipped
	} else {
		step.State = model.StatusSuccess
	}
	if step.ExitCode != 0 || step.Error != "" {
		step.State = model.StatusFailure
	}
	return &step, store.StepUpdate(&step)
}

func UpdateStepToStatusKilled(store model.UpdateStepStore, step model.Step) (*model.Step, error) {
	step.State = model.StatusKilled
	step.Stopped = time.Now().Unix()
	if step.Started == 0 {
		step.Started = step.Stopped
	}
	step.ExitCode = 137
	return &step, store.StepUpdate(&step)
}
