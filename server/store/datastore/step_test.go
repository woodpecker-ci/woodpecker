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

	"go.woodpecker-ci.org/woodpecker/server/model"
	"go.woodpecker-ci.org/woodpecker/server/store/types"
)

func TestStepFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.Step), new(model.Pipeline))
	sess := store.engine.NewSession()

	defer closer()

	steps := []*model.Step{
		{
			UUID:       "8d89104f-d44e-4b45-b86e-17f8b5e74a0e",
			PipelineID: 1000,
			PID:        1,
			PPID:       2,
			Name:       "build",
			State:      model.StatusSuccess,
			Error:      "pc load letter",
			ExitCode:   255,
		},
	}
	assert.NoError(t, store.stepCreate(sess, steps))
	assert.EqualValues(t, 1, steps[0].ID)
	assert.Error(t, store.stepCreate(sess, steps))
	assert.NoError(t, sess.Close())

	step, err := store.StepFind(&model.Pipeline{ID: 1000}, 1)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, steps[0], step)
}

func TestStepChild(t *testing.T) {
	store, closer := newTestStore(t, new(model.Step), new(model.Pipeline))
	defer closer()

	sess := store.engine.NewSession()
	err := store.stepCreate(sess, []*model.Step{
		{
			UUID:       "ea6d4008-8ace-4f8a-ad03-53f1756465d9",
			PipelineID: 1,
			PID:        1,
			PPID:       1,
			State:      "success",
		},
		{
			UUID:       "2bf387f7-2913-4907-814c-c9ada88707c0",
			PipelineID: 1,
			PID:        2,
			PPID:       1,
			Name:       "build",
			State:      "success",
		},
	})
	if err != nil {
		t.Errorf("Unexpected error: insert steps: %s", err)
		return
	}
	_ = sess.Commit()
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

	sess := store.engine.NewSession()
	err := store.stepCreate(sess, []*model.Step{
		{
			UUID:       "2bf387f7-2913-4907-814c-c9ada88707c0",
			PipelineID: 2,
			PID:        1,
			PPID:       1,
			State:      "success",
		},
		{
			UUID:       "4b04073c-1827-4aa4-a5f5-c7b21c5e44a6",
			PipelineID: 1,
			PID:        1,
			PPID:       1,
			State:      "success",
		},
		{
			UUID:       "40aab045-970b-4892-b6df-6f825a7ec97a",
			PipelineID: 1,
			PID:        2,
			PPID:       1,
			Name:       "build",
			State:      "success",
		},
	})
	if err != nil {
		t.Errorf("Unexpected error: insert steps: %s", err)
		return
	}
	_ = sess.Commit()
	steps, err := store.StepList(&model.Pipeline{ID: 1})
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
		UUID:       "fc7c7fd6-553e-480b-8ed7-30d8563d0b79",
		PipelineID: 1,
		PID:        1,
		PPID:       2,
		Name:       "build",
		State:      "pending",
		Error:      "pc load letter",
		ExitCode:   255,
	}
	sess := store.engine.NewSession()
	if err := store.stepCreate(sess, []*model.Step{step}); err != nil {
		t.Errorf("Unexpected error: insert step: %s", err)
		return
	}
	_ = sess.Commit()
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

	sess := store.engine.NewSession()
	defer sess.Close()

	if err := store.stepCreate(sess, []*model.Step{
		{
			UUID:       "4db7e5fc-5312-4d02-9e14-b51b9e3242cc",
			PipelineID: 1,
			PID:        1,
			PPID:       1,
			State:      "running",
			Name:       "build",
		},
	}); err != nil {
		t.Errorf("Unexpected error: insert steps: %s", err)
		return
	}

	// fail due to duplicate pid
	if err := store.stepCreate(sess, []*model.Step{
		{
			UUID:       "c1f33a9e-2a02-4579-95ec-90255d785a12",
			PipelineID: 1,
			PID:        1,
			PPID:       1,
			State:      "success",
			Name:       "clone",
		},
	}); err == nil {
		t.Errorf("Unexpected error: duplicate pid")
	}
}

func TestStepByUUID(t *testing.T) {
	store, closer := newTestStore(t, new(model.Step), new(model.Pipeline))
	defer closer()

	sess := store.engine.NewSession()
	assert.NoError(t, store.stepCreate(sess, []*model.Step{
		{
			UUID:       "4db7e5fc-5312-4d02-9e14-b51b9e3242cc",
			PipelineID: 1,
			PID:        1,
			PPID:       1,
			State:      "running",
			Name:       "build",
		},
		{
			UUID:       "fc7c7fd6-553e-480b-8ed7-30d8563d0b79",
			PipelineID: 4,
			PID:        6,
			PPID:       7,
			Name:       "build",
			State:      "pending",
			Error:      "pc load letter",
			ExitCode:   255,
		},
	}))
	_ = sess.Close()

	step, err := store.StepByUUID("4db7e5fc-5312-4d02-9e14-b51b9e3242cc")
	assert.NoError(t, err)
	assert.NotEmpty(t, step)

	step, err = store.StepByUUID("52feb6f5-8ce2-40c0-9937-9d0e3349c98c")
	assert.ErrorIs(t, err, types.RecordNotExist)
	assert.Empty(t, step)
}

func TestStepLoad(t *testing.T) {
	store, closer := newTestStore(t, new(model.Step))
	defer closer()

	sess := store.engine.NewSession()
	assert.NoError(t, store.stepCreate(sess, []*model.Step{
		{
			UUID:       "4db7e5fc-5312-4d02-9e14-b51b9e3242cc",
			PipelineID: 1,
			PID:        1,
			PPID:       1,
			State:      "running",
			Name:       "build",
		},
		{
			UUID:       "fc7c7fd6-553e-480b-8ed7-30d8563d0b79",
			PipelineID: 4,
			PID:        6,
			PPID:       7,
			Name:       "build",
			State:      "pending",
			Error:      "pc load letter",
			ExitCode:   255,
		},
	}))
	_ = sess.Close()

	step, err := store.StepLoad(1)
	assert.NoError(t, err)
	assert.NotEmpty(t, step)
	assert.Equal(t, step.UUID, "4db7e5fc-5312-4d02-9e14-b51b9e3242cc")

	step, err = store.StepLoad(5)
	assert.ErrorIs(t, err, types.RecordNotExist)
	assert.Empty(t, step)
}
