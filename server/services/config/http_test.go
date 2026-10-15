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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/config"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils"
)

func newTestHTTPService(t *testing.T, handler http.HandlerFunc) config.Service {
	t.Helper()

	_, privEd25519Key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := utils.NewHTTPClient(privEd25519Key, "loopback")
	require.NoError(t, err)

	return config.NewHTTP(server.URL, client, false)
}

func testFetchInput() (*model.Repo, *model.Pipeline, []*types.FileMeta) {
	repo := &model.Repo{FullName: "test/test"}
	pipeline := &model.Pipeline{Ref: "refs/heads/main"}
	oldConfig := []*types.FileMeta{{Name: ".woodpecker.yaml", Data: []byte("steps: []")}}
	return repo, pipeline, oldConfig
}

func TestFetchOK(t *testing.T) {
	service := newTestHTTPService(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"configs":[{"name":"new.yaml","data":"steps: [{name: a, image: alpine}]"}]}`))
	})

	repo, pipeline, oldConfig := testFetchInput()
	configs, err := service.Fetch(t.Context(), nil, &model.User{}, repo, pipeline, oldConfig, false)
	require.NoError(t, err)
	require.Len(t, configs, 1)
	assert.Equal(t, "new.yaml", configs[0].Name)
}

func TestFetchNoContentKeepsOldConfig(t *testing.T) {
	service := newTestHTTPService(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	repo, pipeline, oldConfig := testFetchInput()
	configs, err := service.Fetch(t.Context(), nil, &model.User{}, repo, pipeline, oldConfig, false)
	require.NoError(t, err)
	assert.Equal(t, oldConfig, configs)
}

func TestFetchUnprocessableEntityReturnsErrConfigExtension(t *testing.T) {
	service := newTestHTTPService(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("branch protection rules missing"))
	})

	repo, pipeline, oldConfig := testFetchInput()
	configs, err := service.Fetch(t.Context(), nil, &model.User{}, repo, pipeline, oldConfig, false)
	assert.Nil(t, configs)

	var extErr *config.ErrConfigExtension
	require.ErrorAs(t, err, &extErr)
	assert.Equal(t, "branch protection rules missing", extErr.Message)
	assert.ErrorIs(t, err, &config.ErrConfigExtension{})
}

func TestFetchBadRequestReturnsGenericError(t *testing.T) {
	service := newTestHTTPService(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad payload"))
	})

	repo, pipeline, oldConfig := testFetchInput()
	configs, err := service.Fetch(t.Context(), nil, &model.User{}, repo, pipeline, oldConfig, false)
	assert.Nil(t, configs)
	require.Error(t, err)
	assert.NotErrorIs(t, err, &config.ErrConfigExtension{})
	assert.Contains(t, err.Error(), "bad payload")
}

func TestErrConfigExtension(t *testing.T) {
	err := &config.ErrConfigExtension{Message: "some message"}
	assert.Equal(t, "config extension error: some message", err.Error())
	assert.ErrorIs(t, err, &config.ErrConfigExtension{})
	assert.NotErrorIs(t, errors.New("other"), &config.ErrConfigExtension{})
}
