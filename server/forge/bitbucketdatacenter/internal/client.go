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

package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

const (
	currentUserID = "%s/plugins/servlet/applinks/whoami" // cspell:disable-line
)

type Client struct {
	client *http.Client
	base   string
}

func NewClientWithToken(ctx context.Context, ts oauth2.TokenSource, url string) *Client {
	return &Client{
		client: oauth2.NewClient(ctx, ts),
		base:   url,
	}
}

// FindCurrentUser is returning the current user id - however it is not really part of the API so it is not part of the Bitbucket go client.
func (c *Client) FindCurrentUser(ctx context.Context) (string, error) {
	url := fmt.Sprintf(currentUserID, c.base)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create http request: %w", err)
	}

	resp, err := c.client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", fmt.Errorf("unable to query logged in user id: %w", err)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read data from user id query: %w", err)
	}
	login := string(buf)
	login = strings.ReplaceAll(login, "@", "_") // Apparently the "whoami" endpoint may return the "wrong" username - converting to user slug
	return login, nil
}
