package server_test

import (
	"errors"
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
		files             []*remote.FileMeta
		expectedFileNames []string
		expectedError     bool
	}{
		{
			name:       "No special config with .woodpecker/ folder",
			repoConfig: "",
			files: []*remote.FileMeta{
				{
					Name: ".woodpecker/text.txt",
					Data: []byte{},
				},
				{
					Name: ".woodpecker/release.yml",
					Data: []byte{},
				},
				{
					Name: ".woodpecker/image.png",
					Data: []byte{},
				},
			},
			expectedFileNames: []string{
				".woodpecker/release.yml",
			},
			expectedError: false,
		},
		{
			name:       "No special config with .woodpecker.yml file",
			repoConfig: "",
			files: []*remote.FileMeta{
				{
					Name: ".woodpecker.yml",
					Data: []byte{},
				},
			},
			expectedFileNames: []string{
				".woodpecker.yml",
			},
			expectedError: false,
		},
		{
			name:       "No special config with .drone.yml file",
			repoConfig: "",
			files: []*remote.FileMeta{
				{
					Name: ".drone.yml",
					Data: []byte{},
				},
			},
			expectedFileNames: []string{
				".drone.yml",
			},
			expectedError: false,
		},
		{
			name:              "No special config without files and folder",
			repoConfig:        "",
			files:             []*remote.FileMeta{},
			expectedFileNames: []string{},
			expectedError:     true,
		},
		{
			name:       "Special folder",
			repoConfig: ".my-ci-folder/",
			files: []*remote.FileMeta{
				{
					Name: ".my-ci-folder/test.yml",
					Data: []byte{},
				},
			},
			expectedFileNames: []string{
				".drone.yml",
			},
			expectedError: false,
		},
		{
			name:       "Special subfolder",
			repoConfig: ".my-ci-folder/my-config/",
			files: []*remote.FileMeta{
				{
					Name: ".my-ci-folder/my-config/test.yml",
					Data: []byte{},
				},
			},
			expectedFileNames: []string{
				".drone.yml",
			},
			expectedError: false,
		},
		{
			name:       "Special file",
			repoConfig: ".config.yml",
			files: []*remote.FileMeta{
				{
					Name: ".my-ci-folder/my-config/test.yml",
					Data: []byte{},
				},
			},
			expectedFileNames: []string{
				".drone.yml",
			},
			expectedError: false,
		},
		{
			name:       "Special file inside subfolder",
			repoConfig: ".my-ci-folder/sub-folder/config.yml",
			files: []*remote.FileMeta{
				{
					Name: ".woodpecker.yml",
					Data: []byte{},
				},
				{
					Name: ".woodpecker/test.yml",
					Data: []byte{},
				},
				{
					Name: ".my-ci-folder/sub-folder/test.yml",
					Data: []byte{},
				},
				{
					Name: ".my-ci-folder/sub-folder/config.yml",
					Data: []byte{},
				},
			},
			expectedFileNames: []string{
				".my-ci-folder/sub-folder/config.yml",
			},
			expectedError: false,
		},
		{
			name:       "Special file which does not exists",
			repoConfig: ".config.yml",
			fileMocks: []struct {
				file []byte
				err  error
			}{
				// first call requesting regular woodpecker.yml
				{
					file: nil,
					err:  errors.New("File not found"),
				},
				// fallback file call
				{
					file: []byte{},
					err:  errors.New("File not found"),
				},
			},
			expectedFileNames: []string{},
			expectedError:     true,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			repo := &model.Repo{Owner: "laszlocph", Name: "drone-multipipeline", Config: tt.repoConfig}

			r := new(mocks.Remote)
			for _, fileMock := range tt.files {
				r.On("File", mock.Anything, mock.Anything, mock.Anything, fileMock.Name).Return(fileMock.Data, nil)
				// r.On("Dir", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.dirMock.files, tt.dirMock.err)
			}

			// File(u *model.User, r *model.Repo, b *model.Build, f string) ([]byte, error)
			// Dir(u *model.User, r *model.Repo, b *model.Build, f string) ([]*FileMeta, error)

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
				t.Fatal("expected some other pipeline files", tt.expectedFileNames, files)
			}
		})
	}
}
