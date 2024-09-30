// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/yaronf/httpsign"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/utils/hostmatcher"
)

type Client struct {
	*http.Client
	privateKey crypto.PrivateKey
}

func getHttpClient(allowedHostListValue string) *http.Client {
	timeout := time.Duration(10 * time.Second)

	if allowedHostListValue == "" {
		allowedHostListValue = hostmatcher.MatchBuiltinExternal
	}
	allowedHostMatcher := hostmatcher.ParseHostMatchList("WOODPECKER_ALLOWED_EXTENSIONS_HOSTS", allowedHostListValue)

	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
			DialContext:     hostmatcher.NewDialContext("extensions", allowedHostMatcher, nil),
		},
	}
}

func NewHTTPClient(privateKey crypto.PrivateKey, allowedHostList string) *Client {
	return &Client{
		Client:     getHttpClient(allowedHostList),
		privateKey: privateKey,
	}
}

// Send makes an http request to the given endpoint, writing the input
// to the request body and un-marshaling the output from the response body.
func (e *Client) Send(ctx context.Context, method, path string, in, out any) (int, error) {
	uri, err := url.Parse(path)
	if err != nil {
		return 0, err
	}

	// if we are posting or putting data, we need to write it to the body of the request.
	var buf io.ReadWriter
	if in != nil {
		buf = new(bytes.Buffer)
		jsonErr := json.NewEncoder(buf).Encode(in)
		if jsonErr != nil {
			return 0, jsonErr
		}
	}

	// creates a new http request to the endpoint.
	req, err := http.NewRequestWithContext(ctx, method, uri.String(), buf)
	if err != nil {
		return 0, err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	err = SignHTTPRequest(e.privateKey, req)
	if err != nil {
		return 0, err
	}

	resp, err := e.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, err
		}

		return resp.StatusCode, fmt.Errorf("response: %s", string(body))
	}

	// if no other errors parse and return the json response.
	err = json.NewDecoder(resp.Body).Decode(out)
	return resp.StatusCode, err
}

func signClient(privateKey ed25519.PrivateKey) (*httpsign.Client, error) {
	pubKeyID := "woodpecker-ci-extensions"

	signer, err := httpsign.NewEd25519Signer(privateKey,
		httpsign.NewSignConfig(),
		httpsign.Headers("@request-target", "content-digest")) // The Content-Digest header will be auto-generated
	if err != nil {
		return nil, err
	}
	return httpsign.NewDefaultClient(httpsign.NewClientConfig().SetSignatureName(pubKeyID).SetSigner(signer)), nil // sign requests, don't verify responses
}
