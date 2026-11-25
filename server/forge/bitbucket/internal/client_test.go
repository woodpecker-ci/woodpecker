// Copyright 2026 Woodpecker Authors
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

package internal

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient returns a client for a fake Bitbucket API that answers with pages[page-1].
// Asking for a page that does not exist fails the request, so a loop that keeps
// paging fails the test instead of hanging it.
func newTestClient(t *testing.T, pages []string, requests *int) *Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*requests++
		page := r.URL.Query().Get("page")
		for i, payload := range pages {
			if page == strconv.Itoa(i+1) {
				_, _ = w.Write([]byte(payload))
				return
			}
		}
		http.Error(w, "no page "+page, http.StatusNotFound)
	}))
	t.Cleanup(server.Close)
	return NewClient(t.Context(), server.URL, http.DefaultClient)
}

func TestGetUserWorkspaceMembership(t *testing.T) {
	// the last page of a Bitbucket collection has no "next" link
	pages := []string{
		`{"values":[{"permission":"owner","user":{"nickname":"alice"}}],"next":"https://api.bitbucket.org/2.0/workspaces/ws/permissions?page=2"}`,
		`{"values":[{"permission":"member","user":{"nickname":"bob"}}]}`,
	}

	t.Run("finds the user on a later page", func(t *testing.T) {
		requests := 0
		perm, err := newTestClient(t, pages, &requests).GetUserWorkspaceMembership("ws", "bob")
		require.NoError(t, err)
		assert.Equal(t, "member", perm)
		assert.Equal(t, 2, requests)
	})

	t.Run("stops at the last page if the user is no member", func(t *testing.T) {
		requests := 0
		perm, err := newTestClient(t, pages, &requests).GetUserWorkspaceMembership("ws", "carol")
		require.NoError(t, err)
		assert.Empty(t, perm)
		assert.Equal(t, 2, requests)
	})
}
