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

func TestProcFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.Proc), new(model.Pipeline))
	defer closer()

	procs := []*model.Proc{
		{
			PipelineID: 1000,
			PID:        1,
			PPID:       2,
			PGID:       3,
			Name:       "build",
			State:      model.StatusSuccess,
			Error:      "pc load letter",
			ExitCode:   255,
			Machine:    "localhost",
			Platform:   "linux/amd64",
			Environ:    map[string]string{"GOLANG": "tip"},
		},
	}
	assert.NoError(t, store.ProcCreate(procs))
	assert.EqualValues(t, 1, procs[0].ID)
	assert.Error(t, store.ProcCreate(procs))

	proc, err := store.ProcFind(&model.Pipeline{ID: 1000}, 1)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, procs[0], proc)
}

func TestProcChild(t *testing.T) {
	store, closer := newTestStore(t, new(model.Proc), new(model.Pipeline))
	defer closer()

	err := store.ProcCreate([]*model.Proc{
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
		t.Errorf("Unexpected error: insert procs: %s", err)
		return
	}
	proc, err := store.ProcChild(&model.Pipeline{ID: 1}, 1, "build")
	if err != nil {
		t.Error(err)
		return
	}

	if got, want := proc.PID, 2; got != want {
		t.Errorf("Want proc pid %d, got %d", want, got)
	}
	if got, want := proc.Name, "build"; got != want {
		t.Errorf("Want proc name %s, got %s", want, got)
	}
}

func TestProcList(t *testing.T) {
	store, closer := newTestStore(t, new(model.Proc), new(model.Pipeline))
	defer closer()

	err := store.ProcCreate([]*model.Proc{
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
		t.Errorf("Unexpected error: insert procs: %s", err)
		return
	}
	procs, err := store.ProcList(&model.Pipeline{ID: 1})
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := len(procs), 2; got != want {
		t.Errorf("Want %d procs, got %d", want, got)
	}
}

func TestProcUpdate(t *testing.T) {
	store, closer := newTestStore(t, new(model.Proc), new(model.Pipeline))
	defer closer()

	proc := &model.Proc{
		PipelineID: 1,
		PID:        1,
		PPID:       2,
		PGID:       3,
		Name:       "build",
		State:      "pending",
		Error:      "pc load letter",
		ExitCode:   255,
		Machine:    "localhost",
		Platform:   "linux/amd64",
		Environ:    map[string]string{"GOLANG": "tip"},
	}
	if err := store.ProcCreate([]*model.Proc{proc}); err != nil {
		t.Errorf("Unexpected error: insert proc: %s", err)
		return
	}
	proc.State = "running"
	if err := store.ProcUpdate(proc); err != nil {
		t.Errorf("Unexpected error: update proc: %s", err)
		return
	}
	updated, err := store.ProcFind(&model.Pipeline{ID: 1}, 1)
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := updated.State, model.StatusRunning; got != want {
		t.Errorf("Want proc name %s, got %s", want, got)
	}
}

func TestProcIndexes(t *testing.T) {
	store, closer := newTestStore(t, new(model.Proc), new(model.Pipeline))
	defer closer()

	if err := store.ProcCreate([]*model.Proc{
		{
			PipelineID: 1,
			PID:        1,
			PPID:       1,
			PGID:       1,
			State:      "running",
			Name:       "build",
		},
	}); err != nil {
		t.Errorf("Unexpected error: insert procs: %s", err)
		return
	}

	// fail due to duplicate pid
	if err := store.ProcCreate([]*model.Proc{
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

// TODO: func TestProcCascade(t *testing.T) {}
