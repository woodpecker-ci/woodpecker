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

package config

import (
	"context"
	"crypto"
	"fmt"

	forge_types "github.com/woodpecker-ci/woodpecker/server/forge/types"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/plugins/utils"
)

type http struct {
	endpoint   string
	privateKey crypto.PrivateKey
}

// Same as forge.FileMeta but with json tags and string data
type config struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type requestStructure struct {
	Repo          *model.Repo     `json:"repo"`
	Pipeline      *model.Pipeline `json:"pipeline"`
	Configuration []*config       `json:"configs"`
}

type responseStructure struct {
	Configs []config `json:"configs"`
}

func NewHTTP(endpoint string, privateKey crypto.PrivateKey) Extension {
	return &http{endpoint, privateKey}
}

func (cp *http) IsConfigured() bool {
	return cp.endpoint != ""
}

func (cp *http) FetchConfig(ctx context.Context, repo *model.Repo, pipeline *model.Pipeline, currentFileMeta []*forge_types.FileMeta) (configData []*forge_types.FileMeta, useOld bool, err error) {
	currentConfigs := make([]*config, len(currentFileMeta))
	for i, pipe := range currentFileMeta {
		currentConfigs[i] = &config{Name: pipe.Name, Data: string(pipe.Data)}
	}

	response := new(responseStructure)
	body := requestStructure{Repo: repo, Pipeline: pipeline, Configuration: currentConfigs}
	status, err := utils.Send(ctx, "POST", cp.endpoint, cp.privateKey, body, response)
	if err != nil && status != 204 {
		return nil, false, fmt.Errorf("Failed to fetch config via http (%d) %w", status, err)
	}

	var newFileMeta []*forge_types.FileMeta
	if status != 200 {
		newFileMeta = make([]*forge_types.FileMeta, 0)
	} else {
		newFileMeta = make([]*forge_types.FileMeta, len(response.Configs))
		for i, pipe := range response.Configs {
			newFileMeta[i] = &forge_types.FileMeta{Name: pipe.Name, Data: []byte(pipe.Data)}
		}
	}

	return newFileMeta, status == 204, nil
}
