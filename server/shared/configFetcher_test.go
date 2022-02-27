package shared_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/99designs/httpsignatures-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/plugins/configuration"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/remote/mocks"
	"github.com/woodpecker-ci/woodpecker/server/shared"
)

func TestFetch(t *testing.T) {
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
			name:       "Default config - .woodpecker/",
			repoConfig: "",
			files: []file{{
				name: ".woodpecker/text.txt",
				data: dummyData,
			}, {
				name: ".woodpecker/release.yml",
				data: dummyData,
			}, {
				name: ".woodpecker/image.png",
				data: dummyData,
			}},
			expectedFileNames: []string{
				".woodpecker/release.yml",
			},
			expectedError: false,
		},
		{
			name:       "Override via API with custom config",
			repoConfig: "",
			files: []file{{
				name: ".woodpecker.yml",
				data: dummyData,
			}},
			expectedFileNames: []string{
				".woodpecker.yml",
			},
			expectedError: false,
		},
		{
			name:       "Use old config on 204 response",
			repoConfig: "",
			files: []file{{
				name: ".drone.yml",
				data: dummyData,
			}},
			expectedFileNames: []string{
				".drone.yml",
			},
			expectedError: false,
		},
		{
			name:              "Default config - Empty repo",
			repoConfig:        "",
			files:             []file{},
			expectedFileNames: []string{},
			expectedError:     true,
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
			expectedFileNames: []string{
				".woodpecker/test.yml",
			},
			expectedError: false,
		},
		{
			name:       "Default config - Additional none .yml files",
			repoConfig: "",
			files: []file{{
				name: ".woodpecker/notes.txt",
				data: dummyData,
			}, {
				name: ".woodpecker/image.png",
				data: dummyData,
			}, {
				name: ".woodpecker/test.yml",
				data: dummyData,
			}},
			expectedFileNames: []string{
				".woodpecker/test.yml",
			},
			expectedError: false,
		},
		{
			name:       "Default config - Empty Folder",
			repoConfig: " ",
			files: []file{{
				name: ".woodpecker/.keep",
				data: dummyData,
			}, {
				name: ".woodpecker.yml",
				data: nil,
			}, {
				name: ".drone.yml",
				data: dummyData,
			}},
			expectedFileNames: []string{
				".drone.yml",
			},
			expectedError: false,
		},
		{
			name:       "Special config - folder (ignoring default files)",
			repoConfig: ".my-ci-folder/",
			files: []file{{
				name: ".woodpecker/test.yml",
				data: dummyData,
			}, {
				name: ".woodpecker.yml",
				data: dummyData,
			}, {
				name: ".drone.yml",
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
		{
			name:       "Special config - folder",
			repoConfig: ".my-ci-folder/",
			files: []file{{
				name: ".my-ci-folder/test.yml",
				data: dummyData,
			}},
			expectedFileNames: []string{
				".my-ci-folder/test.yml",
			},
			expectedError: false,
		},
		{
			name:       "Special config - subfolder",
			repoConfig: ".my-ci-folder/my-config/",
			files: []file{{
				name: ".my-ci-folder/my-config/test.yml",
				data: dummyData,
			}},
			expectedFileNames: []string{
				".my-ci-folder/my-config/test.yml",
			},
			expectedError: false,
		},
		{
			name:       "Special config - file",
			repoConfig: ".config.yml",
			files: []file{{
				name: ".config.yml",
				data: dummyData,
			}},
			expectedFileNames: []string{
				".config.yml",
			},
			expectedError: false,
		},
		{
			name:       "Special config - file inside subfolder",
			repoConfig: ".my-ci-folder/sub-folder/config.yml",
			files: []file{{
				name: ".my-ci-folder/sub-folder/config.yml",
				data: dummyData,
			}},
			expectedFileNames: []string{
				".my-ci-folder/sub-folder/config.yml",
			},
			expectedError: false,
		},
		{
			name:              "Special config - empty repo",
			repoConfig:        ".config.yml",
			files:             []file{},
			expectedFileNames: []string{},
			expectedError:     true,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			repo := &model.Repo{Owner: "laszlocph", Name: "multipipeline", Config: tt.repoConfig}

			r := new(mocks.Remote)
			dirs := map[string][]*remote.FileMeta{}
			for _, file := range tt.files {
				r.On("File", mock.Anything, mock.Anything, mock.Anything, mock.Anything, file.name).Return(file.data, nil)
				path := filepath.Dir(file.name)
				if path != "." {
					dirs[path] = append(dirs[path], &remote.FileMeta{
						Name: file.name,
						Data: file.data,
					})
				}
			}

			for path, files := range dirs {
				r.On("Dir", mock.Anything, mock.Anything, mock.Anything, mock.Anything, path).Return(files, nil)
			}

			// if the previous mocks do not match return not found errors
			r.On("File", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("File not found"))
			r.On("Dir", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Directory not found"))

			configFetcher := shared.NewConfigFetcher(
				r,
				configuration.NewAPI("", ""),
				&model.User{Token: "xxx"},
				repo,
				&model.Build{Commit: "89ab7b2d6bfb347144ac7c557e638ab402848fee"},
			)
			files, err := configFetcher.Fetch(context.Background())
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

func TestFetchFromConfigurationAPI(t *testing.T) {
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
				name: ".drone.yml",
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
				name: ".drone.yml",
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

	httpSigSecret := "wykf9frJbGXwSHcJ7AQF4tlfXUo0Tkixh57WPEXMyWVgkxIsAarYa2Hb8UTwPpbqO0N3NueKwjv4DVhPgvQjGur3LuCbiGHbBoaL1X5gZ9oyxD2lBHndoNxifDyNH7tNPw3Lh5lX2MSrWP1yuqHp8Sgm7fX8pLTjaKKFgFIKlODd"

	fixtureHandler := func(w http.ResponseWriter, r *http.Request) {
		// check signature
		signature, err := httpsignatures.FromRequest(r)
		if err != nil {
			http.Error(w, "Invalid or Missing Signature", http.StatusBadRequest)
			return
		}
		if !signature.IsValid(httpSigSecret, r) {
			http.Error(w, "Invalid Signature", http.StatusBadRequest)
			return
		}

		type config struct {
			Name string `json:"name"`
			Data string `json:"data"`
		}

		type incoming struct {
			Repo          *model.Repo  `json:"repo"`
			Build         *model.Build `json:"build"`
			Configuration []*config    `json:"config"`
		}

		var req incoming
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(body, &req)
		if err != nil {
			http.Error(w, "Failed to parse JSON"+err.Error(), http.StatusBadRequest)
			return
		}

		if req.Repo.Name == "Fetch empty" {
			w.WriteHeader(404)
			return
		}

		if req.Repo.Name == "Use old config" {
			w.WriteHeader(204)
			return
		}

		fmt.Fprint(w, `{
			"pipelines": [
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
	configAPI := configuration.NewAPI(ts.URL, httpSigSecret)

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			repo := &model.Repo{Owner: "laszlocph", Name: tt.name, Config: tt.repoConfig} // Using test name as repo name to provide different responses in mock server

			r := new(mocks.Remote)
			dirs := map[string][]*remote.FileMeta{}
			for _, file := range tt.files {
				r.On("File", mock.Anything, mock.Anything, mock.Anything, mock.Anything, file.name).Return(file.data, nil)
				path := filepath.Dir(file.name)
				if path != "." {
					dirs[path] = append(dirs[path], &remote.FileMeta{
						Name: file.name,
						Data: file.data,
					})
				}
			}

			for path, files := range dirs {
				r.On("Dir", mock.Anything, mock.Anything, mock.Anything, mock.Anything, path).Return(files, nil)
			}

			// if the previous mocks do not match return not found errors
			r.On("File", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("File not found"))
			r.On("Dir", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Directory not found"))

			configFetcher := shared.NewConfigFetcher(
				r,
				configAPI,
				&model.User{Token: "xxx"},
				repo,
				&model.Build{Commit: "89ab7b2d6bfb347144ac7c557e638ab402848fee"},
			)
			files, err := configFetcher.Fetch(context.Background())
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
