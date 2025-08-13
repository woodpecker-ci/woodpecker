// Copyright 2025 Woodpecker Authors
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

	t.Run("pull requests", func(t *testing.T) {
		t.Run("malformed pull-request hook", func(t *testing.T) {
			buf := bytes.NewBufferString("[]")
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = http.Header{}
			req.Header.Set(hookEvent, hookPullCreated)

			_, _, err := parseHook(req)
			assert.Error(t, err)
		})

		t.Run("pull-request created", func(t *testing.T) {
			buf := bytes.NewBufferString(fixtures.HookPullRequestCreated)
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = fixtures.HookPullRequestCreatedHeaders

			r, b, err := parseHook(req)
			assert.NoError(t, err)
			if assert.NotNil(t, b) && assert.NotNil(t, r) {
				assert.Equal(t, "6543/collect-webhooks", r.FullName)
				assert.Equal(t, model.EventPull, b.Event)
				assert.Empty(t, b.EventReason)
				assert.Equal(t, "39f188d78e1e", b.Commit)
				assert.Equal(t, "aha", b.Title)
				assert.Equal(t, "", b.Message)
				assert.Equal(t, "6543", b.Author)
			}
		})

		t.Run("pull-request updated", func(t *testing.T) {
			buf := bytes.NewBufferString(fixtures.HookPullPush)
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = fixtures.HookPullPushHeaders

			r, b, err := parseHook(req)
			assert.NoError(t, err)
			if assert.NotNil(t, b) && assert.NotNil(t, r) {
				assert.Equal(t, "6543/collect-webhooks", r.FullName)
				assert.Equal(t, model.EventPull, b.Event)
				assert.Empty(t, b.EventReason)
				assert.Equal(t, "26240d6b7e74", b.Commit)
				assert.Equal(t, "aha", b.Title)
				assert.Equal(t, "some nice ahas", b.Message)
				assert.Equal(t, "6543", b.Author)
			}
		})

		t.Run("pull-request merged", func(t *testing.T) {
			buf := bytes.NewBufferString(fixtures.HookPullMerged)
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = fixtures.HookPullMergedHeaders

			r, b, err := parseHook(req)
			assert.NoError(t, err)
			if assert.NotNil(t, b) && assert.NotNil(t, r) {
				assert.Equal(t, "6543/collect-webhooks", r.FullName)
				assert.Equal(t, "https://bitbucket.org/6543/collect-webhooks/pull-requests/2", b.ForgeURL)
				assert.Equal(t, model.EventPullClosed, b.Event)
				assert.Empty(t, b.EventReason)
				assert.Equal(t, "fc2a2c05765d", b.Commit)
				assert.Equal(t, "aha", b.Title)
				assert.Equal(t, "bha", b.Message)
				assert.Equal(t, "demoaccount2-commits", b.Author)
			}
		})

		t.Run("pull-request rejected", func(t *testing.T) {
			buf := bytes.NewBufferString(fixtures.HookPullRequestRejected)
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = fixtures.HookPullRequestRejectedHeaders

			r, b, err := parseHook(req)
			assert.NoError(t, err)
			if assert.NotNil(t, b) && assert.NotNil(t, r) {
				assert.Equal(t, "6543/collect-webhooks", r.FullName)
				assert.Equal(t, "https://bitbucket.org/6543/collect-webhooks/pull-requests/1", b.ForgeURL)
				assert.Equal(t, model.EventPullClosed, b.Event)
				assert.Empty(t, b.EventReason)
				assert.Equal(t, "d0e829618d28", b.Commit)
				assert.Equal(t, "taerg era senilwen", b.Title)
				assert.Equal(t, "", b.Message)
				assert.Equal(t, "6543", b.Author)
			}
		})

		t.Run("pull-request approved", func(t *testing.T) {
			buf := bytes.NewBufferString(fixtures.HookPullApproved)
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = fixtures.HookPullApprovedHeaders

			r, b, err := parseHook(req)
			assert.NoError(t, err)
			if assert.NotNil(t, b) && assert.NotNil(t, r) {
				assert.Equal(t, "6543/collect-webhooks", r.FullName)
				assert.Equal(t, "https://bitbucket.org/6543/collect-webhooks/pull-requests/2", b.ForgeURL)
				assert.Equal(t, model.EventPullMetadata, b.Event)
				assert.Equal(t, "approved", b.EventReason)
				assert.Equal(t, "26240d6b7e74", b.Commit)
				assert.Equal(t, "aha", b.Title)
				assert.Equal(t, "", b.Message)
				assert.Equal(t, "demoaccount2-commits", b.Author)
			}
		})

		t.Run("pull-request unapproved", func(t *testing.T) {
			buf := bytes.NewBufferString(fixtures.HookPullUnapproved)
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = fixtures.HookPullUnapprovedHeaders

			r, b, err := parseHook(req)
			assert.NoError(t, err)
			if assert.NotNil(t, b) && assert.NotNil(t, r) {
				assert.Equal(t, "6543/collect-webhooks", r.FullName)
				assert.Equal(t, "https://bitbucket.org/6543/collect-webhooks/pull-requests/2", b.ForgeURL)
				assert.Equal(t, model.EventPullMetadata, b.Event)
				assert.Equal(t, "unapproved", b.EventReason)
				assert.Equal(t, "26240d6b7e74", b.Commit)
				assert.Equal(t, "aha", b.Title)
				assert.Equal(t, "", b.Message)
				assert.Equal(t, "demoaccount2-commits", b.Author)
			}
		})

		t.Run("pull-request comment created", func(t *testing.T) {
			buf := bytes.NewBufferString(fixtures.HookPullCommentCreated)
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = fixtures.HookPullCommentCreatedHeaders

			r, b, err := parseHook(req)
			assert.NoError(t, err)
			if assert.NotNil(t, b) && assert.NotNil(t, r) {
				assert.Equal(t, "6543/collect-webhooks", r.FullName)
				assert.Equal(t, "https://bitbucket.org/6543/collect-webhooks/pull-requests/2", b.ForgeURL)
				assert.Equal(t, model.EventPullMetadata, b.Event)
				assert.Equal(t, "comment_created", b.EventReason)
				assert.Equal(t, "26240d6b7e74", b.Commit)
				assert.Equal(t, "aha", b.Title)
				assert.Equal(t, "", b.Message)
				assert.Equal(t, "6543", b.Author)
			}
		})

		t.Run("pull-request changes request created", func(t *testing.T) {
			buf := bytes.NewBufferString(fixtures.HookPullChangesRequestCreated)
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = fixtures.HookPullChangesRequestCreatedHeaders

			r, b, err := parseHook(req)
			assert.NoError(t, err)
			if assert.NotNil(t, b) && assert.NotNil(t, r) {
				assert.Equal(t, "6543/collect-webhooks", r.FullName)
				assert.Equal(t, "https://bitbucket.org/6543/collect-webhooks/pull-requests/3", b.ForgeURL)
				assert.Equal(t, model.EventPullMetadata, b.Event)
				assert.Equal(t, "changes_request_created", b.EventReason)
				assert.Equal(t, "dd1c5b604ee9", b.Commit)
				assert.Equal(t, "hturt eht llet", b.Title)
				assert.Equal(t, "", b.Message)
				assert.Equal(t, "demoaccount2-commits", b.Author)
			}
		})

		t.Run("pull-request changes request removed", func(t *testing.T) {
			buf := bytes.NewBufferString(fixtures.HookPullChangesRequestRemoved)
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = fixtures.HookPullChangesRequestRemovedHeaders

			r, b, err := parseHook(req)
			assert.NoError(t, err)
			if assert.NotNil(t, b) && assert.NotNil(t, r) {
				assert.Equal(t, "6543/collect-webhooks", r.FullName)
				assert.Equal(t, "https://bitbucket.org/6543/collect-webhooks/pull-requests/2", b.ForgeURL)
				assert.Equal(t, model.EventPullMetadata, b.Event)
				assert.Equal(t, "changes_request_removed", b.EventReason)
				assert.Equal(t, "26240d6b7e74", b.Commit)
				assert.Equal(t, "aha", b.Title)
				assert.Equal(t, "", b.Message)
				assert.Equal(t, "demoaccount2-commits", b.Author)
			}
		})

		// TODO: currently BB dont allow us to distinguish
		t.Run("pull-request to draft", func(t *testing.T) {
			buf := bytes.NewBufferString(fixtures.HookPullToDraft)
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = fixtures.HookPullToDraftHeaders

			r, b, err := parseHook(req)
			assert.NoError(t, err)
			if assert.NotNil(t, b) && assert.NotNil(t, r) {
				assert.Equal(t, "6543/collect-webhooks", r.FullName)
				assert.Equal(t, "https://bitbucket.org/6543/collect-webhooks/pull-requests/1", b.ForgeURL)
				assert.Equal(t, model.EventPull, b.Event) // TODO: model.EventPullMetadata
				assert.Equal(t, "", b.EventReason)        // TODO: marked_as_draft
				assert.Equal(t, "d0e829618d28", b.Commit)
				assert.Equal(t, "taerg era senilwen", b.Title)
				assert.Equal(t, "", b.Message)
				assert.Equal(t, "6543", b.Author)
			}
		})

		// TODO: currently BB dont allow us to distinguish
		t.Run("pull-request ready from draft", func(t *testing.T) {
			buf := bytes.NewBufferString(fixtures.HookPullReadyFromDraft)
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = fixtures.HookPullReadyFromDraftHeaders

			r, b, err := parseHook(req)
			assert.NoError(t, err)
			if assert.NotNil(t, b) && assert.NotNil(t, r) {
				assert.Equal(t, "6543/collect-webhooks", r.FullName)
				assert.Equal(t, "https://bitbucket.org/6543/collect-webhooks/pull-requests/1", b.ForgeURL)
				assert.Equal(t, model.EventPull, b.Event) // TODO: model.EventPullMetadata
				assert.Equal(t, "", b.EventReason)        // TODO: marked_as_ready
				assert.Equal(t, "d0e829618d28", b.Commit)
				assert.Equal(t, "taerg era senilwen", b.Title)
				assert.Equal(t, "", b.Message)
				assert.Equal(t, "6543", b.Author)
			}
		})

		// TODO: currently BB dont allow us to distinguish
		t.Run("pull-request review requested", func(t *testing.T) {
			buf := bytes.NewBufferString(fixtures.HookPullReviewRequested)
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = fixtures.HookPullReviewRequestedHeaders

			r, b, err := parseHook(req)
			assert.NoError(t, err)
			if assert.NotNil(t, b) && assert.NotNil(t, r) {
				assert.Equal(t, "6543/collect-webhooks", r.FullName)
				assert.Equal(t, "https://bitbucket.org/6543/collect-webhooks/pull-requests/3", b.ForgeURL)
				assert.Equal(t, model.EventPull, b.Event) // TODO: model.EventPullMetadata
				assert.Equal(t, "", b.EventReason)        // TODO: review_request if we can distinguish
				assert.Equal(t, "dd1c5b604ee9", b.Commit)
				assert.Equal(t, "hturt eht llet", b.Title)
				assert.Equal(t, "", b.Message)
				assert.Equal(t, "6543", b.Author)
			}
		})
	})

	t.Run("push hooks", func(t *testing.T) {
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
			assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
		})

		t.Run("push hook", func(t *testing.T) {
			buf := bytes.NewBufferString(fixtures.HookPush)
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = fixtures.HookPushHeaders

			r, b, err := parseHook(req)
			assert.NoError(t, err)
			assert.Equal(t, "6543/collect-webhooks", r.FullName)
			assert.Equal(t, "859c737a468f8168b257db109295876fd1f5dbd6", b.Commit)
			assert.Equal(t, "b hcus on si ereht\n", b.Message)
			assert.Equal(t, "6543", b.Author)
			assert.Equal(t, model.EventPush, b.Event)
			assert.Empty(t, b.EventReason)
		})
	})
}
