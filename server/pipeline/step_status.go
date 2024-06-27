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
	"go.woodpecker-ci.org/woodpecker/v2/pipeline"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

func UpdateStepStatusToRunning(store store.Store, step model.Step, state rpc.StepState) (*model.Step, error) {
	if step.Finished == 0 && state.Finished == 0 {
		step.Started = state.Started
		step.State = model.StatusRunning
	}
	return &step, store.StepUpdate(&step)
}

func UpdateStepStatusToSkipped(store store.Store, step model.Step, finished int64) (*model.Step, error) {
	step.State = model.StatusSkipped
	if step.Started != 0 {
		step.State = model.StatusSuccess // for daemons that are killed
		step.Finished = finished
	}
	return &step, store.StepUpdate(&step)
}

func UpdateStepStatusToDone(store store.Store, step model.Step, state rpc.StepState) (*model.Step, error) {
	step.Finished = state.Finished
	step.Error = state.Error
	step.ExitCode = state.ExitCode
	step.State = model.StatusSuccess
	if state.Started == 0 {
		step.State = model.StatusSkipped
	}
	if state.ExitCode != 0 || state.Error != "" {
		step.State = model.StatusFailure
	}
	if state.ExitCode == pipeline.ExitCodeKilled {
		step.State = model.StatusKilled
	}

	return &step, store.StepUpdate(&step)
}
