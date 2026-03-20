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

package secret

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

type secretRequestStructure struct {
	Repo     *model.Repo     `json:"repo"`
	Pipeline *model.Pipeline `json:"pipeline"`
	Netrc    *model.Netrc    `json:"netrc"`
}

// NewHTTP returns a new HTTP secret extension client.
func NewHTTP(endpoint string, client *utils.Client) *httpExtension {
	return &httpExtension{endpoint: endpoint, client: client}
}

// SecretListPipeline fetches secrets from an external HTTP extension.
func (h *httpExtension) SecretListPipeline(repo *model.Repo, pipeline *model.Pipeline, netrc *model.Netrc) ([]*model.Secret, error) {
	body := secretRequestStructure{
		Repo:     repo,
		Pipeline: pipeline,
		Netrc:    netrc,
	}

	var response []*model.Secret
	status, err := h.client.Send(context.Background(), net_http.MethodPost, h.endpoint, body, &response)
	if err != nil && status != net_http.StatusNoContent {
		return nil, fmt.Errorf("failed to fetch secrets via http (%d) %w", status, err)
	}

	if status != net_http.StatusOK {
		// 204 No Content means no additional secrets
		return nil, nil
	}

	return response, nil
}
