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
	"time"

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// WorkflowStatus determine workflow status based on corresponding step list.
func WorkflowStatus(steps []*model.Step) model.StatusValue {
	status := model.StatusSuccess

	for _, p := range steps {
		if p.Failure == model.FailureFail || !p.Failing() {
			status = MergeStatusValues(status, p.State)
		}
	}

	return status
}

func UpdateWorkflowStatusToRunning(store store.Store, workflow model.Workflow, _ rpc.WorkflowState) (*model.Workflow, error) {
	// Record the workflow start time from the server clock rather than the
	// agent-supplied state.Started, so that Started and Finished share a single
	// clock and durations stay correct even when the agent clock is skewed
	// (#6808). This also keeps pipeline/workflow times consistent across a
	// multi-agent pipeline where different workflows run on different agents.
	workflow.Started = time.Now().Unix()
	workflow.State = model.StatusRunning
	return &workflow, store.WorkflowUpdate(&workflow)
}

func UpdateWorkflowToStatusSkipped(store store.Store, workflow model.Workflow) (*model.Workflow, error) {
	workflow.State = model.StatusSkipped
	return &workflow, store.WorkflowUpdate(&workflow)
}

func UpdateWorkflowStatusToDone(store store.Store, workflow model.Workflow, state rpc.WorkflowState) (*model.Workflow, error) {
	// Record the finish time from the server clock (see UpdateWorkflowStatusToRunning
	// and #6808). state.Started is still consulted below purely as a presence
	// flag to detect a workflow that was never started (skipped).
	workflow.Finished = time.Now().Unix()
	workflow.Error = state.Error
	if state.Started == 0 {
		workflow.State = model.StatusSkipped
	} else {
		workflow.State = WorkflowStatus(workflow.Children)
	}
	if workflow.Error != "" {
		workflow.State = model.StatusFailure
	}
	if state.Canceled {
		workflow.State = model.StatusKilled
	}
	return &workflow, store.WorkflowUpdate(&workflow)
}
