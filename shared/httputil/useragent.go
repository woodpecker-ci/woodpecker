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

package httputil

import (
	"fmt"
	"net/http"

	"go.woodpecker-ci.org/woodpecker/v3/version"
)

// UserAgentRoundTripper is an http.RoundTripper that sets a custom User-Agent header
// on all outgoing requests.
type UserAgentRoundTripper struct {
	base      http.RoundTripper
	userAgent string
}

// NewUserAgentRoundTripper creates a new RoundTripper that adds the Woodpecker User-Agent
// to all requests. If base is nil, http.DefaultTransport is used.
func NewUserAgentRoundTripper(base http.RoundTripper, component string) *UserAgentRoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	userAgent := fmt.Sprintf("Woodpecker/%s", version.String())
	if component != "" {
		userAgent = fmt.Sprintf("%s (%s)", userAgent, component)
	}

	return &UserAgentRoundTripper{
		base:      base,
		userAgent: userAgent,
	}
}

// RoundTrip implements the http.RoundTripper interface.
func (rt *UserAgentRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	reqClone := req.Clone(req.Context())

	// Set the User-Agent header if not already set
	if reqClone.Header.Get("User-Agent") == "" {
		reqClone.Header.Set("User-Agent", rt.userAgent)
	}

	// Execute the request using the base transport
	return rt.base.RoundTrip(reqClone)
}

// WrapClient wraps an existing http.Client with the UserAgentRoundTripper.
// If client is nil, a new client with default settings is created.
func WrapClient(client *http.Client, component string) *http.Client {
	if client == nil {
		client = &http.Client{}
	}

	client.Transport = NewUserAgentRoundTripper(client.Transport, component)
	return client
}
