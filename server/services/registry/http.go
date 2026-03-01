// Copyright 2025 Woodpecker Authors
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

package registry

import (
	"context"
	"fmt"
	net_http "net/http"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils"
)

type httpExtension struct {
	endpoint string
	client   *utils.Client
}

type requestStructure struct {
	Repo     *model.Repo     `json:"repo"`
	Pipeline *model.Pipeline `json:"pipeline"`
}

type responseStructure struct {
	Registries []*registryData `json:"registries"`
}

type registryData struct {
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// NewHTTP returns a new HTTP registry extension client.
func NewHTTP(endpoint string, client *utils.Client) *httpExtension {
	return &httpExtension{endpoint, client}
}

// RegistryListPipeline fetches registry credentials from an external HTTP extension.
func (h *httpExtension) RegistryListPipeline(ctx context.Context, repo *model.Repo, pipeline *model.Pipeline) ([]*model.Registry, error) {
	response := new(responseStructure)
	body := requestStructure{
		Repo:     repo,
		Pipeline: pipeline,
	}

	status, err := h.client.Send(ctx, net_http.MethodPost, h.endpoint, body, response)
	if err != nil && status != net_http.StatusNoContent {
		return nil, fmt.Errorf("failed to fetch registries via http (%d) %w", status, err)
	}

	if status != net_http.StatusOK {
		// 204 No Content means no additional registries
		return nil, nil
	}

	registries := make([]*model.Registry, len(response.Registries))
	for i, reg := range response.Registries {
		registries[i] = &model.Registry{
			Address:  reg.Address,
			Username: reg.Username,
			Password: reg.Password,
		}
	}

	return registries, nil
}
