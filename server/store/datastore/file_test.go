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
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestFileFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.File), new(model.Step))
	defer closer()

	if err := store.FileCreate(
		&model.File{
			PipelineID: 2,
			StepID:     1,
			Name:       "hello.txt",
			Mime:       "text/plain",
			Size:       11,
		},
		bytes.NewBufferString("hello world"),
	); err != nil {
		t.Errorf("Unexpected error: insert file: %s", err)
		return
	}

	file, err := store.FileFind(&model.Step{ID: 1}, "hello.txt")
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := file.ID, int64(1); got != want {
		t.Errorf("Want file id %d, got %d", want, got)
	}
	if got, want := file.PipelineID, int64(2); got != want {
		t.Errorf("Want file pipeline id %d, got %d", want, got)
	}
	if got, want := file.StepID, int64(1); got != want {
		t.Errorf("Want file step id %d, got %d", want, got)
	}
	if got, want := file.Name, "hello.txt"; got != want {
		t.Errorf("Want file name %s, got %s", want, got)
	}
	if got, want := file.Mime, "text/plain"; got != want {
		t.Errorf("Want file mime %s, got %s", want, got)
	}
	if got, want := file.Size, 11; got != want {
		t.Errorf("Want file size %d, got %d", want, got)
	}

	rc, err := store.FileRead(&model.Step{ID: 1}, "hello.txt")
	if err != nil {
		t.Error(err)
		return
	}
	out, _ := io.ReadAll(rc)
	if got, want := string(out), "hello world"; got != want {
		t.Errorf("Want file data %s, got %s", want, got)
	}
}

func TestFileList(t *testing.T) {
	store, closer := newTestStore(t, new(model.File), new(model.Pipeline))
	defer closer()

	assert.NoError(t, store.FileCreate(
		&model.File{
			PipelineID: 1,
			StepID:     1,
			Name:       "hello.txt",
			Mime:       "text/plain",
			Size:       11,
		},
		bytes.NewBufferString("hello world"),
	))
	assert.NoError(t, store.FileCreate(
		&model.File{
			PipelineID: 1,
			StepID:     1,
			Name:       "hola.txt",
			Mime:       "text/plain",
			Size:       11,
		},
		bytes.NewBufferString("hola mundo"),
	))

	files, err := store.FileList(&model.Pipeline{ID: 1}, &model.ListOptions{Page: 1, PerPage: 50})
	if err != nil {
		t.Errorf("Unexpected error: select files: %s", err)
		return
	}

	if got, want := len(files), 2; got != want {
		t.Errorf("Wanted %d files, got %d", want, got)
	}
}

func TestFileIndexes(t *testing.T) {
	store, closer := newTestStore(t, new(model.File), new(model.Pipeline))
	defer closer()

	if err := store.FileCreate(
		&model.File{
			PipelineID: 1,
			StepID:     1,
			Name:       "hello.txt",
			Size:       11,
			Mime:       "text/plain",
		},
		bytes.NewBufferString("hello world"),
	); err != nil {
		t.Errorf("Unexpected error: insert file: %s", err)
		return
	}

	// fail due to duplicate file name
	if err := store.FileCreate(
		&model.File{
			PipelineID: 1,
			StepID:     1,
			Name:       "hello.txt",
			Mime:       "text/plain",
			Size:       11,
		},
		bytes.NewBufferString("hello world"),
	); err == nil {
		t.Errorf("Unexpected error: duplicate pid")
	}
}

func TestFileCascade(t *testing.T) {
	store, closer := newTestStore(t, new(model.File), new(model.Step), new(model.Pipeline))
	defer closer()

	stepOne := &model.Step{
		PipelineID: 1,
		PID:        1,
		PGID:       1,
		Name:       "build",
		State:      "success",
	}
	err1 := store.StepCreate([]*model.Step{stepOne})
	assert.EqualValues(t, int64(1), stepOne.ID)

	err2 := store.FileCreate(
		&model.File{
			PipelineID: 1,
			StepID:     1,
			Name:       "hello.txt",
			Mime:       "text/plain",
			Size:       11,
		},
		bytes.NewBufferString("hello world"),
	)

	if err1 != nil {
		t.Errorf("Unexpected error: cannot insert step: %s", err1)
	} else if err2 != nil {
		t.Errorf("Unexpected error: cannot insert file: %s", err2)
	}

	if _, err3 := store.StepFind(&model.Pipeline{ID: 1}, 1); err3 != nil {
		t.Errorf("Unexpected error: cannot get inserted step: %s", err3)
	}

	err := store.StepClear(&model.Pipeline{ID: 1, Steps: []*model.Step{stepOne}})
	assert.NoError(t, err)

	file, err4 := store.FileFind(&model.Step{ID: 1}, "hello.txt")
	if err4 == nil {
		t.Errorf("Expected no rows in result set error")
		t.Log(file)
	}
}
