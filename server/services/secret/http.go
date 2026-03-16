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

package secret

import (
	"context"
	"fmt"
	net_http "net/http"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils"
)

type http struct {
	endpoint string
	client   *utils.Client
}

type secretRequestStructure struct {
	Repo     *model.Repo     `json:"repo"`
	Pipeline *model.Pipeline `json:"pipeline"`
	Netrc    *model.Netrc    `json:"netrc"`
}

// NewHTTP returns a new external secret service backed by an HTTP endpoint.
func NewHTTP(endpoint string, client *utils.Client) Service {
	return &http{endpoint: endpoint, client: client}
}

func (h *http) SecretListPipeline(repo *model.Repo, pipeline *model.Pipeline, netrc *model.Netrc) ([]*model.Secret, error) {
	body := secretRequestStructure{
		Repo:     repo,
		Pipeline: pipeline,
		Netrc:    netrc,
	}

	var response []*model.Secret
	status, err := h.client.Send(context.Background(), net_http.MethodPost, h.endpoint, body, &response)
	if err != nil && status != net_http.StatusNoContent {
		return nil, fmt.Errorf("failed to fetch secrets via http (status: %d): %w", status, err)
	}

	if status == net_http.StatusNoContent {
		log.Debug().
			Str("endpoint", h.endpoint).
			Str("repo", repo.FullName).
			Msg("secret endpoint returned 204 No Content, returning no external secrets")
		return nil, nil
	}

	if status != net_http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d from secret endpoint (expected 200 or 204)", status)
	}

	return response, nil
}

// The CRUD methods below are not supported by the external secret service.
// They return errors to indicate that management must be done via the external service itself.

func (h *http) SecretFind(_ *model.Repo, _ string) (*model.Secret, error) {
	return nil, fmt.Errorf("operation not supported by external secret service")
}

func (h *http) SecretList(_ *model.Repo, _ *model.ListOptions) ([]*model.Secret, error) {
	return nil, fmt.Errorf("operation not supported by external secret service")
}

func (h *http) SecretCreate(_ *model.Repo, _ *model.Secret) error {
	return fmt.Errorf("operation not supported by external secret service")
}

func (h *http) SecretUpdate(_ *model.Repo, _ *model.Secret) error {
	return fmt.Errorf("operation not supported by external secret service")
}

func (h *http) SecretDelete(_ *model.Repo, _ string) error {
	return fmt.Errorf("operation not supported by external secret service")
}

func (h *http) OrgSecretFind(_ int64, _ string) (*model.Secret, error) {
	return nil, fmt.Errorf("operation not supported by external secret service")
}

func (h *http) OrgSecretList(_ int64, _ *model.ListOptions) ([]*model.Secret, error) {
	return nil, fmt.Errorf("operation not supported by external secret service")
}

func (h *http) OrgSecretCreate(_ int64, _ *model.Secret) error {
	return fmt.Errorf("operation not supported by external secret service")
}

func (h *http) OrgSecretUpdate(_ int64, _ *model.Secret) error {
	return fmt.Errorf("operation not supported by external secret service")
}

func (h *http) OrgSecretDelete(_ int64, _ string) error {
	return fmt.Errorf("operation not supported by external secret service")
}

func (h *http) GlobalSecretFind(_ string) (*model.Secret, error) {
	return nil, fmt.Errorf("operation not supported by external secret service")
}

func (h *http) GlobalSecretList(_ *model.ListOptions) ([]*model.Secret, error) {
	return nil, fmt.Errorf("operation not supported by external secret service")
}

func (h *http) GlobalSecretCreate(_ *model.Secret) error {
	return fmt.Errorf("operation not supported by external secret service")
}

func (h *http) GlobalSecretUpdate(_ *model.Secret) error {
	return fmt.Errorf("operation not supported by external secret service")
}

func (h *http) GlobalSecretDelete(_ string) error {
	return fmt.Errorf("operation not supported by external secret service")
}
