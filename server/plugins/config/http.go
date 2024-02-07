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

	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/utils"
)

type http struct {
	endpoint   string
	privateKey crypto.PrivateKey
}

type requestStructure struct {
	Repo          *model.Repo     `json:"repo"`
	Pipeline      *model.Pipeline `json:"pipeline"`
	Configuration []*ConfigData   `json:"configs"`
	Netrc         *model.Netrc    `json:"netrc"`
}

type responseStructure struct {
	Configs []*ConfigData `json:"configs"`
}

func NewHTTP(endpoint string, privateKey crypto.PrivateKey) Service {
	return &http{endpoint, privateKey}
}

func (h *http) Fetch(ctx context.Context, forge *forge.Forge, user *model.User, repo *model.Repo, pipeline *model.Pipeline, netrc *model.Netrc) ([]*ConfigData, error) {
	// currentConfigs := make([]*configData, len(currentFileMeta))
	// for i, pipe := range currentFileMeta {
	// 	currentConfigs[i] = &configData{Name: pipe.Name, Data: string(pipe.Data)}
	// }

	response := new(responseStructure)
	body := requestStructure{
		Repo:     repo,
		Pipeline: pipeline,
		// Configuration: currentConfigs,
		Netrc: netrc,
	}

	status, err := utils.Send(ctx, "POST", h.endpoint, h.privateKey, body, response)
	if err != nil && status != 204 {
		return nil, fmt.Errorf("failed to fetch config via http (%d) %w", status, err)
	}

	if status != 200 {
		return []*ConfigData{}, nil
	}

	return response.Configs, nil
}
