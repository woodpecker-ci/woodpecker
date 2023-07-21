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

func TestSecretFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.Secret))
	defer closer()

	err := store.SecretCreate(&model.Secret{
		RepoID: 1,
		Name:   "password",
		Value:  "correct-horse-battery-staple",
		Images: []string{"golang", "node"},
		Events: []model.WebhookEvent{"push", "tag"},
	})
	if err != nil {
		t.Errorf("Unexpected error: insert secret: %s", err)
		return
	}

	secret, err := store.SecretFind(&model.Repo{ID: 1}, "password")
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := secret.RepoID, int64(1); got != want {
		t.Errorf("Want repo id %d, got %d", want, got)
	}
	if got, want := secret.Name, "password"; got != want {
		t.Errorf("Want secret name %s, got %s", want, got)
	}
	if got, want := secret.Value, "correct-horse-battery-staple"; got != want {
		t.Errorf("Want secret value %s, got %s", want, got)
	}
	if got, want := secret.Events[0], model.EventPush; got != want {
		t.Errorf("Want secret event %s, got %s", want, got)
	}
	if got, want := secret.Events[1], model.EventTag; got != want {
		t.Errorf("Want secret event %s, got %s", want, got)
	}
	if got, want := secret.Images[0], "golang"; got != want {
		t.Errorf("Want secret image %s, got %s", want, got)
	}
	if got, want := secret.Images[1], "node"; got != want {
		t.Errorf("Want secret image %s, got %s", want, got)
	}
}

func TestSecretList(t *testing.T) {
	store, closer := newTestStore(t, new(model.Secret))
	defer closer()

	createTestSecrets(t, store)

	list, err := store.SecretList(&model.Repo{ID: 1, OrgID: 12}, false, &model.ListOptions{Page: 1, PerPage: 50})
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestSecretListAll(t *testing.T) {
	store, closer := newTestStore(t, new(model.Secret))
	defer closer()

	createTestSecrets(t, store)

	list, err := store.SecretListAll()
	assert.NoError(t, err)
	assert.Len(t, list, 4)
}

func TestSecretPipelineList(t *testing.T) {
	store, closer := newTestStore(t, new(model.Secret))
	defer closer()

	createTestSecrets(t, store)

	list, err := store.SecretList(&model.Repo{ID: 1, OrgID: 12}, true, &model.ListOptions{Page: 1, PerPage: 50})
	assert.NoError(t, err)
	assert.Len(t, list, 4)
}

func TestSecretUpdate(t *testing.T) {
	store, closer := newTestStore(t, new(model.Secret))
	defer closer()

	secret := &model.Secret{
		RepoID: 1,
		Name:   "foo",
		Value:  "baz",
	}
	if err := store.SecretCreate(secret); err != nil {
		t.Errorf("Unexpected error: insert secret: %s", err)
		return
	}
	secret.Value = "qux"
	assert.EqualValues(t, 1, secret.ID)
	if err := store.SecretUpdate(secret); err != nil {
		t.Errorf("Unexpected error: update secret: %s", err)
		return
	}
	updated, err := store.SecretFind(&model.Repo{ID: 1}, "foo")
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := updated.Value, "qux"; got != want {
		t.Errorf("Want secret value %s, got %s", want, got)
	}
}

func TestSecretDelete(t *testing.T) {
	store, closer := newTestStore(t, new(model.Secret))
	defer closer()

	secret := &model.Secret{
		RepoID: 1,
		Name:   "foo",
		Value:  "baz",
	}
	if err := store.SecretCreate(secret); err != nil {
		t.Errorf("Unexpected error: insert secret: %s", err)
		return
	}

	if err := store.SecretDelete(secret); err != nil {
		t.Errorf("Unexpected error: delete secret: %s", err)
		return
	}
	_, err := store.SecretFind(&model.Repo{ID: 1}, "foo")
	if err == nil {
		t.Errorf("Expect error: sql.ErrNoRows")
		return
	}
}

func TestSecretIndexes(t *testing.T) {
	store, closer := newTestStore(t, new(model.Secret))
	defer closer()

	if err := store.SecretCreate(&model.Secret{
		RepoID: 1,
		Name:   "foo",
		Value:  "bar",
	}); err != nil {
		t.Errorf("Unexpected error: insert secret: %s", err)
		return
	}

	// fail due to duplicate name
	if err := store.SecretCreate(&model.Secret{
		RepoID: 1,
		Name:   "foo",
		Value:  "baz",
	}); err == nil {
		t.Errorf("Unexpected error: duplicate name")
	}
}

func createTestSecrets(t *testing.T, store *storage) {
	assert.NoError(t, store.SecretCreate(&model.Secret{
		OrgID: 12,
		Name:  "usr",
		Value: "sec",
	}))
	assert.NoError(t, store.SecretCreate(&model.Secret{
		RepoID: 1,
		Name:   "foo",
		Value:  "bar",
	}))
	assert.NoError(t, store.SecretCreate(&model.Secret{
		RepoID: 1,
		Name:   "baz",
		Value:  "qux",
	}))
	assert.NoError(t, store.SecretCreate(&model.Secret{
		Name:  "global",
		Value: "val",
	}))
}

func TestOrgSecretFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.Secret))
	defer closer()

	err := store.SecretCreate(&model.Secret{
		OrgID:  12,
		Name:   "password",
		Value:  "correct-horse-battery-staple",
		Images: []string{"golang", "node"},
		Events: []model.WebhookEvent{"push", "tag"},
	})
	if err != nil {
		t.Errorf("Unexpected error: insert secret: %s", err)
		return
	}

	secret, err := store.OrgSecretFind(12, "password")
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := secret.OrgID, int64(12); got != want {
		t.Errorf("Want org_id %d, got %d", want, got)
	}
	if got, want := secret.Name, "password"; got != want {
		t.Errorf("Want secret name %s, got %s", want, got)
	}
	if got, want := secret.Value, "correct-horse-battery-staple"; got != want {
		t.Errorf("Want secret value %s, got %s", want, got)
	}
	if got, want := secret.Events[0], model.EventPush; got != want {
		t.Errorf("Want secret event %s, got %s", want, got)
	}
	if got, want := secret.Events[1], model.EventTag; got != want {
		t.Errorf("Want secret event %s, got %s", want, got)
	}
	if got, want := secret.Images[0], "golang"; got != want {
		t.Errorf("Want secret image %s, got %s", want, got)
	}
	if got, want := secret.Images[1], "node"; got != want {
		t.Errorf("Want secret image %s, got %s", want, got)
	}
}

func TestOrgSecretList(t *testing.T) {
	store, closer := newTestStore(t, new(model.Secret))
	defer closer()

	createTestSecrets(t, store)

	list, err := store.OrgSecretList(12, &model.ListOptions{All: true})
	assert.NoError(t, err)
	assert.Len(t, list, 1)

	assert.True(t, list[0].Organization())
}

func TestGlobalSecretFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.Secret))
	defer closer()

	err := store.SecretCreate(&model.Secret{
		Name:   "password",
		Value:  "correct-horse-battery-staple",
		Images: []string{"golang", "node"},
		Events: []model.WebhookEvent{"push", "tag"},
	})
	if err != nil {
		t.Errorf("Unexpected error: insert secret: %s", err)
		return
	}

	secret, err := store.GlobalSecretFind("password")
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := secret.Name, "password"; got != want {
		t.Errorf("Want secret name %s, got %s", want, got)
	}
	if got, want := secret.Value, "correct-horse-battery-staple"; got != want {
		t.Errorf("Want secret value %s, got %s", want, got)
	}
	if got, want := secret.Events[0], model.EventPush; got != want {
		t.Errorf("Want secret event %s, got %s", want, got)
	}
	if got, want := secret.Events[1], model.EventTag; got != want {
		t.Errorf("Want secret event %s, got %s", want, got)
	}
	if got, want := secret.Images[0], "golang"; got != want {
		t.Errorf("Want secret image %s, got %s", want, got)
	}
	if got, want := secret.Images[1], "node"; got != want {
		t.Errorf("Want secret image %s, got %s", want, got)
	}
}

func TestGlobalSecretList(t *testing.T) {
	store, closer := newTestStore(t, new(model.Secret))
	defer closer()

	createTestSecrets(t, store)

	list, err := store.GlobalSecretList(&model.ListOptions{All: true})
	assert.NoError(t, err)
	assert.Len(t, list, 1)

	assert.True(t, list[0].Global())
}
