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
	"io/ioutil"
	"testing"

	"github.com/woodpecker-ci/woodpecker/model"
)

func TestFileFind(t *testing.T) {
	s := newTest()
	defer func() {
		s.Exec("delete from files")
		s.Close()
	}()

	if err := s.FileCreate(
		&model.File{
			BuildID: 2,
			ProcID:  1,
			Name:    "hello.txt",
			Mime:    "text/plain",
			Size:    11,
		},
		bytes.NewBufferString("hello world"),
	); err != nil {
		t.Errorf("Unexpected error: insert file: %s", err)
		return
	}

	file, err := s.FileFind(&model.Proc{ID: 1}, "hello.txt")
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := file.ID, int64(1); got != want {
		t.Errorf("Want file id %d, got %d", want, got)
	}
	if got, want := file.BuildID, int64(2); got != want {
		t.Errorf("Want file build id %d, got %d", want, got)
	}
	if got, want := file.ProcID, int64(1); got != want {
		t.Errorf("Want file proc id %d, got %d", want, got)
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

	rc, err := s.FileRead(&model.Proc{ID: 1}, "hello.txt")
	if err != nil {
		t.Error(err)
		return
	}
	out, _ := ioutil.ReadAll(rc)
	if got, want := string(out), "hello world"; got != want {
		t.Errorf("Want file data %s, got %s", want, got)
	}
}

func TestFileList(t *testing.T) {
	s := newTest()
	defer func() {
		s.Exec("delete from files")
		s.Close()
	}()

	s.FileCreate(
		&model.File{
			BuildID: 1,
			ProcID:  1,
			Name:    "hello.txt",
			Mime:    "text/plain",
			Size:    11,
		},
		bytes.NewBufferString("hello world"),
	)
	s.FileCreate(
		&model.File{
			BuildID: 1,
			ProcID:  1,
			Name:    "hola.txt",
			Mime:    "text/plain",
			Size:    11,
		},
		bytes.NewBufferString("hola mundo"),
	)

	files, err := s.FileList(&model.Build{ID: 1})
	if err != nil {
		t.Errorf("Unexpected error: select files: %s", err)
		return
	}

	if got, want := len(files), 2; got != want {
		t.Errorf("Wanted %d files, got %d", want, got)
	}
}

func TestFileIndexes(t *testing.T) {
	s := newTest()
	defer func() {
		s.Exec("delete from files")
		s.Close()
	}()

	if err := s.FileCreate(
		&model.File{
			BuildID: 1,
			ProcID:  1,
			Name:    "hello.txt",
			Size:    11,
			Mime:    "text/plain",
		},
		bytes.NewBufferString("hello world"),
	); err != nil {
		t.Errorf("Unexpected error: insert file: %s", err)
		return
	}

	// fail due to duplicate file name
	if err := s.FileCreate(
		&model.File{
			BuildID: 1,
			ProcID:  1,
			Name:    "hello.txt",
			Mime:    "text/plain",
			Size:    11,
		},
		bytes.NewBufferString("hello world"),
	); err == nil {
		t.Errorf("Unexpected error: dupliate pid")
	}
}

// func TestFileCascade(t *testing.T) {
// 	s := newTest()
// 	defer s.Close()
//
//
// 	err1 := s.ProcCreate([]*model.Proc{
// 		{
// 			BuildID: 1,
// 			PID:     1,
// 			PGID:    1,
// 			Name:    "build",
// 			State:   "success",
// 		},
// 	})
// 	err2 := s.FileCreate(
// 		&model.File{
// 			BuildID: 1,
// 			ProcID:  1,
// 			Name:    "hello.txt",
// 			Mime:    "text/plain",
// 			Size:    11,
// 		},
// 		bytes.NewBufferString("hello world"),
// 	)
//
// 	if err1 != nil {
// 		t.Errorf("Unexpected error: cannot insert proc: %s", err1)
// 	} else if err2 != nil {
// 		t.Errorf("Unexpected error: cannot insert file: %s", err2)
// 	}
//
// 	if _, err3 := s.ProcFind(&model.Build{ID: 1}, 1); err3 != nil {
// 		t.Errorf("Unexpected error: cannot get inserted proc: %s", err3)
// 	}
//
// 	db.Exec("delete from procs where proc_id = 1")
//
// 	file, err4 := s.FileFind(&model.Proc{ID: 1}, "hello.txt")
// 	if err4 == nil {
// 		t.Errorf("Expected no rows in result set error")
// 		t.Log(file)
// 	}
// }
