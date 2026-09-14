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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
)

// graphqlURL derives the GraphQL endpoint from the REST API url.
func (c *client) graphqlURL() (string, error) {
	if c.API == defaultAPI {
		return "https://api.github.com/graphql", nil
	}
	// GitHub Enterprise: https://host/api/v3/ -> https://host/api/graphql
	if !strings.HasSuffix(c.API, "v3/") {
		return "", fmt.Errorf("cannot derive graphql endpoint from api url %s", c.API)
	}
	return strings.TrimSuffix(c.API, "v3/") + "graphql", nil
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
	Data *struct {
		// a pointer, so a response that does not carry a repository object at
		// all is distinguishable from one where the repository resolved but the
		// requested path did not exist
		Repository *struct {
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

	url, err := c.graphqlURL()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// reuse the REST client's http client, so the transport (and any client
	// injected into the context) is shared instead of allocating a new
	// connection pool per directory fetch
	restClient, err := c.newClientToken(ctx, token)
	if err != nil {
		return nil, err
	}
	httpClient := restClient.Client()

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

	// only a response that actually resolved the repository can prove the path
	// is missing. Anything else is an unexpected shape and must fall back to
	// the REST path rather than be reported as "config not found", which would
	// silently drop the pipeline
	if result.Data == nil || result.Data.Repository == nil {
		return nil, fmt.Errorf("graphql response for %s did not contain a repository object", path)
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
