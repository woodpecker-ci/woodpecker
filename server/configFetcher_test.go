package server_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/woodpecker-ci/woodpecker/model"
	"github.com/woodpecker-ci/woodpecker/remote"
	"github.com/woodpecker-ci/woodpecker/remote/mocks"
	"github.com/woodpecker-ci/woodpecker/server"
)

func TestFetch(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name              string
		repoConfig        string
		files             []string
		expectedFileNames []string
		expectedError     bool
	}{
		{
			name:       "Default config - .woodpecker/",
			repoConfig: "",
			files: []string{
				".woodpecker/text.txt",
				".woodpecker/release.yml",
				".woodpecker/image.png",
			},
			expectedFileNames: []string{
				".woodpecker/release.yml",
			},
			expectedError: false,
		},
		{
			name:       "Default config - .woodpecker.yml",
			repoConfig: "",
			files: []string{
				".woodpecker.yml",
			},
			expectedFileNames: []string{
				".woodpecker.yml",
			},
			expectedError: false,
		},
		{
			name:       "Default config - .drone.yml",
			repoConfig: "",
			files: []string{
				".drone.yml",
			},
			expectedFileNames: []string{
				".drone.yml",
			},
			expectedError: false,
		},
		{
			name:              "Default config - Empty repo",
			repoConfig:        "",
			files:             []string{},
			expectedFileNames: []string{},
			expectedError:     true,
		},
		{
			name:       "Default config - Additional sub-folders",
			repoConfig: "",
			files: []string{
				".woodpecker/test.yml",
				".woodpecker/sub-folder/config.yml",
			},
			expectedFileNames: []string{
				".woodpecker/test.yml",
			},
			expectedError: false,
		},
		{
			name:       "Default config - Additional none .yml files",
			repoConfig: "",
			files: []string{
				".woodpecker/notes.txt",
				".woodpecker/image.png",
				".woodpecker/test.yml",
			},
			expectedFileNames: []string{
				".woodpecker/test.yml",
			},
			expectedError: false,
		},
		{
			name:       "Special config - folder (ignoring default files)",
			repoConfig: ".my-ci-folder/",
			files: []string{
				".woodpecker/test.yml",
				".woodpecker.yml",
				".drone.yml",
				".my-ci-folder/test.yml",
			},
			expectedFileNames: []string{
				".my-ci-folder/test.yml",
			},
			expectedError: false,
		},
		{
			name:       "Special config - folder",
			repoConfig: ".my-ci-folder/",
			files: []string{
				".my-ci-folder/test.yml",
			},
			expectedFileNames: []string{
				".my-ci-folder/test.yml",
			},
			expectedError: false,
		},
		{
			name:       "Special config - subfolder",
			repoConfig: ".my-ci-folder/my-config/",
			files: []string{
				".my-ci-folder/my-config/test.yml",
			},
			expectedFileNames: []string{
				".my-ci-folder/my-config/test.yml",
			},
			expectedError: false,
		},
		{
			name:       "Special config - file",
			repoConfig: ".config.yml",
			files: []string{
				".config.yml",
			},
			expectedFileNames: []string{
				".config.yml",
			},
			expectedError: false,
		},
		{
			name:       "Special config - file inside subfolder",
			repoConfig: ".my-ci-folder/sub-folder/config.yml",
			files: []string{
				".my-ci-folder/sub-folder/config.yml",
			},
			expectedFileNames: []string{
				".my-ci-folder/sub-folder/config.yml",
			},
			expectedError: false,
		},
		{
			name:              "Special config - empty repo",
			repoConfig:        ".config.yml",
			files:             []string{},
			expectedFileNames: []string{},
			expectedError:     true,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			repo := &model.Repo{Owner: "laszlocph", Name: "drone-multipipeline", Config: tt.repoConfig}

			r := new(mocks.Remote)
			dirs := map[string][]*remote.FileMeta{}
			for _, file := range tt.files {
				r.On("File", mock.Anything, mock.Anything, mock.Anything, file).Return([]byte{}, nil)
				path := filepath.Dir(file)
				dirs[path] = append(dirs[path], &remote.FileMeta{
					Name: file,
					Data: []byte{},
				})
			}

			for path, files := range dirs {
				r.On("Dir", mock.Anything, mock.Anything, mock.Anything, path).Return(files, nil)
			}

			// if the previous mocks do not match return not found errors
			r.On("File", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("File not found"))
			r.On("Dir", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Directory not found"))

			configFetcher := server.NewConfigFetcher(
				r,
				&model.User{Token: "xxx"},
				repo,
				&model.Build{Commit: "89ab7b2d6bfb347144ac7c557e638ab402848fee"},
			)
			files, err := configFetcher.Fetch()
			if tt.expectedError && err == nil {
				t.Fatal("expected an error")
			} else if !tt.expectedError && err != nil {
				t.Fatal("error fetching config:", err)
			}

			matchingFiles := 0
			for _, expectedFileName := range tt.expectedFileNames {
				for _, file := range files {
					if file.Name == expectedFileName {
						matchingFiles += 1
					}
				}
			}

			if matchingFiles != len(tt.expectedFileNames) {
				receivedFileNames := []string{}
				for _, file := range files {
					receivedFileNames = append(receivedFileNames, file.Name)
				}
				t.Fatal("expected some other pipeline files", tt.expectedFileNames, receivedFileNames)
			}
		})
	}
}
