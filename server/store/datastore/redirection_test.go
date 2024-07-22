// Copyright 2022 Woodpecker Authors
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

func TestCreateRedirection(t *testing.T) {
	store, closer := newTestStore(t, new(model.Redirection))
	defer closer()

	redirection := &model.Redirection{
		RepoID:   1,
		FullName: "foo/bar",
	}
	assert.NoError(t, store.CreateRedirection(redirection))
}

func TestHasRedirectionForRepo(t *testing.T) {
	store, closer := newTestStore(t, new(model.Redirection))
	defer closer()

	redirection := &model.Redirection{
		RepoID:   1,
		FullName: "foo/bar",
	}
	assert.NoError(t, store.CreateRedirection(redirection))
	has, err := store.HasRedirectionForRepo(1, "foo/bar")
	assert.NoError(t, err)
	assert.True(t, has)
	has, err = store.HasRedirectionForRepo(1, "foo/baz")
	assert.NoError(t, err)
	assert.False(t, has)
}
