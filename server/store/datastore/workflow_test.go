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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestWorkflowLoad(t *testing.T) {
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
	assert.NoError(t, store.WorkflowsCreate([]*model.Workflow{wf}))
	workflowGet, err := store.WorkflowLoad(1)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, workflowGet.PipelineID)
	assert.Equal(t, 1, workflowGet.PID)
	assert.Len(t, workflowGet.Children, 0)
}

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
	assert.NoError(t, store.WorkflowsCreate([]*model.Workflow{wf}))

	workflowsGet, err := store.WorkflowGetTree(&model.Pipeline{ID: 1})
	assert.NoError(t, err)
	assert.Len(t, workflowsGet, 1)
	workflowGet := workflowsGet[0]
	assert.Equal(t, "woodpecker", workflowGet.Name)
	assert.Len(t, workflowGet.Children, 2)
	assert.Equal(t, 2, workflowGet.Children[0].PID)
	assert.Equal(t, 3, workflowGet.Children[1].PID)
}

func TestWorkflowUpdate(t *testing.T) {
	store, closer := newTestStore(t, new(model.Step), new(model.Pipeline), new(model.Workflow))
	defer closer()

	wf := &model.Workflow{
		PipelineID: 1,
		PID:        1,
		Name:       "woodpecker",
		State:      "pending",
	}
	assert.NoError(t, store.WorkflowsCreate([]*model.Workflow{wf}))
	workflowGet, err := store.WorkflowLoad(1)
	assert.NoError(t, err)

	assert.Equal(t, model.StatusValue("pending"), workflowGet.State)

	wf.State = "success"

	assert.NoError(t, store.WorkflowUpdate(wf))
	workflowGet, err = store.WorkflowLoad(1)
	assert.NoError(t, err)
	assert.Equal(t, model.StatusValue("success"), workflowGet.State)
}

func TestWorkflowCancelPending(t *testing.T) {
	s, close := newTestStore(t, new(model.Workflow), new(model.Step))
	defer close()
	w := &model.Workflow{PipelineID: 1, PID: 1, State: model.StatusPending, Children: []*model.Step{
		{UUID: "pending", PipelineID: 1, PID: 2, PPID: 1, State: model.StatusPending},
		{UUID: "done", PipelineID: 1, PID: 3, PPID: 1, State: model.StatusSuccess, Finished: 7},
	}}
	sibling := &model.Workflow{PipelineID: 1, PID: 4, State: model.StatusPending}
	require.NoError(t, s.WorkflowsCreate([]*model.Workflow{w, sibling}))
	canceled, err := s.WorkflowCancelPending(w.ID, 10)
	require.NoError(t, err)
	require.True(t, canceled)
	tree, err := s.WorkflowGetTree(&model.Pipeline{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, model.StatusCanceled, tree[0].State)
	assert.Zero(t, tree[0].Started)
	assert.EqualValues(t, 10, tree[0].Finished)
	assert.Equal(t, model.StatusCanceled, tree[0].Children[0].State)
	assert.Equal(t, model.StatusSuccess, tree[0].Children[1].State)
	assert.EqualValues(t, 7, tree[0].Children[1].Finished)
	assert.Equal(t, model.StatusPending, tree[1].State)
	w.State = model.StatusRunning
	assert.Error(t, s.WorkflowUpdateIfState(w, model.StatusPending), "canceled workflow cannot initialize")
	canceled, err = s.WorkflowCancelPending(w.ID, 11)
	require.NoError(t, err)
	assert.False(t, canceled)
	sibling.State = model.StatusRunning
	require.NoError(t, s.WorkflowUpdateIfState(sibling, model.StatusPending))
	canceled, err = s.WorkflowCancelPending(sibling.ID, 12)
	require.NoError(t, err)
	assert.False(t, canceled, "initialization won")
}

func TestWorkflowCancellationRaces(t *testing.T) {
	for _, next := range []model.StatusValue{model.StatusRunning, model.StatusSuccess} {
		t.Run(string(next), func(t *testing.T) {
			s, closeStore := newTestStore(t, new(model.Workflow), new(model.Step))
			defer closeStore()
			w := &model.Workflow{PipelineID: 1, PID: 1, State: model.StatusPending}
			require.NoError(t, s.WorkflowsCreate([]*model.Workflow{w}))
			start := make(chan struct{})
			cancelResult := make(chan bool, 1)
			cancelError := make(chan error, 1)
			transitionError := make(chan error, 1)
			go func() { <-start; ok, err := s.WorkflowCancelPending(w.ID, 10); cancelResult <- ok; cancelError <- err }()
			go func() {
				<-start
				copy := *w
				copy.State = next
				transitionError <- s.WorkflowUpdateIfState(&copy, model.StatusPending)
			}()
			close(start)
			canceled := <-cancelResult
			require.NoError(t, <-cancelError)
			err := <-transitionError
			fresh, loadErr := s.WorkflowLoad(w.ID)
			require.NoError(t, loadErr)
			if canceled {
				assert.Error(t, err)
				assert.Equal(t, model.StatusCanceled, fresh.State)
			} else {
				require.NoError(t, err)
				assert.Equal(t, next, fresh.State)
			}
		})
	}
}
