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

package gitlab

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// zeroTimeUnix is time.Time{}.Unix(), the expiry Login stores when the
// token response carries no expires_in.
const zeroTimeUnix = -62135596800

func newRefreshServer(t *testing.T, body string) (*GitLab, *atomic.Int32, *atomic.Pointer[url.Values]) {
	t.Helper()

	var hits atomic.Int32
	var form atomic.Pointer[url.Values]
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
			http.NotFound(w, r)
			return
		}
		hits.Add(1)
		_ = r.ParseForm()
		form.Store(&r.PostForm)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return load(server.URL + "?client_id=test&client_secret=test"), &hits, &form
}

func TestRefreshExchangesRefreshToken(t *testing.T) {
	for name, expiry := range map[string]int64{
		"zero time expiry": zeroTimeUnix,
		"unset expiry":     0,
		"expired expiry":   time.Now().Add(-time.Hour).Unix(),
		"expiring soon":    time.Now().Add(20 * time.Minute).Unix(),
	} {
		t.Run(name, func(t *testing.T) {
			g, hits, form := newRefreshServer(t,
				`{"access_token":"new-access","refresh_token":"new-refresh","token_type":"bearer","expires_in":7200}`)
			user := &model.User{AccessToken: "old-access", RefreshToken: "old-refresh", Expiry: expiry}

			updated, err := g.Refresh(t.Context(), user)
			require.NoError(t, err)

			require.Equal(t, int32(1), hits.Load(), "token endpoint must be called")
			assert.True(t, updated)
			assert.Equal(t, "refresh_token", form.Load().Get("grant_type"))
			assert.Equal(t, "old-refresh", form.Load().Get("refresh_token"))
			assert.Equal(t, "new-access", user.AccessToken)
			assert.Equal(t, "new-refresh", user.RefreshToken)
			assert.Greater(t, user.Expiry, time.Now().Add(time.Hour).Unix())
		})
	}
}
