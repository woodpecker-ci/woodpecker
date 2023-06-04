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

func TestLogCreateFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.Step), new(model.LogEntry))
	defer closer()

	step := model.Step{
		ID: 1,
	}

	logEntries := []*model.LogEntry{
		{
			StepID: step.ID,
			Data:   []byte("hello"),
			Line:   1,
			Time:   0,
		},
		{
			StepID: step.ID,
			Data:   []byte("world"),
			Line:   2,
			Time:   10,
		},
	}

	err := store.LogSave(&step, logEntries)
	if err != nil {
		t.Errorf("Unexpected error: log create: %s", err)
	}

	_logEntries, err := store.LogFind(&step)
	if err != nil {
		t.Errorf("Unexpected error: log create: %s", err)
	}

	if got, want := len(_logEntries), len(logEntries); got != want {
		t.Errorf("Want %d log entries, got %d", want, got)
	}
}

func TestLogAppend(t *testing.T) {
	store, closer := newTestStore(t, new(model.Step), new(model.LogEntry))
	defer closer()

	step := model.Step{
		ID: 1,
	}
	logEntries := []*model.LogEntry{
		{
			StepID: step.ID,
			Data:   []byte("hello"),
			Line:   1,
			Time:   0,
		},
		{
			StepID: step.ID,
			Data:   []byte("world"),
			Line:   2,
			Time:   10,
		},
	}

	if err := store.LogSave(&step, logEntries); err != nil {
		t.Errorf("Unexpected error: log create: %s", err)
	}

	logEntry := &model.LogEntry{
		StepID: step.ID,
		Data:   []byte("allo?"),
		Line:   3,
		Time:   20,
	}

	if err := store.LogAppend(logEntry); err != nil {
		t.Errorf("Unexpected error: log append: %s", err)
	}

	_logEntries, err := store.LogFind(&step)
	if err != nil {
		t.Errorf("Unexpected error: log find: %s", err)
	}

	if got, want := len(_logEntries), len(logEntries)+1; got != want {
		t.Errorf("Want %d log entries, got %d", want, got)
	}
}
