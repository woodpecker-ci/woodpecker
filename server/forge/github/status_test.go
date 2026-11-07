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

package github

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-github/v93/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// TestStatusErrors: how a failed status request tells the server whether,
// and when, to send it again.
func TestStatusErrors(t *testing.T) {
	reset := time.Now().Add(30 * time.Second)
	tests := []struct {
		name      string
		reply     func(w http.ResponseWriter)
		wantErr   bool
		wantFinal bool
		wantAfter time.Duration // at least
	}{
		{name: "created", reply: func(w http.ResponseWriter) { w.WriteHeader(http.StatusCreated) }},
		{name: "server error", reply: func(w http.ResponseWriter) { w.WriteHeader(http.StatusBadGateway) }, wantErr: true},
		{name: "too many requests", reply: func(w http.ResponseWriter) { w.WriteHeader(http.StatusTooManyRequests) }, wantErr: true},
		{name: "unprocessable", reply: func(w http.ResponseWriter) { w.WriteHeader(http.StatusUnprocessableEntity) }, wantErr: true, wantFinal: true},
		{name: "not found", reply: func(w http.ResponseWriter) { w.WriteHeader(http.StatusNotFound) }, wantErr: true, wantFinal: true},
		{
			name: "rate limit",
			reply: func(w http.ResponseWriter) {
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
			},
			wantErr: true, wantAfter: 20 * time.Second,
		},
		{
			name: "secondary rate limit",
			reply: func(w http.ResponseWriter) {
				w.Header().Set("Retry-After", "42")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"message":"You have exceeded a secondary rate limit","documentation_url":"https://docs.github.com/rest/overview/rate-limits-for-the-rest-api#about-secondary-rate-limits"}`))
			},
			wantErr: true, wantAfter: 42 * time.Second,
		},
		{
			name: "dropped connection",
			reply: func(w http.ResponseWriter) {
				if hijacker, ok := w.(http.Hijacker); ok {
					if conn, _, err := hijacker.Hijack(); err == nil {
						_ = conn.Close()
					}
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				tt.reply(w)
			}))
			t.Cleanup(fake.Close)
			gh, err := github.NewClient(github.WithURLs(new(fake.URL+"/"), nil), github.WithHTTPClient(fake.Client()))
			require.NoError(t, err)
			ctx := context.WithValue(t.Context(), githubClientKey, gh)

			c := &client{}
			err = c.Status(ctx, fakeUser, &model.Repo{Owner: "octocat", Name: "Hello-World"},
				&model.Pipeline{Number: 9, Event: model.EventPull, Commit: "abc123"},
				&model.Workflow{Name: "test", State: model.StatusSuccess})
			if !tt.wantErr {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			var se *types.StatusError
			if !errors.As(err, &se) {
				assert.False(t, tt.wantFinal, "a refusal must say it is final")
				assert.Zero(t, tt.wantAfter, "a rate limit must say how long to wait")
				return
			}
			assert.Equal(t, tt.wantFinal, se.Final, "final")
			assert.GreaterOrEqual(t, se.RetryAfter, tt.wantAfter, "retry after")
		})
	}
}
