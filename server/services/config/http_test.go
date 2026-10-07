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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/config"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils"
)

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

// Only 200 and 204 are answers, anything else must not let the pipeline run with a config the extension did not approve.
func TestHTTPExtensionFailsClosedOnUnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{"configs": []}`)
	}))
	defer ts.Close()

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	client, err := utils.NewHTTPClient(privateKey, "loopback")
	require.NoError(t, err)

	fromForge := []*forge_types.FileMeta{{Name: ".woodpecker.yaml", Data: []byte("steps: []")}}
	files, err := config.NewHTTP(ts.URL, client, false).Fetch(t.Context(), nil, nil, &model.Repo{}, &model.Pipeline{}, fromForge, false)
	require.ErrorContains(t, err, "unexpected status code 202")
	assert.Nil(t, files)
}
