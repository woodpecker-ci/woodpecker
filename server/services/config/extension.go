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

package config

import (
	"errors"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// configData same as forge.FileMeta but with json tags and string data.
type configData struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// requestStructure is what a configuration extension receives.
type requestStructure struct {
	Repo          *model.Repo     `json:"repo"`
	Pipeline      *model.Pipeline `json:"pipeline"`
	Netrc         *model.Netrc    `json:"netrc"`
	Configuration []*configData   `json:"configuration,omitempty"`
}

// responseStructure is what a configuration extension answers.
type responseStructure struct {
	Configs []*configData `json:"configs"`
}

func toConfigData(files []*types.FileMeta) []*configData {
	configs := make([]*configData, len(files))
	for i, file := range files {
		configs[i] = &configData{Name: file.Name, Data: string(file.Data)}
	}
	return configs
}

// toFileMeta converts the answer of an extension, which is foreign input and can contain anything.
func toFileMeta(configs []*configData) ([]*types.FileMeta, error) {
	files := make([]*types.FileMeta, len(configs))
	for i, config := range configs {
		if config == nil {
			return nil, errors.New("config extension returned an empty config")
		}
		files[i] = &types.FileMeta{Name: config.Name, Data: []byte(config.Data)}
	}
	return files, nil
}
