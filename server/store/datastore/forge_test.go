// Copyright 2024 Woodpecker Authors
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

func TestForgeCRUD(t *testing.T) {
	store, closer := newTestStore(t, new(model.Forge), new(model.Repo), new(model.User))
	defer closer()

	forge1 := &model.Forge{
		Type:         "github",
		URL:          "https://github.com",
		Client:       "client",
		ClientSecret: "secret",
		SkipVerify:   false,
		AdditionalOptions: map[string]any{
			"foo": "bar",
		},
	}

	// create first forge to play with
	assert.NoError(t, store.ForgeCreate(forge1))
	assert.EqualValues(t, "github", forge1.Type)

	// retrieve it
	forgeOne, err := store.ForgeGet(forge1.ID)
	assert.NoError(t, err)
	assert.EqualValues(t, forge1, forgeOne)

	// change type
	assert.NoError(t, store.ForgeUpdate(&model.Forge{ID: forge1.ID, Type: "gitlab"}))

	// find updated forge by id
	forgeOne, err = store.ForgeGet(forge1.ID)
	assert.NoError(t, err)
	assert.EqualValues(t, "gitlab", forgeOne.Type)

	// create two more forges and repos
	someUser := &model.Forge{Type: "bitbucket"}
	assert.NoError(t, store.ForgeCreate(someUser))
	assert.NoError(t, store.ForgeCreate(&model.Forge{Type: "gitea"}))

	// get all repos for a specific forge
	forges, err := store.ForgeList(&model.ListOptions{All: true})
	assert.NoError(t, err)
	assert.Len(t, forges, 3)

	// delete an forge and check if it's gone
	assert.NoError(t, store.ForgeDelete(forge1))
	_, err = store.ForgeGet(forge1.ID)
	assert.Error(t, err)
}
