// Copyright 2024 Woodpecker Authors
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

package httputil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/version"
)

func TestNewUserAgentRoundTripper(t *testing.T) {
	t.Run("with custom component", func(t *testing.T) {
		rt := NewUserAgentRoundTripper(nil, "test-component")
		assert.NotNil(t, rt)
		assert.NotNil(t, rt.base)
		expectedUA := fmt.Sprintf("Woodpecker/%s (test-component)", version.String())
		assert.Equal(t, expectedUA, rt.userAgent)
	})

	t.Run("without component", func(t *testing.T) {
		rt := NewUserAgentRoundTripper(nil, "")
		assert.NotNil(t, rt)
		expectedUA := fmt.Sprintf("Woodpecker/%s", version.String())
		assert.Equal(t, expectedUA, rt.userAgent)
	})

	t.Run("with custom base transport", func(t *testing.T) {
		customTransport := &http.Transport{}
		rt := NewUserAgentRoundTripper(customTransport, "custom")
		assert.Equal(t, customTransport, rt.base)
	})
}

func TestUserAgentRoundTripper_RoundTrip(t *testing.T) {
	// Create a test server to capture requests
	var capturedUserAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	t.Run("sets user-agent when not present", func(t *testing.T) {
		client := &http.Client{
			Transport: NewUserAgentRoundTripper(nil, "agent"),
		}

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		assert.NoError(t, err)

		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		defer resp.Body.Close()

		expectedUA := fmt.Sprintf("Woodpecker/%s (agent)", version.String())
		assert.Equal(t, expectedUA, capturedUserAgent)
	})

	t.Run("preserves existing user-agent", func(t *testing.T) {
		client := &http.Client{
			Transport: NewUserAgentRoundTripper(nil, "agent"),
		}

		customUA := "CustomUserAgent/1.0"
		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		assert.NoError(t, err)
		req.Header.Set("User-Agent", customUA)

		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, customUA, capturedUserAgent)
	})

	t.Run("does not modify original request", func(t *testing.T) {
		client := &http.Client{
			Transport: NewUserAgentRoundTripper(nil, "test"),
		}

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		assert.NoError(t, err)

		originalUserAgent := req.Header.Get("User-Agent")

		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		defer resp.Body.Close()

		// Original request should remain unchanged
		assert.Equal(t, originalUserAgent, req.Header.Get("User-Agent"))
	})
}

func TestWrapClient(t *testing.T) {
	t.Run("wraps existing client", func(t *testing.T) {
		originalClient := &http.Client{}
		wrappedClient := WrapClient(originalClient, "cli")

		assert.Equal(t, originalClient, wrappedClient)
		assert.IsType(t, &UserAgentRoundTripper{}, wrappedClient.Transport)
	})

	t.Run("creates new client when nil", func(t *testing.T) {
		wrappedClient := WrapClient(nil, "server")

		assert.NotNil(t, wrappedClient)
		assert.IsType(t, &UserAgentRoundTripper{}, wrappedClient.Transport)
	})

	t.Run("preserves existing transport", func(t *testing.T) {
		customTransport := &http.Transport{}
		originalClient := &http.Client{
			Transport: customTransport,
		}

		wrappedClient := WrapClient(originalClient, "test")

		rt, ok := wrappedClient.Transport.(*UserAgentRoundTripper)
		assert.True(t, ok)
		assert.Equal(t, customTransport, rt.base)
	})
}

func TestIntegration_UserAgentInRealRequest(t *testing.T) {
	// Test with a real HTTP server
	var receivedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := WrapClient(nil, "integration-test")

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	assert.NoError(t, err)

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	defer resp.Body.Close()

	userAgent := receivedHeaders.Get("User-Agent")
	assert.NotEmpty(t, userAgent)
	assert.Contains(t, userAgent, "Woodpecker/")
	assert.Contains(t, userAgent, "(integration-test)")
}
