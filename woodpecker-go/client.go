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

package woodpeckergo

import (
	"net/http"
	"net/url"

	httptransport "github.com/go-openapi/runtime/client"

	apiClient "go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/client"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../docs/openapi.json

type Client struct {
	*apiClient.ClientWithResponses
	uri       string
	transport *httptransport.Runtime
}

// New returns a client at the specified url.
func New(uri string) (*Client, error) {
	return NewWithClient(uri, http.DefaultClient)
}

// NewWithClient returns a client at the specified url.
func NewWithClient(_uri string, httpClient *http.Client) (*Client, error) {
	uri, err := url.Parse(_uri)
	if err != nil {
		return nil, err
	}

	transport := httptransport.NewWithClient(uri.Host, uri.Path, []string{"https", "http"}, httpClient)

	client, err := apiClient.NewClientWithResponses(uri.Host, apiClient.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	return &Client{
		uri:                 _uri,
		transport:           transport,
		ClientWithResponses: client,
	}, nil
}
