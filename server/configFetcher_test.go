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
		name         string
		repoConfig   string
		repoFallback bool
		fileMocks    []struct {
			file []byte
			err  error
		}
		dirMock struct {
			files []*remote.FileMeta
			err   error
		}
		expectedFileNames []string
		expectedError     bool
	}{
		{
			name:         "Single .woodpecker.yml file",
			repoConfig:   ".woodpecker.yml",
			repoFallback: false,
			fileMocks: []struct {
				file []byte
				err  error
			}{
				{
					file: []byte{},
					err:  nil,
				},
			},
			expectedFileNames: []string{
				".woodpecker.yml",
			},
			expectedError: false,
		},
		{
			name:         "Folder .woodpecker/",
			repoConfig:   ".woodpecker/",
			repoFallback: false,
			dirMock: struct {
				files []*remote.FileMeta
				err   error
			}{
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
				err: nil,
			},
			expectedFileNames: []string{
				".woodpecker/release.yml",
			},
			expectedError: false,
		},
		{
			name:         "Requesting woodpecker-file but using fallback",
			repoConfig:   ".woodpecker.yml",
			repoFallback: true,
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
					err:  nil,
				},
			},
			expectedFileNames: []string{
				".drone.yml",
			},
			expectedError: false,
		},
		{
			name:         "Requesting folder but using fallback",
			repoConfig:   ".woodpecker/",
			repoFallback: true,
			fileMocks: []struct {
				file []byte
				err  error
			}{
				{
					file: []byte{},
					err:  nil,
				},
			},
			dirMock: struct {
				files []*remote.FileMeta
				err   error
			}{
				files: []*remote.FileMeta{},
				err:   errors.New("Dir not found"),
			},
			expectedFileNames: []string{
				".drone.yml",
			},
			expectedError: false,
		},
		{
			name:         "Not found and disabled fallback",
			repoConfig:   ".woodpecker.yml",
			repoFallback: false,
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
			repo := &model.Repo{Owner: "laszlocph", Name: "drone-multipipeline", Config: tt.repoConfig, Fallback: tt.repoFallback}

			r := new(mocks.Remote)
			for _, fileMock := range tt.fileMocks {
				r.On("File", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fileMock.file, fileMock.err).Once()
			}
			r.On("Dir", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.dirMock.files, tt.dirMock.err)

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
