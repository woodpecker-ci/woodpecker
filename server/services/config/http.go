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
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils"
)

type httpService struct {
	endpoint     string
	client       *utils.Client
	includeNetrc bool
}

// configData same as forge.FileMeta but with json tags and string data.
type configData struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type requestStructure struct {
	Repo          *model.Repo     `json:"repo"`
	Pipeline      *model.Pipeline `json:"pipeline"`
	Netrc         *model.Netrc    `json:"netrc"`
	Configuration []*configData   `json:"configuration,omitempty"`
}

type responseStructure struct {
	Configs []*configData `json:"configs"`
}

func NewHTTP(endpoint string, client *utils.Client, includeNetrc bool) Service {
	return &httpService{endpoint, client, includeNetrc}
}

func (h *httpService) Fetch(ctx context.Context, forge forge.Forge, user *model.User, repo *model.Repo, pipeline *model.Pipeline, oldConfigData []*types.FileMeta, _ bool) ([]*types.FileMeta, error) {
	configuration := make([]*configData, len(oldConfigData))
	for i, oldConfig := range oldConfigData {
		configuration[i] = &configData{Name: oldConfig.Name, Data: string(oldConfig.Data)}
	}

	response := new(responseStructure)
	body := requestStructure{
		Repo:          repo,
		Pipeline:      pipeline,
		Configuration: configuration,
	}

	if h.includeNetrc {
		netrc, err := forge.Netrc(user, repo)
		if err != nil {
			return nil, fmt.Errorf("could not get Netrc data from forge: %w", err)
		}
		body.Netrc = netrc
	}

	status, err := h.client.Send(ctx, http.MethodPost, h.endpoint, body, response)
	if err != nil && status != http.StatusNoContent {
		return nil, fmt.Errorf("failed to fetch config via http (status: %d): %w", status, err)
	}

	// handle 204 - no new config available, return old config without error
	if status == http.StatusNoContent {
		log.Debug().
			Str("endpoint", h.endpoint).
			Str("repo", repo.FullName).
			Msg("config endpoint returned 204 No Content, using fallback config")
		return oldConfigData, nil
	}

	// unexpected non-success status code
	if status != http.StatusOK {
		return oldConfigData, fmt.Errorf("unexpected status code %d from config endpoint (expected 200 or 204)", status)
	}

	fileMetaList := make([]*types.FileMeta, len(response.Configs))
	for i, config := range response.Configs {
		fileMetaList[i] = &types.FileMeta{Name: config.Name, Data: []byte(config.Data)}
	}

	return fileMetaList, nil
}
