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

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestLogCreateFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.Proc), new(model.Logs))
	defer closer()

	proc := model.Proc{
		ID: 1,
	}
	buf := bytes.NewBufferString("echo hi")
	err := store.LogSave(&proc, buf)
	if err != nil {
		t.Errorf("Unexpected error: log create: %s", err)
	}

	rc, err := store.LogFind(&proc)
	if err != nil {
		t.Errorf("Unexpected error: log create: %s", err)
	}

	defer rc.Close()
	out, _ := io.ReadAll(rc)
	if got, want := string(out), "echo hi"; got != want {
		t.Errorf("Want log data %s, got %s", want, got)
	}
}

func TestLogUpdate(t *testing.T) {
	store, closer := newTestStore(t, new(model.Proc), new(model.Logs))
	defer closer()

	proc := model.Proc{
		ID: 1,
	}
	buf1 := bytes.NewBufferString("echo hi")
	buf2 := bytes.NewBufferString("echo allo?")
	err1 := store.LogSave(&proc, buf1)
	err2 := store.LogSave(&proc, buf2)
	if err1 != nil {
		t.Errorf("Unexpected error: log create: %s", err1)
	}
	if err2 != nil {
		t.Errorf("Unexpected error: log update: %s", err2)
	}

	rc, err := store.LogFind(&proc)
	if err != nil {
		t.Errorf("Unexpected error: log create: %s", err)
	}

	defer rc.Close()
	out, _ := io.ReadAll(rc)
	if got, want := string(out), "echo allo?"; got != want {
		t.Errorf("Want log data %s, got %s", want, got)
	}
}
