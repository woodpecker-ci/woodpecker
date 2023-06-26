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

package bitbucket

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	shared_utils "github.com/woodpecker-ci/woodpecker/shared/utils"
	"golang.org/x/oauth2"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/forge"
	"github.com/woodpecker-ci/woodpecker/server/forge/bitbucket/internal"
	"github.com/woodpecker-ci/woodpecker/server/forge/common"
	forge_types "github.com/woodpecker-ci/woodpecker/server/forge/types"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

// Bitbucket cloud endpoints.
const (
	DefaultAPI = "https://api.bitbucket.org"
	DefaultURL = "https://bitbucket.org"
)

// Opts are forge options for bitbucket
type Opts struct {
	Client string
	Secret string
}

type config struct {
	API    string
	url    string
	Client string
	Secret string
}

// New returns a new forge Configuration for integrating with the Bitbucket
// repository hosting service at https://bitbucket.org
func New(opts *Opts) (forge.Forge, error) {
	return &config{
		API:    DefaultAPI,
		url:    DefaultURL,
		Client: opts.Client,
		Secret: opts.Secret,
	}, nil
	// TODO: add checks
}

// Name returns the string name of this driver
func (c *config) Name() string {
	return "bitbucket"
}

// URL returns the root url of a configured forge
func (c *config) URL() string {
	return c.url
}

// Login authenticates an account with Bitbucket using the oauth2 protocol. The
// Bitbucket account details are returned when the user is successfully authenticated.
func (c *config) Login(ctx context.Context, w http.ResponseWriter, req *http.Request) (*model.User, error) {
	config := c.newConfig(server.Config.Server.Host)

	// get the OAuth errors
	if err := req.FormValue("error"); err != "" {
		return nil, &forge_types.AuthError{
			Err:         err,
			Description: req.FormValue("error_description"),
			URI:         req.FormValue("error_uri"),
		}
	}

	// get the OAuth code
	code := req.FormValue("code")
	if len(code) == 0 {
		http.Redirect(w, req, config.AuthCodeURL("woodpecker"), http.StatusSeeOther)
		return nil, nil
	}

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	client := internal.NewClient(ctx, c.API, config.Client(ctx, token))
	curr, err := client.FindCurrent()
	if err != nil {
		return nil, err
	}
	return convertUser(curr, token), nil
}

// Auth uses the Bitbucket oauth2 access token and refresh token to authenticate
// a session and return the Bitbucket account login.
func (c *config) Auth(ctx context.Context, token, secret string) (string, error) {
	client := c.newClientToken(ctx, token, secret)
	user, err := client.FindCurrent()
	if err != nil {
		return "", err
	}
	return user.Login, nil
}

// Refresh refreshes the Bitbucket oauth2 access token. If the token is
// refreshed the user is updated and a true value is returned.
func (c *config) Refresh(ctx context.Context, user *model.User) (bool, error) {
	config := c.newConfig("")
	source := config.TokenSource(
		ctx, &oauth2.Token{RefreshToken: user.Secret})

	token, err := source.Token()
	if err != nil || len(token.AccessToken) == 0 {
		return false, err
	}

	user.Token = token.AccessToken
	user.Secret = token.RefreshToken
	user.Expiry = token.Expiry.UTC().Unix()
	return true, nil
}

// Teams returns a list of all team membership for the Bitbucket account.
func (c *config) Teams(ctx context.Context, u *model.User) ([]*model.Team, error) {
	return shared_utils.Paginate(func(page int) ([]*model.Team, error) {
		opts := &internal.ListWorkspacesOpts{
			PageLen: 100,
			Page:    page,
			Role:    "member",
		}
		resp, err := c.newClient(ctx, u).ListWorkspaces(opts)
		if err != nil {
			return nil, err
		}
		return convertWorkspaceList(resp.Values), nil
	})
}

// Repo returns the named Bitbucket repository.
func (c *config) Repo(ctx context.Context, u *model.User, remoteID model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
	if remoteID.IsValid() {
		name = string(remoteID)
	}
	repos, err := c.Repos(ctx, u)
	if err != nil {
		return nil, err
	}
	if len(owner) == 0 {
		for _, repo := range repos {
			if string(repo.ForgeRemoteID) == name {
				owner = repo.Owner
			}
		}
	}
	client := c.newClient(ctx, u)
	repo, err := client.FindRepo(owner, name)
	if err != nil {
		return nil, err
	}
	perm, err := client.GetPermission(repo.FullName)
	if err != nil {
		return nil, err
	}
	return convertRepo(repo, perm), nil
}

// Repos returns a list of all repositories for Bitbucket account, including
// organization repositories.
func (c *config) Repos(ctx context.Context, u *model.User) ([]*model.Repo, error) {
	client := c.newClient(ctx, u)

	var all []*model.Repo

	resp, err := client.ListWorkspaces(&internal.ListWorkspacesOpts{
		PageLen: 100,
		Role:    "member",
	})
	if err != nil {
		return all, err
	}

	for _, workspace := range resp.Values {
		repos, err := client.ListReposAll(workspace.Slug)
		if err != nil {
			return all, err
		}
		for _, repo := range repos {
			perm, err := client.GetPermission(repo.FullName)
			if err != nil {
				return nil, err
			}

			all = append(all, convertRepo(repo, perm))
		}
	}
	return all, nil
}

// File fetches the file from the Bitbucket repository and returns its contents.
func (c *config) File(ctx context.Context, u *model.User, r *model.Repo, p *model.Pipeline, f string) ([]byte, error) {
	config, err := c.newClient(ctx, u).FindSource(r.Owner, r.Name, p.Commit, f)
	if err != nil {
		return nil, err
	}
	return []byte(*config), err
}

func (c *config) Dir(_ context.Context, _ *model.User, _ *model.Repo, _ *model.Pipeline, _ string) ([]*forge_types.FileMeta, error) {
	return nil, forge_types.ErrNotImplemented
}

// Status creates a pipeline status for the Bitbucket commit.
func (c *config) Status(ctx context.Context, user *model.User, repo *model.Repo, pipeline *model.Pipeline, _ *model.Step) error {
	status := internal.PipelineStatus{
		State: convertStatus(pipeline.Status),
		Desc:  common.GetPipelineStatusDescription(pipeline.Status),
		Key:   "Woodpecker",
		URL:   common.GetPipelineStatusLink(repo, pipeline, nil),
	}
	return c.newClient(ctx, user).CreateStatus(repo.Owner, repo.Name, pipeline.Commit, &status)
}

// Activate activates the repository by registering repository push hooks with
// the Bitbucket repository. Prior to registering hook, previously created hooks
// are deleted.
func (c *config) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	rawurl, err := url.Parse(link)
	if err != nil {
		return err
	}
	_ = c.Deactivate(ctx, u, r, link)

	return c.newClient(ctx, u).CreateHook(r.Owner, r.Name, &internal.Hook{
		Active: true,
		Desc:   rawurl.Host,
		Events: []string{"repo:push"},
		URL:    link,
	})
}

// Deactivate deactivates the repository be removing repository push hooks from
// the Bitbucket repository.
func (c *config) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	client := c.newClient(ctx, u)

	hooks, err := client.ListHooks(r.Owner, r.Name, &internal.ListOpts{})
	if err != nil {
		return err
	}
	hook := matchingHooks(hooks.Values, link)
	if hook != nil {
		return client.DeleteHook(r.Owner, r.Name, hook.UUID)
	}
	return nil
}

// Netrc returns a netrc file capable of authenticating Bitbucket requests and
// cloning Bitbucket repositories.
func (c *config) Netrc(u *model.User, _ *model.Repo) (*model.Netrc, error) {
	return &model.Netrc{
		Machine:  "bitbucket.org",
		Login:    "x-token-auth",
		Password: u.Token,
	}, nil
}

// Branches returns the names of all branches for the named repository.
func (c *config) Branches(ctx context.Context, u *model.User, r *model.Repo, _ *model.ListOptions) ([]string, error) {
	bitbucketBranches, err := c.newClient(ctx, u).ListBranches(r.Owner, r.Name)
	if err != nil {
		return nil, err
	}

	branches := make([]string, 0)
	for _, branch := range bitbucketBranches {
		branches = append(branches, branch.Name)
	}
	return branches, nil
}

// BranchHead returns the sha of the head (latest commit) of the specified branch
func (c *config) BranchHead(_ context.Context, _ *model.User, _ *model.Repo, _ string) (string, error) {
	// TODO(1138): missing implementation
	return "", forge_types.ErrNotImplemented
}

func (c *config) PullRequests(_ context.Context, _ *model.User, _ *model.Repo, _ *model.ListOptions) ([]*model.PullRequest, error) {
	return nil, forge_types.ErrNotImplemented
}

// Hook parses the incoming Bitbucket hook and returns the Repository and
// Pipeline details. If the hook is unsupported nil values are returned.
func (c *config) Hook(_ context.Context, req *http.Request) (*model.Repo, *model.Pipeline, error) {
	return parseHook(req)
}

// OrgMembership returns if user is member of organization and if user
// is admin/owner in this organization.
func (c *config) OrgMembership(ctx context.Context, u *model.User, owner string) (*model.OrgPerm, error) {
	perm, err := c.newClient(ctx, u).GetUserWorkspaceMembership(owner, u.Login)
	if err != nil {
		return nil, err
	}
	return &model.OrgPerm{Member: perm != "", Admin: perm == "owner"}, nil
}

// helper function to return the bitbucket oauth2 client
func (c *config) newClient(ctx context.Context, u *model.User) *internal.Client {
	if u == nil {
		return c.newClientToken(ctx, "", "")
	}
	return c.newClientToken(ctx, u.Token, u.Secret)
}

// helper function to return the bitbucket oauth2 client
func (c *config) newClientToken(ctx context.Context, token, secret string) *internal.Client {
	return internal.NewClientToken(
		ctx,
		c.API,
		c.Client,
		c.Secret,
		&oauth2.Token{
			AccessToken:  token,
			RefreshToken: secret,
		},
	)
}

// helper function to return the bitbucket oauth2 config
func (c *config) newConfig(redirect string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.Client,
		ClientSecret: c.Secret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/site/oauth2/authorize", c.url),
			TokenURL: fmt.Sprintf("%s/site/oauth2/access_token", c.url),
		},
		RedirectURL: fmt.Sprintf("%s/authorize", redirect),
	}
}

// helper function to return matching hooks.
func matchingHooks(hooks []*internal.Hook, rawurl string) *internal.Hook {
	link, err := url.Parse(rawurl)
	if err != nil {
		return nil
	}
	for _, hook := range hooks {
		hookurl, err := url.Parse(hook.URL)
		if err == nil && hookurl.Host == link.Host {
			return hook
		}
	}
	return nil
}
