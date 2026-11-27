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

package exec

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/shared/httputil"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

const (
	configExtensionTimeout      = 10 * time.Second
	maxConfigExtensionErrorBody = 4096
)

// The request and response match server/services/config/http.go. The cli may not
// import the server packages, so the envelope is declared here and the repo and
// pipeline use the types of the go client.
type (
	configExtensionRequest struct {
		Repo          *woodpecker.Repo       `json:"repo"`
		Pipeline      *woodpecker.Pipeline   `json:"pipeline"`
		Restart       bool                   `json:"restart"`
		Configuration []*configExtensionFile `json:"configuration,omitempty"`
	}

	configExtensionResponse struct {
		Configs []*configExtensionFile `json:"configs"`
	}

	configExtensionFile struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}
)

func newConfigExtensionRequest(m *metadata.Metadata, local []*builder.YamlFile) *configExtensionRequest {
	fullName := m.Repo.Name
	if m.Repo.Owner != "" {
		fullName = m.Repo.Owner + "/" + m.Repo.Name
	}

	// TODO: the go client types lack clone_url_ssh, deploy_task, changed_files,
	// pr_labels, pr_milestone, pr_draft and cron, so they are not sent. Set them
	// once https://github.com/woodpecker-ci/woodpecker/pull/7588 is merged.
	req := &configExtensionRequest{
		Repo: &woodpecker.Repo{
			ID:           m.Repo.ID,
			Owner:        m.Repo.Owner,
			Name:         m.Repo.Name,
			FullName:     fullName,
			ForgeURL:     m.Repo.ForgeURL,
			Clone:        m.Repo.CloneURL,
			Branch:       m.Repo.Branch,
			IsSCMPrivate: m.Repo.Private,
			Trusted: woodpecker.TrustedConfiguration{
				Network:  m.Repo.Trusted.Network,
				Volumes:  m.Repo.Trusted.Volumes,
				Security: m.Repo.Trusted.Security,
			},
		},
		Pipeline: &woodpecker.Pipeline{
			Number:    m.Curr.Number,
			Parent:    m.Curr.Parent,
			Event:     string(m.Curr.Event),
			Created:   m.Curr.Created,
			Started:   m.Curr.Started,
			Deploy:    m.Curr.DeployTo,
			Commit:    m.Curr.Commit.Sha,
			Branch:    m.Curr.Commit.Branch,
			Ref:       m.Curr.Commit.Ref,
			Refspec:   m.Curr.Commit.Refspec,
			Author:    m.Curr.Commit.Author.Name,
			Avatar:    m.Curr.Avatar,
			Email:     m.Curr.Commit.Author.Email,
			Message:   m.Curr.Commit.Message,
			Timestamp: m.Curr.Commit.Timestamp,
			ForgeURL:  m.Curr.ForgeURL,
		},
	}
	for _, y := range local {
		req.Configuration = append(req.Configuration, &configExtensionFile{Name: y.Name, Data: string(y.Data)})
	}
	return req
}

// configExtensionDoer sends a request to the configuration extension. It is the
// signing client when signing is on, and a plain client otherwise.
type configExtensionDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// configExtensionClient returns the client for config extension requests. With
// sign set, the request is signed with the CLI's ed25519 key, and the matching
// public key is written to stderr so the extension can be configured with it.
func configExtensionClient(sign bool) (configExtensionDoer, error) {
	if !sign {
		return &http.Client{Transport: httputil.NewUserAgentRoundTripper(nil, "cli")}, nil
	}
	privateKey, err := configExtensionPrivateKey()
	if err != nil {
		return nil, err
	}
	if err := provideConfigExtensionPublicKey(privateKey); err != nil {
		return nil, err
	}
	return newConfigExtensionClient(privateKey)
}

// fetchConfigExtension asks the configuration extension at endpoint for the
// workflows of the pipeline described by m. The local files are sent as the
// repository's config files and are returned when the extension responds with
// 204 No Content, the same as the server does. Requests are signed unless sign
// is false.
func fetchConfigExtension(ctx context.Context, endpoint string, m *metadata.Metadata, local []*builder.YamlFile, sign bool) ([]*builder.YamlFile, error) {
	client, err := configExtensionClient(sign)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(newConfigExtensionRequest(m, local))
	if err != nil {
		return nil, fmt.Errorf("could not encode config extension request: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, configExtensionTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("could not create config extension request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch config from extension: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warn().Err(err).Msg("could not close config extension response body")
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNoContent:
		if len(local) == 0 {
			return nil, errors.New("config extension responded with 204 No Content and there are no local workflow files to fall back to")
		}
		log.Debug().Str("endpoint", endpoint).Msg("config extension responded with 204 No Content, using the local workflow files")
		return local, nil
	default:
		msg, readErr := io.ReadAll(io.LimitReader(resp.Body, maxConfigExtensionErrorBody))
		if readErr != nil {
			return nil, fmt.Errorf("config extension responded with status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("config extension responded with status %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	var out configExtensionResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("could not decode config extension response: %w", err)
	}
	if len(out.Configs) == 0 {
		return nil, errors.New("config extension returned no configs")
	}

	yamls := make([]*builder.YamlFile, 0, len(out.Configs))
	for _, config := range out.Configs {
		yamls = append(yamls, &builder.YamlFile{Name: config.Name, Data: []byte(config.Data)})
	}
	return yamls, nil
}
