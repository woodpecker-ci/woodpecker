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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/bitbucket"
)

const (
	get  = "GET"
	post = "POST"
	del  = "DELETE"
)

const (
	pathUser        = "%s/2.0/user/"
	pathEmails      = "%s/2.0/user/emails"
	pathPermissions = "%s/2.0/user/permissions/repositories?q=repository.full_name=%q"
	pathTeams       = "%s/2.0/teams/?%s"
	pathRepo        = "%s/2.0/repositories/%s/%s"
	pathRepos       = "%s/2.0/repositories/%s?%s"
	pathHook        = "%s/2.0/repositories/%s/%s/hooks/%s"
	pathHooks       = "%s/2.0/repositories/%s/%s/hooks?%s"
	pathSource      = "%s/2.0/repositories/%s/%s/src/%s/%s"
	pathStatus      = "%s/2.0/repositories/%s/%s/commit/%s/statuses/build"
)

type Client struct {
	*http.Client
	base string
	ctx  context.Context
}

func NewClient(ctx context.Context, url string, client *http.Client) *Client {
	return &Client{
		Client: client,
		base:   url,
		ctx:    ctx,
	}
}

func NewClientToken(ctx context.Context, url, client, secret string, token *oauth2.Token) *Client {
	config := &oauth2.Config{
		ClientID:     client,
		ClientSecret: secret,
		Endpoint:     bitbucket.Endpoint,
	}
	return NewClient(ctx, url, config.Client(ctx, token))
}

func (c *Client) FindCurrent() (*Account, error) {
	out := new(Account)
	uri := fmt.Sprintf(pathUser, c.base)
	_, err := c.do(uri, get, nil, out)
	return out, err
}

func (c *Client) ListEmail() (*EmailResp, error) {
	out := new(EmailResp)
	uri := fmt.Sprintf(pathEmails, c.base)
	_, err := c.do(uri, get, nil, out)
	return out, err
}

func (c *Client) ListTeams(opts *ListTeamOpts) (*AccountResp, error) {
	out := new(AccountResp)
	uri := fmt.Sprintf(pathTeams, c.base, opts.Encode())
	_, err := c.do(uri, get, nil, out)
	return out, err
}

func (c *Client) FindRepo(owner, name string) (*Repo, error) {
	out := new(Repo)
	uri := fmt.Sprintf(pathRepo, c.base, owner, name)
	_, err := c.do(uri, get, nil, out)
	return out, err
}

func (c *Client) ListRepos(account string, opts *ListOpts) (*RepoResp, error) {
	out := new(RepoResp)
	uri := fmt.Sprintf(pathRepos, c.base, account, opts.Encode())
	_, err := c.do(uri, get, nil, out)
	return out, err
}

func (c *Client) ListReposAll(account string) ([]*Repo, error) {
	var page = 1
	var repos []*Repo

	for {
		resp, err := c.ListRepos(account, &ListOpts{Page: page, PageLen: 100})
		if err != nil {
			return repos, err
		}
		repos = append(repos, resp.Values...)
		if len(resp.Next) == 0 {
			break
		}
		page = resp.Page + 1
	}
	return repos, nil
}

func (c *Client) FindHook(owner, name, id string) (*Hook, error) {
	out := new(Hook)
	uri := fmt.Sprintf(pathHook, c.base, owner, name, id)
	_, err := c.do(uri, get, nil, out)
	return out, err
}

func (c *Client) ListHooks(owner, name string, opts *ListOpts) (*HookResp, error) {
	out := new(HookResp)
	uri := fmt.Sprintf(pathHooks, c.base, owner, name, opts.Encode())
	_, err := c.do(uri, get, nil, out)
	return out, err
}

func (c *Client) CreateHook(owner, name string, hook *Hook) error {
	uri := fmt.Sprintf(pathHooks, c.base, owner, name, "")
	_, err := c.do(uri, post, hook, nil)
	return err
}

func (c *Client) DeleteHook(owner, name, id string) error {
	uri := fmt.Sprintf(pathHook, c.base, owner, name, id)
	_, err := c.do(uri, del, nil, nil)
	return err
}

func (c *Client) FindSource(owner, name, revision, path string) (*string, error) {
	uri := fmt.Sprintf(pathSource, c.base, owner, name, revision, path)
	return c.do(uri, get, nil, nil)
}

func (c *Client) CreateStatus(owner, name, revision string, status *BuildStatus) error {
	uri := fmt.Sprintf(pathStatus, c.base, owner, name, revision)
	_, err := c.do(uri, post, status, nil)
	return err
}

func (c *Client) GetPermission(fullName string) (*RepoPerm, error) {
	out := new(RepoPermResp)
	uri := fmt.Sprintf(pathPermissions, c.base, fullName)
	_, err := c.do(uri, get, nil, out)

	if err != nil {
		return nil, err
	}

	if len(out.Values) == 0 {
		return nil, fmt.Errorf("no permissions in repository %s", fullName)
	} else {
		return out.Values[0], nil
	}
}

func (c *Client) do(rawurl, method string, in, out interface{}) (*string, error) {
	uri, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	// if we are posting or putting data, we need to
	// write it to the body of the request.
	var buf io.ReadWriter
	if in != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(in)
		if err != nil {
			return nil, err
		}
	}

	// creates a new http request to bitbucket.
	req, err := http.NewRequestWithContext(c.ctx, method, uri.String(), buf)
	if err != nil {
		return nil, err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// if an error is encountered, parse and return the
	// error response.
	if resp.StatusCode > http.StatusPartialContent {
		err := Error{}
		json.NewDecoder(resp.Body).Decode(&err)
		err.Status = resp.StatusCode
		return nil, err
	}

	// if a json response is expected, parse and return
	// the json response.
	if out != nil {
		return nil, json.NewDecoder(resp.Body).Decode(out)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)

	return &bodyString, nil
}
