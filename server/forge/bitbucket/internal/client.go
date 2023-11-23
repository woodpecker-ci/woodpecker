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
	"net/http"
	"net/url"

	shared_utils "go.woodpecker-ci.org/woodpecker/shared/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/bitbucket"
)

const (
	get  = "GET"
	post = "POST"
	del  = "DELETE"
)

const (
	pathUser          = "%s/2.0/user/"
	pathEmails        = "%s/2.0/user/emails"
	pathPermissions   = "%s/2.0/user/permissions/repositories?q=repository.full_name=%q"
	pathWorkspace     = "%s/2.0/workspaces/?%s"
	pathRepo          = "%s/2.0/repositories/%s/%s"
	pathRepos         = "%s/2.0/repositories/%s?%s"
	pathHook          = "%s/2.0/repositories/%s/%s/hooks/%s"
	pathHooks         = "%s/2.0/repositories/%s/%s/hooks?%s"
	pathSource        = "%s/2.0/repositories/%s/%s/src/%s/%s"
	pathStatus        = "%s/2.0/repositories/%s/%s/commit/%s/statuses/build"
	pathBranches      = "%s/2.0/repositories/%s/%s/refs/branches?%s"
	pathOrgPerms      = "%s/2.0/workspaces/%s/permissions?%s"
	pathPullRequests  = "%s/2.0/repositories/%s/%s/pullrequests"
	pathBranchCommits = "%s/2.0/repositories/%s/%s/commits/%s"
	pathDir           = "%s/2.0/repositories/%s/%s/src/%s%s"
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

func (c *Client) ListWorkspaces(opts *ListWorkspacesOpts) (*WorkspacesResp, error) {
	out := new(WorkspacesResp)
	uri := fmt.Sprintf(pathWorkspace, c.base, opts.Encode())
	_, err := c.do(uri, get, nil, out)
	return out, err
}

func (c *Client) FindRepo(owner, name string) (*Repo, error) {
	out := new(Repo)
	uri := fmt.Sprintf(pathRepo, c.base, owner, name)
	_, err := c.do(uri, get, nil, out)
	return out, err
}

func (c *Client) ListRepos(workspace string, opts *ListOpts) (*RepoResp, error) {
	out := new(RepoResp)
	uri := fmt.Sprintf(pathRepos, c.base, workspace, opts.Encode())
	_, err := c.do(uri, get, nil, out)
	return out, err
}

func (c *Client) ListReposAll(workspace string) ([]*Repo, error) {
	return shared_utils.Paginate(func(page int) ([]*Repo, error) {
		resp, err := c.ListRepos(workspace, &ListOpts{Page: page, PageLen: 100})
		if err != nil {
			return nil, err
		}
		return resp.Values, nil
	})
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

func (c *Client) CreateStatus(owner, name, revision string, status *PipelineStatus) error {
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
	}
	return out.Values[0], nil
}

func (c *Client) ListBranches(owner, name string, opts *ListOpts) ([]*Branch, error) {
	out := new(BranchResp)
	uri := fmt.Sprintf(pathBranches, c.base, owner, name, opts.Encode())
	_, err := c.do(uri, get, nil, out)
	return out.Values, err
}

func (c *Client) GetBranchHead(owner, name, branch string) (string, error) {
	out := new(CommitsResp)
	uri := fmt.Sprintf(pathBranchCommits, c.base, owner, name, branch)
	_, err := c.do(uri, get, nil, out)
	if err != nil {
		return "", err
	}
	if len(out.Values) == 0 {
		return "", fmt.Errorf("no commits in branch %s", branch)
	}
	return out.Values[0].Hash, nil
}

func (c *Client) GetUserWorkspaceMembership(workspace, user string) (string, error) {
	out := new(WorkspaceMembershipResp)
	opts := &ListOpts{Page: 1, PageLen: 100}
	for {
		uri := fmt.Sprintf(pathOrgPerms, c.base, workspace, opts.Encode())
		_, err := c.do(uri, get, nil, out)
		if err != nil {
			return "", err
		}
		for _, m := range out.Values {
			if m.User.Nickname == user {
				return m.Permission, nil
			}
		}
		if len(out.Next) == 0 {
			break
		}
		opts.Page++
	}
	return "", nil
}

func (c *Client) ListPullRequests(owner, name string, opts *ListOpts) ([]*PullRequest, error) {
	out := new(PullRequestResp)
	uri := fmt.Sprintf(pathPullRequests, c.base, owner, name)
	_, err := c.do(uri, get, opts.Encode(), out)
	return out.Values, err
}

func (c *Client) GetWorkspace(name string) (*Workspace, error) {
	out := new(Workspace)
	uri := fmt.Sprintf(pathWorkspace, c.base, name)
	_, err := c.do(uri, get, nil, out)
	return out, err
}

func (c *Client) GetRepoFiles(owner, name, revision, path string, page *string) (*DirResp, error) {
	out := new(DirResp)
	uri := fmt.Sprintf(pathDir, c.base, owner, name, revision, path)
	if page != nil {
		uri += "?page=" + *page
	}
	_, err := c.do(uri, get, nil, out)
	return out, err
}

func (c *Client) do(rawurl, method string, in, out any) (*string, error) {
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
		_ = json.NewDecoder(resp.Body).Decode(&err)
		err.Status = resp.StatusCode
		return nil, err
	}

	// if a json response is expected, parse and return
	// the json response.
	if out != nil {
		return nil, json.NewDecoder(resp.Body).Decode(out)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)

	return &bodyString, nil
}
