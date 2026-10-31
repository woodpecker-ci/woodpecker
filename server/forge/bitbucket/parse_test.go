// // Copyright 2018 Drone.IO Inc.
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

package bitbucket

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucket/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestParseHook(t *testing.T) {
	t.Run("unsupported hook", func(t *testing.T) {
		_, r, b, err := parseHook("issue:created", []byte(fixtures.HookPush))
		assert.Nil(t, r)
		assert.Nil(t, b)
		assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
	})

	t.Run("malformed pull-request hook", func(t *testing.T) {
		_, _, _, err := parseHook(hookPullCreated, []byte("[]"))
		assert.Error(t, err)
	})

	t.Run("pull-request", func(t *testing.T) {
		pr, r, b, err := parseHook(hookPullCreated, []byte(fixtures.HookPull))
		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Equal(t, "user_name/repo_name", r.FullName)
		assert.Equal(t, model.EventPull, b.Event)
		assert.Equal(t, "d3022fc0ca3d", b.Commit)
	})

	t.Run("pull-request merged", func(t *testing.T) {
		pr, r, b, err := parseHook(hookPullMerged, []byte(fixtures.HookPullRequestMerged))
		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Equal(t, "anbraten/test-2", r.FullName)
		assert.Equal(t, model.EventPullClosed, b.Event)
		assert.Equal(t, "006704dbeab2", b.Commit)
	})

	t.Run("pull-request closed", func(t *testing.T) {
		pr, r, b, err := parseHook(hookPullDeclined, []byte(fixtures.HookPullRequestDeclined))
		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Equal(t, "anbraten/test-2", r.FullName)
		assert.Equal(t, model.EventPullClosed, b.Event)
		assert.Equal(t, "f90e18fc9d45", b.Commit)
	})

	t.Run("malformed push", func(t *testing.T) {
		_, _, _, err := parseHook(hookPush, []byte("[]"))
		assert.Error(t, err)
	})

	t.Run("missing commit sha", func(t *testing.T) {
		_, r, b, err := parseHook(hookPush, []byte(fixtures.HookPushEmptyHash))
		assert.Nil(t, r)
		assert.Nil(t, b)
		assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
	})

	t.Run("push hook", func(t *testing.T) {
		pr, r, b, err := parseHook(hookPush, []byte(fixtures.HookPush))
		assert.NoError(t, err)
		assert.Nil(t, pr)
		assert.Equal(t, "martinherren1984/publictestrepo", r.FullName)
		assert.Equal(t, "https://bitbucket.org/martinherren1984/publictestrepo", r.Clone)
		assert.Equal(t, "c14c1bb05dfb1fdcdf06b31485fff61b0ea44277", b.Commit)
		assert.Equal(t, "a\n", b.Message)
	})
}
