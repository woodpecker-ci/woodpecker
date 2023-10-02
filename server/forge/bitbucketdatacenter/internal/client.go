// Copyright 2022 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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

	"github.com/mrjones/oauth"
	"github.com/rs/zerolog/log"
)

const (
	currentUserID = "%s/plugins/servlet/applinks/whoami"
)

type Client struct {
	client      *http.Client
	base        string
	accessToken string
	ctx         context.Context
}

func NewClientWithToken(ctx context.Context, url string, consumer *oauth.Consumer, AccessToken string) *Client {
	var token oauth.AccessToken
	token.Token = AccessToken
	client, err := consumer.MakeHttpClient(&token)
	if err != nil {
		log.Err(err).Msg("")
	}

	return &Client{
		client:      client,
		base:        url,
		accessToken: AccessToken,
		ctx:         ctx,
	}
}

func (c *Client) FindCurrentUser() (string, error) {
	resp, err := c.doGet(fmt.Sprintf(currentUserID, c.base))
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
	return login, nil
}

// Helper function to help create get
func (c *Client) doGet(url string) (*http.Response, error) {
	log.Trace().Msgf("do GET from %s", url)
	request, err := http.NewRequestWithContext(c.ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json")
	return c.client.Do(request)
}
