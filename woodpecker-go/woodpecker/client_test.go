// Copyright 2022 Woodpecker Authors
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

package woodpecker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogLevel(t *testing.T) {
	logLevel := "warn"
	fixtureHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var ll LogLevel
			require.NoError(t, json.NewDecoder(r.Body).Decode(&ll))
			logLevel = ll.Level
		}

		_, err := fmt.Fprintf(w, `{
			"log-level": "%s"
	}`, logLevel)
		assert.NoError(t, err)
	}

	ts := httptest.NewServer(http.HandlerFunc(fixtureHandler))
	defer ts.Close()

	client := NewClient(ts.URL, http.DefaultClient)

	curLvl, err := client.LogLevel(t.Context())
	assert.NoError(t, err)
	assert.True(t, strings.EqualFold(curLvl.Level, logLevel))

	newLvl, err := client.SetLogLevel(t.Context(), &LogLevel{Level: "trace"})
	assert.NoError(t, err)
	assert.True(t, strings.EqualFold(newLvl.Level, logLevel))
}

func TestRequestCanceledByContext(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		started <- struct{}{}
		// never answer on our own, so only the context can end the request
		<-release
	}))
	defer ts.Close()
	defer close(release)

	client := NewClient(ts.URL, http.DefaultClient)

	tests := []struct {
		name string
		call func(ctx context.Context) error
	}{
		{
			name: "json response",
			call: func(ctx context.Context) error {
				_, err := client.LogLevel(ctx)
				return err
			},
		},
		{
			name: "raw response",
			call: func(ctx context.Context) error {
				_, err := client.PipelineMetadata(ctx, 1, 1)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancelCause(t.Context())
			defer cancel(nil)

			errCh := make(chan error, 1)
			go func() { errCh <- tt.call(ctx) }()

			<-started
			cancel(nil)

			select {
			case err := <-errCh:
				assert.ErrorIs(t, err, context.Canceled)
			case <-time.After(5 * time.Second):
				t.Fatal("request was not canceled together with the context")
			}
		})
	}
}

func TestVersion(t *testing.T) {
	fixtureHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/version", r.URL.Path)
		_, err := fmt.Fprint(w, `{"source":"https://github.com/woodpecker-ci/woodpecker","version":"3.15.0"}`)
		assert.NoError(t, err)
	}

	ts := httptest.NewServer(http.HandlerFunc(fixtureHandler))
	defer ts.Close()

	client := NewClient(ts.URL, http.DefaultClient)

	version, err := client.Version(t.Context())
	require.NoError(t, err)
	assert.Equal(t, &Version{
		Source:  "https://github.com/woodpecker-ci/woodpecker",
		Version: "3.15.0",
	}, version)
}
