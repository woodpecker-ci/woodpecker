// Copyright 2022 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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

package datastore

import (
	"testing"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestWorkflowGetTree(t *testing.T) {
	store, closer := newTestStore(t, new(model.Step), new(model.Pipeline), new(model.Workflow))
	defer closer()

	wf := &model.Workflow{
		PipelineID: 1,
		PID:        1,
		Name:       "woodpecker",
		Children: []*model.Step{
			{
				UUID:       "ea6d4008-8ace-4f8a-ad03-53f1756465d9",
				PipelineID: 1,
				PID:        2,
				PPID:       1,
				State:      "success",
			},
			{
				UUID:       "2bf387f7-2913-4907-814c-c9ada88707c0",
				PipelineID: 1,
				PID:        3,
				PPID:       1,
				Name:       "build",
				State:      "success",
			},
		},
	}
	err := store.WorkflowsCreate([]*model.Workflow{wf})
	if err != nil {
		t.Errorf("Unexpected error: insert steps: %s", err)
		return
	}

	workflowsGet, err := store.WorkflowGetTree(&model.Pipeline{ID: 1})
	if err != nil {
		t.Error(err)
		return
	}

	if got, want := len(workflowsGet), 1; got != want {
		t.Errorf("Want workflow len %d, got %d", want, got)
		return
	}
	workflowGet := workflowsGet[0]
	if got, want := workflowGet.Name, "woodpecker"; got != want {
		t.Errorf("Want workflow name %s, got %s", want, got)
	}
	if got, want := len(workflowGet.Children), 2; got != want {
		t.Errorf("Want children len %d, got %d", want, got)
		return
	}
	if got, want := workflowGet.Children[0].PID, 2; got != want {
		t.Errorf("Want children len %d, got %d", want, got)
	}
	if got, want := workflowGet.Children[1].PID, 3; got != want {
		t.Errorf("Want children len %d, got %d", want, got)
	}
}
