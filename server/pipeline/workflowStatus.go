// Copyright 2023 Woodpecker Authors
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
	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func UpdateWorkflowToStatusStarted(store model.UpdateWorkflowStore, workflow model.Workflow, state rpc.State) (*model.Workflow, error) {
	workflow.Started = state.Started
	workflow.State = model.StatusRunning
	return &workflow, store.WorkflowUpdate(&workflow)
}

func UpdateWorkflowToStatusSkipped(store model.UpdateWorkflowStore, workflow model.Workflow) (*model.Workflow, error) {
	workflow.State = model.StatusSkipped
	return &workflow, store.WorkflowUpdate(&workflow)
}

func UpdateWorkflowStatusToDone(store model.UpdateWorkflowStore, workflow model.Workflow, state rpc.State) (*model.Workflow, error) {
	workflow.Stopped = state.Finished
	workflow.Error = state.Error
	if state.Started == 0 {
		workflow.State = model.StatusSkipped
	} else {
		workflow.State = model.WorkflowStatus(workflow.Children)
	}
	if workflow.Error != "" {
		workflow.State = model.StatusFailure
	}
	return &workflow, store.WorkflowUpdate(&workflow)
}
