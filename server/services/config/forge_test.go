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

package config_test

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/config"
	"go.woodpecker-ci.org/woodpecker/v3/shared/constant"
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
		wantErr           bool
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
			wantErr: false,
		},
		{
			name:       "Default config with .yaml - .woodpecker/",
			repoConfig: "",
			files: []file{{
				name: ".woodpecker/text.txt",
				data: dummyData,
			}, {
				name: ".woodpecker/release.yaml",
				data: dummyData,
			}, {
				name: ".woodpecker/image.png",
				data: dummyData,
			}},
			expectedFileNames: []string{
				".woodpecker/release.yaml",
			},
			wantErr: false,
		},
		{
			name:       "Default config with .yaml, .yml mix - .woodpecker/",
			repoConfig: "",
			files: []file{{
				name: ".woodpecker/text.txt",
				data: dummyData,
			}, {
				name: ".woodpecker/release.yaml",
				data: dummyData,
			}, {
				name: ".woodpecker/other.yml",
				data: dummyData,
			}, {
				name: ".woodpecker/image.png",
				data: dummyData,
			}},
			expectedFileNames: []string{
				".woodpecker/release.yaml",
				".woodpecker/other.yml",
			},
			wantErr: false,
		},
		{
			name:       "Default config check .woodpecker.yaml before .woodpecker.yml",
			repoConfig: "",
			files: []file{{
				name: ".woodpecker.yaml",
				data: dummyData,
			}, {
				name: ".woodpecker.yml",
				data: dummyData,
			}},
			expectedFileNames: []string{
				".woodpecker.yaml",
			},
			wantErr: false,
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
			wantErr: false,
		},
		{
			name:       "Use old config on 204 response",
			repoConfig: "",
			files: []file{{
				name: ".woodpecker.yaml",
				data: dummyData,
			}},
			expectedFileNames: []string{
				".woodpecker.yaml",
			},
			wantErr: false,
		},
		{
			name:              "Default config - Empty repo",
			repoConfig:        "",
			files:             []file{},
			expectedFileNames: []string{},
			wantErr:           true,
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
			wantErr: false,
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
			wantErr: false,
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
				name: ".woodpecker.yaml",
				data: dummyData,
			}},
			expectedFileNames: []string{
				".woodpecker.yaml",
			},
			wantErr: false,
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
			wantErr: false,
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
			wantErr: false,
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
			wantErr: false,
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
			wantErr: false,
		},
		{
			name:              "Special config - empty repo",
			repoConfig:        ".config.yml",
			files:             []file{},
			expectedFileNames: []string{},
			wantErr:           true,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			repo := &model.Repo{Owner: "laszlocph", Name: "multipipeline", Config: tt.repoConfig}

			f := new(mocks.MockForge)
			dirs := make(map[string][]*forge_types.FileMeta)
			for _, file := range tt.files {
				f.On("File", mock.Anything, mock.Anything, mock.Anything, mock.Anything, file.name).Once().Return(file.data, nil)
				path := filepath.Dir(file.name)
				if path != "." {
					dirs[path] = append(dirs[path], &forge_types.FileMeta{
						Name: file.name,
						Data: file.data,
					})
				}
			}

			for path, files := range dirs {
				f.On("Dir", mock.Anything, mock.Anything, mock.Anything, mock.Anything, path).Once().Return(files, nil)
			}

			// if the previous mocks do not match return not found errors
			f.On("File", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("file not found"))
			f.On("Dir", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("directory not found"))

			configFetcher := config.NewForge(
				time.Second*3,
				3,
				constant.DefaultConfigOrder,
				[]string{".yaml", ".yml"},
			)
			files, err := configFetcher.Fetch(
				t.Context(),
				f,
				&model.User{AccessToken: "xxx"},
				repo,
				&model.Pipeline{Commit: "89ab7b2d6bfb347144ac7c557e638ab402848fee"},
				nil,
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

// configsForge is a forge that can fetch all config candidates at once.
type configsForge struct {
	*mocks.MockForge
	configs    func(candidates []string) ([][]*forge_types.FileMeta, error)
	candidates []string
}

var _ forge.ConfigsFetcher = (*configsForge)(nil)

func (f *configsForge) Configs(_ context.Context, _ *model.User, _ *model.Repo, _ *model.Pipeline, candidates []string) ([][]*forge_types.FileMeta, error) {
	f.candidates = candidates
	return f.configs(candidates)
}

func TestFetchConfigsAtOnce(t *testing.T) {
	t.Parallel()

	file := func(name, data string) *forge_types.FileMeta {
		return &forge_types.FileMeta{Name: name, Data: []byte(data)}
	}

	tests := []struct {
		name               string
		repoConfig         string
		configs            [][]*forge_types.FileMeta
		expectedCandidates []string
		expectedFileNames  []string
		notFound           bool
	}{
		{
			name: "folder with pipeline files",
			configs: [][]*forge_types.FileMeta{
				{file(".woodpecker/b.yml", "b"), file(".woodpecker/a.yaml", "a"), file(".woodpecker/readme.md", "r")},
				{file(".woodpecker.yaml", "root")},
				nil,
			},
			expectedFileNames: []string{".woodpecker/a.yaml", ".woodpecker/b.yml"},
		},
		{
			name: "folder without pipeline files falls through to the file",
			configs: [][]*forge_types.FileMeta{
				{file(".woodpecker/readme.md", "r")},
				{file(".woodpecker.yaml", "root")},
				nil,
			},
			expectedFileNames: []string{".woodpecker.yaml"},
		},
		{
			name:              ".woodpecker.yaml before .woodpecker.yml",
			configs:           [][]*forge_types.FileMeta{nil, {file(".woodpecker.yaml", "yaml")}, {file(".woodpecker.yml", "yml")}},
			expectedFileNames: []string{".woodpecker.yaml"},
		},
		{
			name:              "empty file is skipped",
			configs:           [][]*forge_types.FileMeta{nil, {file(".woodpecker.yaml", "")}, {file(".woodpecker.yml", "yml")}},
			expectedFileNames: []string{".woodpecker.yml"},
		},
		{
			name:     "nothing found",
			configs:  [][]*forge_types.FileMeta{nil, nil, nil},
			notFound: true,
		},
		{
			name:               "user defined config",
			repoConfig:         ".ci/pipeline.yml",
			configs:            [][]*forge_types.FileMeta{{file(".ci/pipeline.yml", "user")}},
			expectedCandidates: []string{".ci/pipeline.yml"},
			expectedFileNames:  []string{".ci/pipeline.yml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// no File or Dir expectations: the candidates fetched at once must be used
			f := &configsForge{
				MockForge: mocks.NewMockForge(t),
				configs: func([]string) ([][]*forge_types.FileMeta, error) {
					return tt.configs, nil
				},
			}

			files, err := config.NewForge(time.Second*3, 1, constant.DefaultConfigOrder, []string{".yaml", ".yml"}).Fetch(
				t.Context(), f, &model.User{AccessToken: "xxx"},
				&model.Repo{Owner: "laszlocph", Name: "multipipeline", Config: tt.repoConfig},
				&model.Pipeline{Commit: "89ab7b2d6bfb347144ac7c557e638ab402848fee"}, nil, false,
			)

			expectedCandidates := tt.expectedCandidates
			if expectedCandidates == nil {
				expectedCandidates = constant.DefaultConfigOrder
			}
			assert.Equal(t, expectedCandidates, f.candidates)

			if tt.notFound {
				assert.ErrorIs(t, err, &forge_types.ErrConfigNotFound{})
				return
			}
			require.NoError(t, err)
			fileNames := make([]string, len(files))
			for i := range files {
				fileNames[i] = files[i].Name
			}
			assert.Equal(t, tt.expectedFileNames, fileNames)
		})
	}
}

func TestFetchConfigsAtOnceFallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		configs [][]*forge_types.FileMeta
		err     error
	}{
		{name: "error", err: errors.New("graphql request failed with status 502")},
		{name: "result does not match the candidates", configs: [][]*forge_types.FileMeta{nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := &configsForge{
				MockForge: mocks.NewMockForge(t),
				configs: func([]string) ([][]*forge_types.FileMeta, error) {
					return tt.configs, tt.err
				},
			}
			f.On("Dir", mock.Anything, mock.Anything, mock.Anything, mock.Anything, ".woodpecker").Return(nil, &forge_types.ErrConfigNotFound{})
			f.On("File", mock.Anything, mock.Anything, mock.Anything, mock.Anything, ".woodpecker.yaml").Return([]byte("root"), nil)

			files, err := config.NewForge(time.Second*3, 1, constant.DefaultConfigOrder, []string{".yaml", ".yml"}).Fetch(
				t.Context(), f, &model.User{AccessToken: "xxx"},
				&model.Repo{Owner: "laszlocph", Name: "multipipeline"},
				&model.Pipeline{Commit: "89ab7b2d6bfb347144ac7c557e638ab402848fee"}, nil, false,
			)
			require.NoError(t, err)
			require.Len(t, files, 1)
			assert.Equal(t, ".woodpecker.yaml", files[0].Name)
			assert.Equal(t, "root", string(files[0].Data))
		})
	}
}
