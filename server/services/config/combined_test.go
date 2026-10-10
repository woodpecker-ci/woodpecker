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

package config_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yaronf/httpsign"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/config"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils"
	"go.woodpecker-ci.org/woodpecker/v3/shared/constant"
)

func TestFetchFromConfigService(t *testing.T) {
	t.Parallel()

	type file struct {
		name string
		data []byte
	}

	dummyData := []byte("TEST")

	testTable := []struct {
		name              string
		repoConfig        string
		files             []file
		expectedFileNames []string
		wantErr           bool
	}{
		{
			name:              "External Fetch empty repo",
			repoConfig:        "",
			files:             []file{},
			expectedFileNames: []string{"override1", "override2", "override3"},
			wantErr:           false,
		},
		{
			name:       "Default config - Additional sub-folders",
			repoConfig: "",
			files: []file{{
				name: ".woodpecker/test.yml",
				data: dummyData,
			}, {
				name: ".woodpecker/sub-folder/config.yml",
				data: dummyData,
			}},
			expectedFileNames: []string{"override1", "override2", "override3"},
			wantErr:           false,
		},
		{
			name:       "Fetch empty",
			repoConfig: " ",
			files: []file{{
				name: ".woodpecker/.keep",
				data: dummyData,
			}, {
				name: ".woodpecker.yml",
				data: nil,
			}, {
				name: ".woodpecker.yaml",
				data: dummyData,
			}},
			expectedFileNames: []string{},
			wantErr:           true,
		},
		{
			name:       "Use old config",
			repoConfig: ".my-ci-folder/",
			files: []file{{
				name: ".woodpecker/test.yml",
				data: dummyData,
			}, {
				name: ".woodpecker.yml",
				data: dummyData,
			}, {
				name: ".woodpecker.yaml",
				data: dummyData,
			}, {
				name: ".my-ci-folder/test.yml",
				data: dummyData,
			}},
			expectedFileNames: []string{
				".my-ci-folder/test.yml",
			},
			wantErr: false,
		},
	}

	pubEd25519Key, privEd25519Key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err, "can't generate ed25519 key pair")

	fixtureHandler := func(w http.ResponseWriter, r *http.Request) {
		// check signature
		pubKeyID := "woodpecker-ci-extensions"

		verifier, err := httpsign.NewEd25519Verifier(pubEd25519Key,
			httpsign.NewVerifyConfig(),
			httpsign.Headers("@request-target", "content-digest")) // The Content-Digest header will be auto-generated
		assert.NoError(t, err)

		err = httpsign.VerifyRequest(pubKeyID, *verifier, r)
		if err != nil {
			http.Error(w, "Invalid signature", http.StatusBadRequest)
			return
		}

		type config struct {
			Name string `json:"name"`
			Data string `json:"data"`
		}

		type incoming struct {
			Repo          *model.Repo     `json:"repo"`
			Build         *model.Pipeline `json:"pipeline"`
			Configuration []*config       `json:"config"`
		}

		var req incoming
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(body, &req)
		if err != nil {
			http.Error(w, "Failed to parse JSON"+err.Error(), http.StatusBadRequest)
			return
		}

		if req.Repo.Name == "Fetch empty" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if req.Repo.Name == "Use old config" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		fmt.Fprint(w, `{
			"configs": [
					{
							"name": "override1",
							"data": "some new pipelineconfig \n pipe, pipe, pipe"
					},
					{
							"name": "override2",
							"data": "some new pipelineconfig \n pipe, pipe, pipe"
					},
					{
							"name": "override3",
							"data": "some new pipelineconfig \n pipe, pipe, pipe"
					}
			]
}`)
	}

	ts := httptest.NewServer(http.HandlerFunc(fixtureHandler))
	defer ts.Close()

	client, err := utils.NewHTTPClient(privEd25519Key, "loopback")
	require.NoError(t, err)

	httpFetcher := config.NewHTTP(ts.URL+"/", client, true)

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			repo := &model.Repo{Owner: "laszlocph", Name: tt.name, Config: tt.repoConfig} // Using test name as repo name to provide different responses in mock server

			f := new(mocks.MockForge)
			dirs := map[string][]*forge_types.FileMeta{}
			for _, file := range tt.files {
				f.On("File", mock.Anything, mock.Anything, mock.Anything, mock.Anything, file.name).Return(file.data, nil)
				path := filepath.Dir(file.name)
				if path != "." {
					dirs[path] = append(dirs[path], &forge_types.FileMeta{
						Name: file.name,
						Data: file.data,
					})
				}
			}

			for path, files := range dirs {
				f.On("Dir", mock.Anything, mock.Anything, mock.Anything, mock.Anything, path).Return(files, nil)
			}

			// if the previous mocks do not match return not found errors
			f.On("File", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, &forge_types.ErrConfigNotFound{})
			f.On("Dir", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, &forge_types.ErrConfigNotFound{})

			f.On("Netrc", mock.Anything, mock.Anything).Return(&model.Netrc{Machine: "mock", Login: "mock", Password: "mock"}, nil)

			forgeFetcher := config.NewForge(time.Second*3, 3, constant.DefaultConfigOrder, []string{".yaml", ".yml"})
			configFetcher := config.NewCombined(forgeFetcher, httpFetcher)
			files, err := configFetcher.Fetch(
				t.Context(),
				f,
				&model.User{AccessToken: "xxx"},
				repo,
				&model.Pipeline{Commit: "89ab7b2d6bfb347144ac7c557e638ab402848fee"},
				[]*forge_types.FileMeta{},
				false,
			)
			if tt.wantErr {
				require.Error(t, err, "expected an error")
			} else {
				require.NoError(t, err, "error fetching config")
			}

			matchingFiles := make([]string, len(files))
			for i := range files {
				matchingFiles[i] = files[i].Name
			}
			assert.ElementsMatch(t, tt.expectedFileNames, matchingFiles, "expected some other pipeline files")
		})
	}
}

// fetchFunc is a config service that answers with whatever the test needs.
type fetchFunc func(old []*forge_types.FileMeta) ([]*forge_types.FileMeta, error)

func (f fetchFunc) Fetch(_ context.Context, _ forge.Forge, _ *model.User, _ *model.Repo, _ *model.Pipeline, old []*forge_types.FileMeta, _ bool) ([]*forge_types.FileMeta, error) {
	return f(old)
}

func TestCombinedErrorHandling(t *testing.T) {
	central := []*forge_types.FileMeta{{Name: "central", Data: []byte("steps: []")}}
	notFound := fetchFunc(func([]*forge_types.FileMeta) ([]*forge_types.FileMeta, error) {
		return nil, &forge_types.ErrConfigNotFound{Configs: []string{".woodpecker.yaml"}}
	})
	keep := fetchFunc(func(old []*forge_types.FileMeta) ([]*forge_types.FileMeta, error) { return old, nil })
	fetch := func(services ...config.Service) ([]*forge_types.FileMeta, error) {
		return config.NewCombined(services...).Fetch(t.Context(), nil, nil, &model.Repo{}, &model.Pipeline{}, nil, false)
	}

	t.Run("a failure is not covered up by the next service", func(t *testing.T) {
		forgeDown := errors.New("forge is down")
		asked := false

		files, err := fetch(
			fetchFunc(func([]*forge_types.FileMeta) ([]*forge_types.FileMeta, error) { return nil, forgeDown }),
			fetchFunc(func(old []*forge_types.FileMeta) ([]*forge_types.FileMeta, error) {
				asked = true
				return old, nil
			}),
		)
		require.ErrorIs(t, err, forgeDown)
		assert.Nil(t, files)
		assert.False(t, asked, "the next service must not be asked to continue after a failure")
	})

	t.Run("a repo without config can get one from the next service", func(t *testing.T) {
		files, err := fetch(notFound, fetchFunc(func([]*forge_types.FileMeta) ([]*forge_types.FileMeta, error) { return central, nil }))
		require.NoError(t, err)
		assert.Equal(t, central, files)
	})

	t.Run("a repo without config stays without if no service provides one", func(t *testing.T) {
		files, err := fetch(notFound, keep)
		require.ErrorIs(t, err, &forge_types.ErrConfigNotFound{})
		assert.Empty(t, files)
	})
}
