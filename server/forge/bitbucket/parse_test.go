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

func TestParseHook(t *testing.T) {
	t.Run("unsupported hook", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPush)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, "issue:created")

		_, r, b, err := parseHook(req)
		assert.Nil(t, r)
		assert.Nil(t, b)
		assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
	})

	t.Run("malformed pull-request hook", func(t *testing.T) {
		buf := bytes.NewBufferString("[]")
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullCreated)

		_, _, _, err := parseHook(req)
		assert.Error(t, err)
	})

	t.Run("pull-request hook without a repository", func(t *testing.T) {
		buf := bytes.NewBufferString("{}")
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullCreated)

		_, r, b, err := parseHook(req)
		assert.Nil(t, r)
		assert.Nil(t, b)
		assert.ErrorIs(t, err, errIncompleteHook)
	})

	t.Run("push hook with a full name that has no owner", func(t *testing.T) {
		buf := bytes.NewBufferString(`{"repository":{"full_name":"nameonly"},"push":{"changes":[{"new":{"name":"main","type":"branch","target":{"hash":"abc"}}}]}}`)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPush)

		_, r, b, err := parseHook(req)
		assert.Nil(t, r)
		assert.Nil(t, b)
		assert.ErrorIs(t, err, errIncompleteHook)
	})

	t.Run("pull-request", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPull)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullCreated)

		pr, r, bs, err := parseHook(req)
		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Len(t, bs, 1)
		b := bs[0]
		assert.Equal(t, "user_name/repo_name", r.FullName)
		assert.Equal(t, model.EventPull, b.Event)
		assert.Equal(t, "d3022fc0ca3d", b.Commit)
	})

	t.Run("pull-request merged", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPullRequestMerged)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullMerged)

		pr, r, bs, err := parseHook(req)
		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Len(t, bs, 1)
		b := bs[0]
		assert.Equal(t, "anbraten/test-2", r.FullName)
		assert.Equal(t, model.EventPullClosed, b.Event)
		assert.Equal(t, "006704dbeab2", b.Commit)
	})

	t.Run("pull-request closed", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPullRequestDeclined)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullDeclined)

		pr, r, bs, err := parseHook(req)
		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Len(t, bs, 1)
		b := bs[0]
		assert.Equal(t, "anbraten/test-2", r.FullName)
		assert.Equal(t, model.EventPullClosed, b.Event)
		assert.Equal(t, "f90e18fc9d45", b.Commit)
	})

	t.Run("malformed push", func(t *testing.T) {
		buf := bytes.NewBufferString("[]")
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPush)

		_, _, _, err := parseHook(req)
		assert.Error(t, err)
	})

	t.Run("missing commit sha", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPushEmptyHash)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPush)

		_, r, b, err := parseHook(req)
		assert.Nil(t, r)
		assert.Nil(t, b)
		assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
	})

	t.Run("push hook", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPush)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPush)

		pr, r, bs, err := parseHook(req)
		assert.NoError(t, err)
		assert.Nil(t, pr)
		assert.Len(t, bs, 1)
		b := bs[0]
		assert.Equal(t, "martinherren1984/publictestrepo", r.FullName)
		assert.Equal(t, "https://bitbucket.org/martinherren1984/publictestrepo", r.Clone)
		assert.Equal(t, "c14c1bb05dfb1fdcdf06b31485fff61b0ea44277", b.Commit)
		assert.Equal(t, "a\n", b.Message)
	})

	t.Run("push hook with several changes", func(t *testing.T) {
		// One `git push` of two branches, a deleted branch and a tag: Bitbucket sends them as one
		// webhook. Every change with a new commit becomes a pipeline, in the order Bitbucket lists
		// them; the deletion starts none.
		buf := bytes.NewBufferString(fixtures.HookPushMultipleChanges)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPush)

		pr, r, bs, err := parseHook(req)
		assert.NoError(t, err)
		assert.Nil(t, pr)
		assert.Equal(t, "martinherren1984/publictestrepo", r.FullName)
		if assert.Len(t, bs, 3) {
			assert.Equal(t, model.EventPush, bs[0].Event)
			assert.Equal(t, "refs/heads/develop", bs[0].Ref)
			assert.Equal(t, "develop", bs[0].Branch)
			assert.Equal(t, "1111111111111111111111111111111111111111", bs[0].Commit)

			assert.Equal(t, model.EventPush, bs[1].Event)
			assert.Equal(t, "refs/heads/release/1.2", bs[1].Ref)
			assert.Equal(t, "release/1.2", bs[1].Branch)
			assert.Equal(t, "2222222222222222222222222222222222222222", bs[1].Commit)

			assert.Equal(t, model.EventTag, bs[2].Event)
			assert.Equal(t, "refs/tags/v1.0.0", bs[2].Ref)
			assert.Equal(t, "3333333333333333333333333333333333333333", bs[2].Commit)
		}
	})
}
