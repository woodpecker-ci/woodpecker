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

package datastore

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func TestLogCreateFindDelete(t *testing.T) {
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

	assert.NoError(t, store.LogAppend(&step, logEntries))

	// we want to find our inserted logs
	_logEntries, err := store.LogFind(&step)
	assert.NoError(t, err)
	assert.Len(t, _logEntries, len(logEntries))

	// delete and check
	assert.NoError(t, store.LogDelete(&step))
	_logEntries, err = store.LogFind(&step)
	assert.NoError(t, err)
	assert.Len(t, _logEntries, 0)
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

	assert.NoError(t, store.LogAppend(&step, logEntries))

	logEntry := &model.LogEntry{
		StepID: step.ID,
		Data:   []byte("allo?"),
		Line:   3,
		Time:   20,
	}

	assert.NoError(t, store.LogAppend(&step, []*model.LogEntry{logEntry}))

	_logEntries, err := store.LogFind(&step)
	assert.NoError(t, err)
	assert.Len(t, _logEntries, len(logEntries)+1)
}
