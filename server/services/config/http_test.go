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

package config_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/config"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils"
)

func TestHTTPFetchSendsRestart(t *testing.T) {
	_, privEd25519Key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	var got struct {
		Restart *bool `json:"restart"`
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.Restart = nil
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client, err := utils.NewHTTPClient(privEd25519Key, "loopback")
	require.NoError(t, err)
	fetcher := config.NewHTTP(ts.URL, client, false)

	for _, restart := range []bool{false, true} {
		_, err := fetcher.Fetch(t.Context(), nil, &model.User{}, &model.Repo{FullName: "owner/repo"}, &model.Pipeline{Number: 1}, nil, restart)
		require.NoError(t, err)
		require.NotNil(t, got.Restart, "request body has no restart field")
		assert.Equal(t, restart, *got.Restart)
	}
}

// An extension is a foreign service, whatever it answers must not be able to crash the server.
func TestHTTPExtensionRejectsNullConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"configs": [null]}`)
	}))
	defer ts.Close()

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	client, err := utils.NewHTTPClient(privateKey, "loopback")
	require.NoError(t, err)

	files, err := config.NewHTTP(ts.URL, client, false).Fetch(t.Context(), nil, nil, &model.Repo{}, &model.Pipeline{}, nil, false)
	require.ErrorContains(t, err, "empty config")
	assert.Nil(t, files)
}

// Failures of the extension must not look like forge errors in the UI.
func TestHTTPExtensionErrorType(t *testing.T) {
	// the client does not trust the test cert, like an endpoint with a broken TLS setup
	ts := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer ts.Close()

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	client, err := utils.NewHTTPClient(privateKey, "loopback")
	require.NoError(t, err)

	_, err = config.NewHTTP(ts.URL, client, false).Fetch(t.Context(), nil, nil, &model.Repo{}, &model.Pipeline{}, nil, false)
	pipelineErr, ok := errors.AsType[*pipeline_errors.PipelineError](err)
	require.True(t, ok, "expected a pipeline error, got: %v", err)
	assert.Equal(t, pipeline_errors.PipelineErrorTypeExtension, pipelineErr.Type)
	assert.True(t, strings.HasPrefix(pipelineErr.Message, "config extension: could not reach endpoint: "), pipelineErr.Message)
}
