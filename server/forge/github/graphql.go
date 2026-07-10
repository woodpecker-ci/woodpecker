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

package github

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/shared/httputil"
)

// graphqlURL derives the GraphQL endpoint from the REST API url.
func (c *client) graphqlURL() string {
	if c.API == defaultAPI {
		return "https://api.github.com/graphql"
	}
	// GitHub Enterprise: https://host/api/v3/ -> https://host/api/graphql
	return strings.TrimSuffix(c.API, "v3/") + "graphql"
}

// dirQuery fetches a directory including the contents of all its files in a
// single request.
const dirQuery = `query($owner: String!, $name: String!, $expression: String!) {
  repository(owner: $owner, name: $name) {
    object(expression: $expression) {
      __typename
      ... on Tree {
        entries {
          name
          type
          mode
          object {
            ... on Blob {
              text
              isTruncated
              isBinary
            }
          }
        }
      }
    }
  }
}`

type graphqlDirResponse struct {
	Data struct {
		Repository struct {
			Object *struct {
				TypeName string `json:"__typename"`
				Entries  []struct {
					Name   string `json:"name"`
					Type   string `json:"type"`
					Mode   int    `json:"mode"`
					Object struct {
						Text        *string `json:"text"`
						IsTruncated bool    `json:"isTruncated"`
						IsBinary    *bool   `json:"isBinary"`
					} `json:"object"`
				} `json:"entries"`
			} `json:"object"`
		} `json:"repository"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// dirFromGraphQL fetches all files of a repository directory with a single
// GraphQL request instead of one REST request per file, so a directory with
// many files neither multiplies the fetch latency nor the exposure to
// transient forge errors.
func (c *client) dirFromGraphQL(ctx context.Context, token, owner, name, ref, path string) ([]*forge_types.FileMeta, error) {
	body, err := json.Marshal(map[string]any{
		"query": dirQuery,
		"variables": map[string]string{
			"owner":      owner,
			"name":       name,
			"expression": ref + ":" + strings.TrimPrefix(path, "/"),
		},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.graphqlURL(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	baseTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	if c.SkipVerify {
		baseTransport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	httpClient := &http.Client{Transport: httputil.NewUserAgentRoundTripper(baseTransport, "forge-github")}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graphql request failed with status %d", resp.StatusCode)
	}

	var result graphqlDirResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("graphql: %s", result.Errors[0].Message)
	}

	object := result.Data.Repository.Object
	if object == nil {
		// the path does not exist at this ref
		return nil, &forge_types.ErrConfigNotFound{Configs: []string{path}}
	}
	if object.TypeName != "Tree" {
		return nil, fmt.Errorf("%s is not a directory at this ref", path)
	}

	// symlink mode in git trees, see https://git-scm.com/book/en/v2/Git-Internals-Git-Objects
	const gitModeSymlink = 0o120000

	files := make([]*forge_types.FileMeta, 0, len(object.Entries))
	for _, entry := range object.Entries {
		// nested directories are not descended into
		if entry.Type != "blob" {
			continue
		}
		if entry.Mode == gitModeSymlink {
			// the graphql api returns the link target instead of the linked
			// content, the per-file REST fallback resolves symlinks properly
			return nil, fmt.Errorf("%s in %s is a symlink", entry.Name, path)
		}
		if entry.Object.Text == nil || entry.Object.IsTruncated || (entry.Object.IsBinary != nil && *entry.Object.IsBinary) {
			return nil, fmt.Errorf("file %s in %s is binary or too large for the graphql api", entry.Name, path)
		}
		files = append(files, &forge_types.FileMeta{
			Name: path + "/" + entry.Name,
			Data: []byte(*entry.Object.Text),
		})
	}
	return files, nil
}
