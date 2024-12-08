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

package woodpeckergo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/client"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		client  *http.Client
		wantErr bool
	}{
		{
			name:    "Valid URI",
			uri:     "http://example.com",
			client:  http.DefaultClient,
			wantErr: false,
		},
		{
			name:    "Invalid URI",
			uri:     "://invalid-uri",
			client:  http.DefaultClient,
			wantErr: true,
		},
		{
			name:    "Empty URI",
			uri:     "",
			client:  http.DefaultClient,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewWithClient(tt.uri, tt.client)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, client)
		})
	}
}

func TestClient_canCall(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/version" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"version":"test"}`))
		}
	}))
	defer ts.Close()

	serverURL := ts.URL

	tests := []struct {
		name string
		test func(t *testing.T, serverURL string) (*http.Response, error)
	}{
		{
			name: "with raw http.Response",
			test: func(_ *testing.T, serverURL string) (*http.Response, error) {
				hc := http.Client{}
				c, err := NewWithClient(serverURL, &hc)
				if err != nil {
					return nil, err
				}
				return c.GetVersion(context.Background())
			},
		},
		{
			name: "with parsed response body",
			test: func(_ *testing.T, serverURL string) (*http.Response, error) {
				hc := http.Client{}
				c, err := client.NewClientWithResponses(serverURL, client.WithHTTPClient(&hc))
				if err != nil {
					return nil, err
				}
				resp, err := c.GetVersionWithResponse(context.Background())
				return resp.HTTPResponse, err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.test(t, serverURL)
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		})
	}
}
