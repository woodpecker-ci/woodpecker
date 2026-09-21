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
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestRefreshExchangesStillValidToken(t *testing.T) {
	var tokenRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
			http.NotFound(w, r)
			return
		}
		tokenRequests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"new-access","token_type":"bearer","expires_in":3600,"refresh_token":"new-refresh"}`)
	}))
	defer server.Close()

	forge, err := New(1, Opts{URL: server.URL, OAuthClientID: "id", OAuthClientSecret: "secret"})
	assert.NoError(t, err)

	user := &model.User{
		AccessToken:  "old-access",
		RefreshToken: "old-refresh",
		Expiry:       time.Now().Add(5 * time.Minute).UTC().Unix(),
	}

	updated, err := forge.(*GitLab).Refresh(t.Context(), user)
	assert.NoError(t, err)
	assert.True(t, updated)

	assert.Equal(t, int32(1), tokenRequests.Load(), "token endpoint must be called exactly once")
	assert.Equal(t, "new-access", user.AccessToken)
	assert.Equal(t, "new-refresh", user.RefreshToken)
	assert.WithinDuration(t, time.Now().Add(time.Hour), time.Unix(user.Expiry, 0), 2*time.Minute)
}
