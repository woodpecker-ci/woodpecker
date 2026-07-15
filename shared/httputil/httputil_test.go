// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httputil

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	require.NoError(t, err)
	return u
}

// readCookie extracts a single named cookie from the recorder's response.
func readCookie(t *testing.T, w *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	require.FailNowf(t, "cookie not found", "no cookie named %q in response", name)
	return nil
}

func TestIsHTTPS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  *http.Request
		want bool
	}{
		{
			name: "url scheme https",
			req:  &http.Request{URL: mustParseURL(t, "https://example.com")},
			want: true,
		},
		{
			name: "tls connection state set",
			req:  &http.Request{URL: mustParseURL(t, "http://example.com"), TLS: &tls.ConnectionState{}},
			want: true,
		},
		{
			name: "proto prefix HTTPS",
			req:  &http.Request{URL: mustParseURL(t, "http://example.com"), Proto: "HTTPS/1.1"},
			want: true,
		},
		{
			name: "x-forwarded-proto https",
			req: &http.Request{
				URL:    mustParseURL(t, "http://example.com"),
				Header: http.Header{"X-Forwarded-Proto": []string{"https"}},
			},
			want: true,
		},
		{
			name: "plain http",
			req:  &http.Request{URL: mustParseURL(t, "http://example.com"), Proto: "HTTP/1.1"},
			want: false,
		},
		{
			name: "x-forwarded-proto http",
			req: &http.Request{
				URL:    mustParseURL(t, "http://example.com"),
				Header: http.Header{"X-Forwarded-Proto": []string{"http"}},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, IsHTTPS(tt.req))
		})
	}
}

func TestSetCookie(t *testing.T) {
	t.Parallel()

	t.Run("secure flag follows request scheme", func(t *testing.T) {
		t.Parallel()

		t.Run("https request sets secure cookie", func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			r := &http.Request{URL: mustParseURL(t, "https://example.com")}

			SetCookie(w, r, "token", "value")

			c := readCookie(t, w, "token")
			assert.Equal(t, "value", c.Value)
			assert.Equal(t, "/", c.Path)
			assert.Equal(t, "example.com", c.Domain)
			assert.True(t, c.HttpOnly)
			assert.True(t, c.Secure)
		})

		t.Run("http request sets insecure cookie", func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			r := &http.Request{URL: mustParseURL(t, "http://example.com")}

			SetCookie(w, r, "token", "value")

			c := readCookie(t, w, "token")
			assert.False(t, c.Secure)
			assert.True(t, c.HttpOnly)
		})
	})
}

func TestDelCookie(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := &http.Request{URL: mustParseURL(t, "https://example.com")}

	DelCookie(w, r, "token")

	c := readCookie(t, w, "token")
	assert.Equal(t, "deleted", c.Value)
	assert.Equal(t, "/", c.Path)
	assert.Equal(t, "example.com", c.Domain)
	// negative MaxAge expires the cookie immediately
	assert.Equal(t, -1, c.MaxAge)
}
