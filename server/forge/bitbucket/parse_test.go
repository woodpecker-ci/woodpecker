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
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucket/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func Test_parseHook(t *testing.T) {
	t.Run("unsupported hook", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPush)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, "issue:created")

		r, b, err := parseHook(req)
		assert.Nil(t, r)
		assert.Nil(t, b)
		assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
	})

	t.Run("malformed pull-request hook", func(t *testing.T) {
		buf := bytes.NewBufferString("[]")
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullCreated)

		_, _, err := parseHook(req)
		assert.Error(t, err)
	})

	t.Run("pull-request", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPull)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullCreated)

		r, b, err := parseHook(req)
		assert.NoError(t, err)
		assert.Equal(t, "user_name/repo_name", r.FullName)
		assert.Equal(t, model.EventPull, b.Event)
		assert.Equal(t, "d3022fc0ca3d", b.Commit.SHA)
	})

	t.Run("pull-request merged", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPullRequestMerged)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullMerged)

		r, b, err := parseHook(req)
		assert.NoError(t, err)
		assert.Equal(t, "anbraten/test-2", r.FullName)
		assert.Equal(t, model.EventPullClosed, b.Event)
		assert.Equal(t, "006704dbeab2", b.Commit.SHA)
	})

	t.Run("pull-request closed", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPullRequestDeclined)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullDeclined)

		r, b, err := parseHook(req)
		assert.NoError(t, err)
		assert.Equal(t, "anbraten/test-2", r.FullName)
		assert.Equal(t, model.EventPullClosed, b.Event)
		assert.Equal(t, "f90e18fc9d45", b.Commit.SHA)
	})

	t.Run("malformed push", func(t *testing.T) {
		buf := bytes.NewBufferString("[]")
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPush)

		_, _, err := parseHook(req)
		assert.Error(t, err)
	})

	t.Run("missing commit sha", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPushEmptyHash)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPush)

		r, b, err := parseHook(req)
		assert.Nil(t, r)
		assert.Nil(t, b)
		assert.NoError(t, err)
	})

	t.Run("push hook", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPush)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPush)

		r, b, err := parseHook(req)
		assert.NoError(t, err)
		assert.Equal(t, "martinherren1984/publictestrepo", r.FullName)
		assert.Equal(t, "https://bitbucket.org/martinherren1984/publictestrepo", r.Clone)
		assert.Equal(t, "c14c1bb05dfb1fdcdf06b31485fff61b0ea44277", b.Commit.SHA)
		assert.Equal(t, "a\n", b.Commit.Message)
	})
}
