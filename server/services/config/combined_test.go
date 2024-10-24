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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yaronf/httpsign"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/config"
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
		expectedError     bool
	}{
		{
			name:              "External Fetch empty repo",
			repoConfig:        "",
			files:             []file{},
			expectedFileNames: []string{"override1", "override2", "override3"},
			expectedError:     false,
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
			expectedError:     false,
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
			expectedError:     true,
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
			expectedError: false,
		},
	}

	pubEd25519Key, privEd25519Key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal("can't generate ed25519 key pair")
	}

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
	httpFetcher := config.NewHTTP(ts.URL+"/", privEd25519Key)

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			repo := &model.Repo{Owner: "laszlocph", Name: tt.name, Config: tt.repoConfig} // Using test name as repo name to provide different responses in mock server

			f := new(mocks.Forge)
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
			f.On("File", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("file not found"))
			f.On("Dir", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("directory not found"))

			f.On("Netrc", mock.Anything, mock.Anything).Return(&model.Netrc{Machine: "mock", Login: "mock", Password: "mock"}, nil)

			forgeFetcher := config.NewForge(time.Second*3, 3)
			configFetcher := config.NewCombined(forgeFetcher, httpFetcher)
			files, err := configFetcher.Fetch(
				context.Background(),
				f,
				&model.User{AccessToken: "xxx"},
				repo,
				&model.Pipeline{Commit: "89ab7b2d6bfb347144ac7c557e638ab402848fee"},
				[]*forge_types.FileMeta{},
				false,
			)
			if tt.expectedError && err == nil {
				t.Fatal("expected an error")
			} else if !tt.expectedError && err != nil {
				t.Fatal("error fetching config:", err)
			}

			matchingFiles := make([]string, len(files))
			for i := range files {
				matchingFiles[i] = files[i].Name
			}
			assert.ElementsMatch(t, tt.expectedFileNames, matchingFiles, "expected some other pipeline files")
		})
	}
}
