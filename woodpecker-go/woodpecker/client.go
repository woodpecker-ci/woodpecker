// Copyright 2022 Woodpecker Authors
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

package woodpecker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	pathLogLevel = "%s/api/log-level"

	//nolint:godot
	// TODO: implement endpoints
	// pathFeed           = "%s/api/user/feed"
	// pathVersion        = "%s/version"
)

type ClientError struct {
	StatusCode int
	Message    string
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("client error %d: %s", e.StatusCode, e.Message)
}

type client struct {
	client *http.Client
	addr   string
}

// New returns a client at the specified url.
func New(uri string) Client {
	return &client{http.DefaultClient, strings.TrimSuffix(uri, "/")}
}

// NewClient returns a client at the specified url.
func NewClient(uri string, cli *http.Client) Client {
	return &client{cli, strings.TrimSuffix(uri, "/")}
}

// SetClient sets the http.Client.
func (c *client) SetClient(client *http.Client) {
	c.client = client
}

// SetAddress sets the server address.
func (c *client) SetAddress(addr string) {
	c.addr = addr
}

// LogLevel returns the current logging level.
func (c *client) LogLevel() (*LogLevel, error) {
	out := new(LogLevel)
	uri := fmt.Sprintf(pathLogLevel, c.addr)
	err := c.get(uri, out)
	return out, err
}

// SetLogLevel sets the logging level of the server.
func (c *client) SetLogLevel(in *LogLevel) (*LogLevel, error) {
	out := new(LogLevel)
	uri := fmt.Sprintf(pathLogLevel, c.addr)
	err := c.post(uri, in, out)
	return out, err
}

//
// HTTP request helper functions.
//

// Helper function for making an http GET request.
func (c *client) get(rawURL string, out any) error {
	return c.do(rawURL, http.MethodGet, nil, out)
}

// Helper function for making an http POST request.
func (c *client) post(rawURL string, in, out any) error {
	return c.do(rawURL, http.MethodPost, in, out)
}

// Helper function for making an http PATCH request.
func (c *client) patch(rawURL string, in, out any) error {
	return c.do(rawURL, http.MethodPatch, in, out)
}

// Helper function for making an http DELETE request.
func (c *client) delete(rawURL string) error {
	return c.do(rawURL, http.MethodDelete, nil, nil)
}

// Helper function to make an http request.
func (c *client) do(rawURL, method string, in, out any) error {
	body, err := c.open(rawURL, method, in)
	if err != nil {
		return err
	}
	defer body.Close()
	if out != nil {
		return json.NewDecoder(body).Decode(out)
	}
	return nil
}

// Helper function to open an http request.
func (c *client) open(rawURL, method string, in any) (io.ReadCloser, error) {
	uri, err := url.Parse(rawURL)
	if err != nil {
		return nil, &ClientError{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}
	}
	req, err := http.NewRequest(method, uri.String(), nil)
	if err != nil {
		return nil, &ClientError{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}
	}
	if in != nil {
		decoded, decodeErr := json.Marshal(in)
		if decodeErr != nil {
			return nil, &ClientError{
				StatusCode: http.StatusInternalServerError,
				Message:    decodeErr.Error(),
			}
		}
		buf := bytes.NewBuffer(decoded)
		req.Body = io.NopCloser(buf)
		req.ContentLength = int64(len(decoded))
		req.Header.Set("Content-Length", strconv.Itoa(len(decoded)))
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, &ClientError{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}
	}
	if resp.StatusCode > http.StatusPartialContent {
		defer resp.Body.Close()
		out, _ := io.ReadAll(resp.Body)
		return nil, &ClientError{
			StatusCode: resp.StatusCode,
			Message:    string(out),
		}
	}
	return resp.Body, nil
}

// mapValues converts a map to `url.Values`.
func mapValues(params map[string]string) url.Values {
	values := url.Values{}
	for key, val := range params {
		values.Add(key, val)
	}
	return values
}
