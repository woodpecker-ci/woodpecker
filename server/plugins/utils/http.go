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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-ap/httpsig"
)

// Send makes an http request to the given endpoint, writing the input
// to the request body and un-marshaling the output from the response body.
func Send(ctx context.Context, method, path string, privateKey crypto.PrivateKey, in, out any) (int, error) {
	uri, err := url.Parse(path)
	if err != nil {
		return 0, err
	}

	// if we are posting or putting data, we need to write it to the body of the request.
	var buf io.ReadWriter
	if in != nil {
		buf = new(bytes.Buffer)
		jsonerr := json.NewEncoder(buf).Encode(in)
		if jsonerr != nil {
			return 0, jsonerr
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

	err = SignHTTPRequest(privateKey, req)
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, err
		}

		return resp.StatusCode, fmt.Errorf("Response: %s", string(body))
	}

	// if no other errors parse and return the json response.
	err = json.NewDecoder(resp.Body).Decode(out)
	return resp.StatusCode, err
}

func SignHTTPRequest(privateKey crypto.PrivateKey, req *http.Request) error {
	pubKeyID := "woodpecker-ci-plugins"

	signer := httpsig.NewEd25519Signer(pubKeyID, privateKey, nil)

	return signer.Sign(req)
}
