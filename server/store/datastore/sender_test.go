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

func TestSenderFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.Sender))
	defer closer()

	err := store.SenderCreate(&model.Sender{
		RepoID: 1,
		Login:  "octocat",
		Allow:  true,
		Block:  false,
	})
	if err != nil {
		t.Errorf("Unexpected error: insert secret: %s", err)
		return
	}

	sender, err := store.SenderFind(&model.Repo{ID: 1}, "octocat")
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := sender.RepoID, int64(1); got != want {
		t.Errorf("Want repo id %d, got %d", want, got)
	}
	if got, want := sender.Login, "octocat"; got != want {
		t.Errorf("Want sender login %s, got %s", want, got)
	}
	if got, want := sender.Allow, true; got != want {
		t.Errorf("Want sender allow %v, got %v", want, got)
	}
}

func TestSenderList(t *testing.T) {
	store, closer := newTestStore(t, new(model.Sender))
	defer closer()

	assert.NoError(t, store.SenderCreate(&model.Sender{
		RepoID: 1,
		Login:  "octocat",
		Allow:  true,
		Block:  false,
	}))
	assert.NoError(t, store.SenderCreate(&model.Sender{
		RepoID: 1,
		Login:  "defunkt",
		Allow:  true,
		Block:  false,
	}))

	list, err := store.SenderList(&model.Repo{ID: 1})
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := len(list), 2; got != want {
		t.Errorf("Want %d senders, got %d", want, got)
	}
}

func TestSenderUpdate(t *testing.T) {
	store, closer := newTestStore(t, new(model.Sender))
	defer closer()

	sender := &model.Sender{
		RepoID: 1,
		Login:  "octocat",
		Allow:  true,
		Block:  false,
	}
	if err := store.SenderCreate(sender); err != nil {
		t.Errorf("Unexpected error: insert sender: %s", err)
		return
	}
	assert.EqualValues(t, 1, sender.ID)
	sender.Allow = false
	if err := store.SenderUpdate(sender); err != nil {
		t.Errorf("Unexpected error: update sender: %s", err)
		return
	}
	updated, err := store.SenderFind(&model.Repo{ID: 1}, "octocat")
	assert.NoError(t, err)
	assert.False(t, updated.Allow)
}

func TestSenderIndexes(t *testing.T) {
	store, closer := newTestStore(t, new(model.Sender))
	defer closer()

	if err := store.SenderCreate(&model.Sender{
		RepoID: 1,
		Login:  "octocat",
		Allow:  true,
		Block:  false,
	}); err != nil {
		t.Errorf("Unexpected error: insert sender: %s", err)
		return
	}

	// fail due to duplicate name
	if err := store.SenderCreate(&model.Sender{
		RepoID: 1,
		Login:  "octocat",
		Allow:  true,
		Block:  false,
	}); err == nil {
		t.Errorf("Unexpected error: duplicate login")
	}
}
