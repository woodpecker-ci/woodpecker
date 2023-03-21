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

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestStepFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.Step), new(model.Pipeline))
	defer closer()

	steps := []*model.Step{
		{
			PipelineID: 1000,
			PID:        1,
			PPID:       2,
			PGID:       3,
			Name:       "build",
			State:      model.StatusSuccess,
			Error:      "pc load letter",
			ExitCode:   255,
			AgentID:    1,
			Platform:   "linux/amd64",
			Environ:    map[string]string{"GOLANG": "tip"},
		},
	}
	assert.NoError(t, store.StepCreate(steps))
	assert.EqualValues(t, 1, steps[0].ID)
	assert.Error(t, store.StepCreate(steps))

	step, err := store.StepFind(&model.Pipeline{ID: 1000}, 1)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, steps[0], step)
}

func TestStepChild(t *testing.T) {
	store, closer := newTestStore(t, new(model.Step), new(model.Pipeline))
	defer closer()

	err := store.StepCreate([]*model.Step{
		{
			PipelineID: 1,
			PID:        1,
			PPID:       1,
			PGID:       1,
			State:      "success",
		},
		{
			PipelineID: 1,
			PID:        2,
			PGID:       2,
			PPID:       1,
			Name:       "build",
			State:      "success",
		},
	})
	if err != nil {
		t.Errorf("Unexpected error: insert steps: %s", err)
		return
	}
	step, err := store.StepChild(&model.Pipeline{ID: 1}, 1, "build")
	if err != nil {
		t.Error(err)
		return
	}

	if got, want := step.PID, 2; got != want {
		t.Errorf("Want step pid %d, got %d", want, got)
	}
	if got, want := step.Name, "build"; got != want {
		t.Errorf("Want step name %s, got %s", want, got)
	}
}

func TestStepList(t *testing.T) {
	store, closer := newTestStore(t, new(model.Step), new(model.Pipeline))
	defer closer()

	err := store.StepCreate([]*model.Step{
		{
			PipelineID: 2,
			PID:        1,
			PPID:       1,
			PGID:       1,
			State:      "success",
		},
		{
			PipelineID: 1,
			PID:        1,
			PPID:       1,
			PGID:       1,
			State:      "success",
		},
		{
			PipelineID: 1,
			PID:        2,
			PGID:       2,
			PPID:       1,
			Name:       "build",
			State:      "success",
		},
	})
	if err != nil {
		t.Errorf("Unexpected error: insert steps: %s", err)
		return
	}
	steps, err := store.StepList(&model.Pipeline{ID: 1}, &model.PaginationData{Page: 1, PerPage: 50})
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := len(steps), 2; got != want {
		t.Errorf("Want %d steps, got %d", want, got)
	}
}

func TestStepUpdate(t *testing.T) {
	store, closer := newTestStore(t, new(model.Step), new(model.Pipeline))
	defer closer()

	step := &model.Step{
		PipelineID: 1,
		PID:        1,
		PPID:       2,
		PGID:       3,
		Name:       "build",
		State:      "pending",
		Error:      "pc load letter",
		ExitCode:   255,
		AgentID:    1,
		Platform:   "linux/amd64",
		Environ:    map[string]string{"GOLANG": "tip"},
	}
	if err := store.StepCreate([]*model.Step{step}); err != nil {
		t.Errorf("Unexpected error: insert step: %s", err)
		return
	}
	step.State = "running"
	if err := store.StepUpdate(step); err != nil {
		t.Errorf("Unexpected error: update step: %s", err)
		return
	}
	updated, err := store.StepFind(&model.Pipeline{ID: 1}, 1)
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := updated.State, model.StatusRunning; got != want {
		t.Errorf("Want step name %s, got %s", want, got)
	}
}

func TestStepIndexes(t *testing.T) {
	store, closer := newTestStore(t, new(model.Step), new(model.Pipeline))
	defer closer()

	if err := store.StepCreate([]*model.Step{
		{
			PipelineID: 1,
			PID:        1,
			PPID:       1,
			PGID:       1,
			State:      "running",
			Name:       "build",
		},
	}); err != nil {
		t.Errorf("Unexpected error: insert steps: %s", err)
		return
	}

	// fail due to duplicate pid
	if err := store.StepCreate([]*model.Step{
		{
			PipelineID: 1,
			PID:        1,
			PPID:       1,
			PGID:       1,
			State:      "success",
			Name:       "clone",
		},
	}); err == nil {
		t.Errorf("Unexpected error: duplicate pid")
	}
}

// TODO: func TestStepCascade(t *testing.T) {}
