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

package config

import (
	"context"
	"crypto/ed25519"
	"fmt"
	net_http "net/http"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/utils"
)

type http struct {
	endpoint   string
	privateKey ed25519.PrivateKey
}

// configData same as forge.FileMeta but with json tags and string data.
type configData struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type requestStructure struct {
	Repo     *model.Repo     `json:"repo"`
	Pipeline *model.Pipeline `json:"pipeline"`
	Netrc    *model.Netrc    `json:"netrc"`
}

type responseStructure struct {
	Configs []*configData `json:"configs"`
}

func NewHTTP(endpoint string, privateKey ed25519.PrivateKey) Service {
	return &http{endpoint, privateKey}
}

func (h *http) Fetch(ctx context.Context, forge forge.Forge, user *model.User, repo *model.Repo, pipeline *model.Pipeline, oldConfigData []*types.FileMeta, _ bool) ([]*types.FileMeta, error) {
	netrc, err := forge.Netrc(user, repo)
	if err != nil {
		return nil, fmt.Errorf("could not get Netrc data from forge: %w", err)
	}

	response := new(responseStructure)
	body := requestStructure{
		Repo:     repo,
		Pipeline: pipeline,
		Netrc:    netrc,
	}

	status, err := utils.Send(ctx, net_http.MethodPost, h.endpoint, h.privateKey, body, response)
	if err != nil && status != 204 {
		return nil, fmt.Errorf("failed to fetch config via http (%d) %w", status, err)
	}

	if status != net_http.StatusOK {
		return oldConfigData, nil
	}

	fileMetaList := make([]*types.FileMeta, len(response.Configs))
	for i, config := range response.Configs {
		fileMetaList[i] = &types.FileMeta{Name: config.Name, Data: []byte(config.Data)}
	}

	return fileMetaList, nil
}
