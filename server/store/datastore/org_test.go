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

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestOrgCRUD(t *testing.T) {
	store, closer := newTestStore(t, new(model.Org), new(model.Repo), new(model.Secret), new(model.Config), new(model.Perm), new(model.Registry), new(model.Redirection), new(model.Pipeline))
	defer closer()

	org1 := &model.Org{
		Name:    "someAwesomeOrg",
		IsUser:  false,
		Private: true,
	}

	// create first org to play with
	assert.NoError(t, store.OrgCreate(org1))
	assert.EqualValues(t, "someawesomeorg", org1.Name)

	// retrieve it
	orgOne, err := store.OrgGet(org1.ID)
	assert.NoError(t, err)
	assert.EqualValues(t, org1, orgOne)

	// change name
	assert.NoError(t, store.OrgUpdate(&model.Org{ID: org1.ID, Name: "RenamedOrg"}))

	// force a name duplication and fail
	assert.Error(t, store.OrgCreate(&model.Org{Name: "reNamedorg"}))

	// find updated org by name
	orgOne, err = store.OrgFindByName("renamedorG")
	assert.NoError(t, err)
	assert.NotEqualValues(t, org1, orgOne)
	assert.EqualValues(t, org1.ID, orgOne.ID)
	assert.EqualValues(t, false, orgOne.IsUser)
	assert.EqualValues(t, false, orgOne.Private)
	assert.EqualValues(t, "renamedorg", orgOne.Name)

	// create two more orgs and repos
	someUser := &model.Org{Name: "some_other_u", IsUser: true}
	assert.NoError(t, store.OrgCreate(someUser))
	assert.NoError(t, store.OrgCreate(&model.Org{Name: "some_other_org"}))
	assert.NoError(t, store.CreateRepo(&model.Repo{UserID: 1, Owner: "some_other_u", Name: "abc", FullName: "some_other_u/abc", OrgID: someUser.ID}))
	assert.NoError(t, store.CreateRepo(&model.Repo{UserID: 1, Owner: "some_other_u", Name: "xyz", FullName: "some_other_u/xyz", OrgID: someUser.ID}))
	assert.NoError(t, store.CreateRepo(&model.Repo{UserID: 1, Owner: "renamedorg", Name: "567", FullName: "renamedorg/567", OrgID: orgOne.ID}))

	// get all repos for a specific org
	repos, err := store.OrgRepoList(someUser, &model.ListOptions{All: true})
	assert.NoError(t, err)
	assert.Len(t, repos, 2)

	// delete an org and check if it's gone
	assert.NoError(t, store.OrgDelete(org1.ID))
	assert.Error(t, store.OrgDelete(org1.ID))
}
